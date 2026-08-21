/*
Copyright © 2025 Samuel Carswell <samuelrcarswell@gmail.com>
*/
package cmd

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/samcarswell/troc/cmd"
	"github.com/samcarswell/troc/config"
	"github.com/samcarswell/troc/core"
	"github.com/samcarswell/troc/data"
	"github.com/samcarswell/troc/notify"
	"github.com/samcarswell/troc/opts"
	"github.com/spf13/cobra"
)

var nameOpt = "name"
var notifyOpt = "notify"

// While *OrExit is useful for most commands, exec actually needs to
// try it's best to recover: it should try to get least get a message to the slack channel notifying of a failure.
// For now I'm just going to assume there's no errors. But this should
// be revisited before a v1
var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "Run a job",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		logger := slog.Default()
		jobName := opts.GetStringOptOrExit(cmd, nameOpt)
		notifyOpt := opts.GetBoolOptOrExit(cmd, notifyOpt)
		conf := config.GetConfig()
		queries := config.GetDatabase(cmd.Context())

		// TODO: validate notify system setup based on notify.system

		logFile := config.GetLogFileOrExit(logger, cmd.Context())

		if len(args) == 0 {
			core.LogErrorAndExit(logger, errors.New("must provide args"))
		}
		completedRun := execRun(
			cmd.Context(),
			logger,
			jobName,
			conf,
			queries,
			logFile,
			args,
		)

		data := core.RunShow{
			ID:            completedRun.Run.ID,
			JobName:       completedRun.Job.Name,
			StartTime:     core.FormatTime(completedRun.Run.StartTime, conf.LocalTime),
			EndTime:       core.FormatTime(completedRun.Run.EndTime.Time, conf.LocalTime),
			LogFile:       completedRun.Run.LogFile,
			SystemLogFile: completedRun.Run.ExecLogFile,
			Status:        completedRun.Run.Status,
			Pid:           core.FormatPid(completedRun.Run.Pid),
			Duration:      core.FormatDuration(completedRun.Run.StartTime, completedRun.Run.EndTime.Time),
		}

		notifyData := core.RunNotify{
			Run: data,
			Job: core.JobShow{
				ID:               completedRun.Job.ID,
				Name:             completedRun.Job.Name,
				NotifyLogContent: completedRun.Job.NotifyLogContent,
			},
			StatusConfig: conf.Notify.Status,
			Hostname:     conf.Notify.Hostname,
		}

		if notifyOpt {
			logger.Info("Sending notify message")
			ok, err := notify.NotifyRun(
				conf,
				notifyData,
				logger,
			)
			if err != nil {
				core.LogErrorAndExit(logger, err, errors.New("unable to notify"))
			}
			if !ok {
				logger.Error("command was run, but notification was unable to be sent")
			}
		}

		core.PrintJson(data)
	},
}

func init() {
	cmd.RootCmd.AddCommand(execCmd)

	execCmd.Flags().String(nameOpt, "", "Name of job to execute. Will create it if it does not exist")
	if err := execCmd.MarkFlagRequired(nameOpt); err != nil {
		core.LogErrorAndExit(slog.Default(), err)
	}
	execCmd.Flags().Bool(notifyOpt, false, "Notifies of the exec success")
}

func execRun(
	ctx context.Context,
	logger *slog.Logger,
	jobName string,
	conf config.Config,
	db *data.Queries,
	logFile string,
	args []string,
) data.GetRunRow {
	jobRow, err := db.GetJob(ctx, jobName)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Print("Job not registered. Creating new job with name " + jobName)
		} else {
			core.LogErrorAndExit(logger, err)
		}
	}
	if jobRow == (data.GetJobRow{}) {
		id, err := db.CreateJob(context.Background(), data.CreateJobParams{
			Name:             jobName,
			NotifyLogContent: false,
		})
		if err != nil {
			if err.Error() == "constraint failed: UNIQUE constraint failed: jobs.name (2067)" {
				logger.Info("Job with name " + jobName + " has been created by another process: using it")
				jobRow, err = db.GetJob(ctx, jobName)
				if err != nil {
					core.LogErrorAndExit(logger, err)
				}

			} else {
				core.LogErrorAndExit(logger, err)
			}
		} else {
			jobRow.Job.Name = jobName
			jobRow.Job.ID = id
		}
	}

	isRunning, err := db.IsJobRunning(context.Background(), jobName)
	if err != nil || isRunning {
		if err != nil {
			logger.Error(err.Error())
		}
		return skipRun(
			jobRow.Job,
			logFile,
			db,
			context.Background(),
			logger,
		)
	}

	stdout, err := os.CreateTemp(conf.LogDir, jobName+".*.log")
	if err != nil {
		core.LogErrorAndExit(logger, err, errors.New("unable to create log file"))
	}
	stdoutLog, err := os.OpenFile(stdout.Name(), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		core.LogErrorAndExit(logger, err, errors.New("unable to open log file"))
	}

	logger.Info("Run log created at: " + stdout.Name())
	runId, err := db.StartRun(context.Background(), data.StartRunParams{
		JobID:       jobRow.Job.ID,
		LogFile:     stdout.Name(),
		ExecLogFile: logFile,
	})
	if err != nil {
		if err.Error() == "constraint failed: already a run for this job with status Running (1811)" {
			logger.Error(err.Error())
			return skipRun(
				jobRow.Job,
				logFile,
				db,
				context.Background(),
				logger,
			)
		}
		core.LogErrorAndExit(logger, err, errors.New("unable to start run"))
	}
	core.LogRunCreated(logger, runId, jobName)

	cmdArgs := []string{"-c"}
	cmdArgs = append(cmdArgs, args[0])
	runCmd := exec.Command("/bin/sh", cmdArgs...)
	runCmd.Stdout = stdoutLog
	runCmd.Stderr = stdoutLog

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-c
		if sig == syscall.SIGTERM {
			err := runCmd.Process.Signal(syscall.SIGTERM)
			if err != nil {
				logger.Error("Failed to send SIGTERM to run.")
			}
		}
	}()

	status := core.RunStatusSucceeded
	err = runCmd.Start()
	if err != nil {
		logger.Error("Failed to start run: " + err.Error())
		status = core.RunStatusFailed
	} else {
		core.LogRunStarted(logger, runId, jobName, runCmd.Process.Pid)
		err := db.UpdateRunPid(ctx, data.UpdateRunPidParams{
			ID: runId,
			Pid: sql.NullInt64{
				Int64: int64(runCmd.Process.Pid),
				Valid: true,
			},
		})
		if err != nil {
			sigtermErr := runCmd.Process.Signal(syscall.SIGTERM)
			if sigtermErr != nil {
				logger.Error("Failed to send SIGTERM to process.")
			}
			status = core.RunStatusFailed
		}
		err = runCmd.Wait()
		if err != nil {
			if strings.HasPrefix(err.Error(), "signal: ") {
				core.LogRunTerminated(logger, runId, jobName, err.Error())
				status = core.RunStatusTerminated
			} else {
				logger.Error("Error occurred during run", "error", err)
				status = core.RunStatusFailed
			}
		}
	}

	err = db.EndRun(context.Background(), data.EndRunParams{
		Status: string(status),
		ID:     runId,
	})
	if err != nil {
		core.LogErrorAndExit(logger, err, errors.New("unable to end run"))
	}
	core.LogRunCompleted(logger, runId, jobName, status)

	completedRun, err := db.GetRun(ctx, runId)
	if err != nil {
		core.LogErrorAndExit(logger, err, errors.New("unable to get completed run"))
	}
	return completedRun
}

func skipRun(
	job data.Job,
	execLogFile string,
	queries *data.Queries,
	ctx context.Context,
	logger *slog.Logger,
) data.GetRunRow {
	id, err := queries.SkipRun(ctx, data.SkipRunParams{
		JobID:       job.ID,
		ExecLogFile: execLogFile,
	})
	if err != nil {
		// TODO: need a standard function here to deal with errors and communicate to slack
		core.LogErrorAndExit(logger, err, errors.New("unable to skip run"))
	}
	run, err := queries.GetRun(ctx, id)
	if err != nil {
		// TODO: need a standard function here to deal with errors and communicate to slack
		core.LogErrorAndExit(logger, err, errors.New("unable to get updated run"))
	}
	core.LogRunSkipped(logger, id, job.Name)
	row, err := queries.GetRun(ctx, run.Run.ID)
	if err != nil {
		core.LogErrorAndExit(logger, err, errors.New("unable to get updated run"))
	}
	return row
}
