package cmd

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"strconv"

	"github.com/samcarswell/troc/config"
	"github.com/samcarswell/troc/core"
	"github.com/samcarswell/troc/opts"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show details of a run",
	Run: func(cmd *cobra.Command, args []string) {
		logger := slog.Default()
		runId := opts.GetInt64OrExit(cmd, "run-id")
		queries := config.GetDatabase(cmd.Context())
		conf := config.GetConfig()

		runRow, err := queries.GetRun(context.Background(), runId)
		if err != nil {
			if err == sql.ErrNoRows {
				core.LogErrorAndExit(logger, errors.New("run with id "+strconv.FormatInt(runId, 10)+" not found"))
			} else {
				core.LogErrorAndExit(logger, err)
			}
		}
		data := core.RunShow{
			ID:            runRow.Run.ID,
			JobName:       runRow.Job.Name,
			StartTime:     core.FormatTime(runRow.Run.StartTime, conf.LocalTime),
			EndTime:       core.FormatTime(runRow.Run.EndTime.Time, conf.LocalTime),
			LogFile:       runRow.Run.LogFile,
			SystemLogFile: runRow.Run.ExecLogFile,
			Status:        runRow.Run.Status,
			Pid:           core.FormatPid(runRow.Run.Pid),
			Duration:      core.FormatDuration(runRow.Run.StartTime, runRow.Run.EndTime.Time),
			IsArchived:    runRow.Run.IsArchived,
		}
		core.PrintJson(data)
	},
}

func init() {
	RunCmd.AddCommand(showCmd)

	showCmd.Flags().Int64P("run-id", "r", 0, "Run id")
	if err := showCmd.MarkFlagRequired("run-id"); err != nil {
		core.LogErrorAndExit(slog.Default(), err)
	}
}
