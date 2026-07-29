package cmd

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/samcarswell/troc/config"
	"github.com/samcarswell/troc/core"
	"github.com/samcarswell/troc/opts"
	"github.com/spf13/cobra"
)

var format string

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List jobs",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return opts.FormatTableOptValidate(cmd, format)
	},
	Run: func(cmd *cobra.Command, args []string) {
		queries := config.GetDatabase(cmd.Context())

		jobRows, err := queries.GetJobs(context.Background())
		if err != nil {
			core.LogErrorAndExit(slog.Default(), err, errors.New("unable to get jobs"))
		}
		var rows = []core.JobShow{}
		for _, job := range jobRows {
			rows = append(rows, core.JobShow{
				ID:               job.Job.ID,
				Name:             job.Job.Name,
				NotifyLogContent: job.Job.NotifyLogContent,
			})
		}

		t := core.NewTable(rows, rowConv, []string{
			"ID", "Name", "Notify Log Content",
		})
		t.Print(core.OutputFormat(format))
	},
}

func rowConv(row core.JobShow, _ core.OutputFormat) table.Row {
	return table.Row{
		row.ID,
		row.Name,
		row.NotifyLogContent,
	}
}

func init() {
	JobCmd.AddCommand(listCmd)
	opts.FormatTableOpt(listCmd, &format)
}
