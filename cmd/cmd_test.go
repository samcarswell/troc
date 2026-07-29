package cmd

import (
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/samcarswell/troc/core"
	"github.com/samcarswell/troc/test"
	"github.com/stretchr/testify/assert"
)

var dir string

var trocExe string

func TestMain(m *testing.M) {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	buildDir := path.Join(pwd, "..")
	dir, err = os.MkdirTemp(os.TempDir(), "build")
	if err != nil {
		log.Fatal(err)
	}
	cmd := exec.Command("./build", dir)
	cmd.Dir = buildDir
	err = cmd.Run()
	if err != nil {
		panic(err)
	}
	trocExe = path.Join(dir, "troc")
	defer os.RemoveAll(dir)
	exitVal := m.Run()
	os.Exit(exitVal)
}

func Test_Version(t *testing.T) {
	cli := test.NewTrocCli(t, trocExe)
	cmd := cli.Base.Version()
	cmd.Run()

	versionString := strings.TrimSpace(cmd.Stdout.String())
	if versionString == "troc version development" || versionString == "" {
		t.Fatal("Build did not provide version")
	}
}

func Test_FirstRunDefaultSettings(t *testing.T) {
	cli := test.NewTrocCli(t, trocExe)
	exec := cli.Base.Exec("first-job", "echo 'Hello!'")
	exec.Run()

	var runInfo core.RunShow
	err := json.Unmarshal(exec.Stdout.Bytes(), &runInfo)
	if err != nil {
		panic(err)
	}
	assert.FileExists(t, runInfo.LogFile)
	assert.FileExists(t, runInfo.SystemLogFile)
	test.AssertFileContents(t, "Hello!\n", runInfo.LogFile)
	assert.Equal(t, int64(1), runInfo.ID)
	assert.Equal(t, "first-job", runInfo.JobName)
	assert.Equal(t, string(core.RunStatusSucceeded), runInfo.Status)
	// TODO: should check the times
	assert.NotEqual(t, "", runInfo.Pid)
	assert.NotEqual(t, "", runInfo.StartTime)
	assert.NotEqual(t, "", runInfo.EndTime)
}

func Test_Kill(t *testing.T) {
	cli := test.NewTrocCli(t, trocExe)
	killedRunCmd := cli.Base.Exec("first-job", "echo 'Started'; sleep 60; echo 'Finished'")
	killedRunCmd.Start()

	time.Sleep(10 * time.Millisecond)

	logRunning := killedRunCmd.ExecLogOrFail()
	runStartedEvent := test.GetEventOrFail(t, core.EventRunStarted, logRunning)
	killRun := cli.Base.Run.Kill(runStartedEvent.RunId)
	killRun.Run()

	runKill := killRun.ExecLogOrFail()
	time.Sleep(10 * time.Millisecond)
	sigtermEvent := test.GetEventOrFail(t, core.EventRunSigterm, runKill)
	assert.Equal(t, runStartedEvent.RunPid, sigtermEvent.RunPid)

	time.Sleep(10 * time.Millisecond)

	logKilled := killedRunCmd.ExecLogOrFail()
	time.Sleep(10 * time.Millisecond)
	terminatedEvent := test.GetEventOrFail(t, core.EventRunTerminated, logKilled)
	assert.Equal(t, runStartedEvent.RunId, terminatedEvent.RunId)
	assert.Equal(t, runStartedEvent.JobName, terminatedEvent.JobName)

	runCmd := cli.Base.Run.List()
	runCmd.Run()
	run := test.CmdConv[[]core.RunShow](runCmd)[0]
	assert.Equal(t, string(core.RunStatusTerminated), run.Status)
}

func Test_ParentEnvAccessibleToRun(t *testing.T) {
	err := os.Setenv("TEST_ENVVAR", "test-value")
	if err != nil {
		panic(err)
	}
	cli := test.NewTrocCli(t, trocExe)
	exec := cli.Base.Exec("test-env", "echo $TEST_ENVVAR")
	exec.Run()

	var runInfo core.RunShow
	err = json.Unmarshal(exec.Stdout.Bytes(), &runInfo)
	if err != nil {
		panic(err)
	}
	assert.FileExists(t, runInfo.LogFile)
	assert.FileExists(t, runInfo.SystemLogFile)
	test.AssertFileContents(t, "test-value\n", runInfo.LogFile)
}

func Test_ArchiveRun(t *testing.T) {
	cli := test.NewTrocCli(t, trocExe)
	exec := cli.Base.Exec("test-env", "echo 'working'")
	exec.Run()
	var runInfo core.RunShow
	err := json.Unmarshal(exec.Stdout.Bytes(), &runInfo)
	if err != nil {
		panic(err)
	}

	assert.Equal(t, false, runInfo.IsArchived)

	runCmd := cli.Base.Run.List()
	runCmd.Run()
	runs := test.CmdConv[[]core.RunShow](runCmd)
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

	log := archiveCmd.ExecLogOrFail()
	test.AssertLogHasInfo(t, "Run "+strconv.Itoa(int(runInfo.ID))+" successfully archived.", log)

	runCmd2 := cli.Base.Run.List()
	runCmd2.Run()
	runs2 := test.CmdConv[[]core.RunShow](runCmd2)
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
	runs3 := test.CmdConv[[]core.RunShow](runCmd3)
	var run3 core.RunShow
	for _, r := range runs3 {
		if r.ID == runInfo.ID {
			run3 = r
			break
		}
	}
	assert.NotEmpty(t, run3)
}
