/*
Copyright © 2024 Robert Bauernfeind
*/
package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/log"
	"github.com/manifoldco/promptui"
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
	Long:  `add new cron - sudo required!`,
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

	var cronPath = "/etc/cron.d/" + cronName
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
	if _, err := os.Stat(cronPath); err == nil {
		prompt := promptui.Prompt{
			Label:     "Do want to rename the file?",
			IsConfirm: true,
		}

		if result, err := prompt.Run(); err != nil {
			return err
		} else {
			if result == "y" {
				prompt := promptui.Prompt{
					Label: "New name",
				}
				if newName, err := prompt.Run(); err != nil {
					return err
				} else {
					cronPath = "/etc/cron.d/" + newName
				}
			} else {
				return err
			}
		}
	}

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
		user           = "root"
		backupCfg      = backupConfig{
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

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Enter cron name").Value(&cronName).Validate(required),
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
				// huh.NewInput().Title("Enter user").Value(&user).,
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
				huh.NewInput().Title("Output directory").Value(&backupCfg.output).Validate(required),
				huh.NewInput().Title("Max file age (default \"\" => no delete)").Value(&backupCfg.maxFileAge).Placeholder(""),
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
