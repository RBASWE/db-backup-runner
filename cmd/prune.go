/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	pruneMaxAge = ""
)

func prune(outputDir string, maxAgeAsString string) (int, error) {
	maxAge, err := time.ParseDuration(maxAgeAsString)
	var pruneCount = 0
	if err != nil {
		return pruneCount, err
	}

	files, err := filepath.Glob(filepath.Join(outputDir, "*.sql"))
	if err != nil {
		return pruneCount, err
	}
	now := time.Now()

	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			return pruneCount, err
		}
		if now.Sub(info.ModTime()) > maxAge {
			err = os.Remove(file)
			if err != nil {
				return pruneCount, err
			}

			pruneCount++
		}
	}
	return pruneCount, nil
}

// pruneCmd represents the prune command
var pruneCmd = &cobra.Command{
	Use:        "prune",
	Short:      "Prune old backup files",
	Long:       `Prune all backup files older then given max file age. REMOVE PERMANENTLY!`,
	Args:       cobra.ExactArgs(1),
	ArgAliases: []string{"directory"},
	Run: func(cmd *cobra.Command, args []string) {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Are you sure you want to perform this action? (y/n): ")
		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			return
		}

		response = strings.TrimSpace(response)
		if strings.ToLower(response) == "y" {
			if _, err := prune(args[0], cmd.Flag("max-file-age").Value.String()); err != nil {
				log.Fatalln(err)
			}
		} else {
			fmt.Println("Action canceled.")
		}
	},
}

func init() {
	rootCmd.AddCommand(pruneCmd)

	// pruneCmd.Flags().StringVarP(&pruneDirectory, "directory", "d", ".", "Backup directory")
	pruneCmd.Flags().StringVar(&pruneMaxAge, "max-file-age", ".", "Max file age. files will automatically get deleted on backup cmd run. Valid time units are 'ns', 'us' (or 'µs'), 'ms', 's', 'm', 'h'")

	var requiredFlags = []string{"max-file-age"}
	for _, flag := range requiredFlags {
		pruneCmd.MarkFlagRequired(flag)
	}

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pruneCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pruneCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
