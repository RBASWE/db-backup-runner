/*
Copyright © 2024 Robert Bauernfeind
*/
package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

type backupConfig struct {
	dbType     string
	host       string
	port       string
	user       string
	password   string
	dbName     string
	output     string
	maxFileAge string
}

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add new cron - sudo required!",
	Long:  `Add new cron - sudo required!`,
	Run: func(cmd *cobra.Command, args []string) {
		if name, cron, err := askForCron(); err != nil {
			log.Error(err)
			// log.Fatal(err)
		} else {
			if err := writeCron(name, cron); err != nil {
				log.Error(err)
				// log.Fatal(err)
			}
		}
	},
}

func writeCron(cronName string, cron string) error {
	if runtime.GOOS != "linux" {
		return errors.New("this function is only supported on Linux")
	}

	var cronPath = filepath.Join(CronDir, cronName)
	file, err := os.CreateTemp("", "dbbackuprunner_"+cronName)
	if err != nil {
		fmt.Printf("Error opening or creating file: %v\n", err)
		return err
	}
	// Write the cron job to the file
	_, err = file.WriteString(cron)
	if err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
		return err
	}
	defer file.Close()

	// Move the file to the cron directory
	// TODO better solution?
	cmd := exec.Command("sudo", "mv", file.Name(), cronPath)
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
		return err
	}

	cmd = exec.Command("sudo", "chown", "root:root", cronPath)
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error changing file owner:", err)
		return err
	}

	return nil
}

func askForCron() (cronName string, cron string, err error) {

	var (
		pasteCron      string
		cronExpression string
		dbConnectionOk = false
		user           = "root"
		backupCfg      = backupConfig{
			dbType:     "pgsql",
			host:       "192.168.35.43",
			port:       "5432",
			user:       "admin",
			password:   "admin",
			dbName:     "gseven",
			output:     "/home/rbaswe/backups",
			maxFileAge: "",
		}
		fullCron string
	)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Enter cron name").Value(&cronName).Validate(func(s string) error {
				if err := required(s); err != nil {
					return err
				}
				if err := checkFileName(s); err != nil {
					return err
				}
				return nil
			}),
			huh.NewSelect[string]().Title("Do you want to paste a cron?").
				Options(
					huh.NewOption[string]("Yes", "yes"),
					huh.NewOption[string]("No", "no").Selected(true),
					huh.NewOption[string]("Read from file", "file"),
				).Value(&pasteCron),
		),
	)

	if err := form.Run(); err != nil {
		return "", "", err
	}

	cronName = CronFilePrefix + "_" + cronName

	// TODO add valitadors to prompts
	switch pasteCron {
	case "yes":
		cronInput := huh.NewInput().
			Title("Enter cron expression").
			Description("Format: [cronExpression: * * * * *] [user: root] [task: pathToExecutable]").
			Value(&fullCron).
			Validate(required)
		if err := cronInput.Run(); err != nil {
			return "", "", err
		}

		log.Info(fullCron)
		// return cronName, fullCron, nil
	case "no":
		cronBuilderForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().Title("Enter cron expression").Value(&cronExpression).Validate(func(s string) error {
					if err := required(s); err != nil {
						return err
					}
					if err = validateCronExpression(s); err != nil {
						return err
					}

					return nil
				}),
				huh.NewSelect[string]().Title("Select database type").
					Options(
						// huh.NewOption[string]("MySQL", "mysql"),
						huh.NewOption[string]("PostgreSQL", "pgsql").Selected(true),
					).Value(&backupCfg.dbType),

				huh.NewInput().Title("Database host").Value(&backupCfg.host).Validate(required),
				huh.NewInput().Title("Database port").Value(&backupCfg.port).Validate(required),
				huh.NewInput().Title("Database user").Value(&backupCfg.user).Validate(required),
				huh.NewInput().Title("Database password").Value(&backupCfg.password).Validate(required).EchoMode(huh.EchoModePassword),
				huh.NewInput().Title("Database name").Value(&backupCfg.dbName).Validate(required),
				huh.NewInput().Title("Output directory").Value(&backupCfg.output).Validate(func(s string) error {
					if err := required(s); err != nil {
						return err
					}

					// Check if the directory exists
					if _, err := os.Stat(s); os.IsNotExist(err) {
						// Prompt the user with a confirmation dialog
						createDir := false
						huh.NewConfirm().Title("Directory does not exist. Create it?").Value(&createDir).Run()

						if createDir {
							// Attempt to create the directory
							if err := os.MkdirAll(s, os.ModePerm); err != nil {
								return err
							}
						} else {
							log.Info("output directory does not exist and was not created")
						}
					}
					return nil
				}),
				huh.NewInput().Title("Max file age (default \"\" => no delete)").Value(&backupCfg.maxFileAge).Placeholder(""),
				huh.NewConfirm().Title("Test connection?").Value(&dbConnectionOk).Validate(func(b bool) error {
					log.Info("Testing connection to PostgreSQL database")

					if b {
						if err := dryRun(backupCfg); err != nil {
							return errors.New("connection failed, check you credentials")
						}
						log.Info("connection successful")
						return nil
					} else {
						return nil
					}
				}),
			),
		)

		if err := cronBuilderForm.Run(); err != nil {
			return "", "", err
		}

		var command = ""
		if binaryPath, err := os.Executable(); err != nil {
			return "", "", err
		} else {
			command = binaryPath
		}

		command += " " + backupCfg.dbType + " --host " + backupCfg.host + " --port " + backupCfg.port + " --user " + backupCfg.user + " --pw " + backupCfg.password + " --database " + backupCfg.dbName + " --output " + backupCfg.output

		if backupCfg.maxFileAge != "" {
			command += " --max-file-age " + backupCfg.maxFileAge
		}

		fullCron = cronExpression + " " + user + " " + command

	case "file":
		return "", "", errors.New("option not supported yet")
	}

	fullCron += "\n"
	log.Info(fullCron)

	ok := false
	confirmCron := huh.NewConfirm().Title("Confirm cron expression").Value(&ok)
	if err := confirmCron.Run(); err != nil {
		return "", "", err
	} else {
		if ok {
			return cronName, fullCron, nil
		} else {
			return cronName, "", errors.New("cron expression not confirmed")
		}
	}
}

func init() {
	CronRootCmd.AddCommand(addCmd)
}

// Form validators
func required(s string) error {
	if s == "" {
		return errors.New("required")
	}
	return nil
}

func validateCronExpression(s string) error {
	match, err := regexp.MatchString(`^(\*|[0-5]?\d)(/(\*|[1-5]?\d))?\s+(\*|[0-5]?\d)(-(\*|[0-5]?\d))?(/(\*|[1-5]?\d))?\s+(\*|[0-2]?\d|3[0-1])(-(\*|[0-2]?\d|3[0-1]))?(/(\*|[1-5]?\d))?\s+(\*|[1-9]|1[0-2])(-(\*|[1-9]|1[0-2]))?(/(\*|[1-5]?\d))?\s+(\*|[0-6]|7)(-(\*|[0-6]|7))?(/(\*|[1-5]?\d))?$`, s)
	if err != nil {
		return err
	}
	if !match {
		return errors.New("invalid cron expression")
	}

	return nil
}

func checkFileName(s string) error {
	if _, err := os.Stat(filepath.Join(CronDir, CronFilePrefix+"_"+s)); err != nil {
		return nil
	} else {
		return errors.New("cron already exists")
	}
}

func dryRun(backupCfg backupConfig) error {
	switch backupCfg.dbType {
	case "pgsql":
		os.Setenv("PGPASSWORD", backupCfg.password)
		cmd := exec.Command("psql", "-h", backupCfg.host, "-p", backupCfg.port, "-U", backupCfg.user, "-d", backupCfg.dbName)

		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}
