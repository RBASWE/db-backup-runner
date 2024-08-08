package cmd

import (
	"os"

	cron "github.com/RBASWE/db-backup-runner/cmd/cron"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "db-backup-runner",
	Short: "CLI Tool for auto database backups",
	Long:  `With this CLI tool multiple runners can be configured to generate database backups`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(cron.CronRootCmd)
}
