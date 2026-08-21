package notify

import (
	"testing"

	"github.com/samcarswell/troc/core"
)

func Test_getNotifyTextSlack(t *testing.T) {
	testData := []struct {
		name             string
		jobName          string
		runId            int64
		runStatus        core.RunStatus
		logFile          string
		notifyLogContent bool
		hostname         string
		expected         string
		showEmoji        bool
		tagStatuses      core.StatusConfig
	}{
		{"success", "test-1", 34, core.RunStatusSucceeded, "", false, "server1.com",
			`*test-1@server1.com:34* - ✅ Succeeded`, true, core.StatusConfig{}},
		{"failed", "test-2", 10000000, core.RunStatusFailed, "", false, "server1.com",
			`*test-2@server1.com:10000000* - ❌ Failed <!channel>`, true, core.StatusConfig{Failed: true}},
		{"empty-hostname", "test-4", 34, core.RunStatusSucceeded, "", false, "",
			`*test-4:34* - ✅ Succeeded`, true, core.StatusConfig{}},
		{"notify-log-content", "test-5", 34, core.RunStatusSucceeded, "testdata/example.log", true, "",
			`*test-5:34* - ✅ Succeeded
` + "```" + "\nLine one of log\nLine two of log\n```", true, core.StatusConfig{}},
		{"notify-log-content-file-does-not-exist", "test-6", 34, core.RunStatusSucceeded, "", true, "", "*test-6:34* - ✅ Succeeded", true, core.StatusConfig{}},
		{"success", "test-7", 34, core.RunStatusSucceeded, "", false, "server1.com",
			`*test-7@server1.com:34* - Succeeded`, false, core.StatusConfig{}},
		{"success-tag-config", "test-8", 10000000, core.RunStatusSucceeded, "", false, "server1.com",
			`*test-8@server1.com:10000000* - ✅ Succeeded <!channel>`, true, core.StatusConfig{Succeeded: true}},
		{"skipped-tag-config", "test-9", 10000000, core.RunStatusSkipped, "", false, "server1.com",
			`*test-9@server1.com:10000000* - ⚠️ Skipped <!channel>`, true, core.StatusConfig{Skipped: true}},
		{"running-tag-config", "test-10", 10000000, core.RunStatusRunning, "", false, "server1.com",
			`*test-10@server1.com:10000000* - 🚀 Running <!channel>`, true, core.StatusConfig{Running: true}},
		{"terminated-tag-config", "test-11", 10000000, core.RunStatusTerminated, "", false, "server1.com",
			`*test-11@server1.com:10000000* - 💥 Terminated <!channel>`, true, core.StatusConfig{Terminated: true}},
	}

	for _, d := range testData {
		t.Run(d.name, func(t *testing.T) {
			notifyStr := getNotifyTextSlack(
				core.RunNotify{
					Run: core.RunShow{
						ID:      d.runId,
						LogFile: d.logFile,
						Status:  string(d.runStatus),
					},
					Job: core.JobShow{
						Name:             d.jobName,
						NotifyLogContent: d.notifyLogContent,
					},
				},
				d.tagStatuses,
				d.hostname,
				d.showEmoji,
			)
			if notifyStr != d.expected {
				t.Error("Expected")
				t.Error(d.expected)
				t.Error("Actual")
				t.Fatal(notifyStr)
			}
		})
	}
}

func Test_getNotifyTextNtfy(t *testing.T) {
	testData := []struct {
		name             string
		jobName          string
		runId            int64
		runStatus        core.RunStatus
		logFile          string
		notifyLogContent bool
		hostname         string
		expected         string
		showEmoji        bool
	}{
		{"success", "test-1", 34, core.RunStatusSucceeded, "", false, "server1.com",
			`*test-1@server1.com:34* - ✅ Succeeded`, true},
		{"failed", "test-2", 10000000, core.RunStatusFailed, "", false, "server1.com",
			`*test-2@server1.com:10000000* - ❌ Failed`, true},
		{"empty-hostname", "test-4", 34, core.RunStatusSucceeded, "", false, "",
			`*test-4:34* - ✅ Succeeded`, true},
		{"notify-log-content", "test-5", 34, core.RunStatusSucceeded, "testdata/example.log", true, "",
			`*test-5:34* - ✅ Succeeded
` + "```" + "\nLine one of log\nLine two of log\n```", true},
		{"notify-log-content-file-does-not-exist", "test-6", 34, core.RunStatusSucceeded, "", true, "", "*test-6:34* - ✅ Succeeded", true},
		{"success", "test-7", 34, core.RunStatusSucceeded, "", false, "server1.com",
			`*test-7@server1.com:34* - Succeeded`, false},
		{"success-tag-config", "test-8", 10000000, core.RunStatusSucceeded, "", false, "server1.com",
			`*test-8@server1.com:10000000* - ✅ Succeeded`, true},
		{"skipped-tag-config", "test-9", 10000000, core.RunStatusSkipped, "", false, "server1.com",
			`*test-9@server1.com:10000000* - ⚠️ Skipped`, true},
		{"running-tag-config", "test-10", 10000000, core.RunStatusRunning, "", false, "server1.com",
			`*test-10@server1.com:10000000* - 🚀 Running`, true},
		{"terminated-tag-config", "test-11", 10000000, core.RunStatusTerminated, "", false, "server1.com",
			`*test-11@server1.com:10000000* - 💥 Terminated`, true},
	}

	for _, d := range testData {
		t.Run(d.name, func(t *testing.T) {
			notifyStr := getNotifyTextNtfy(
				core.RunNotify{
					Run: core.RunShow{
						ID:      d.runId,
						LogFile: d.logFile,
						Status:  string(d.runStatus),
					},
					Job: core.JobShow{
						Name:             d.jobName,
						NotifyLogContent: d.notifyLogContent,
					},
				},
				d.hostname,
				d.showEmoji,
			)
			if notifyStr != d.expected {
				t.Error("Expected")
				t.Error(d.expected)
				t.Error("Actual")
				t.Fatal(notifyStr)
			}
		})
	}
}
