package test

import (
	"path"
	"strconv"
	"strings"
	"testing"

	"github.com/samcarswell/troc/core"
	"github.com/stretchr/testify/assert"
)

func Test_Version(t *testing.T) {
	cli := NewTrocCli(t, instance.TrocExe)
	cmd := cli.Base.Version()
	cmd.Run()

	versionString := strings.TrimSpace(cmd.Stdout.String())
	if versionString == "troc version development" || versionString == "" {
		t.Fatal("Build did not provide version")
	}
}

func Test_FirstRunDefaultSettings(t *testing.T) {
	cli := NewTrocCli(t, instance.TrocExe)
	exec := cli.Base.Exec("first-job", "echo 'Hello!'")
	exec.Run()

	runInfo := exec.ParseRun(t)
	assert.FileExists(t, runInfo.LogFile)
	assert.FileExists(t, runInfo.SystemLogFile)
	AssertFileContents(t, "Hello!\n", runInfo.LogFile)
	assert.Equal(t, int64(1), runInfo.ID)
	assert.Equal(t, "first-job", runInfo.JobName)
	assert.Equal(t, string(core.RunStatusSucceeded), runInfo.Status)
	// TODO: should check the times
	assert.NotEqual(t, "", runInfo.Pid)
	assert.NotEqual(t, "", runInfo.StartTime)
	assert.NotEqual(t, "", runInfo.EndTime)
}

func Test_Kill(t *testing.T) {
	cli := NewTrocCli(t, instance.TrocExe)
	killedRunCmd := cli.Base.Exec("first-job", "echo 'Started'; sleep 60; echo 'Finished'")
	killedRunCmd.Start()

	runStartedEvent := PollUntilEventOrFail(t, killedRunCmd, core.EventRunStarted)
	killRun := cli.Base.Run.Kill(runStartedEvent.RunId)
	killRun.Run()

	sigtermEvent := PollUntilEventOrFail(t, killRun, core.EventRunSigterm)
	assert.Equal(t, runStartedEvent.RunPid, sigtermEvent.RunPid)

	terminatedEvent := PollUntilEventOrFail(t, killedRunCmd, core.EventRunTerminated)
	assert.Equal(t, runStartedEvent.RunId, terminatedEvent.RunId)
	assert.Equal(t, runStartedEvent.JobName, terminatedEvent.JobName)

	runCmd := cli.Base.Run.List()
	runCmd.Run()
	run := CmdConv[[]core.RunShow](t, runCmd)[0]
	assert.Equal(t, string(core.RunStatusTerminated), run.Status)
}

func Test_ParentEnvAccessibleToRun(t *testing.T) {
	t.Setenv("TEST_ENVVAR", "test-value")
	cli := NewTrocCli(t, instance.TrocExe)
	exec := cli.Base.Exec("test-env", "echo $TEST_ENVVAR")
	exec.Run()

	runInfo := exec.ParseRun(t)
	assert.FileExists(t, runInfo.LogFile)
	assert.FileExists(t, runInfo.SystemLogFile)
	AssertFileContents(t, "test-value\n", runInfo.LogFile)
}

func Test_ArchiveRun(t *testing.T) {
	cli := NewTrocCli(t, instance.TrocExe)
	exec := cli.Base.Exec("test-env", "echo 'working'")
	exec.Run()
	runInfo := exec.ParseRun(t)

	assert.Equal(t, false, runInfo.IsArchived)

	runCmd := cli.Base.Run.List()
	runCmd.Run()
	runs := CmdConv[[]core.RunShow](t, runCmd)
	var run1 core.RunShow
	for _, r := range runs {
		if r.ID == runInfo.ID {
			run1 = r
			break
		}
	}
	assert.NotEmpty(t, run1)

	archiveCmd := cli.Base.Run.Archive(runInfo.ID)
	archiveCmd.Run()

	log := archiveCmd.ExecLogOrFail(t)
	AssertLogHasInfo(t, "Run "+strconv.Itoa(int(runInfo.ID))+" successfully archived.", log)

	runCmd2 := cli.Base.Run.List()
	runCmd2.Run()
	runs2 := CmdConv[[]core.RunShow](t, runCmd2)
	var run2 core.RunShow
	for _, r := range runs2 {
		if r.ID == runInfo.ID {
			run2 = r
			break
		}
	}
	assert.Empty(t, run2)

	runCmd3 := cli.Base.Run.ListArchived()
	runCmd3.Run()
	runs3 := CmdConv[[]core.RunShow](t, runCmd3)
	var run3 core.RunShow
	for _, r := range runs3 {
		if r.ID == runInfo.ID {
			run3 = r
			break
		}
	}
	assert.NotEmpty(t, run3)
}

func Test_ExecRunNonExistentJob(t *testing.T) {
	cli := NewTrocCli(t, instance.TrocExe)
	logdir := t.TempDir()
	t.Setenv("TROC_LOGDIR", logdir)
	name := UniqueIdentifier()

	run := cli.Base.Exec(name, "./testdata/script-passes")
	run.Run()
	runInfo := run.ParseRun(t)
	log := run.ExecLogOrFail(t)

	job := cli.Base.Job.List()
	job.Run()
	jobs := CmdConv[[]core.JobShow](t, job)
	assert.Equal(t, 1, len(jobs))

	assert.Equal(t, name, runInfo.JobName)
	AssertFileExists(t, runInfo.SystemLogFile)
	AssertFileExists(t, runInfo.LogFile)
	assert.Equal(t, "Succeeded", runInfo.Status)
	AssertFileContents(t, "Output line 1\nOutput line 2\n", runInfo.LogFile)
	AssertLogHasInfo(t, "Job not registered. Creating new job with name "+name, log)
	AssertLogHasInfo(t, "Run log created at: "+runInfo.LogFile, log)
	runCreated := GetEventOrFail(t, core.EventRunCreated, log)
	assert.Equal(t, int64(1), runCreated.RunId)
	runCompleted := GetEventOrFail(t, core.EventRunCompleted, log)
	assert.Equal(t, string(core.RunStatusSucceeded), runCompleted.RunStatus)
	AssertLogDoesNotHaveInfo(t, "Sending notify message", log)
	AssertFileInDirectory(t, logdir, runInfo.LogFile)

	assert.Equal(t, false, jobs[0].NotifyLogContent)
}

func Test_ExecRunExistentJob(t *testing.T) {
	cli := NewTrocCli(t, instance.TrocExe)
	logdir := t.TempDir()
	t.Setenv("TROC_LOGDIR", logdir)
	name := UniqueIdentifier()

	job := cli.Base.Job.Add(name, false)
	job.Run()
	jobInfo := job.ParseJob(t)

	run := cli.Base.Exec(name, "./testdata/script-passes")
	run.Run()
	runInfo := run.ParseRun(t)
	log := run.ExecLogOrFail(t)

	assert.Equal(t, jobInfo.Name, runInfo.JobName)
	AssertFileExists(t, runInfo.SystemLogFile)
	AssertFileExists(t, runInfo.LogFile)
	assert.Equal(t, "Succeeded", runInfo.Status)
	AssertFileContents(t, "Output line 1\nOutput line 2\n", runInfo.LogFile)
	AssertLogDoesNotHaveInfo(t, "Job not registered. Creating new job with name "+name, log)
	AssertLogHasInfo(t, "Run log created at: "+runInfo.LogFile, log)
	runCreated := GetEventOrFail(t, core.EventRunCreated, log)
	assert.Equal(t, int64(1), runCreated.RunId)
	runCompleted := GetEventOrFail(t, core.EventRunCompleted, log)
	assert.Equal(t, string(core.RunStatusSucceeded), runCompleted.RunStatus)
	AssertLogDoesNotHaveInfo(t, "Sending notify message", log)

	assert.Equal(t, name, jobInfo.Name)
	assert.Equal(t, false, jobInfo.NotifyLogContent)
}

func Test_ExecRunScriptFails(t *testing.T) {
	cli := NewTrocCli(t, instance.TrocExe)
	name := UniqueIdentifier()
	run := cli.Base.Exec(name, "./testdata/script-fails")
	run.Run()
	runInfo := run.ParseRun(t)
	log := run.ExecLogOrFail(t)

	assert.Equal(t, "Failed", runInfo.Status)
	runCompleted := GetEventOrFail(t, core.EventRunCompleted, log)
	assert.Equal(t, string(core.RunStatusFailed), runCompleted.RunStatus)
	AssertFileContents(t, "This script will fail\n", runInfo.LogFile)
	AssertLogDoesNotHaveInfo(t, "Sending notify message", log)
}

func Test_ExecCapturesStdoutAndStderr(t *testing.T) {
	cli := NewTrocCli(t, instance.TrocExe)
	name := UniqueIdentifier()
	run := cli.Base.Exec(name, "./testdata/script-stdout-stderr")
	run.Run()
	runInfo := run.ParseRun(t)
	log := run.ExecLogOrFail(t)

	assert.Equal(t, "Succeeded", runInfo.Status)
	runCompleted := GetEventOrFail(t, core.EventRunCompleted, log)
	assert.Equal(t, string(core.RunStatusSucceeded), runCompleted.RunStatus)
	AssertFileContents(t, "Logging to stdout\nLogging to stderr\n", runInfo.LogFile)
	AssertLogDoesNotHaveInfo(t, "Sending notify message", log)
}

func Test_ExecComplexCommand(t *testing.T) {
	cli := NewTrocCli(t, instance.TrocExe)
	name := UniqueIdentifier()
	run := cli.Base.Exec(name, "echo \"Testing again...\" && echo \"and again...\" | awk '{ print toupper($0) }'")
	run.Run()
	runInfo := run.ParseRun(t)
	log := run.ExecLogOrFail(t)

	assert.Equal(t, "Succeeded", runInfo.Status)
	runCompleted := GetEventOrFail(t, core.EventRunCompleted, log)
	assert.Equal(t, string(core.RunStatusSucceeded), runCompleted.RunStatus)
	AssertFileContents(t, "Testing again...\nAND AGAIN...\n", runInfo.LogFile)
}

func Test_CustomNotify(t *testing.T) {
	cli := NewTrocCli(t, instance.TrocExe)
	jobName := UniqueIdentifier()
	notifyName := UniqueIdentifier()
	tmpdir := t.TempDir()
	file := path.Join(tmpdir, "output")
	t.Setenv("TROC_NOTIFY_SYSTEM", notifyName)
	t.Setenv("TROC_NOTIFY_CUSTOM_0_NAME", notifyName)
	t.Setenv("TROC_NOTIFY_CUSTOM_0_COMMAND", "./testdata/script-custom-notify")
	t.Setenv("TROC_NOTIFY_CUSTOM_0_ENVVARS_0", "CUSTOM_FILE="+file)

	run := cli.Base.ExecNotify(jobName, "echo 'done'")
	run.Run()
	runInfo := run.ParseRun(t)
	AssertFileContents(t, strconv.FormatInt(runInfo.ID, 10)+":"+jobName+" - "+runInfo.Status+"\n", file)
}

func Test_CustomNotifyFails(t *testing.T) {
	cli := NewTrocCli(t, instance.TrocExe)
	jobName := UniqueIdentifier()
	notifyName := UniqueIdentifier()
	t.Setenv("TROC_NOTIFY_SYSTEM", notifyName)
	t.Setenv("TROC_NOTIFY_CUSTOM_0_NAME", notifyName)
	t.Setenv("TROC_NOTIFY_CUSTOM_0_COMMAND", "./testdata/script-custom-notify-fails")

	run := cli.Base.ExecNotify(jobName, "echo 'done'")
	run.RunFail()
	runInfo := run.ParseRun(t)
	assert.Equal(t, "Succeeded", runInfo.Status)
	log := run.ExecLogOrFail(t)
	AssertLogHasError(t, "unable to wait for custom notify command for "+notifyName+"\nexit status 1\nunable to notify", log)
	AssertLogHasError(t, "command was run, but notification was unable to be sent", log)
}
