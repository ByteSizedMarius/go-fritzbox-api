package unsafe

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ByteSizedMarius/go-fritzbox-api/v2"
)

const bytesPerMB = 1_000_000

// TrafficStatistics is a struct that holds the traffic statistics of the fritzbox.
type TrafficStatistics struct {
	LastMonth TrafficForDuration `json:"LastMonth"`
	ThisWeek  TrafficForDuration `json:"ThisWeek"`
	Today     TrafficForDuration `json:"Today"`
	Yesterday TrafficForDuration `json:"Yesterday"`
	ThisMonth TrafficForDuration `json:"ThisMonth"`
}

// String returns a string representation of the TrafficStatistics struct.
func (ts *TrafficStatistics) String() string {
	return fmt.Sprintf(
		"LastMonth:\n%s\nThisWeek:\n%s\nToday:\n%s\nYesterday:\n%s\nThisMonth:\n%s",
		&ts.LastMonth, &ts.ThisWeek, &ts.Today, &ts.Yesterday, &ts.ThisMonth,
	)
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

// String returns a string representation of the TrafficForDuration struct.
func (tfd *TrafficForDuration) String() string {
	return fmt.Sprintf("\tMB Sent: %d\n\tMB Received: %d", tfd.MBSent, tfd.MBReceived)
}

// GetTrafficStats returns the traffic statistics from the Fritz!Box.
func GetTrafficStats(c *fritzbox.Client) (ts TrafficStatistics, err error) {
	periods := []string{"Today", "Yesterday", "ThisWeek", "ThisMonth", "LastMonth"}
	fields := []string{"BytesSentHigh", "BytesSentLow", "BytesReceivedHigh", "BytesReceivedLow"}

	data := fritzbox.Values{"sid": c.SID()}
	for _, period := range periods {
		for _, field := range fields {
			key := period + field
			data[key] = "inetstat:status/" + period + "/" + field
		}
	}

	_, resp, err := c.AhaRequestString(http.MethodGet, "query.lua", data)
	if err != nil {
		return
	}

	var raw map[string]string
	if err = json.Unmarshal([]byte(resp), &raw); err != nil {
		return
	}

	ts.Today = extractTrafficForDuration(raw, "Today")
	ts.Yesterday = extractTrafficForDuration(raw, "Yesterday")
	ts.ThisWeek = extractTrafficForDuration(raw, "ThisWeek")
	ts.ThisMonth = extractTrafficForDuration(raw, "ThisMonth")
	ts.LastMonth = extractTrafficForDuration(raw, "LastMonth")
	ts.calc()

	return
}

func extractTrafficForDuration(raw map[string]string, period string) TrafficForDuration {
	return TrafficForDuration{
		BytesSentHigh:     raw[period+"BytesSentHigh"],
		BytesSentLow:      raw[period+"BytesSentLow"],
		BytesReceivedHigh: raw[period+"BytesReceivedHigh"],
		BytesReceivedLow:  raw[period+"BytesReceivedLow"],
	}
}

func (ts *TrafficStatistics) calc() {
	ts.LastMonth.calc()
	ts.ThisWeek.calc()
	ts.Today.calc()
	ts.Yesterday.calc()
	ts.ThisMonth.calc()
}

func (tfd *TrafficForDuration) calc() {
	tfd.MBSent = int(combineBytes(tfd.BytesSentHigh, tfd.BytesSentLow) / bytesPerMB)
	tfd.MBReceived = int(combineBytes(tfd.BytesReceivedHigh, tfd.BytesReceivedLow) / bytesPerMB)
}

// combineBytes converts high/low 32-bit parts into a single 64-bit value.
// If parsing fails, the corresponding part is treated as 0.
func combineBytes(high, low string) uint64 {
	hI, _ := strconv.ParseUint(high, 10, 64)
	lI, _ := strconv.ParseUint(low, 10, 64)
	return (hI << 32) + lI
}
