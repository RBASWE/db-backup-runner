package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

var dbHost string
var dbPort string
var dbUser string
var dbName string
var dbPassword string

func pgsqlBackup() {
	fmt.Println("Add new database runner")

	outputFile := "/home/rbaswe/backups/db_dump" + strconv.Itoa(time.Now().Second()) + ".sql"

	os.Setenv("PGPASSWORD", dbPassword)
	cmd := exec.Command("pg_dump", "-h", dbHost, "-p", dbPort, "-U", dbUser, "-d", dbName, "-f", outputFile)

	var outbuf, errbuf bytes.Buffer
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf

	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error on pg_dump: %v\n", err)
		fmt.Printf("output: %s\n", outbuf.String())
		fmt.Printf("error output: %s\n", errbuf.String())
		return
	}

	fmt.Println("Dump created:", outputFile)
}

var pgsqlCmd = &cobra.Command{
	Use:   "pgsql",
	Short: "Run a 'PGSQL' backup",
	Long:  `Run a 'PGSQL' backup`,
	Run: func(cmd *cobra.Command, args []string) {
		pgsqlBackup()
	},
}

func init() {
	pgsqlCmd.Flags().StringVar(&dbHost, "host", "", "Database host")
	pgsqlCmd.Flags().StringVarP(&dbPort, "port", "p", "", "Database port")
	pgsqlCmd.Flags().StringVarP(&dbUser, "user", "u", "", "Database user")
	pgsqlCmd.Flags().StringVarP(&dbName, "database", "d", "", "Database name")
	pgsqlCmd.Flags().StringVar(&dbPassword, "pw", "", "Database password")

	rootCmd.AddCommand(pgsqlCmd)
}
