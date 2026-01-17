package aha

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/ByteSizedMarius/go-fritzbox-api/v2"
	"github.com/clbanning/mxj"
)

// Temperature is the struct for the temperature capability.
type Temperature struct {
	CapName string
	Celsius string `json:"celsius"`
	Offset  string `json:"offset"`
	device  *Device
}

// TemperatureStats is the struct for the temperature statistics.
// It contains the temperature values, often from the last 24 hours, the amount of values and the amount of seconds between measurements.
type TemperatureStats struct {
	Values                     []float64
	AmountOfValues             int
	SecondsBetweenMeasurements int
}

// String returns a string representation of the temperature statistics
func (ts TemperatureStats) String() string {
	return fmt.Sprintf(
		"Temperature-Stats: {Amount of Values: %d, Amount of Seconds between Measurements: %d, Measurements: %v}",
		ts.AmountOfValues, ts.SecondsBetweenMeasurements, ts.Values,
	)
}

// Name returns the name of the capability
func (t *Temperature) Name() string {
	return t.CapName
}

// String returns a string representation of the capability
func (t *Temperature) String() string {
	return fmt.Sprintf("%s: {Celsius: %f, Offset: %f}", t.CapName, t.GetCelsiusNumeric(), t.GetOffsetNumeric())
}

// Device returns the device the capability belongs to
func (t *Temperature) Device() *Device {
	return t.device
}

func (t *Temperature) fromJSON(m map[string]json.RawMessage, d *Device) (Capability, error) {
	err := json.Unmarshal(m["temperature"], &t)
	if err != nil {
		return t, err
	}

	t.device = d
	return t, nil
}

// Reload fetches the current device and updates the current capability
func (t *Temperature) Reload(c *fritzbox.Client) error {
	return GetDeviceInfos(c, t.Device().Identifier, t)
}

// GetCelsiusNumeric returns the temperature reading in float converted to the usual format (eg. 21.5)
func (t *Temperature) GetCelsiusNumeric() float64 {
	return toTemp(t.Celsius)
}

// GetOffsetNumeric returns the temperature offset set for the device in float converted to the usual format (eg. 21.5)
func (t *Temperature) GetOffsetNumeric() float64 {
	return toTemp(t.Offset)
}

// DECTGetCelsiusNumeric is the same as GetCelsiusNumeric, but it will fetch the current value from the fritzbox and update the local state of the device before returning.
func (t *Temperature) DECTGetCelsiusNumeric(c *fritzbox.Client) (float64, error) {
	resp, err := dectGetter(c, "gettemperature", t)
	if err != nil {
		return 0, err
	}

	t.Celsius = resp
	return t.GetCelsiusNumeric(), nil
}

// DECTGetDeviceStats returns the temperatures measured from the device in the last 24 hours
func (t *Temperature) DECTGetDeviceStats(c *fritzbox.Client) (ts TemperatureStats, err error) {
	data := fritzbox.Values{
		"sid":       c.SID(),
		"ain":       t.Device().Identifier,
		"switchcmd": "getbasicdevicestats",
	}
	_, resp, err := c.AhaRequestString(http.MethodGet, "webservices/homeautoswitch.lua", data)
	if err != nil {
		return ts, err
	}

	mv, err := mxj.NewMapXml([]byte(resp))
	if err != nil {
		return ts, err
	}
	mv, err = mv.NewMap("devicestats.temperature.stats:stats")
	if err != nil {
		return ts, err
	}
	ets := extTemperatureStats{}
	if err = mv.Struct(&ets); err != nil {
		return ts, err
	}

	ts.AmountOfValues, _ = strconv.Atoi(ets.Stats.Count)
	ts.SecondsBetweenMeasurements, _ = strconv.Atoi(ets.Stats.Grid)
	for _, tv := range strings.Split(ets.Stats.Text, ",") {
		ts.Values = append(ts.Values, toTemp(tv))
	}
	return ts, nil
}

func toTemp(o string) float64 {
	s, _ := strconv.ParseFloat(o, 64)
	if s != 0 {
		s /= 10
	}
	return s
}

type extTemperatureStats struct {
	Stats struct {
		Text  string `json:"#text"`
		Count string `json:"-count"`
		Grid  string `json:"-grid"`
	} `json:"stats"`
}
