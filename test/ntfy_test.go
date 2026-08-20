package test

import (
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/samcarswell/troc/core"
	"github.com/stretchr/testify/assert"
)

func Test_NotifyFailure(t *testing.T) {
	ins := GetInstance()
	testData := []struct {
		notifyStatusFailed bool
		ntfyPriority       int
	}{
		{false, 1},
		{true, 3},
	}
	t.Setenv("TROC_NOTIFY_SYSTEM", "ntfy")
	t.Setenv("TROC_NOTIFY_NTFY_TOKEN", ins.NtfyToken)
	t.Setenv("TROC_NOTIFY_NTFY_DOMAIN", ins.NtfyUri)
	jobName := uuid.New().String()
	randomFile := uuid.New().String() + ".file"
	cli := NewTrocCli(t, ins.TrocExe)
	for _, d := range testData {
		topic := uuid.New().String()
		hostname := uuid.New().String()
		t.Setenv("TROC_NOTIFY_NTFY_TOPIC", topic)
		t.Setenv("TROC_NOTIFY_HOSTNAME", hostname)
		t.Setenv("TROC_NOTIFY_STATUS_FAILED", strconv.FormatBool(d.notifyStatusFailed))
		exec := cli.Base.ExecNotify(jobName, "cat "+randomFile)
		exec.Run()
		runInfo := exec.ParseRun(t)
		msg := NtfyPollMsgs(topic, ins.NtfyToken, ins.NtfyUri)[0]
		assert.Equal(t, "*"+jobName+"@"+hostname+":"+strconv.FormatInt(runInfo.ID, 10)+"* - ❌ Failed", msg.Message)
		assert.Equal(t, d.ntfyPriority, msg.Priority)
	}
}

func Test_NotifySucceeded(t *testing.T) {
	ins := GetInstance()
	testData := []struct {
		notifyStatusSucceeded bool
		ntfyPriority          int
	}{
		{false, 1},
		{true, 3},
	}
	t.Setenv("TROC_NOTIFY_SYSTEM", "ntfy")
	t.Setenv("TROC_NOTIFY_NTFY_TOKEN", ins.NtfyToken)
	t.Setenv("TROC_NOTIFY_NTFY_DOMAIN", ins.NtfyUri)
	jobName := uuid.New().String()
	cli := NewTrocCli(t, ins.TrocExe)
	for _, d := range testData {
		topic := uuid.New().String()
		hostname := uuid.New().String()
		t.Setenv("TROC_NOTIFY_NTFY_TOPIC", topic)
		t.Setenv("TROC_NOTIFY_HOSTNAME", hostname)
		t.Setenv("TROC_NOTIFY_STATUS_SUCCEEDED", strconv.FormatBool(d.notifyStatusSucceeded))
		exec := cli.Base.ExecNotify(jobName, "echo hello")
		exec.Run()
		runInfo := exec.ParseRun(t)
		msg := NtfyPollMsgs(topic, ins.NtfyToken, ins.NtfyUri)[0]
		assert.Equal(t, "*"+jobName+"@"+hostname+":"+strconv.FormatInt(runInfo.ID, 10)+"* - ✅ Succeeded", msg.Message)
		assert.Equal(t, d.ntfyPriority, msg.Priority)
	}
}

func Test_NotifySkipped(t *testing.T) {
	ins := GetInstance()
	testData := []struct {
		notifyStatusSkipped bool
		ntfyPriority        int
	}{
		{false, 1},
		{true, 3},
	}
	t.Setenv("TROC_NOTIFY_SYSTEM", "ntfy")
	t.Setenv("TROC_NOTIFY_NTFY_TOKEN", ins.NtfyToken)
	t.Setenv("TROC_NOTIFY_NTFY_DOMAIN", ins.NtfyUri)
	cli := NewTrocCli(t, ins.TrocExe)
	for _, d := range testData {
		topic := uuid.New().String()
		hostname := uuid.New().String()
		jobName := uuid.New().String()
		cli.Base.Job.Add(jobName, false).Run()
		t.Setenv("TROC_NOTIFY_NTFY_TOPIC", topic)
		t.Setenv("TROC_NOTIFY_HOSTNAME", hostname)
		t.Setenv("TROC_NOTIFY_STATUS_SKIPPED", strconv.FormatBool(d.notifyStatusSkipped))
		exec1 := cli.Base.ExecNotify(jobName, "sleep 2; echo done")
		exec1.Start()
		runStartedEvent := PollUntilEventOrFail(t, exec1, core.EventRunStarted)

		exec2 := cli.Base.ExecNotify(jobName, "echo done")
		exec2.Run()
		runinfo := exec2.ParseRun(t)
		msgs := NtfyPollMsgs(topic, ins.NtfyToken, ins.NtfyUri)
		assert.Equal(t, 1, len(msgs))
		msg := msgs[0]
		assert.Equal(t, "*"+jobName+"@"+hostname+":"+strconv.FormatInt(runinfo.ID, 10)+"* - ⚠️ Skipped", msg.Message)
		assert.Equal(t, d.ntfyPriority, msg.Priority)

		cli.Base.Run.Kill(runStartedEvent.RunId).Run()
	}
}

func Test_NotifyTerminated(t *testing.T) {
	ins := GetInstance()
	testData := []struct {
		notifyStatusTerminated bool
		ntfyPriority           int
	}{
		{false, 1},
		{true, 3},
	}
	t.Setenv("TROC_NOTIFY_SYSTEM", "ntfy")
	t.Setenv("TROC_NOTIFY_NTFY_TOKEN", ins.NtfyToken)
	t.Setenv("TROC_NOTIFY_NTFY_DOMAIN", ins.NtfyUri)
	cli := NewTrocCli(t, ins.TrocExe)
	for _, d := range testData {
		topic := uuid.New().String()
		hostname := uuid.New().String()
		jobName := uuid.New().String()
		cli.Base.Job.Add(jobName, false).Run()
		t.Setenv("TROC_NOTIFY_NTFY_TOPIC", topic)
		t.Setenv("TROC_NOTIFY_HOSTNAME", hostname)
		t.Setenv("TROC_NOTIFY_STATUS_TERMINATED", strconv.FormatBool(d.notifyStatusTerminated))
		exec := cli.Base.ExecNotify(jobName, "sleep 60; echo hello")
		exec.Start()
		runStartedEvent := PollUntilEventOrFail(t, exec, core.EventRunStarted)

		cli.Base.Run.Kill(runStartedEvent.RunId).Run()
		exec.Wait()

		runInfo := exec.ParseRun(t)
		msg := NtfyPollMsgs(topic, ins.NtfyToken, ins.NtfyUri)[0]
		assert.Equal(t, "*"+jobName+"@"+hostname+":"+strconv.FormatInt(runInfo.ID, 10)+"* - 💥 Terminated", msg.Message)
		assert.Equal(t, d.ntfyPriority, msg.Priority)
	}
}

func Test_NotifyEmoji(t *testing.T) {
	ins := GetInstance()
	testData := []struct {
		showEmoji bool
		emojiText string
	}{
		{false, ""},
		{true, " ✅"},
	}
	t.Setenv("TROC_NOTIFY_SYSTEM", "ntfy")
	t.Setenv("TROC_NOTIFY_NTFY_TOKEN", ins.NtfyToken)
	t.Setenv("TROC_NOTIFY_NTFY_DOMAIN", ins.NtfyUri)
	jobName := uuid.New().String()
	cli := NewTrocCli(t, ins.TrocExe)
	for _, d := range testData {
		topic := uuid.New().String()
		hostname := uuid.New().String()
		t.Setenv("TROC_NOTIFY_NTFY_TOPIC", topic)
		t.Setenv("TROC_NOTIFY_HOSTNAME", hostname)
		t.Setenv("TROC_DISPLAY_EMOJI", strconv.FormatBool(d.showEmoji))
		exec := cli.Base.ExecNotify(jobName, "echo hello")
		exec.Run()
		runInfo := exec.ParseRun(t)
		msg := NtfyPollMsgs(topic, ins.NtfyToken, ins.NtfyUri)[0]
		assert.Equal(t, "*"+jobName+"@"+hostname+":"+strconv.FormatInt(runInfo.ID, 10)+"* -"+d.emojiText+" Succeeded", msg.Message)
	}
}

// TODO: test for multi line command output to be formatted correctly
