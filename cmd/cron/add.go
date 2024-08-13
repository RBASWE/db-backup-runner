/*
Copyright © 2024 Robert Bauernfeind
*/
package cmd

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/RBASWE/db-backup-runner/logger"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type BackupConfig struct {
	CronExpression string   `yaml:"cronExpression"`
	Db             DbConfig `yaml:"db"`
	Output         string   `yaml:"output"`
	MaxFileAge     string   `yaml:"maxFileAge"`
	User           string   `yaml:"execUser"`
}

type DbConfig struct {
	DbType   string `yaml:"type"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DbName   string `yaml:"database"`
}

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add new cron - sudo required!",
	Long:  `Add new cron - sudo required!`,
	Run: func(cmd *cobra.Command, args []string) {
		if name, cron, err := askForCronCfg(); err != nil {
			log.Error(err)
			logger.FileLogger.Error("Error on add", "err", err)
		} else {
			if err := writeCron(name, cron); err != nil {
				log.Error(err)
				logger.FileLogger.Error("Error on add", "err", err)
			}
		}
	},
}

func writeCron(cronName string, cron string) error {
	cronName = CronFilePrefix + "_" + cronName

	if runtime.GOOS != "linux" {
		return errors.New("this function is only supported on Linux")
	}

	var cronPath = filepath.Join(CronDir, cronName)
	file, err := os.CreateTemp("", "dbbackuprunner_"+cronName)
	if err != nil {
		log.Error("Error opening or creating file", "err", err)
		return err
	}
	// Write the cron job to the file
	_, err = file.WriteString(cron)
	if err != nil {
		log.Error("Error writing to file", "err", err)
		return err

	}
	defer file.Close()

	// Move the file to the cron directory
	// TODO better solution?
	cmd := exec.Command("sudo", "mv", file.Name(), cronPath)
	if err := cmd.Run(); err != nil {
		log.Error("Error writing to file", "err", err)
		return err
	}

	cmd = exec.Command("sudo", "chown", "root:root", cronPath)
	err = cmd.Run()
	if err != nil {
		log.Error("Error changing file owner", "err", err)
		return err
	}

	return nil
}

// Linux
func askForCronCfg() (cronName string, cron string, err error) {
	var (
		pasteCron string
		backupCfg = BackupConfig{
			CronExpression: "0 0 * * *",
			Db: DbConfig{
				DbType:   "pgsql",
				Host:     "",
				Port:     "5432",
				User:     "",
				Password: "",
				DbName:   "",
			},
			Output:     "",
			MaxFileAge: "",
			User:       "",
		}
		fullCron string
	)

	for {
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().Title("Enter cron name").Value(&cronName).Validate(func(s string) error {
					if err := required(s); err != nil {
						return err
					}
					if err := checkFileName(s); err != nil {
						return err
					}
					return nil
				}),
				huh.NewSelect[string]().Title("Do you want to paste a cron?").
					Options(
						huh.NewOption[string]("Yes", "yes"),
						huh.NewOption[string]("No", "no").Selected(true),
						huh.NewOption[string]("Read from file", "file"),
					).Value(&pasteCron),
			),
		)

		if err := form.Run(); err != nil {
			return "", "", err
		}

		// TODO add valitadors to prompts
		switch pasteCron {
		case "yes":
			cronInput := huh.NewInput().
				Title("Enter cron expression").
				Description("Format: [cronExpression: * * * * *] [user: root] [task: pathToExecutable]").
				Value(&fullCron).
				Validate(required)
			if err := cronInput.Run(); err != nil {
				return "", "", err
			}

			log.Info(fullCron)
			// return cronName, fullCron, nil
		case "no":
			if err := huh.NewInput().Title("Enter cron expression").Value(&backupCfg.CronExpression).Validate(func(s string) error {
				if err := required(s); err != nil {
					return err
				}
				if err = validateCronExpression(s); err != nil {
					return err
				}

				return nil
			}).Run(); err != nil {
				return "", "", err
			}

			for {
				cronBuilderForm := huh.NewForm(
					huh.NewGroup(
						huh.NewSelect[string]().Title("Select database type").
							Options(
								// huh.NewOption[string]("MySQL", "mysql"),
								huh.NewOption[string]("PostgreSQL", "pgsql").Selected(true),
							).Value(&backupCfg.Db.DbType),

						huh.NewInput().Title("Database host").Value(&backupCfg.Db.Host).Validate(required),
						huh.NewInput().Title("Database port").Value(&backupCfg.Db.Port).Validate(required),
						huh.NewInput().Title("Database user").Value(&backupCfg.Db.User).Validate(required),
						huh.NewInput().Title("Database password").Value(&backupCfg.Db.Password).Validate(required).EchoMode(huh.EchoModePassword),
						huh.NewInput().Title("Database name").Value(&backupCfg.Db.DbName).Validate(required),
					),
				)

				log.Info("Enter database credentials")
				if err := cronBuilderForm.Run(); err != nil {
					return "", "", err
				}
				var testConn = false
				huh.NewConfirm().Title("Test connection?").Value(&testConn).Run()
				if testConn {
					log.Info("Testing connection...")
					if err := dryRun(backupCfg.Db); err != nil {
						log.Error("Error testing connection, check your credentials", "err", err)
						var tryAgain = false
						huh.NewConfirm().Title("Try again?").Value(&tryAgain).Run()
						if !tryAgain {
							break
						}
					} else {
						// dbConnectionOk = true
						log.Info("Connection successful!")
						break
					}
				}
			}

			for {
				var proceed = true
				if err := huh.NewInput().Title("Output directory").Value(&backupCfg.Output).Validate(required).Run(); err != nil {
					return "", "", err
				}

				// Check if the directory exists
				if _, err := os.Stat(backupCfg.Output); os.IsNotExist(err) {
					proceed = false
					// Prompt the user with a confirmation dialog
					createDir := false
					huh.NewConfirm().Title(backupCfg.Output + `
					Directory does not exist. Create it?`).Value(&createDir).Run()

					if createDir {
						// Attempt to create the directory
						if err := os.MkdirAll(backupCfg.Output, os.ModePerm); err != nil {
							return "", "", err
						}
						proceed = true
					} else {
						var tryAgain = false
						huh.NewConfirm().Title("Try again?").Value(&tryAgain).Run()
						proceed = !tryAgain
					}
				}

				if proceed {
					break
				}
			}

			if err := huh.NewInput().Title("Max file age (default \"\" => no delete)").Value(&backupCfg.MaxFileAge).Placeholder("").Run(); err != nil {
				return "", "", err
			}
		case "file":
			// =================================================
			// =============== READ FROM FILE ==================
			// =================================================
			var filePath string
			huh.NewInput().Title("Config filepath (.yml)").Value(&filePath).Validate(func(s string) error {
				if _, err := os.Stat(s); os.IsNotExist(err) {
					return errors.New("file does not exist")
				}

				if err = importBackupConfigFile(filePath, &backupCfg); err != nil {
					return err
				}

				var emptyProperties []string
				if backupCfg.CronExpression == "" {
					emptyProperties = append(emptyProperties, "cronExpression")
				}
				if backupCfg.Db.DbType == "" {
					emptyProperties = append(emptyProperties, "db.type")
				}
				if backupCfg.Db.Host == "" {
					emptyProperties = append(emptyProperties, "db.host")
				}
				if backupCfg.Db.Port == "" {
					emptyProperties = append(emptyProperties, "db.port")
				}
				if backupCfg.Db.User == "" {
					emptyProperties = append(emptyProperties, "db.user")
				}
				if backupCfg.Db.Password == "" {
					emptyProperties = append(emptyProperties, "db.password")
				}
				if backupCfg.Db.DbName == "" {
					emptyProperties = append(emptyProperties, "db.database")
				}
				if backupCfg.Output == "" {
					emptyProperties = append(emptyProperties, "db.output")
				}
				if len(emptyProperties) > 0 {
					return errors.New("the following properties are empty: " + strings.Join(emptyProperties, ", "))
				}

				if _, err := os.Stat(backupCfg.Output); os.IsNotExist(err) {
					return errors.New("output file does not exist")
				}

				if err := validateCronExpression(backupCfg.CronExpression); err != nil {
					return errors.New("invalid cron expression")
				}

				if err := dryRun(backupCfg.Db); err != nil {
					return errors.New("error testing connection, check your credentials")
				}

				return nil
			}).Run()
		}

		var command = ""
		if binaryPath, err := os.Executable(); err != nil {
			return "", "", err
		} else {
			command = binaryPath
		}

		command += createCommand(backupCfg)

		if backupCfg.MaxFileAge != "" {
			command += " --max-file-age " + backupCfg.MaxFileAge
		}

		fullCron = backupCfg.CronExpression + " " + backupCfg.User + " " + command + "\n"

		log.Info(fullCron)

		ok := false
		confirmCron := huh.NewConfirm().Title("Confirm cron expression").Value(&ok)
		if err := confirmCron.Run(); err != nil {
			return "", "", err
		} else {
			if ok {
				return cronName, fullCron, nil
			} else {
				log.Info("Try again")
			}
		}
	}
}

func init() {
	CronRootCmd.AddCommand(addCmd)
}

func createCommand(backupCfg BackupConfig) string {
	return " " + backupCfg.Db.DbType + " --host " + backupCfg.Db.Host + " --port " + backupCfg.Db.Port + " --user " + backupCfg.Db.User + " --pw " + backupCfg.Db.Password + " --database " + backupCfg.Db.DbName + " --output " + backupCfg.Output
}

// Form validators
func required(s string) error {
	if s == "" {
		return errors.New("required")
	}
	return nil
}

func validateCronExpression(s string) error {
	match, err := regexp.MatchString(`^(\*|[0-5]?\d)(/(\*|[1-5]?\d))?\s+(\*|[0-5]?\d)(-(\*|[0-5]?\d))?(/(\*|[1-5]?\d))?\s+(\*|[0-2]?\d|3[0-1])(-(\*|[0-2]?\d|3[0-1]))?(/(\*|[1-5]?\d))?\s+(\*|[1-9]|1[0-2])(-(\*|[1-9]|1[0-2]))?(/(\*|[1-5]?\d))?\s+(\*|[0-6]|7)(-(\*|[0-6]|7))?(/(\*|[1-5]?\d))?$`, s)
	if err != nil {
		return err
	}
	if !match {
		return errors.New("invalid cron expression")
	}

	return nil
}

func checkFileName(s string) error {
	if _, err := os.Stat(filepath.Join(CronDir, CronFilePrefix+"_"+s)); err != nil {
		return nil
	} else {
		return errors.New("cron already exists")
	}
}

func dryRun(dbConfig DbConfig) error {
	switch dbConfig.DbType {
	case "pgsql":
		os.Setenv("PGPASSWORD", dbConfig.Password)
		cmd := exec.Command("psql", "-h", dbConfig.Host, "-p", dbConfig.Port, "-U", dbConfig.User, "-d", dbConfig.DbName)

		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}

func importBackupConfigFile(cfgFilePath string, backupCfg *BackupConfig) error {
	var yamlFile []byte
	var err error
	if yamlFile, err = os.ReadFile(cfgFilePath); err != nil {
		return err
	}

	if err := yaml.Unmarshal(yamlFile, &backupCfg); err != nil {
		return err
	}

	return nil
}
