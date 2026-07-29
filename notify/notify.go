package notify

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/samcarswell/troc/config"
	"github.com/samcarswell/troc/core"
)

type slackPost struct {
	Channel string `json:"channel"`
	Text    string `json:"text"`
}

type slackResp struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
}

type RunNotifyInfo struct {
	Name             string
	NotifyLogContent bool
	Id               int64
	Status           core.RunStatus
	LogFile          string
}

const slackPostMessage = "https://slack.com/api/chat.postMessage"

func NotifyRun(
	conf config.Config,
	run RunNotifyInfo,
) (bool, error) {
	slackStr := getNotifyText(
		run,
		conf.Notify.Status,
		conf.Notify.Hostname,
		conf.Display.Emoji,
	)
	return notifySlack(conf.Notify.Slack, slackStr)
}

// Returns the notification test for a run.
// This is designed to ignore incorrect inputs; ensuring a notification is sent
// is critical; if it's missing some information, that's acceptable.
func getNotifyText(
	run RunNotifyInfo,
	tagStatuses config.StatusConfig,
	hostname string,
	showEmoji bool,
) string {
	return "*" + run.Name + hostnameIfExists(hostname) + ":" +
		strconv.FormatInt(run.Id, 10) + "* - " +
		core.FormatStatus(run.Status, showEmoji) +
		tagChannelIfStatusConfigured(run.Status, tagStatuses) +
		logFileAndOutput(run.NotifyLogContent, run.LogFile)
}

func logFileAndOutput(notifyLogContent bool, logFile string) string {
	if !notifyLogContent {
		return ""
	}
	if logFile == "" {
		return ""
	}
	logContent, err := os.ReadFile(logFile)
	if err != nil {
		log.Printf("Unable to read logfile: %s. Notify message will omit it.", logFile)
		return ""
	}
	return "\n" + "```\n" + string(logContent) + "```"
}

func hostnameIfExists(hostname string) string {
	if hostname == "" {
		return ""
	}
	return "@" + hostname
}

func tagChannelIfStatusConfigured(
	status core.RunStatus,
	tagStatuses config.StatusConfig,
) string {
	if (status == core.RunStatusRunning && tagStatuses.Running) ||
		(status == core.RunStatusSkipped && tagStatuses.Skipped) ||
		(status == core.RunStatusSucceeded && tagStatuses.Succeeded) ||
		(status == core.RunStatusFailed && tagStatuses.Failed) ||
		(status == core.RunStatusTerminated && tagStatuses.Terminated) {
		return " <!channel>"
	}
	return ""
}

func notifySlack(slackConf config.SlackConfig, text string) (bool, error) {
	postJson, err := json.Marshal(slackPost{
		Channel: slackConf.Channel,
		Text:    text,
	})
	if err != nil {
		return false, err
	}

	r, err := http.NewRequest("POST", slackPostMessage, bytes.NewBuffer(postJson))
	r.Header.Add("Authorization", "Bearer "+slackConf.Token)
	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("charset", "utf-8")

	client := &http.Client{}
	res, err := client.Do(r)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	slackResp := &slackResp{}
	err = json.NewDecoder(res.Body).Decode(slackResp)
	if err != nil {
		return false, err
	}

	if res.StatusCode != 200 || !slackResp.Ok {
		return false, errors.New(slackResp.Error)
	}

	return true, nil
}
