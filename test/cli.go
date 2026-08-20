package test

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"testing"

	"github.com/samcarswell/troc/core"
)

type TrocJob struct {
	Exe string
}

type TrocRun struct {
	Exe string
}

type TrocBase struct {
	Exe string
	Run TrocRun
	Job TrocJob
}

type TrocCli struct {
	Exe  string
	Base TrocBase
}

type TrocCmd struct {
	Cmd    *exec.Cmd
	Stdout *bytes.Buffer
	Stderr *bytes.Buffer
}

func (c TrocCmd) ParseRun(t *testing.T) core.RunShow {
	var runInfo core.RunShow
	err := json.Unmarshal(c.Stdout.Bytes(), &runInfo)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	return runInfo
}

// Creates CLI with new database/config. Should only be called once per tests.
func NewTrocCli(t *testing.T, exe string) *TrocCli {
	tc := TrocCli{
		Exe: exe,
		Base: TrocBase{
			Exe: exe,
			Run: TrocRun{
				Exe: exe,
			},
			Job: TrocJob{
				Exe: exe,
			},
		},
	}
	SetupEnv(t)
	return &tc
}

func (t TrocJob) Add(name string, notifyLog bool) TrocCmd {
	args := []string{"job", "add", "--name", name}
	if notifyLog {
		args = append(args, "--notify-log")
	}
	return getCmd(t.Exe, args)
}

func (t TrocRun) List() TrocCmd {
	return getCmd(t.Exe, []string{"run", "list", "-f", "json"})
}

func (t TrocRun) ListArchived() TrocCmd {
	return getCmd(t.Exe, []string{"run", "list", "-f", "json", "--include-archived"})
}

func (t TrocRun) Kill(runId int64) TrocCmd {
	return getCmd(t.Exe, []string{"run", "kill", "-r", strconv.FormatInt(runId, 10), "--force"})
}

func (t TrocRun) Archive(runId int64) TrocCmd {
	return getCmd(t.Exe, []string{"run", "archive", "-r", strconv.FormatInt(runId, 10)})
}

func (t TrocBase) Exec(name string, script string) TrocCmd {
	return getCmd(t.Exe, []string{"exec", "--name", name, script})
}

func (t TrocBase) ExecNotify(name string, script string) TrocCmd {
	return getCmd(t.Exe, []string{"exec", "--notify", "--name", name, script})
}

func (t TrocBase) Version() TrocCmd {
	return getCmd(t.Exe, []string{"--version"})
}

func getCmd(exe string, args []string) TrocCmd {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := exec.Command(exe, args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return TrocCmd{
		Cmd:    cmd,
		Stdout: stdout,
		Stderr: stderr,
	}
}

func CmdConv[T any](t TrocCmd) T {
	var val T
	err := json.Unmarshal(t.Stdout.Bytes(), &val)
	if err != nil {
		panic(err)
	}
	return val
}

func (t TrocCmd) Run() {
	err := t.Cmd.Run()
	if err != nil {
		log.Fatalln(string(t.Stderr.Bytes()[:]))
		log.Fatalln(err)
		panic(err)
	}
}

func (t TrocCmd) Start() {
	err := t.Cmd.Start()
	if err != nil {
		panic(err)
	}
}

func (t TrocCmd) Wait() {
	err := t.Cmd.Wait()
	if err != nil {
		panic(err)
	}
}

// TODO: remove this. We should be using troc to handle this. Even though this
// Should work. Probably need a separate test to ensure that a sigterm from troc
// flows through to the child process
// func (t TrocCmd) Term() {
// 	err := t.Cmd.Process.Signal(syscall.SIGINT)
// 	if err != nil {
// 		panic(err)
// 	}
// }

func (t TrocCmd) ExecLogOrFail() Log {
	log, err := NewLogFromBuffer(*t.Stderr)
	if err != nil {
		panic(err)
	}
	return log
}

func SetupEnv(t *testing.T) {
	confDir, _ := os.MkdirTemp(os.TempDir(), "config")
	logDir := t.TempDir()
	lockDir := t.TempDir()
	setenv("TROC_CONFIG_PATH", confDir)
	setenv("TROC_DATABASE", path.Join(confDir, "troc.db"))
	setenv("TROC_LOGDIR", logDir)
	setenv("TROC_LOCKDIR", lockDir)
	setenv("TROC_LOGJSON", "true")
	t.Logf("Config directory: %s", confDir)
}

func setenv(key string, value string) {
	err := os.Setenv(key, value)
	if err != nil {
		panic(err)
	}
}
