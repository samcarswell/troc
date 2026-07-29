package config

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"io/fs"
	"log"
	"log/slog"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/samcarswell/troc/core"
	"github.com/samcarswell/troc/data"
	"github.com/spf13/viper"
	_ "modernc.org/sqlite"

	"github.com/amacneil/dbmate/v2/pkg/dbmate"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/sqlite"
)

const ConfigDatabasePath = "database"
const ConfigLogDir = "logdir"

func expandDir(path string) (string, error) {
	if strings.HasSuffix(path, "~") {
		path = strings.Replace(path, "~", "$HOME", 1)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	tmpDir := os.TempDir()

	mapper := func(placeholderName string) string {
		switch placeholderName {
		case "HOME":
			return homeDir
		case "TMPDIR":
			return tmpDir
		}
		return ""
	}
	return os.Expand(path, mapper), nil
}

func CreateAndReadConfig(
	confDir string,
	confName string,
	confType string,
) {
	expandedConfigDir, err := expandDir(confDir)
	if err != nil {
		core.LogErrorAndExit(slog.Default(), err, errors.New("unable to expand configuration directory "+confDir))
	}
	err = viper.ReadInConfig()
	if err != nil {
		var confNotFoundErr viper.ConfigFileNotFoundError
		if errors.As(err, &confNotFoundErr) {
			log.Println("Creating config directory at " + expandedConfigDir)
			err := os.MkdirAll(expandedConfigDir, os.ModePerm)
			if err != nil {
				core.LogErrorAndExit(slog.Default(), err, errors.New("unable to create configuration directory "+expandedConfigDir))
			}
			log.Println("Creating initial config file at " + expandedConfigDir + "/" + confName + "." + confType)
			err = viper.SafeWriteConfig()
			if err != nil {
				core.LogErrorAndExit(slog.Default(), err, errors.New("unable to write initial config"))
			}
		} else {
			core.LogErrorAndExit(slog.Default(), err, errors.New("unable to read config"))
		}
	}
}

type dbLogger struct {
	Logger *slog.Logger
}

func (dl dbLogger) Write(p []byte) (n int, err error) {
	dl.Logger.Info(strings.TrimRight(string(p), "\n"))
	return len(p), nil
}

func CreateOrUpdateDatabase(
	migrations fs.FS,
	ctx context.Context,
	dbPath string,
	migrationsDir string,
) *sql.DB {
	logger := slog.Default()
	fileName := path.Base(dbPath)
	dir := path.Dir(dbPath)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		core.LogErrorAndExit(logger, err, errors.New("unable to create database path "+dir))
	}
	u, _ := url.Parse("sqlite3:///" + dbPath)
	dbMateDb := dbmate.New(u)
	dbMateDb.FS = migrations
	dbMateDb.AutoDumpSchema = false
	dbMateDb.MigrationsDir = []string{migrationsDir}
	dblog := dbLogger{
		Logger: logger,
	}
	dbMateDb.Log = dblog

	err = dbMateDb.CreateAndMigrate()
	if err != nil {
		core.LogErrorAndExit(logger, err, errors.New("unable to update database"))
	}
	db, err := sql.Open("sqlite", path.Join(dir, fileName)+"?mode=rw")
	if err != nil {
		core.LogErrorAndExit(logger, err, errors.New("unable to open database"))
	}

	confDbErrMsg := "error configuring database connection"

	for range 10 {
		_, err = db.Exec("PRAGMA busy_timeout=10000;")
		if err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if err != nil {
		core.LogErrorAndExit(logger, err, errors.New(confDbErrMsg+": busy_timeout"))
	}
	_, err = db.Exec("PRAGMA journal_mode=WAL;")
	if err != nil {
		core.LogErrorAndExit(logger, err, errors.New(confDbErrMsg+": journal_mode"))
	}
	_, err = db.Exec("PRAGMA synchronous=NORMAL;")
	if err != nil {
		core.LogErrorAndExit(logger, err, errors.New(confDbErrMsg+": synchronous"))
	}
	_, err = db.Exec("PRAGMA cache_size=-20000;")
	if err != nil {
		core.LogErrorAndExit(logger, err, errors.New(confDbErrMsg+": cache_size"))
	}
	_, err = db.Exec("PRAGMA foreign_keys=true;")
	if err != nil {
		core.LogErrorAndExit(logger, err, errors.New(confDbErrMsg+": foreign_keys"))
	}
	return db
}

func GetDatabase(ctx context.Context) *data.Queries {
	migrations, ok := MigrationsFromContext(ctx)
	if !ok {
		core.LogErrorAndExit(slog.Default(), errors.New("could not get migrations"))
	}
	dbPath := viper.GetString(ConfigDatabasePath)
	if dbPath == "" {
		core.LogErrorAndExit(slog.Default(), errors.New("database config value is empty"))
	}
	expandedPath, err := expandDir(dbPath)
	if err != nil {
		core.LogErrorAndExit(slog.Default(), err, errors.New("unable to expand database path"))
	}

	return data.New(CreateOrUpdateDatabase(
		migrations,
		ctx,
		expandedPath,
		"./db/migrations",
	))
}

func GetLogFileOrExit(logger *slog.Logger, ctx context.Context) string {
	logFile, ok := LogFileFromContext(ctx)
	if !ok {
		core.LogErrorAndExit(slog.Default(), errors.New("unable to get logFile from context"))
	}
	return logFile
}

type CleanConfig struct {
	Days int
}

type NotifyConfig struct {
	Hostname string
	Slack    SlackConfig
	Status   StatusConfig
}

type SlackConfig struct {
	Token   string
	Channel string
}

type StatusConfig struct {
	Succeeded  bool
	Failed     bool
	Running    bool
	Skipped    bool
	Terminated bool
}

type ColorConfig struct {
	Status StatusConfig
}

type DisplayConfig struct {
	Emoji bool
	Color ColorConfig
}

type Config struct {
	Database  string
	LockDir   string
	LogDir    string
	Clean     CleanConfig
	Notify    NotifyConfig
	LocalTime bool
	Display   DisplayConfig
}

func GetConfig() Config {
	return Config{
		Database: viper.GetString("database"),
		LockDir:  viper.GetString("lockdir"),
		LogDir:   viper.GetString("logdir"),
		Clean: CleanConfig{
			Days: viper.GetInt("clean.days"),
		},
		LocalTime: viper.GetBool("localtime"),
		Notify: NotifyConfig{
			Hostname: viper.GetString("notify.hostname"),
			Slack: SlackConfig{
				Token:   viper.GetString("notify.slack.token"),
				Channel: viper.GetString("notify.slack.channel"),
			},
			Status: StatusConfig{
				Succeeded:  viper.GetBool("notify.status.succeeded"),
				Failed:     viper.GetBool("notify.status.failed"),
				Running:    viper.GetBool("notify.status.running"),
				Skipped:    viper.GetBool("notify.status.skipped"),
				Terminated: viper.GetBool("notify.status.terminated"),
			},
		},
		Display: DisplayConfig{
			Emoji: viper.GetBool("display.emoji"),
			Color: ColorConfig{
				Status: StatusConfig{
					Succeeded:  viper.GetBool("display.color.status.succeeded"),
					Failed:     viper.GetBool("display.color.status.failed"),
					Running:    viper.GetBool("display.color.status.running"),
					Skipped:    viper.GetBool("display.color.status.skipped"),
					Terminated: viper.GetBool("display.color.status.terminated"),
				},
			},
		},
	}
}

type loggerKey struct{}
type logFileKey struct{}
type migrationsKey struct{}

func ContextWithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

func LoggerFromContext(ctx context.Context) (*slog.Logger, bool) {
	dbConn, ok := ctx.Value(loggerKey{}).(*slog.Logger)
	return dbConn, ok
}

func ContextWithLogFile(ctx context.Context, logFile string) context.Context {
	return context.WithValue(ctx, logFileKey{}, logFile)
}

func LogFileFromContext(ctx context.Context) (string, bool) {
	logFile, ok := ctx.Value(logFileKey{}).(string)
	return logFile, ok
}

func ContextWithMigrations(ctx context.Context, migrations embed.FS) context.Context {
	return context.WithValue(ctx, migrationsKey{}, migrations)
}

func MigrationsFromContext(ctx context.Context) (embed.FS, bool) {
	migrations, ok := ctx.Value(migrationsKey{}).(embed.FS)
	return migrations, ok
}
