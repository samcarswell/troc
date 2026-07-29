package cmd

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"strconv"

	"github.com/samcarswell/troc/cmd/lib"
	"github.com/samcarswell/troc/config"
	"github.com/samcarswell/troc/core"
	"github.com/samcarswell/troc/opts"
	"github.com/spf13/cobra"
)

var archiveCmd = &cobra.Command{
	Use:   "archive",
	Short: "Archive a run",
	Long: `
Archive a run.
This command will set is_archived=true on the run, 
	and remove any log files associated with it.
`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := slog.Default()
		runId := opts.GetInt64OrExit(cmd, "run-id")
		queries := config.GetDatabase(cmd.Context())
		runRow, err := queries.GetRun(context.Background(), runId)
		if err != nil {
			if err == sql.ErrNoRows {
				core.LogErrorAndExit(logger, errors.New("run with id "+strconv.FormatInt(runId, 10)+" not found"))
			} else {
				core.LogErrorAndExit(logger, err)
			}
		}

		err = lib.ArchiveRun(cmd.Context(), queries, runRow.Run, logger)
		if err != nil {
			core.LogErrorAndExit(logger, err, errors.New("unable to archive run"))
		}
	},
}

func init() {
	RunCmd.AddCommand(archiveCmd)

	archiveCmd.Flags().Int64P("run-id", "r", 0, "Run id")
	if err := archiveCmd.MarkFlagRequired("run-id"); err != nil {
		core.LogErrorAndExit(slog.Default(), err)
	}
}
