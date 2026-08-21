/*
Copyright © 2025 Samuel Carswell <samuelrcarswell@gmail.com>
*/
package cmd

import (
	"embed"
	"errors"
	"log/slog"
	"os"
	"path"
	"strings"

	"github.com/samcarswell/troc/config"
	"github.com/samcarswell/troc/core"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cliName = "troc"
var configPath = "$HOME/.config/troc"
var configName = "config"
var configType = "yaml"
var version = "development"

var RootCmd = &cobra.Command{
	Use:     cliName,
	Version: version,
	Short:   "troc - simple job monitoring",
	Long: `troc - simple job monitoring
	
https://github.com/samcarswell/troc
	`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		setupContext(cmd)
	},
}

// Most of this file needs to be moved to config/config.go
func setupContext(cmd *cobra.Command) {
	conf := config.GetConfig()
	err := os.MkdirAll(conf.LogDir, os.ModePerm)
	if err != nil {
		core.LogErrorAndExit(slog.Default(), err, errors.New("unable to create logdir"))
	}
	var l *slog.Logger
	if cmd.CommandPath() == cliName+" exec" {
		// If we're executing a job, we need to log to file
		logFile, err := core.CreateSyslog(conf.LogDir)
		if err != nil {
			core.LogErrorAndExit(slog.Default(), err, errors.New("unable to create trocsys log"))
		}
		l = slog.New(slog.NewMultiHandler(
			slog.Default().Handler(),
			slog.NewJSONHandler(logFile, core.GetSlogHandlerOptions()),
		))
		core.SetDefaultSlogLogger(l)
		l.Info("Logging to " + logFile.Name())
		cmd.SetContext(config.ContextWithLogFile(cmd.Context(), logFile.Name()))
	} else {
		l = slog.Default()
	}

	cmd.SetContext(config.ContextWithLogger(cmd.Context(), l))
	cmd.SetContext(config.ContextWithMigrations(cmd.Context(), Migrations))
}

var Migrations embed.FS

func Execute(migrations embed.FS) {
	Migrations = migrations
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	RootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
	viper.SetEnvPrefix("TROC")
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.SetDefault("logjson", getLogType())
	core.SetDefaultSlogLoggerInit(viper.GetBool("logjson"))
	logger := slog.Default()

	hostname, err := os.Hostname()
	if err != nil {
		logger.Error("Unable to read hostname. Defaulting to unknown-hostname")
		hostname = "unknown-hostname"
	}

	homedir, err := os.UserHomeDir()
	if err != nil {
		core.LogErrorAndExit(logger, err, errors.New("unable to get user home dir"))
	}
	viper.SetDefault("database", path.Join(homedir, ".config", "troc", "troc.db"))
	viper.SetDefault("logdir", os.TempDir())
	viper.SetDefault("clean.days", 30)
	viper.SetDefault("notify.hostname", hostname)
	viper.SetDefault("localtime", true)
	viper.SetDefault("display.emoji", true)
	viper.SetDefault("display.color.status.succeeded", false)
	viper.SetDefault("display.color.status.failed", false)
	viper.SetDefault("display.color.status.running", false)
	viper.SetDefault("display.color.status.skipped", false)
	viper.SetDefault("display.color.status.terminated", false)
	viper.SetDefault("notify.status.succeeded", false)
	viper.SetDefault("notify.status.failed", true)
	viper.SetDefault("notify.status.running", false)
	viper.SetDefault("notify.status.skipped", false)
	viper.SetDefault("notify.status.terminated", true)
	viper.SetDefault("notify.system", "slack")
	viper.SetDefault("notify.ntfy.domain", "https://ntfy.sh")

	confPath, ok := os.LookupEnv("TROC_CONFIG_PATH")
	if !ok {
		confPath = configPath
	}

	viper.AddConfigPath(confPath)
	viper.SetConfigName(configName)
	viper.SetConfigType(configType)
	viper.AutomaticEnv()
	err = config.CreateAndReadConfig(confPath, configName, configType)
	if err != nil {
		core.LogErrorAndExit(logger, err, errors.New("unable to create configuration"))
	}
}

func getLogType() bool {
	logJson, ok := os.LookupEnv("TROC_LOGJSON")
	if !ok {
		return false
	}
	return strings.ToLower(logJson) == "true"
}
