package test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/samcarswell/troc/config"
	"github.com/samcarswell/troc/core"
	"github.com/samcarswell/troc/data"
	"github.com/stretchr/testify/assert"
)

var (
	_, b, _, _  = runtime.Caller(0)
	basepath    = filepath.Dir(b)
	letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")
)

type logRow struct {
	Time      time.Time      `json:"time"`
	Level     string         `json:"level"`
	Msg       string         `json:"msg"`
	Event     string         `json:"event"`
	RunId     int64          `json:"run_id"`
	JobName   string         `json:"job_name"`
	RunStatus string         `json:"run_status"`
	RunPid    int            `json:"run_pid"`
	Data      map[string]any `json:"data"`
}

func MigrationsDir() string {
	return path.Join(basepath, "..", "db", "migrations")
}

func CreateDb(ctx context.Context, t *testing.T) *data.Queries {
	dir := t.TempDir()
	dbPath := path.Join(dir, "troc.db")
	t.Log("Creating database at " + dbPath)
	f := os.DirFS(MigrationsDir())
	return data.New(config.CreateOrUpdateDatabase(
		f,
		context.Background(),
		dbPath,
		".", // Not sure why this works. Passing the correct path doesn't work
	))
}

func CreateSysLogFile(t *testing.T) (string, *slog.Logger) {
	logFile, err := core.CreateSyslog(t.TempDir())
	if err != nil {
		t.Error(err)
	}
	t.Log("Created syslog file at " + logFile.Name())
	l := CreateJsonLogger(logFile)
	slog.SetDefault(l)

	return logFile.Name(), l
}

func UniqueIdentifier() string {
	b := make([]rune, 10)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func CreateJsonLogger(logFile *os.File) *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.MultiWriter(logFile, os.Stderr), &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

func GetInfoLogLineStartingWith(t *testing.T, text string, log Log) logRow {
	for _, row := range log.Rows {
		if row.Level == "INFO" && strings.HasPrefix(row.Msg, text) {
			return row
		}
	}
	t.Fatalf("log line with level %s and message: %s not found", "INFO", text)
	return logRow{}
}

func AssertFileExists(t *testing.T, path string) {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		t.Fatalf("file does not exist at %s", path)
	}
}

func AssertFileInDirectory(t *testing.T, dir string, path string) {
	fileDir := filepath.Dir(path)
	assert.Equal(t, dir, fileDir)
}

func AssertFileContents(t *testing.T, expected string, path string) {
	file, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err.Error())
	}
	assert.Equal(t, expected, string(file))
}

func PollUntilEventOrFail(
	t *testing.T,
	cmd TrocCmd,
	event core.Event,
) *logRow {
	timeout := 5000
	delay := 10
	for i := 0; i < timeout/delay; i++ {
		log := cmd.ExecLogOrFail(t)
		for _, row := range log.Rows {
			if row.Event == string(event) {
				return &row
			}
		}
		time.Sleep(time.Millisecond * time.Duration(delay))
	}
	t.Fatalf("%s", "could not find event "+string(event))
	return nil
}

func GetEventOrFail(t *testing.T, event core.Event, log Log) *logRow {
	for _, row := range log.Rows {
		if row.Event == string(event) {
			return &row
		}
	}
	t.Fatalf("log does not contain event %s", string(event))

	return nil
}

func AssertLogHasInfo(t *testing.T, text string, log Log) {
	assertLogHasLine(t, "INFO", text, log)
}

func AssertLogHasWarn(t *testing.T, text string, log Log) {
	assertLogHasLine(t, "WARN", text, log)
}

func AssertLogHasError(t *testing.T, text string, log Log) {
	assertLogHasLine(t, "ERROR", text, log)
}

func AssertLogDoesNotHaveInfo(t *testing.T, text string, log Log) {
	assertLogDoesNotHaveLine(t, "INFO", text, log)
}

func AssertLogDoesNotHaveWarn(t *testing.T, text string, log Log) {
	assertLogDoesNotHaveLine(t, "WARN", text, log)
}

func assertLogHasLine(t *testing.T, level string, text string, log Log) {
	for _, row := range log.Rows {
		if row.Level == level && row.Msg == text {
			return
		}
	}
	t.Fatalf("log does not contain line with level %s and message: %s", level, text)
}

func assertLogDoesNotHaveLine(t *testing.T, level string, text string, log Log) {
	for _, row := range log.Rows {
		if row.Level == level && row.Msg == text {
			t.Fatalf("log contains line with level %s and message: %s", level, text)
		}
	}
}

type Log struct {
	Rows []logRow
}

func NewLogFromBuffer(data bytes.Buffer) (Log, error) {
	r := bytes.NewReader(data.Bytes())
	return NewLogFromReader(r)
}

func NewLogFromFileOrFail(path string) Log {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	defer file.Close()
	log, err := NewLogFromReader(file)
	if err != nil {
		panic(err)
	}
	return log
}

func NewLogFromReader(reader io.Reader) (Log, error) {
	var log = Log{}
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		var row logRow
		err := json.Unmarshal([]byte(scanner.Text()), &row)
		if err != nil {
			return Log{}, nil
		}
		err = json.Unmarshal([]byte(scanner.Text()), &row.Data)
		if err != nil {
			return Log{}, nil
		}
		delete(row.Data, "time")
		delete(row.Data, "level")
		delete(row.Data, "msg")
		delete(row.Data, "event")
		delete(row.Data, "run_id")
		delete(row.Data, "job_name")
		delete(row.Data, "run_status")
		delete(row.Data, "run_pid")
		log.Rows = append(log.Rows, row)
	}
	return log, nil
}
