package notify

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/samcarswell/troc/config"
	"github.com/samcarswell/troc/core"
	"github.com/samcarswell/troc/data"
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

const markdownBold = "*"
const markdownPreformattedText = "```"
const slackPostMessage = "https://slack.com/api/chat.postMessage"
const slackTagChannel = "<!channel>"

func NotifyRun(
	conf config.Config,
	run data.GetRunRow,
	logger *slog.Logger,
) (bool, error) {
	switch conf.Notify.System {
	case config.ConfigNotifySystemSlack:
		return notifySlack(conf.Notify.Slack, getNotifyTextSlack(
			run,
			conf.Notify.Status,
			conf.Notify.Hostname,
			conf.Display.Emoji,
		))
	case config.ConfigNotifySystemNfty:
		return notifyNfty(
			conf.Notify.Ntfy,
			getNotifyTextNtfy(
				run,
				conf.Notify.Hostname,
				conf.Display.Emoji,
			),
			core.RunStatus(run.Run.Status),
			conf.Notify.Status,
		)

	default:
		var systemConf *config.CustomNotifySystemConfigItem
		var found = false
		for _, row := range conf.Notify.Custom {
			if row.Name == conf.Notify.System {
				systemConf = &row
				found = true
			}
		}
		if !found {
			return false, errors.New("unknown notification system: " + conf.Notify.System)
		}
		logger.Info("Executing custom notifier: " + conf.Notify.System)

		var runNotifyCmd *exec.Cmd
		if strings.ContainsAny(systemConf.Command, " ") {
			cmds := strings.Split(systemConf.Command, " ")
			runNotifyCmd = exec.Command(cmds[0], cmds[1:]...)
		} else {
			runNotifyCmd = exec.Command(systemConf.Command)
		}

		var jsonBytes, err = json.Marshal(run)
		if err != nil {
			return false, err
		}

		runNotifyCmd.Env = systemConf.EnvVars

		runNotifyCmd.Stdin = bytes.NewReader(jsonBytes)
		runNotifyCmd.Stdout = notifyLogger{
			Logger: logger,
			Name:   systemConf.Name,
		}
		runNotifyCmd.Stderr = notifyLogger{
			Logger: logger,
			Name:   systemConf.Name,
		}
		err = runNotifyCmd.Start()
		if err != nil {
			return false, errors.Join(errors.New("unable to start custom notify command for "+systemConf.Name), err)
		}
		err = runNotifyCmd.Wait()
		if err != nil {
			return false, errors.Join(errors.New("unable to wait for custom notify command for "+systemConf.Name), err)
		}
	}
	return true, nil
}

type notifyLogger struct {
	Logger *slog.Logger
	Name   string
}

func (dl notifyLogger) Write(p []byte) (n int, err error) {
	dl.Logger.Info(dl.Name + ": " + strings.TrimRight(string(p), "\n"))
	return len(p), nil
}

// Returns the notification text for a run.
// This is designed to ignore incorrect inputs; ensuring a notification is sent
// is critical; if it's missing some information, that's acceptable.
func getNotifyTextSlack(
	run data.GetRunRow,
	tagStatuses config.StatusConfig,
	hostname string,
	showEmoji bool,
) string {
	status := core.RunStatus(run.Run.Status)
	return markdownBold + run.Job.Name + hostnameIfExists(hostname) + ":" +
		strconv.FormatInt(run.Run.ID, 10) + markdownBold + " - " +
		core.FormatStatus(core.RunStatus(status), showEmoji) +
		tagChannelIfStatusConfigured(status, tagStatuses) +
		logFileAndOutput(run.Job.NotifyLogContent, run.Run.LogFile)
}

func getNotifyTextNtfy(
	run data.GetRunRow,
	hostname string,
	showEmoji bool,
) string {
	status := core.RunStatus(run.Run.Status)
	return markdownBold + run.Job.Name + hostnameIfExists(hostname) + ":" +
		strconv.FormatInt(run.Run.ID, 10) + markdownBold + " - " +
		core.FormatStatus(core.RunStatus(status), showEmoji) +
		logFileAndOutput(run.Job.NotifyLogContent, run.Run.LogFile)
}

func logFileAndOutput(
	notifyLogContent bool,
	logFile string,
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
	return "\n" + markdownPreformattedText + "\n" + string(logContent) + markdownPreformattedText
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
	if shouldNotifyBePriority(status, tagStatuses) {
		return " " + slackTagChannel
	}
	return ""
}

func shouldNotifyBePriority(
	status core.RunStatus,
	statusConfig config.StatusConfig,
) bool {
	if (status == core.RunStatusRunning && statusConfig.Running) ||
		(status == core.RunStatusSkipped && statusConfig.Skipped) ||
		(status == core.RunStatusSucceeded && statusConfig.Succeeded) ||
		(status == core.RunStatusFailed && statusConfig.Failed) ||
		(status == core.RunStatusTerminated && statusConfig.Terminated) {
		return true
	}
	return false
}

func notifyNfty(
	conf config.NtfyConfig,
	text string,
	status core.RunStatus,
	statusConfig config.StatusConfig,
) (bool, error) {
	if conf.Topic == "" {
		return false, errors.New("notify.ntfy.topic must be set")
	}

	r, _ := http.NewRequest("POST", conf.Domain+"/"+conf.Topic, strings.NewReader(text))
	r.Header.Set("Markdown", "yes")
	if conf.Token != "" {
		r.Header.Set("Authorization", "Bearer "+conf.Token)
	}
	if shouldNotifyBePriority(status, statusConfig) {
		r.Header.Set("Priority", "3")
	} else {
		r.Header.Set("Priority", "1")
	}

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
