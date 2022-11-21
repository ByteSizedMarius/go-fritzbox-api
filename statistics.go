package go_fritzbox_api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type TrafficStatistics struct {
	LastMonth TrafficForDuration `json:"LastMonth"`
	ThisWeek  TrafficForDuration `json:"ThisWeek"`
	Today     TrafficForDuration `json:"Today"`
	Yesterday TrafficForDuration `json:"Yesterday"`
	ThisMonth TrafficForDuration `json:"ThisMonth"`
}

func (ts TrafficStatistics) String() string {
	return fmt.Sprintf("LastMonth:\n%s\nThisWeek:\n%s\nToday:\n%s\nYesterday:\n%s\nThisMonth:\n%s", ts.LastMonth, ts.ThisWeek, ts.Today, ts.Yesterday, ts.ThisMonth)
}

type TrafficForDuration struct {
	BytesSentHigh     string `json:"BytesSentHigh"`
	BytesSentLow      string `json:"BytesSentLow"`
	BytesReceivedHigh string `json:"BytesReceivedHigh"`
	BytesReceivedLow  string `json:"BytesReceivedLow"`
}

func (tfd TrafficForDuration) String() string {
	return fmt.Sprintf("\tBytesSentHigh: %s\n\tBytesSentLow: %s\n\tBytesReceivedHigh: %s\n\tBytesReceivedLow: %s", tfd.BytesSentHigh, tfd.BytesSentHigh, tfd.BytesReceivedHigh, tfd.BytesReceivedLow)
}

func (c *Client) GetTrafficStats() (ts TrafficStatistics, err error) {
	data := url.Values{
		"sid": {c.session.Sid},
	}

	resp, err := c.doRequest(http.MethodPost, "internet/inetstat_counter.lua", data)
	if err != nil {
		return
	}

	body, err := getBody(resp)
	if err != nil {
		return
	}

	if !strings.Contains(body, "updateVolume() {") {
		err = fmt.Errorf("unsuccessful")
		return
	}

	statsdata := body[strings.Index(body, "const data = {")+13:strings.Index(body, "};")] + "}"
	err = json.Unmarshal([]byte(statsdata), &ts)
	return
}
