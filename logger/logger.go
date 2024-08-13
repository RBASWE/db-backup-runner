package logger

import (
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
)

var FileLogger *log.Logger

func init() {
	var file *os.File
	var err error
	var logFile string

	if logFileEnv := os.Getenv("LOG_FILE_PATH"); logFileEnv != "" {
		logFile = logFileEnv
	} else {
		logFile = filepath.Join(os.TempDir(), "db-backup-runner.log")
	}

	if file, err = os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600); err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	FileLogger = log.NewWithOptions(file, log.Options{
		ReportCaller:    true,
		ReportTimestamp: true,
		Level:           log.InfoLevel,
		Prefix:          "main-logger 💾",
		TimeFormat:      "2006-01-02 15:04:05",
		CallerOffset:    1,
	})

	FileLogger.Info("Logging to file: " + logFile)
}
