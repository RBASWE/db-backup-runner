/*
Copyright © 2024 Robert Bauernfeind
*/
package cmd

import (
	"github.com/RBASWE/db-backup-runner/logger"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

// logCmd represents the log command
var logCmd = &cobra.Command{
	Use:   "log",
	Short: "output logpath",
	Long:  `output logpath`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("logpath", "path", logger.LogFile)
	},
}

func init() {
	rootCmd.AddCommand(logCmd)
}
