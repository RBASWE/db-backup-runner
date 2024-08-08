# Backup CLI Tool

This tool automatically creates a database dump and stores it on the file system. If configured, a maximum file age (`maxFileAge`) can be set, which will remove files older than the specified age.

## Usage

To use the tool, run the following command:

```sh
db-backup-runner pgsql --host "192.168.35.43" -p 5432 -u admin --pw admin -d gseven -o "/home/rbaswe/backups"
```

### Parameters

> **`--host`**: The database host (e.g., "192.168.35.43")  
> **`--port`**, `-p`: The database port (e.g., 5432)  
> **`--user`**, `-u`: The database user (e.g., "admin")  
> **`--pw`**: The database password (e.g., "admin")  
> **`--database`**, `-d`: The database name (e.g., "gseven")  
> **`--output`**, `-o`: The output directory for the backup files (e.g., "/home/rbaswe/backups")  

#### optional

> `--max -file-age`: The maximum age of backup files in days (e.g., 7s)

## Prune
Delete all files older than the specified age.

```sh
db-backup-runner prune [backupDir] --max-file-age 720h # 720h => 30days
```

## Install

```sh
go install github.com/rbaswe/db-backup-runner@latest
```

### Windows


in order for windws to find the executable, you need to add the go bin path to the system path.  
Go to `Environment Variables` > `System Variables` > `Path` > `Edit` > `New` and then paste the go bin path.  (e.g., `%USERPROFILE%\go\bin`)
This should be done automatically when installing go.

Since this script uses `pg_dump` to create the backup, you need to install [postgresql](https://www.postgresql.org/download/windows/) and add the `bin` folder to the system path.  
Again go to `Environment Variables` > `System Variables` > `Path` > `Edit` > `New` and paste the path to the bin folder of postgresql. (e.g., `C:\Program Files\PostgreSQL\16\bin`)

### Linux

Same in linux, the PATH and GOPATH environment variables need to be set in order to find  
the executables.

* export PATH=$PATH:$HOME/go/bin
* export GOPATH=$HOME/go

In order to get the pgsqlclient package which is used by the script use following command to install

```sh
sudo apt install postgresql-client
```

## Cronjob example

```sh
* * * * * /home/rbaswe/go/bin/db-backup-runner pgsql --host 192.168.35.43 -u admin --pw admin -d gseven -p 5432 --max-file-age 10m -o /home/rbaswe/backups/
```
