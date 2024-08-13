package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/RBASWE/db-backup-runner/logger"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var dbHost string
var dbPort string
var dbUser string
var dbName string
var dbPassword string
var outputDir string
var maxAge = ""

func pgsqlBackup(dbHost string, dbPort string, dbUser string, dbName string, dbPassword string, outputDir string, maxAge string) error {
	now := time.Now()
	outputFile := filepath.Join(filepath.Clean(outputDir), "pg_dump_"+now.Format("20060102_150405")+".sql")

	var cmd *exec.Cmd
	// check if pg_dump is installed
	if runtime.GOOS == "windows" {
		// Windows-specific logic to check if pg_dump is installed
		cmd = exec.Command("where", "pg_dump")
		if err := cmd.Run(); err != nil {
			// log.Fatal("PostgreSQL client is not installed. Please install it first.")
			return errors.New("postgreSQL client is not installed. Please install it first")
		}
	} else {
		// Linux
		cmd = exec.Command("dpkg", "-s", "postgresql-client-16")
		if err := cmd.Run(); err != nil {
			// log.Fatal("PostgreSQL client is not installed. Please install it first. [sudo apt install postgresql-client]")
			return errors.New("postgresql-client-16 is not installed. Please install it first. [sudo apt install postgresql-client]")
		}
	}

	os.Setenv("PGPASSWORD", dbPassword)
	cmd = exec.Command("pg_dump", "-h", dbHost, "-p", dbPort, "-U", dbUser, "-d", dbName, "-f", outputFile)

	var outbuf, errbuf bytes.Buffer
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf

	if err := cmd.Run(); err != nil {
		return fmt.Errorf(
			"error on pg_dump: %v\noutput: %s\nerror output: %s",
			err, outbuf.String(), errbuf.String(),
		)
	}

	// Remove files older than ***, time defined in attributes
	if maxAge != "" {
		pruneCount, err := prune(outputDir, maxAge)
		if err != nil {
			log.Error("Error on pruneBackups", "error", err)
			return err
		}

		log.Warn("Pruned old backups", "count", pruneCount)
		logger.FileLogger.Warn("Pruned old backups", "count", pruneCount)
	}

	log.Info("Dump created:", "dump file", outputFile)
	msg := dbName + " | Dump created:"

	logger.FileLogger.Info(msg, "dump file", outputFile)
	return nil
}

var pgsqlCmd = &cobra.Command{
	Use:   "pgsql",
	Short: "Run a 'PGSQL' backup",
	Long:  `Run a 'PGSQL' backup`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := pgsqlBackup(dbHost, dbPort, dbUser, dbName, dbPassword, outputDir, maxAge); err != nil {
			logger.FileLogger.Error("Error on pgsqlBackup", "error", err)
			log.Error("Error on pgsqlBackup", "error", err)
		}
	},
}

func init() {
	pgsqlCmd.Flags().StringVar(&dbHost, "host", "", "Database host")
	pgsqlCmd.Flags().StringVar(&dbPassword, "pw", "", "Database password")
	pgsqlCmd.Flags().StringVarP(&dbPort, "port", "p", "", "Database port")
	pgsqlCmd.Flags().StringVarP(&dbUser, "user", "u", "", "Database user")
	pgsqlCmd.Flags().StringVarP(&dbName, "database", "d", "", "Database name")
	pgsqlCmd.Flags().StringVarP(&outputDir, "output", "o", "", "File output directory")
	pgsqlCmd.Flags().StringVarP(&maxAge, "max-file-age", "a", "", "OPTIONAL - Max file age. files will automatically get deleted on backup cmd run. Valid time units are 'ns', 'us' (or 'µs'), 'ms', 's', 'm', 'h'")

	var requiredFlags = []string{"host", "port", "user", "database", "pw", "output"}
	for _, flag := range requiredFlags {
		pgsqlCmd.MarkFlagRequired(flag)
	}

	rootCmd.AddCommand(pgsqlCmd)
}
