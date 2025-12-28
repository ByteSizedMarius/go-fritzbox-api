package unsafe

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ByteSizedMarius/go-fritzbox-api"
)

// LogDateFormat is the date/time format used by the Fritz!Box event log.
// Default format (DD.MM.YY HH:MM:SS) should work for all locales, but this
// is configurable in case it doesn't.
var LogDateFormat = "02.01.06 15:04:05"

// LogMessage is the internal representation of a log message in the Fritz!Box API.
type LogMessage struct {
	DateTime time.Time
	Date     string `json:"date"`
	Group    string `json:"group"`
	ID       int    `json:"id"`
	Message  string `json:"msg"`
	Time     string `json:"time"`
}

// GetEventLog returns all log events from the Fritz!Box.
func GetEventLog(c *fritzbox.Client) (logEvents []LogMessage, err error) {
	data := fritzbox.Values{
		"sid":  c.SID(),
		"page": "log",
	}

	_, resp, err := c.AhaRequestString(http.MethodPost, "data.lua", data)
	if err != nil {
		return
	}

	v, err := fritzbox.ValueFromJsonPath(resp, []string{"data"})
	if err != nil {
		return
	}
	jb, err := json.Marshal(v["log"])
	if err != nil {
		return
	}
	err = json.Unmarshal(jb, &logEvents)
	if err != nil {
		return
	}
	for i := range logEvents {
		logEvents[i].DateTime, err = time.Parse(
			LogDateFormat,
			fmt.Sprintf("%s %s", logEvents[i].Date, logEvents[i].Time),
		)
		if err != nil {
			return
		}
	}
	return
}

// GetEventLogUntil returns all log messages newer than the given id.
// Returns an error if the id is not found in the log.
func GetEventLogUntil(c *fritzbox.Client, id int) (logEvents []LogMessage, err error) {
	tmp, err := GetEventLog(c)
	if err != nil {
		return
	}

	for i, ev := range tmp {
		if ev.ID == id {
			return tmp[:i], nil
		}
	}

	return nil, errors.New("log event id not found")
}
