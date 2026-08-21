package test

import (
	"bufio"
	"encoding/json"
	"net/http"
)

type NtfyMsg struct {
	ID          string `json:"id"`
	Time        int    `json:"time"`
	Expires     int    `json:"expires"`
	Event       string `json:"event"`
	Topic       string `json:"topic"`
	Message     string `json:"message"`
	Priority    int    `json:"priority"`
	ContentType string `json:"content_type"`
}

func NtfyPollMsgs(
	topic string,
	token string,
	url string,
) []NtfyMsg {

	r, _ := http.NewRequest("GET", url+"/"+topic+"/json?poll=1", nil)
	r.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	res, err := client.Do(r)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode > 299 {
		panic(err)
	}
	var msgs []NtfyMsg

	scanner := bufio.NewScanner(res.Body)

	for scanner.Scan() {
		var msg NtfyMsg
		err = json.Unmarshal(scanner.Bytes(), &msg)
		if err != nil {
			panic(err)
		}
		msgs = append(msgs, msg)
	}
	return msgs
}
