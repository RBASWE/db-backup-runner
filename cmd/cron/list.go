/*
Copyright © 2024 Robert Bauernfeind
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/charmbracelet/log"
	"github.com/mergestat/timediff"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all active cronjobs",
	Long:  `List all active cronjobs`,
	Run: func(cmd *cobra.Command, args []string) {
		if fileNames, err := readCronFiles(); err != nil {
			log.Error("Error reading cron files", "err", err)
		} else {
			log.Debug("FileNames", "files", fileNames)
			if err := displayFiles(fileNames); err != nil {
				log.Error("Error displaying files", "err", err)
			}
		}
	},
}

func readCronFiles() (fileNames []string, err error) {
	if files, err := os.ReadDir(CronDir); err != nil {
		return []string{}, err
	} else {
		for _, file := range files {
			if strings.HasPrefix(file.Name(), CronFilePrefix) {
				fileNames = append(fileNames, file.Name())
				log.Debug("Found file", "file", file.Name())
			}
		}

		return fileNames, nil
	}
}

func displayFiles(fileNames []string) error {
	rows := [][]string{}
	for _, fileName := range fileNames {
		if fileInfo, err := os.Stat(filepath.Join(CronDir, fileName)); err != nil {
			return err
		} else {
			diff := time.Since(fileInfo.ModTime())
			timeDiff := timediff.TimeDiff(time.Now().Add(-diff))

			rows = append(rows, []string{strings.Split(fileName, "_")[1], timeDiff, fileInfo.ModTime().Format(time.RFC3339Nano)})
		}
	}

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("99"))).
		Headers("Cron", "Created", "Created at").
		Rows(rows...)

	fmt.Println(t)
	return nil
}

func init() {
	CronRootCmd.AddCommand(listCmd)
}
