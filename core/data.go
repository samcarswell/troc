package core

type RunStatus string

const (
	RunStatusRunning    RunStatus = "Running"
	RunStatusSkipped    RunStatus = "Skipped"
	RunStatusSucceeded  RunStatus = "Succeeded"
	RunStatusFailed     RunStatus = "Failed"
	RunStatusTerminated RunStatus = "Terminated"
)

type RunShow struct {
	ID            int64  `json:"id"`
	JobName       string `json:"job_name"`
	StartTime     string `json:"start_time"`
	EndTime       string `json:"end_time"`
	LogFile       string `json:"log_file"`
	SystemLogFile string `json:"system_log_file"`
	Status        string `json:"status"`
	Duration      string `json:"duration"`
	Pid           string `json:"pid"`
	IsArchived    bool   `json:"is_archived"`
}

type JobShow struct {
	ID               int64  `json:"id"`
	Name             string `json:"name"`
	NotifyLogContent bool   `json:"notify_log_content"`
}
