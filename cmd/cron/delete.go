/*
Copyright © 2024 Robert Bauernfeind
*/
package cmd

import (
	"os/exec"
	"path/filepath"

	"github.com/RBASWE/db-backup-runner/logger"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var cronName string

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete running cronjob - sudo required",
	Long:  `Delete running cronjob - sudo required`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := delete(cronName); err != nil {
			logger.FileLogger.Error("Error on delete", "err", err)
			log.Error("Error on delete", "err", err)
		}
	},
}

func delete(cronName string) error {
	delete := false
	if err := huh.NewConfirm().Title("Proceed with delete?").Value(&delete).Run(); err != nil {
		return err
	}

	if err := exec.Command("sudo", "rm", filepath.Join(CronDir, CronFilePrefix+"_"+cronName)).Run(); err != nil {
		return err
	}

	log.Info("Deleted cronjob:" + cronName)

	return nil
}

func init() {
	deleteCmd.Flags().StringVarP(&cronName, "cronName", "c", "", "Name of the cronjob")
	deleteCmd.MarkFlagRequired("cronName")
	CronRootCmd.AddCommand(deleteCmd)
}
