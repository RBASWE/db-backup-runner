package cmd

import "testing"

func TestAddCronJob(t *testing.T) {
	tests := []struct {
		cronName string
		cron     string
	}{
		// TODO: Add test cases.
		{"test", "*/1 * * * *"},
	}
	for _, tt := range tests {
		t.Run(tt.cronName, func(t *testing.T) {
			if err := writeCron(tt.cronName, tt.cron); err != nil {
				t.Fail()
			}
		})
	}
}

func TestFileImport(t *testing.T) {
	var filePath = "/home/rbaswe/workspace/db-backup-cli/sample.yaml"
	var cfg BackupConfig
	if err := importBackupConfigFile(filePath, &cfg); err != nil {
		t.Fail()
	} else {
		t.Log(cfg)
	}
}
