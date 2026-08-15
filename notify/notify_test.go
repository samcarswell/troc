package notify

import (
	"testing"

	"github.com/samcarswell/troc/config"
	"github.com/samcarswell/troc/core"
	"github.com/samcarswell/troc/data"
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
		tagStatuses      config.StatusConfig
	}{
		{"success", "test-1", 34, core.RunStatusSucceeded, "", false, "server1.com",
			`*test-1@server1.com:34* - ✅ Succeeded`, true, config.StatusConfig{}},
		{"failed", "test-2", 10000000, core.RunStatusFailed, "", false, "server1.com",
			`*test-2@server1.com:10000000* - ❌ Failed <!channel>`, true, config.StatusConfig{Failed: true}},
		{"empty-hostname", "test-4", 34, core.RunStatusSucceeded, "", false, "",
			`*test-4:34* - ✅ Succeeded`, true, config.StatusConfig{}},
		{"notify-log-content", "test-5", 34, core.RunStatusSucceeded, "testdata/example.log", true, "",
			`*test-5:34* - ✅ Succeeded
` + "```" + "\nLine one of log\nLine two of log\n```", true, config.StatusConfig{}},
		{"notify-log-content-file-does-not-exist", "test-6", 34, core.RunStatusSucceeded, "", true, "", "*test-6:34* - ✅ Succeeded", true, config.StatusConfig{}},
		{"success", "test-7", 34, core.RunStatusSucceeded, "", false, "server1.com",
			`*test-7@server1.com:34* - Succeeded`, false, config.StatusConfig{}},
		{"success-tag-config", "test-8", 10000000, core.RunStatusSucceeded, "", false, "server1.com",
			`*test-8@server1.com:10000000* - ✅ Succeeded <!channel>`, true, config.StatusConfig{Succeeded: true}},
		{"skipped-tag-config", "test-9", 10000000, core.RunStatusSkipped, "", false, "server1.com",
			`*test-9@server1.com:10000000* - ⚠️ Skipped <!channel>`, true, config.StatusConfig{Skipped: true}},
		{"running-tag-config", "test-10", 10000000, core.RunStatusRunning, "", false, "server1.com",
			`*test-10@server1.com:10000000* - 🚀 Running <!channel>`, true, config.StatusConfig{Running: true}},
		{"terminated-tag-config", "test-11", 10000000, core.RunStatusTerminated, "", false, "server1.com",
			`*test-11@server1.com:10000000* - 💥 Terminated <!channel>`, true, config.StatusConfig{Terminated: true}},
	}

	for _, d := range testData {
		t.Run(d.name, func(t *testing.T) {
			notifyStr := getNotifyText(
				data.GetRunRow{
					Run: data.Run{
						ID:      d.runId,
						LogFile: d.logFile,
						Status:  string(d.runStatus),
					},
					Job: data.Job{
						Name:             d.jobName,
						NotifyLogContent: d.notifyLogContent,
					},
				},
				d.tagStatuses,
				d.hostname,
				d.showEmoji,
				SlackOpts,
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

func Test_getNotifyTextCampfire(t *testing.T) {
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
		tagStatuses      config.StatusConfig
	}{
		{"success", "test-1", 34, core.RunStatusSucceeded, "", false, "server1.com",
			`<b>test-1@server1.com:34</b> - ✅ Succeeded`, true, config.StatusConfig{}},
		{"failed", "test-2", 10000000, core.RunStatusFailed, "", false, "server1.com",
			`<b>test-2@server1.com:10000000</b> - ❌ Failed`, true, config.StatusConfig{Failed: true}},
		{"empty-hostname", "test-4", 34, core.RunStatusSucceeded, "", false, "",
			`<b>test-4:34</b> - ✅ Succeeded`, true, config.StatusConfig{}},
		{"notify-log-content", "test-5", 34, core.RunStatusSucceeded, "testdata/example.log", true, "",
			`<b>test-5:34</b> - ✅ Succeeded
` + "<pre>" + "\nLine one of log\nLine two of log\n</pre>", true, config.StatusConfig{}},
		{"notify-log-content-file-does-not-exist", "test-6", 34, core.RunStatusSucceeded, "", true, "", "<b>test-6:34</b> - ✅ Succeeded", true, config.StatusConfig{}},
		{"success", "test-7", 34, core.RunStatusSucceeded, "", false, "server1.com",
			`<b>test-7@server1.com:34</b> - Succeeded`, false, config.StatusConfig{}},
		{"success-tag-config", "test-8", 10000000, core.RunStatusSucceeded, "", false, "server1.com",
			`<b>test-8@server1.com:10000000</b> - ✅ Succeeded`, true, config.StatusConfig{Succeeded: true}},
		{"skipped-tag-config", "test-9", 10000000, core.RunStatusSkipped, "", false, "server1.com",
			`<b>test-9@server1.com:10000000</b> - ⚠️ Skipped`, true, config.StatusConfig{Skipped: true}},
		{"running-tag-config", "test-10", 10000000, core.RunStatusRunning, "", false, "server1.com",
			`<b>test-10@server1.com:10000000</b> - 🚀 Running`, true, config.StatusConfig{Running: true}},
		{"terminated-tag-config", "test-11", 10000000, core.RunStatusTerminated, "", false, "server1.com",
			`<b>test-11@server1.com:10000000</b> - 💥 Terminated`, true, config.StatusConfig{Terminated: true}},
	}

	for _, d := range testData {
		t.Run(d.name, func(t *testing.T) {
			notifyStr := getNotifyText(
				data.GetRunRow{
					Run: data.Run{
						ID:      d.runId,
						LogFile: d.logFile,
						Status:  string(d.runStatus),
					},
					Job: data.Job{
						Name:             d.jobName,
						NotifyLogContent: d.notifyLogContent,
					},
				},
				d.tagStatuses,
				d.hostname,
				d.showEmoji,
				CampfireOpts,
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
