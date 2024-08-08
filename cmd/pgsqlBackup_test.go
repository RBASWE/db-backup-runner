package cmd

import (
	"testing"
)

var (
	testDbHost     = "192.168.35.43"
	testDbPort     = "5432"
	testDbUser     = "admin"
	testDbPassword = "admin"
	testDbName     = "gseven"
	testOutputDir  = "/home/rbaswe/backups/"
)

func TestPgslBackupexec(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
		{"test"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pgsqlBackup(testDbHost, testDbPort, testDbUser, testDbName, testDbPassword, testOutputDir, "")
			if err != nil {
				t.Errorf("pgsqlBackup() error = %v", err)
				return
			}
		})
	}
}

func TestPgsqlBackupFileDeletion(t *testing.T) {
	var filePath = "/home/rbaswe/backups/"

	pruneCount, err := prune(filePath, "30m")
	t.Log(pruneCount)
	if err != nil {
		t.Fatal(err)
	}
}
