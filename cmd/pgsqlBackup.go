package cmd

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

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
	// if dbHost == "" || dbPort == "" || dbUser == "" || dbName == "" || dbPassword == "" || outputDir == "" {
	// 	return fmt.Errorf("missing required arguments")
	// }

	now := time.Now()
	outputFile := filepath.Join(filepath.Clean(outputDir), "pg_dump "+now.Format("2006-01-02 15:04:05")+".sql")

	// check if pg_dump is installed
	cmd := exec.Command("dpkg", "-s", "postgresql-client")
	if err := cmd.Run(); err != nil {
		log.Fatal("PostgreSQL client is not installed. Please install it first. [sudo apt install postgresql-client]")
		return err
	}

	os.Setenv("PGPASSWORD", dbPassword)
	cmd = exec.Command("pg_dump", "-h", dbHost, "-p", dbPort, "-U", dbUser, "-d", dbName, "-f", outputFile)

	var outbuf, errbuf bytes.Buffer
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error on pg_dump: %v\n", err)
		fmt.Printf("output: %s\n", outbuf.String())
		fmt.Printf("error output: %s\n", errbuf.String())
		return err
	}

	// Remove files older than ***, time defined in attributes
	// ...

	if maxAge != "" {
		pruneCount, err := prune(outputDir, maxAge)
		if err != nil {
			fmt.Printf("Error on pruneBackups: %v\n", err)
			return err
		}
		fmt.Printf("Pruned %d old backups\n", pruneCount)
	}

	fmt.Println("Dump created:", outputFile)
	return nil
}

var pgsqlCmd = &cobra.Command{
	Use:   "pgsql",
	Short: "Run a 'PGSQL' backup",
	Long:  `Run a 'PGSQL' backup`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := pgsqlBackup(dbHost, dbPort, dbUser, dbName, dbPassword, outputDir, maxAge); err != nil {
			log.Fatal(err)
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
