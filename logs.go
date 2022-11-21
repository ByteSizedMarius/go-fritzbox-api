package go_fritzbox_api

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type rawLog [][]string

type LogMessage struct {
	Message string
	Date    time.Time
}

func (c *Client) GetEventLogUntil(l LogMessage) (logEvents []LogMessage, err error) {
	tmp, err := c.GetEventLog()
	if err != nil {
		return
	}

	for i, ev := range tmp {
		if ev.MD5() == l.MD5() {
			logEvents = tmp[:i]
			break
		}
	}

	return
}

func (c *Client) GetEventLog() (logEvents []LogMessage, err error) {
	if err = c.checkExpiry(); err != nil {
		return
	}

	data := url.Values{
		"sid":  {c.session.Sid},
		"page": {"log"},
	}

	resp, err := c.doRequest(http.MethodPost, "data.lua", data)
	if err != nil {
		return
	}

	body, err := getBody(resp)
	if err != nil {
		return
	}

	body = getFromOffset(body, "\"log\":[", 6)
	body = getUntil(body, ",\"filter")

	logs := rawLog{}
	err = json.Unmarshal([]byte(body), &logs)

	logEvents = make([]LogMessage, 0)

	for _, ev := range logs {
		var date time.Time
		date, err = time.Parse("02.01.06 15:04:05", ev[0]+" "+ev[1])
		if err != nil {
			fmt.Println(fmt.Errorf("error on Date parse: " + ev[0] + " " + ev[1]))
		}

		logEvents = append(logEvents, LogMessage{
			Message: ev[2],
			Date:    date,
		})
	}

	return
}

func (l LogMessage) MD5() string {
	return fmt.Sprintf("%x", md5.Sum([]byte(l.Message+l.Date.String())))
}
