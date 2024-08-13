package logger

import (
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
)

var FileLogger *log.Logger
var LogFile string

func init() {
	var file *os.File
	var err error

	if logFileEnv := os.Getenv("LOG_FILE_PATH"); logFileEnv != "" {
		LogFile = logFileEnv
	} else {
		LogFile = filepath.Join(os.TempDir(), "db-backup-runner.log")
	}

	if file, err = os.OpenFile(LogFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600); err != nil {
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
}
