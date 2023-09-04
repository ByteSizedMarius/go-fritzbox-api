package go_fritzbox_api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type rawLog [][]string

type LogMessage struct {
	DateTime time.Time
	Date     string `json:"date"`
	Group    string `json:"group"`
	ID       int    `json:"id"`
	Message  string `json:"msg"`
	Time     string `json:"time"`
}

func (c *Client) GetEventLog() (logEvents []LogMessage, err error) {
	if err = c.checkExpiry(); err != nil {
		return
	}

	data := url.Values{
		"sid":  {c.session.Sid},
		"page": {"log"},
	}

	resp, err := c.doRequest(http.MethodPost, "data.lua", data, true)
	if err != nil {
		return
	}

	body, err := getBody(resp)
	if err != nil {
		return
	}

	v, err := valueFromJson(body, []string{"data"})
	if err != nil {
		return
	}

	jb, err := json.Marshal(v["log"])
	if err != nil {
		return
	}

	err = json.Unmarshal(jb, &logEvents)
	for i := range logEvents {
		logEvents[i].DateTime, err = time.Parse("02.01.06 15:04:05", logEvents[i].Date+" "+logEvents[i].Time)
		if err != nil {
			fmt.Println(fmt.Errorf("error on date parse: " + logEvents[i].Date + " " + logEvents[i].Time))
		}
	}

	return
}

// GetEventLogUntil returns all log messages newer than the given id.
func (c *Client) GetEventLogUntil(id int) (logEvents []LogMessage, err error) {
	tmp, err := c.GetEventLog()
	if err != nil {
		return
	}

	for i, ev := range tmp {
		if ev.ID == id {
			logEvents = tmp[:i]
			break
		}
	}

	return
}
