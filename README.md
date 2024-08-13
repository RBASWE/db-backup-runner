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

### Prune
Delete all files older than the specified age.

```sh
db-backup-runner prune [backupDir] --max-file-age 720h # 720h => 30days
```

### Using Cronjobs

To automate the backup process, you can use cronjobs. The cli tool provides a command `cron add` to add a cronjob to your system.  
so typing  
```sh
db-backup-runner cron add
```
will open up a input form to fill in the details.

This command will add a file containing the cronjob expression to the systems dir `/etc/cron.d/`.  
All files inside this directory will be executed by the cron daemon.

**in order to use this command, you need to have root privileges.**


#### list active crons
```sh
db-backup-runner cron list  
```

#### remove cron
```sh
db-backup-runner cron delete -c [name]
```

## Logs
The tool logs all actions to the file `/tmp/db-backup-runner.log` by default. The logpath can be modified by the `LOG_FILE_PATH` environment variable.

## Install

```sh
go install github.com/rbaswe/db-backup-runner@latest
```

### Windows (only on time backups supported, no automated tasks)

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

### Using the docker image
The can also be used in a docker container.  
To do so, run the following command:

```sh
docker build -it db-backup-runner .
```
```sh
docker run --rm -d --name dbbr db-backup-runner 
docker exec -it dbbr /bin/bash
```

The cli tool will be available inside the container with all need dependencies.

## Cronjob example

```sh
* * * * * /home/rbaswe/go/bin/db-backup-runner pgsql --host 192.168.35.43 -u admin --pw admin -d gseven -p 5432 --max-file-age 10m -o /home/rbaswe/backups/
```
