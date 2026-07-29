package test

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path"
	"strconv"
	"testing"
)

type TrocRun struct {
	Exe string
}

type TrocBase struct {
	Exe string
	Run TrocRun
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

// Creates CLI with new database/config. Should only be called once per tests.
func NewTrocCli(t *testing.T, exe string) *TrocCli {
	tc := TrocCli{
		Exe: exe,
		Base: TrocBase{
			Exe: exe,
			Run: TrocRun{
				Exe: exe,
			},
		},
	}
	SetupEnv(t)
	return &tc
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

func (t TrocCmd) ExecLogOrFail() Log {
	log, err := NewLogFromBuffer(*t.Stderr)
	if err != nil {
		panic(err)
	}
	return log
}

func SetupEnv(t *testing.T) {
	confDir := t.TempDir()
	logDir := t.TempDir()
	lockDir := t.TempDir()
	setenv("TROC_CONFIG_PATH", confDir)
	setenv("TROC_DATABASE", path.Join(confDir, "troc.db"))
	setenv("TROC_LOGDIR", logDir)
	setenv("TROC_LOCKDIR", lockDir)
	setenv("TROC_LOGJSON", "true")
}

func setenv(key string, value string) {
	err := os.Setenv(key, value)
	if err != nil {
		panic(err)
	}
}
