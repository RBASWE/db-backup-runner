/*
Copyright © 2024 Robert Bauernfeind
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// cronCmd represents the cron command
var CronRootCmd = &cobra.Command{
	Use:   "cron",
	Short: "Manage crons for backups",
	Long:  `Manage crons for backups`,
	// Run: func(cmd *cobra.Command, args []string) {
	// 	fmt.Println("cron called")
	// },
}
