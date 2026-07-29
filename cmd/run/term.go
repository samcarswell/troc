package cmd

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"strconv"

	"github.com/samcarswell/troc/config"
	"github.com/samcarswell/troc/core"
	"github.com/samcarswell/troc/data"
	"github.com/samcarswell/troc/opts"
	"github.com/spf13/cobra"
)

var termCmd = &cobra.Command{
	Use:   "term",
	Short: "Manually terminate a run",
	Long: `
Manually terminate a run.

This command is to handle situations where a run has been killed via SIGKILL.
In this case troc cannot gracefully fail the run, and will be left in a state of 'running'.
	
This does NOT stop the original process if it is still running; 
if it is still running it will update the run again once it completes.

This command will fail if the run is in any other state than 'running'.
`,
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
		if runRow.Run.Status != string(core.RunStatusRunning) {
			core.LogErrorAndExit(logger, errors.New("run must be in a running state to manually fail"))
		}
		err = queries.EndRun(cmd.Context(), data.EndRunParams{
			Status: string(core.RunStatusTerminated),
			ID:     runId,
		})
		if err != nil {
			core.LogErrorAndExit(logger, err)
		}

		core.LogRunManuallyTerminated(logger, runId, runRow.Job.Name)

		updRunRow, err := queries.GetRun(context.Background(), runId)
		if err != nil {
			core.LogErrorAndExit(logger, err)
		}

		data := core.RunShow{
			ID:            updRunRow.Run.ID,
			JobName:       updRunRow.Job.Name,
			StartTime:     core.FormatTime(updRunRow.Run.StartTime, conf.LocalTime),
			EndTime:       core.FormatTime(updRunRow.Run.EndTime.Time, conf.LocalTime),
			LogFile:       updRunRow.Run.LogFile,
			SystemLogFile: updRunRow.Run.ExecLogFile,
			Status:        updRunRow.Run.Status,
			Duration:      core.FormatDuration(updRunRow.Run.StartTime, updRunRow.Run.EndTime.Time),
			Pid:           core.FormatPid(updRunRow.Run.Pid),
		}
		core.PrintJson(data)
	},
}

func init() {
	RunCmd.AddCommand(termCmd)

	termCmd.Flags().Int64P("run-id", "r", 0, "Run id")
	if err := termCmd.MarkFlagRequired("run-id"); err != nil {
		core.LogErrorAndExit(slog.Default(), err)
	}
}
