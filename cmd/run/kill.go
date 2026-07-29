package cmd

import (
	"bufio"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/samcarswell/troc/config"
	"github.com/samcarswell/troc/core"
	"github.com/samcarswell/troc/opts"
	"github.com/spf13/cobra"
)

var killCmd = &cobra.Command{
	Use:   "kill",
	Short: "Kill a run",
	Long: `
Kill a run.
This command will lookup the PID of a run and send a SIGTERM signal to it.
`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := slog.Default()
		runId := opts.GetInt64OrExit(cmd, "run-id")
		force := opts.GetBoolOptOrExit(cmd, "force")
		queries := config.GetDatabase(cmd.Context())

		runRow, err := queries.GetRun(context.Background(), runId)
		if err != nil {
			if err == sql.ErrNoRows {
				core.LogErrorAndExit(logger, errors.New("run with id "+strconv.FormatInt(runId, 10)+" not found"))
			} else {
				core.LogErrorAndExit(logger, err)
			}
		}
		if runRow.Run.Status != string(core.RunStatusRunning) {
			core.LogErrorAndExit(logger, errors.New("run must be in a running state to kill it"))
		}
		if !runRow.Run.Pid.Valid {
			core.LogErrorAndExit(logger, errors.New("run does not have a PID associated with it"))
		}

		logger.Info("Attempting to kill run " + strconv.FormatInt(runId, 10) + " with PID " + core.FormatPid(runRow.Run.Pid) + ". Ensure that the PID is as expected.")

		if !force {
			fmt.Fprintf(os.Stderr, "%s (%s) ", "Are you sure?", "y/n")
			r := bufio.NewReader(os.Stdin)
			var s string
			s, _ = r.ReadString('\n')
			s = strings.ToLower(strings.TrimSpace(s))
			if s != "y" && s != "yes" {
				logger.Info("Cancelling")
				return
			}
		} else {
			logger.Info("--force flag provided. Skipping confirmation")
		}

		err = syscall.Kill(int(runRow.Run.Pid.Int64), syscall.SIGTERM)
		if err != nil {
			core.LogErrorAndExit(logger, err, errors.New("unable to kill run"))
		}
		core.LogRunSentSigterm(logger, runId, runRow.Job.Name, int(runRow.Run.Pid.Int64))
	},
}

func init() {
	RunCmd.AddCommand(killCmd)

	killCmd.Flags().Int64P("run-id", "r", 0, "Run id")
	killCmd.Flags().Bool("force", false, "Force kill")
	if err := killCmd.MarkFlagRequired("run-id"); err != nil {
		core.LogErrorAndExit(slog.Default(), err)
	}
}
