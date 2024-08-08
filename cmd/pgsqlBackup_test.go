package cmd

import (
	"testing"
)

type dbBackupCfg struct {
	host       string
	port       string
	user       string
	password   string
	dbName     string
	output     string
	maxFileAge string
}

func TestPgslBackupexec(t *testing.T) {
	tests := []struct {
		cfg dbBackupCfg
	}{
		{
			cfg: dbBackupCfg{
				host:       "192.168.35.43", // local test db
				port:       "5432",
				user:       "admin",
				password:   "admin",
				dbName:     "gseven",
				output:     "/home/rbaswe/backups/",
				maxFileAge: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.cfg.dbName, func(t *testing.T) {
			var cfg = tt.cfg
			err := pgsqlBackup(cfg.host, cfg.port, cfg.user, cfg.dbName, cfg.password, cfg.output, cfg.maxFileAge)
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
