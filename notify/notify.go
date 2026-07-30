package notify

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/samcarswell/troc/config"
	"github.com/samcarswell/troc/core"
)

type notifySystemOpts struct {
	BoldStart             string
	BoldEnd               string
	PreformattedTextStart string
	PreformattedTextEnd   string
	TagChannel            string
}

var SlackOpts = notifySystemOpts{
	BoldStart:             "*",
	BoldEnd:               "*",
	PreformattedTextStart: "```",
	PreformattedTextEnd:   "```",
	TagChannel:            " <!channel>",
}

var CampfireOpts = notifySystemOpts{
	BoldStart:             "<b>",
	BoldEnd:               "</b>",
	PreformattedTextStart: "<pre>",
	PreformattedTextEnd:   "</pre>",
	TagChannel:            "",
}

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
	switch conf.Notify.System {
	case "slack":
		return notifySlack(conf.Notify.Slack, getNotifyTextSlack(
			run,
			conf.Notify.Status,
			conf.Notify.Hostname,
			conf.Display.Emoji,
			SlackOpts,
		))
	case "campfire":
		return notifyCampfire(conf.Notify.Campfire, getNotifyTextSlack(
			run,
			conf.Notify.Status,
			conf.Notify.Hostname,
			conf.Display.Emoji,
			CampfireOpts,
		))
	default:
		// TODO: this should actually fail at config parsing
		return false, errors.New("unknown system: " + conf.Notify.System)
	}
}

// Returns the notification test for a run.
// This is designed to ignore incorrect inputs; ensuring a notification is sent
// is critical; if it's missing some information, that's acceptable.
func getNotifyTextSlack(
	run RunNotifyInfo,
	tagStatuses config.StatusConfig,
	hostname string,
	showEmoji bool,
	opts notifySystemOpts,
) string {
	return opts.BoldStart + run.Name + hostnameIfExists(hostname) + ":" +
		strconv.FormatInt(run.Id, 10) + opts.BoldEnd + " - " +
		core.FormatStatus(run.Status, showEmoji) +
		tagChannelIfStatusConfigured(run.Status, tagStatuses, opts) +
		logFileAndOutput(run.NotifyLogContent, run.LogFile, opts)
}

func logFileAndOutput(
	notifyLogContent bool,
	logFile string,
	opts notifySystemOpts,
) string {
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
	return "\n" + opts.PreformattedTextStart + "\n" + string(logContent) + opts.PreformattedTextEnd
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
	opts notifySystemOpts,
) string {
	if (status == core.RunStatusRunning && tagStatuses.Running) ||
		(status == core.RunStatusSkipped && tagStatuses.Skipped) ||
		(status == core.RunStatusSucceeded && tagStatuses.Succeeded) ||
		(status == core.RunStatusFailed && tagStatuses.Failed) ||
		(status == core.RunStatusTerminated && tagStatuses.Terminated) {
		return opts.TagChannel
	}
	return ""
}

func notifyCampfire(conf config.CampfireConfig, text string) (bool, error) {
	r, _ := http.NewRequest("POST", conf.Domain+"/rooms/"+conf.RoomId+"/"+conf.Token+"/messages", strings.NewReader(text))

	client := &http.Client{}
	res, err := client.Do(r)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode > 299 {
		return false, errors.New("failed to notity: " + res.Status)
	}

	return true, nil
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
