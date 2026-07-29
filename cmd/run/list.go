package cmd

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/samcarswell/troc/config"
	"github.com/samcarswell/troc/core"
	"github.com/samcarswell/troc/data"
	"github.com/samcarswell/troc/opts"
	"github.com/spf13/cobra"
)

var includeArchivedOpt = "include-archived"
var nameOpt = "name"
var statusField = "Status"
var format string

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists runs",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return opts.FormatTableOptValidate(cmd, format)
	},
	Run: func(cmd *cobra.Command, args []string) {
		logger := slog.Default()
		conf := config.GetConfig()
		jobName := opts.GetStringOptOrExit(cmd, nameOpt)
		includeArchived := opts.GetBoolOptOrExit(cmd, includeArchivedOpt)
		queries := config.GetDatabase(cmd.Context())

		if jobName != "" {
			_, err := queries.GetJob(cmd.Context(), jobName)
			if err != nil {
				if err == sql.ErrNoRows {
					core.LogErrorAndExit(logger, errors.New("job with name '"+jobName+"' not found"))
				} else {
					core.LogErrorAndExit(logger, err)
				}
			}
		}

		runRows, err := queries.GetRuns(cmd.Context(), data.GetRunsParams{
			Column1: jobName,
			Column2: includeArchived,
		})
		if err != nil {
			core.LogErrorAndExit(slog.Default(), err)
		}
		var rows = []core.RunShow{}

		for _, runRow := range runRows {
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
			}
			rows = append(rows, data)
		}

		t := core.NewTable(rows, rowConv(conf), []string{
			"ID",
			"Job Name",
			"Start Time",
			"End Time",
			"Log File",
			"Exec Log File",
			statusField,
		})
		statusTransformer := text.Transformer(func(val any) string {
			if status, ok := val.(string); ok {
				color := text.FgWhite
				switch status {
				case core.FormatStatus(core.RunStatusSucceeded, conf.Display.Emoji):
					if conf.Display.Color.Status.Succeeded {
						color = text.FgGreen
					}
				case core.FormatStatus(core.RunStatusFailed, conf.Display.Emoji):
					if conf.Display.Color.Status.Failed {
						color = text.FgHiRed
					}
				case core.FormatStatus(core.RunStatusRunning, conf.Display.Emoji):
					if conf.Display.Color.Status.Running {
						color = text.FgCyan
					}
				case core.FormatStatus(core.RunStatusSkipped, conf.Display.Emoji):
					if conf.Display.Color.Status.Skipped {
						color = text.FgYellow
					}
				case core.FormatStatus(core.RunStatusTerminated, conf.Display.Emoji):
					if conf.Display.Color.Status.Terminated {
						color = text.FgHiMagenta
					}
				}
				return color.Sprintf("%s", status)
			}
			return fmt.Sprint(val)
		})

		t.Table.SetColumnConfigs([]table.ColumnConfig{
			{
				Name:        statusField,
				Transformer: statusTransformer,
			},
		})
		t.Print(core.OutputFormat(format))
	},
}

func rowConv(conf config.Config) func(core.RunShow, core.OutputFormat) table.Row {
	return func(row core.RunShow, format core.OutputFormat) table.Row {
		var status string
		if format == core.FormatPretty {
			status = core.FormatStatus(core.RunStatus(row.Status), conf.Display.Emoji)
		} else {
			status = row.Status
		}
		return table.Row{
			row.ID,
			row.JobName,
			row.StartTime,
			row.EndTime,
			row.LogFile,
			row.SystemLogFile,
			status,
		}
	}
}

func init() {
	RunCmd.AddCommand(listCmd)

	listCmd.Flags().String(nameOpt, "", "Name of job to filter on")
	listCmd.Flags().Bool(includeArchivedOpt, false, "Include archived runs")
	opts.FormatTableOpt(listCmd, &format)
}
