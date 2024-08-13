/*
Copyright © 2024 Robert Bauernfeind
*/
package cmd

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"

	"github.com/RBASWE/db-backup-runner/logger"
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
		if name, cron, err := askForCronCfg(); err != nil {
			log.Error(err)
			logger.FileLogger.Error("Error on add", "err", err)
		} else {
			if err := writeCron(name, cron); err != nil {
				log.Error(err)
				logger.FileLogger.Error("Error on add", "err", err)
			}
		}
	},
}

func writeCron(cronName string, cron string) error {
	cronName = CronFilePrefix + "_" + cronName

	if runtime.GOOS != "linux" {
		return errors.New("this function is only supported on Linux")
	}

	var cronPath = filepath.Join(CronDir, cronName)
	file, err := os.CreateTemp("", "dbbackuprunner_"+cronName)
	if err != nil {
		log.Error("Error opening or creating file", "err", err)
		return err
	}
	// Write the cron job to the file
	_, err = file.WriteString(cron)
	if err != nil {
		log.Error("Error writing to file", "err", err)
		return err

	}
	defer file.Close()

	// Move the file to the cron directory
	// TODO better solution?
	cmd := exec.Command("sudo", "mv", file.Name(), cronPath)
	if err := cmd.Run(); err != nil {
		log.Error("Error writing to file", "err", err)
		return err
	}

	cmd = exec.Command("sudo", "chown", "root:root", cronPath)
	err = cmd.Run()
	if err != nil {
		log.Error("Error changing file owner", "err", err)
		return err
	}

	return nil
}

// Linux
func askForCronCfg() (cronName string, cron string, err error) {
	var (
		pasteCron      string
		cronExpression string
		// dbConnectionOk = false
		user      = "root"
		backupCfg = backupConfig{
			dbType:     "pgsql",
			host:       "",
			port:       "5432",
			user:       "",
			password:   "",
			dbName:     "",
			output:     "",
			maxFileAge: "",
		}
		fullCron string
	)

	for {
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
			if err := huh.NewInput().Title("Enter cron expression").Value(&cronExpression).Validate(func(s string) error {
				if err := required(s); err != nil {
					return err
				}
				if err = validateCronExpression(s); err != nil {
					return err
				}

				return nil
			}).Run(); err != nil {
				return "", "", err
			}

			for {
				cronBuilderForm := huh.NewForm(
					huh.NewGroup(
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
					),
				)

				log.Info("Enter database credentials")
				if err := cronBuilderForm.Run(); err != nil {
					return "", "", err
				}
				var testConn = false
				huh.NewConfirm().Title("Test connection?").Value(&testConn).Run()
				if testConn {
					log.Info("Testing connection...")
					if err := dryRun(backupCfg); err != nil {
						log.Error("Error testing connection, check your credentials", "err", err)
						var tryAgain = false
						huh.NewConfirm().Title("Try again?").Value(&tryAgain).Run()
						if !tryAgain {
							break
						}
					} else {
						// dbConnectionOk = true
						log.Info("Connection successful!")
						break
					}
				}
			}

			for {
				var proceed = true
				if err := huh.NewInput().Title("Output directory").Value(&backupCfg.output).Validate(required).Run(); err != nil {
					return "", "", err
				}

				// Check if the directory exists
				if _, err := os.Stat(backupCfg.output); os.IsNotExist(err) {
					proceed = false
					// Prompt the user with a confirmation dialog
					createDir := false
					huh.NewConfirm().Title(backupCfg.output + `
					Directory does not exist. Create it?`).Value(&createDir).Run()

					if createDir {
						// Attempt to create the directory
						if err := os.MkdirAll(backupCfg.output, os.ModePerm); err != nil {
							return "", "", err
						}
						proceed = true
					} else {
						var tryAgain = false
						huh.NewConfirm().Title("Try again?").Value(&tryAgain).Run()
						proceed = !tryAgain
					}
				}

				if proceed {
					break
				}
			}

			if err := huh.NewInput().Title("Max file age (default \"\" => no delete)").Value(&backupCfg.maxFileAge).Placeholder("").Run(); err != nil {
				return "", "", err
			}

			var command = ""
			if binaryPath, err := os.Executable(); err != nil {
				return "", "", err
			} else {
				command = binaryPath
			}

			command += createCommand(backupCfg)

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
				log.Info("Try again")
			}
		}
	}
}

func init() {
	CronRootCmd.AddCommand(addCmd)
}

func createCommand(backupCfg backupConfig) string {
	return " " + backupCfg.dbType + " --host " + backupCfg.host + " --port " + backupCfg.port + " --user " + backupCfg.user + " --pw " + backupCfg.password + " --database " + backupCfg.dbName + " --output " + backupCfg.output
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
