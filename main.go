/*
Copyright © 2024 Robert Bauernfeind  r.bauernfeind@braun.at
*/
package main

import (
	"github.com/RBASWE/db-backup-runner/cmd"
	"github.com/RBASWE/db-backup-runner/logger"
)

func main() {
	cmd.Execute()
	logger.FileLogger.Info("Hello World")
}
