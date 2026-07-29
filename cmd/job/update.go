package cmd

import (
	"database/sql"
	"errors"
	"log"
	"log/slog"

	"github.com/samcarswell/troc/config"
	"github.com/samcarswell/troc/core"
	"github.com/samcarswell/troc/data"
	"github.com/samcarswell/troc/opts"
	"github.com/spf13/cobra"
)

var notifyLogOpt = "notify-log"
var newNameOpt = "new-name"

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update job",
	Run: func(cmd *cobra.Command, args []string) {
		logger := slog.Default()
		jobName := opts.GetStringOptOrExit(cmd, "name")
		queries := config.GetDatabase(cmd.Context())

		job, err := queries.GetJob(cmd.Context(), jobName)
		if err != nil {
			if err == sql.ErrNoRows {
				core.LogErrorAndExit(logger, errors.New("job with name '"+jobName+"' not found"))
			} else {
				core.LogErrorAndExit(logger, err)
			}
		}

		if cmd.Flags().Changed(notifyLogOpt) {
			job.Job.NotifyLogContent = opts.GetBoolOptOrExit(cmd, notifyLogOpt)
		}
		if cmd.Flags().Changed(newNameOpt) {
			job.Job.Name = opts.GetStringOptOrExit(cmd, newNameOpt)
		}

		err = queries.UpdateJob(cmd.Context(), data.UpdateJobParams{
			ID:               job.Job.ID,
			Name:             job.Job.Name,
			NotifyLogContent: job.Job.NotifyLogContent,
		})

		if err != nil {
			core.LogErrorAndExit(logger, err, errors.New("unable to update job"))
		}

		logger.Info("Job updated")
	},
}

func init() {
	JobCmd.AddCommand(updateCmd)
	updateCmd.Flags().String("name", "", "Job Name (required)")
	updateCmd.Flags().String(newNameOpt, "", "New job Name")
	updateCmd.Flags().Bool(notifyLogOpt, false, "Includes the raw log output rather than the log filename in notification messages (default false)")
	if err := updateCmd.MarkFlagRequired("name"); err != nil {
		log.Fatalf("Unable to mark name as required %s", err)
	}
}
