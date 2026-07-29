/*
Copyright © 2025 Samuel Carswell <samuelrcarswell@gmail.com>
*/
package cmd

import (
	"database/sql"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/samcarswell/troc/cmd"
	"github.com/samcarswell/troc/cmd/lib"
	"github.com/samcarswell/troc/config"
	"github.com/samcarswell/troc/core"
	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Archive old runs",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		logger := slog.Default()
		conf := config.GetConfig()
		cleanTime := time.Now().UTC().AddDate(0, 0, -conf.Clean.Days)
		logger.Info("Archiving runs before " + strconv.Itoa(conf.Clean.Days) + " days ago: " + cleanTime.String())

		queries := config.GetDatabase(ctx)
		runs, err := queries.GetNonArchivedRunsBeforeDate(ctx, sql.NullTime{
			Time:  cleanTime,
			Valid: true,
		})
		if err != nil && err != sql.ErrNoRows {
			core.LogErrorAndExit(logger, err)
		}
		if err == sql.ErrNoRows || len(runs) == 0 {
			logger.Info("No runs found.")
		}
		for _, run := range runs {
			err := lib.ArchiveRun(ctx, queries, run.Run, logger)
			if err != nil {
				logger.Error("Failed to archive run", "err", err)
				continue
			}
		}

		err = clearMiscSystemLogs(conf.LogDir, cleanTime, logger)
		if err != nil {
			core.LogErrorAndExit(logger, err)
		}
	},
}

func clearMiscSystemLogs(dir string, cleanTime time.Time, logger *slog.Logger) error {
	pattern := path.Join(dir, "trocsys_*")

	files, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}
	for _, file := range files {
		segments := strings.Split(file, "_")
		if len(segments) != 3 {
			logger.Warn("log file not recognised: " + file + ". skipping")
			continue
		}
		subSegments := strings.Split(segments[2], ".")
		if len(subSegments) != 2 {
			logger.Warn("file not recognised: " + file + ". skipping")
			continue
		}
		ts, err := time.Parse("20060102T150405", subSegments[0])
		if err != nil {
			logger.Error("unable to parse file: " + file + ". skipping")
			continue
		}
		if cleanTime.After(ts.UTC()) {
			logger.Info("removing syslog file " + file)
			err = os.Remove(file)
			if err != nil {
				logger.Error("unable to delete syslog file " + file)
				continue
			}
			logger.Info("deleted syslog file " + file)
		}
	}
	return nil
}

func init() {
	cmd.RootCmd.AddCommand(cleanCmd)
}
