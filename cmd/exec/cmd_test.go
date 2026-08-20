package cmd

import (
	"context"
	"testing"
	"time"

	"github.com/samcarswell/troc/config"
	"github.com/samcarswell/troc/core"
	"github.com/samcarswell/troc/data"
	"github.com/samcarswell/troc/test"
	"github.com/stretchr/testify/assert"
)

// TODO: move these tests to e2e tests in /cmd/cmd_test.go
// TODO: do this as apart of this release ^
func Test_execRunNonExistentJob(t *testing.T) {
	ctx := context.Background()
	db := test.CreateDb(ctx, t)
	jobName := test.UniqueIdentifer()
	logFile, logger := test.CreateSysLogFile(t)
	conf := config.Config{
		LockDir: t.TempDir(),
		LogDir:  t.TempDir(),
	}
	run := execRun(ctx, logger, jobName, conf, db, logFile, []string{"./testdata/script-passes"})
	dbJob, err := db.GetJob(ctx, jobName)
	if err != nil {
		t.Fatal(err.Error())
	}
	dbRun, err := db.GetRun(ctx, 1)
	if err != nil {
		t.Fatal(err.Error())
	}

	execLog := test.NewLogFromFileOrFail(run.Run.ExecLogFile)
	assert.Equal(t, jobName, run.Job.Name)
	assert.Equal(t, false, run.Job.NotifyLogContent)
	test.AssertFileExists(t, run.Run.ExecLogFile)
	test.AssertFileExists(t, run.Run.LogFile)
	assert.Equal(t, "Succeeded", run.Run.Status)
	test.AssertFileContents(t, "Output line 1\nOutput line 2\n", run.Run.LogFile)
	test.AssertLogHasInfo(t, "Job not registered. Creating new job with name "+jobName, execLog)
	test.AssertLogHasInfo(t, "Run log created at: "+run.Run.LogFile, execLog)
	runCreated := test.GetEventOrFail(t, core.EventRunCreated, execLog)
	assert.Equal(t, int64(1), runCreated.RunId)
	runCompleted := test.GetEventOrFail(t, core.EventRunCompleted, execLog)
	assert.Equal(t, string(core.RunStatusSucceeded), runCompleted.RunStatus)
	test.AssertLogDoesNotHaveInfo(t, "Sending notify message", execLog)
	test.AssertFileInDirectory(t, conf.LogDir, run.Run.LogFile)
	// lockRow := test.GetInfoLogLineStartingWith(t, "Created job lock at ", execLog)
	// lockFile := strings.Split(lockRow.Msg, "Created job lock at ")[1]
	// test.AssertFileInDirectory(t, conf.LockDir, lockFile)

	assert.Equal(t, int64(1), dbJob.Job.ID)
	assert.Equal(t, false, dbJob.Job.NotifyLogContent)
	assert.Equal(t, jobName, dbJob.Job.Name)

	assert.Equal(t, int64(1), dbRun.Run.ID)
	assert.Equal(t, int64(1), dbRun.Run.JobID)
	assert.Equal(t, "Succeeded", dbRun.Run.Status)
	assert.Equal(t, run.Run.LogFile, dbRun.Run.LogFile)
	assert.Equal(t, run.Run.ExecLogFile, dbRun.Run.ExecLogFile)
}

func Test_execRunExistentJob(t *testing.T) {
	ctx := context.Background()
	db := test.CreateDb(ctx, t)
	jobName := test.UniqueIdentifer()
	logFile, logger := test.CreateSysLogFile(t)
	conf := config.Config{
		LockDir: t.TempDir(),
		LogDir:  t.TempDir(),
	}

	jobId, err := db.CreateJob(ctx, data.CreateJobParams{
		Name:             jobName,
		NotifyLogContent: false,
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	run := execRun(ctx, logger, jobName, conf, db, logFile, []string{"./testdata/script-passes"})
	dbJob, err := db.GetJob(ctx, jobName)
	if err != nil {
		t.Fatal(err.Error())
	}
	dbRun, err := db.GetRun(ctx, 1)
	if err != nil {
		t.Fatal(err.Error())
	}

	execLog := test.NewLogFromFileOrFail(run.Run.ExecLogFile)
	assert.Equal(t, jobName, run.Job.Name)
	assert.Equal(t, jobId, run.Job.ID)
	assert.Equal(t, false, run.Job.NotifyLogContent)
	test.AssertFileExists(t, run.Run.ExecLogFile)
	test.AssertFileExists(t, run.Run.LogFile)
	assert.Equal(t, "Succeeded", run.Run.Status)
	test.AssertFileContents(t, "Output line 1\nOutput line 2\n", run.Run.LogFile)
	test.AssertLogDoesNotHaveInfo(t, "Job not registered. Creating new job with name "+jobName, execLog)
	test.AssertLogHasInfo(t, "Run log created at: "+run.Run.LogFile, execLog)
	runCreated := test.GetEventOrFail(t, core.EventRunCreated, execLog)
	assert.Equal(t, int64(1), runCreated.RunId)
	runCompleted := test.GetEventOrFail(t, core.EventRunCompleted, execLog)
	assert.Equal(t, string(core.RunStatusSucceeded), runCompleted.RunStatus)
	test.AssertLogDoesNotHaveInfo(t, "Sending notify message", execLog)

	assert.Equal(t, int64(1), dbJob.Job.ID)
	assert.Equal(t, false, dbJob.Job.NotifyLogContent)
	assert.Equal(t, jobName, dbJob.Job.Name)

	assert.Equal(t, int64(1), dbRun.Run.ID)
	assert.Equal(t, jobId, dbRun.Run.JobID)
	assert.Equal(t, "Succeeded", dbRun.Run.Status)
	assert.Equal(t, run.Run.LogFile, dbRun.Run.LogFile)
	assert.Equal(t, run.Run.ExecLogFile, dbRun.Run.ExecLogFile)
}

func Test_execRunScriptFails(t *testing.T) {
	ctx := context.Background()
	db := test.CreateDb(ctx, t)
	jobName := test.UniqueIdentifer()
	logFile, logger := test.CreateSysLogFile(t)
	conf := config.Config{
		LockDir: t.TempDir(),
		LogDir:  t.TempDir(),
	}
	run := execRun(ctx, logger, jobName, conf, db, logFile, []string{"./testdata/script-fails"})
	dbRun, err := db.GetRun(ctx, 1)
	if err != nil {
		t.Fatal(err.Error())
	}

	execLog := test.NewLogFromFileOrFail(run.Run.ExecLogFile)
	assert.Equal(t, "Failed", run.Run.Status)
	assert.Equal(t, "Failed", dbRun.Run.Status)
	runCompleted := test.GetEventOrFail(t, core.EventRunCompleted, execLog)
	assert.Equal(t, string(core.RunStatusFailed), runCompleted.RunStatus)
	test.AssertFileContents(t, "This script will fail\n", run.Run.LogFile)
	test.AssertLogDoesNotHaveInfo(t, "Sending notify message", execLog)
}

func Test_execRunStdoutStderr(t *testing.T) {
	ctx := context.Background()
	db := test.CreateDb(ctx, t)
	jobName := test.UniqueIdentifer()
	logFile, logger := test.CreateSysLogFile(t)
	conf := config.Config{
		LockDir: t.TempDir(),
		LogDir:  t.TempDir(),
	}
	run := execRun(ctx, logger, jobName, conf, db, logFile, []string{"./testdata/script-stdout-stderr"})
	dbRun, err := db.GetRun(ctx, 1)
	if err != nil {
		t.Fatal(err.Error())
	}

	execLog := test.NewLogFromFileOrFail(run.Run.ExecLogFile)
	assert.Equal(t, "Succeeded", run.Run.Status)
	assert.Equal(t, "Succeeded", dbRun.Run.Status)
	runCompleted := test.GetEventOrFail(t, core.EventRunCompleted, execLog)
	assert.Equal(t, string(core.RunStatusSucceeded), runCompleted.RunStatus)
	test.AssertFileContents(t, "Logging to stdout\nLogging to stderr\n", run.Run.LogFile)
	test.AssertLogDoesNotHaveInfo(t, "Sending notify message", execLog)
}

func Test_execRunSkippedRun(t *testing.T) {
	blocked := make(chan data.GetRunRow)
	ctx := context.Background()
	db := test.CreateDb(ctx, t)
	jobName := test.UniqueIdentifer()
	logFile1, logger1 := test.CreateSysLogFile(t)
	logFile2, logger2 := test.CreateSysLogFile(t)
	conf := config.Config{
		LockDir: t.TempDir(),
		LogDir:  t.TempDir(),
	}
	go func() {
		blocked <- execRun(ctx, logger1, jobName, conf, db, logFile1, []string{"./testdata/script-sleeps"})
	}()
	time.Sleep(100 * time.Millisecond)
	skippedRun := execRun(ctx, logger2, jobName, conf, db, logFile2, []string{"./testdata/script-passes"})
	successfulRun := <-blocked
	runs, err := db.GetRuns(ctx, data.GetRunsParams{
		Column1: "",
		Column2: false,
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	skippedLog := test.NewLogFromFileOrFail(skippedRun.Run.ExecLogFile)
	successfulLog := test.NewLogFromFileOrFail(successfulRun.Run.ExecLogFile)
	assert.Equal(t, 2, len(runs))
	assert.Equal(t, "Skipped", skippedRun.Run.Status)
	assert.Equal(t, "", skippedRun.Run.LogFile)
	runSkipped := test.GetEventOrFail(t, core.EventRunSkipped, skippedLog)
	assert.Equal(t, skippedRun.Run.ID, runSkipped.RunId)
	assert.Equal(t, "Succeeded", successfulRun.Run.Status)
	runCompleted := test.GetEventOrFail(t, core.EventRunCompleted, successfulLog)
	assert.Equal(t, string(core.RunStatusSucceeded), runCompleted.RunStatus)
	test.AssertLogDoesNotHaveInfo(t, "Sending notify message", skippedLog)
	test.AssertLogDoesNotHaveInfo(t, "Sending notify message", successfulLog)
}

func Test_execRunComplexCommand(t *testing.T) {
	ctx := context.Background()
	db := test.CreateDb(ctx, t)
	jobName := test.UniqueIdentifer()
	logFile, logger := test.CreateSysLogFile(t)
	conf := config.Config{
		LockDir: t.TempDir(),
		LogDir:  t.TempDir(),
	}
	run := execRun(ctx, logger, jobName, conf, db, logFile, []string{"echo \"Testing again...\" && echo \"and again...\" | awk '{ print toupper($0) }'"})
	dbRun, err := db.GetRun(ctx, 1)
	if err != nil {
		t.Fatal(err.Error())
	}

	execLog := test.NewLogFromFileOrFail(run.Run.ExecLogFile)
	assert.Equal(t, "Succeeded", run.Run.Status)
	assert.Equal(t, "Succeeded", dbRun.Run.Status)
	runCompleted := test.GetEventOrFail(t, core.EventRunCompleted, execLog)
	assert.Equal(t, string(core.RunStatusSucceeded), runCompleted.RunStatus)
	test.AssertFileContents(t, "Testing again...\nAND AGAIN...\n", run.Run.LogFile)
}
