/*
Copyright © 2024 Robert Bauernfeind
*/
package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"

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
	Short: "Add new cron",
	Long:  `add new cron`,
	Run: func(cmd *cobra.Command, args []string) {
		if name, cron, err := askForCron(); err != nil {
			log.Fatal(err)
		} else {
			if err := writeCron(name, cron); err != nil {
				log.Fatal(err)
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
	// TODO check if file already exists
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
		cronExpression string
		user           = "root"
		backupCfg      = backupConfig{
			dbType:     "",
			host:       "",
			port:       "",
			user:       "",
			password:   "",
			dbName:     "",
			output:     "",
			maxFileAge: "",
		}
		fullCron     string
		promptSelect promptui.Select
		promptInput  promptui.Prompt
	)

	// TODO add valitadors to prompts

	promptInput = promptui.Prompt{
		Label:    "Cron name",
		Validate: nil,
	}
	if cronName, err = promptInput.Run(); err != nil {
		return "", "", err
	}

	promptSelect = promptui.Select{
		Label: "Paste cron?",
		Items: []string{"Yes", "No", "Read from file"},
	}

	i, _, err := promptSelect.Run()
	if err != nil {
		return "", "", err
	}

	switch i {
	case 0:
		promptInput = promptui.Prompt{
			Label:    "Paste cron",
			Validate: nil,
		}
		if fullCron, err = promptInput.Run(); err != nil {
			return "", "", err
		}

		return cronName, fullCron, nil
	case 1:
		var validator = func(input string) error {
			if input == "" {
				return errors.New("cannot be empty")
			}
			return nil
		}
		type input struct {
			label        string
			value        *string
			inputType    string // prompt or select
			options      []string
			defaultValue string
			validator    func(input string) error
		}

		inputs := []input{
			{
				label:        "Cron expression (* * * * *)",
				value:        &cronExpression,
				inputType:    "prompt",
				defaultValue: "* * * * *",
				validator:    validator,
			},
			{
				label:     "Database type",
				value:     &backupCfg.dbType,
				inputType: "select",
				options:   []string{"pgsql"}, // Add more options as needed
				validator: validator,
			},
			{
				label:        "Database host",
				value:        &backupCfg.host,
				inputType:    "prompt",
				defaultValue: "192.168.35.43",
				validator:    validator,
			},
			{
				label:        "Database port",
				value:        &backupCfg.port,
				inputType:    "prompt",
				defaultValue: "5432",
				validator:    validator,
			},
			{
				label:        "Database user",
				value:        &backupCfg.user,
				inputType:    "prompt",
				defaultValue: "admin",
				validator:    validator,
			},
			{
				label:        "Database password",
				value:        &backupCfg.password,
				inputType:    "prompt",
				defaultValue: "admin",
				validator:    validator,
			},
			{
				label:        "Database name",
				value:        &backupCfg.dbName,
				inputType:    "prompt",
				defaultValue: "gseven",
				validator:    validator,
			},
			{
				label:        "Output directory",
				value:        &backupCfg.output,
				inputType:    "prompt",
				defaultValue: "/home/rbaswe/backups",
				validator:    validator,
			},
			{
				label:        "Max file age (default \"\" => no delete)",
				value:        &backupCfg.maxFileAge,
				inputType:    "prompt",
				defaultValue: "",
				validator:    validator,
			},
		}

		for _, input := range inputs {
			if input.inputType == "select" {
				promptSelect = promptui.Select{
					Label: input.label,
					Items: input.options,
				}
				if _, *input.value, err = promptSelect.Run(); err != nil {
					return "", "", err
				}
			} else {

				promptInput = promptui.Prompt{
					Label:    input.label,
					Validate: input.validator,
					Default:  input.defaultValue,
				}
				if *input.value, err = promptInput.Run(); err != nil {
					return "", "", err
				}
			}
		}
	case 2:
		return "", "", errors.New("option not supported yet")
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

	fullCron = cronExpression + " " + user + " " + command + "\n"

	fmt.Println("Cron expression: ", fullCron)

	promptInput = promptui.Prompt{
		Label:     "Please check the cron expresion before confirming",
		Validate:  nil,
		Default:   fullCron,
		IsConfirm: true,
	}

	if confirm, err := promptInput.Run(); err != nil {
		return "", "", err
	} else {
		if confirm == "y" {
			return cronName, fullCron, nil

		} else {
			return cronName, "", errors.New("cron expression not confirmed")
		}
	}
}

func init() {
	CronRootCmd.AddCommand(addCmd)
}
