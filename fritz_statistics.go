package go_fritzbox_api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type TrafficStatistics struct {
	LastMonth TrafficForDuration `json:"LastMonth"`
	ThisWeek  TrafficForDuration `json:"ThisWeek"`
	Today     TrafficForDuration `json:"Today"`
	Yesterday TrafficForDuration `json:"Yesterday"`
	ThisMonth TrafficForDuration `json:"ThisMonth"`
}

func (ts *TrafficStatistics) String() string {
	return fmt.Sprintf("LastMonth:\n%s\nThisWeek:\n%s\nToday:\n%s\nYesterday:\n%s\nThisMonth:\n%s", &ts.LastMonth, &ts.ThisWeek, &ts.Today, &ts.Yesterday, &ts.ThisMonth)
}

// The TrafficForDuration struct contains network statistics. Only MBSent and MBReceived are real values. The rest are raw values returned by the backend.
type TrafficForDuration struct {
	BytesSentHigh     string `json:"BytesSentHigh"`
	BytesSentLow      string `json:"BytesSentLow"`
	BytesReceivedHigh string `json:"BytesReceivedHigh"`
	BytesReceivedLow  string `json:"BytesReceivedLow"`
	MBSent            int
	MBReceived        int
}

func (tfd *TrafficForDuration) String() string {
	return fmt.Sprintf("\tMB Sent: %d\n\tMB Received: %d", tfd.MBSent, tfd.MBReceived)
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
	ts.calc()

	return
}

func (ts *TrafficStatistics) calc() {
	ts.LastMonth.calc()
	ts.ThisWeek.calc()
	ts.Today.calc()
	ts.Yesterday.calc()
	ts.ThisMonth.calc()
}

func (tfd *TrafficForDuration) calc() {
	tfd.MBSent = int(combineBytes(tfd.BytesSentHigh, tfd.BytesSentLow) / 1000000)
	tfd.MBReceived = int(combineBytes(tfd.BytesReceivedHigh, tfd.BytesReceivedLow) / 1000000)
}

func combineBytes(high, low string) float64 {
	hI, err := strconv.Atoi(high)
	if err != nil {
		hI = 0
	}

	lI, err := strconv.Atoi(low)
	if err != nil {
		lI = 0
	}

	return float64(hI)*float64(4294967296) + float64(lI)
}
