/*
Copyright © 2024 Robert Bauernfeind
*/
package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

type backupConfig struct {
	dbType     string // pgsql, mysql, ...
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
			fmt.Println(cron, name)
			if err := writeCron(name, cron); err != nil {
				log.Fatal(err)
			}
		}
	},
}

func writeCron(cronName string, cron string) error {
	var cronPath = "/etc/cron.d/" + cronName

	file, err := os.OpenFile(cronPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		fmt.Printf("Error opening or creating file: %v\n", err)
		return err
	}
	defer file.Close()

	// Write the cron job to the file
	_, err = file.WriteString(cron)
	if err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
		return err
	}

	return nil
}

func askForCron() (cronName string, cron string, err error) {
	var (
		cronExpression string
		user           string
		backupCfg      backupConfig
		fullCron       string
		promptSelect   promptui.Select
		promptInput    promptui.Prompt
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
		Items: []string{"Yes", "No"},
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
		type input struct {
			label        string
			value        *string
			inputType    string // prompt or select
			options      []string
			defaultValue string
		}

		inputs := []input{
			{
				label:        "Cron expression (* * * * *)",
				value:        &cronExpression,
				inputType:    "prompt",
				defaultValue: "* * * * *",
			},
			{
				label:        "Execution user",
				value:        &user,
				inputType:    "prompt",
				defaultValue: "root",
			},
			{
				label:     "Database type",
				value:     &backupCfg.dbType,
				inputType: "select",
				options:   []string{"pgsql"}, // Add more options as needed
			},
			{
				label:        "Database host",
				value:        &backupCfg.host,
				inputType:    "prompt",
				defaultValue: "192.168.35.43",
			},
			{
				label:        "Database port",
				value:        &backupCfg.port,
				inputType:    "prompt",
				defaultValue: "5432",
			},
			{
				label:        "Database user",
				value:        &backupCfg.user,
				inputType:    "prompt",
				defaultValue: "admin",
			},
			{
				label:        "Database password",
				value:        &backupCfg.password,
				inputType:    "prompt",
				defaultValue: "admin",
			},
			{
				label:        "Database name",
				value:        &backupCfg.dbName,
				inputType:    "prompt",
				defaultValue: "gseven",
			},
			{
				label:        "Output directory",
				value:        &backupCfg.output,
				inputType:    "prompt",
				defaultValue: "/home/rbaswe/backups",
			},
			{
				label:        "Max file age (default \"\" => no delete)",
				value:        &backupCfg.maxFileAge,
				inputType:    "prompt",
				defaultValue: "",
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
					Validate: nil,
					Default:  input.defaultValue,
				}
				if *input.value, err = promptInput.Run(); err != nil {
					return "", "", err
				}
			}
		}
	}

	var goBin = os.Getenv("GOPATH")
	command := goBin + "/bin/db-backup-runner "

	command += backupCfg.dbType + " --host " + backupCfg.host + " --port " + backupCfg.port + " --user " + backupCfg.user + " --pw " + backupCfg.password + " --database " + backupCfg.dbName + " --output " + backupCfg.output

	if backupCfg.maxFileAge != "" {
		command += " --max-file-age " + backupCfg.maxFileAge
	}

	fullCron = cronExpression + " " + user + " " + command

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

// paste cron?
// 	yes => past cron and save
// 	no => promt for cron
// 		=> cron expression
// 		=> user
// 		=> command
// => get parameters
// 		=> show cron
// 		=> confirm
// 			yes => save cron
// 			no => discard cron
