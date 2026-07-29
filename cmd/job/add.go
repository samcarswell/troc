package cmd

import (
	"errors"
	"log"
	"log/slog"
	"strconv"
	"strings"

	"github.com/samcarswell/troc/config"
	"github.com/samcarswell/troc/core"
	"github.com/samcarswell/troc/data"
	"github.com/samcarswell/troc/opts"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add job",
	Run: func(cmd *cobra.Command, args []string) {
		logger := slog.Default()
		jobName := opts.GetStringOptOrExit(cmd, "name")
		notifyLog := opts.GetBoolOptOrExit(cmd, "notify-log")
		queries := config.GetDatabase(cmd.Context())

		newJobId, err := queries.CreateJob(cmd.Context(), data.CreateJobParams{
			Name:             jobName,
			NotifyLogContent: notifyLog,
		})
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint failed: jobs.name") {
				core.LogErrorAndExit(logger, err, errors.New("job with name "+jobName+" already exists"))
			}
			core.LogErrorAndExit(logger, err, errors.New("unable to create job"))
		}

		logger.Info("Job created with ID " + strconv.FormatInt(newJobId, 10) + " and name " + jobName)
	},
}

func init() {
	JobCmd.AddCommand(addCmd)
	addCmd.Flags().String("name", "", "Job Name (required)")
	addCmd.Flags().Bool("notify-log", false, "Includes the raw log output rather than the log filename in notification messages (default false)")
	if err := addCmd.MarkFlagRequired("name"); err != nil {
		log.Fatalf("Unable to mark name as required %s", err)
	}
}
