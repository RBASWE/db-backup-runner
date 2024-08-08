# Backup CLI Tool

This tool automatically creates a database dump and stores it on the file system. If configured, a maximum file age (`maxFileAge`) can be set, which will remove files older than the specified age.

## Usage

To use the tool, run the following command:

```sh
db-backup-runner pgsql --host "192.168.35.43" -p 5432 -u admin --pw admin -d gseven -o "/home/rbaswe/backups"
```

### Parameters

> `--host`, `-h`: The database host (e.g., "192.168.35.43")
> `--port`, `-p`: The database port (e.g., 5432)
> `--user`, `-u`: The database user (e.g., "admin")
> `--password`, `--pw`: The database password (e.g., "admin")
> `--database`, `-d`: The database name (e.g., "gseven")
> `--output`, `-o`: The output directory for the backup files (e.g., "/home/rbaswe/backups")

#### optional

> `--max -file-age`: The maximum age of backup files in days (e.g., 7s)

## Prune
Delete all files older than the specified age.

```sh
db-backup-runner prune [backupDir] --max-file-age 720h # 720h => 30days
```