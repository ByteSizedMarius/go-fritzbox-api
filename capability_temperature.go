package go_fritzbox_api

import (
	"encoding/json"
	"fmt"
	"github.com/clbanning/mxj/v2"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Temperature struct {
	CapName string
	Celsius string `json:"celsius"`
	Offset  string `json:"offset"`
	device  *Device
}

type TemperatureStats struct {
	Values                     []float64
	AmountOfValues             int
	SecondsBetweenMeasurements int
}

// Reload fetches the current device and updates the current capability
func (t *Temperature) Reload(c *Client) error {
	tt, err := getDeviceInfosFromCapability(c, t)
	if err != nil {
		return err
	}

	// update current capability
	th := tt.(Temperature)
	*t = th
	return nil
}

// GetCelsiusNumeric returns the temperature reading in float converted to the usual format (eg. 21.5)
func (t Temperature) GetCelsiusNumeric() float64 {
	return toTemp(t.Celsius)
}

// GetOffset returns the temperature offset set for the device in float converted to the usual format (eg. 21.5)
func (t Temperature) GetOffset() float64 {
	return toTemp(t.Offset)
}

// DECTGetCelsiusNumeric is the same as GetIstNumeric, but it will fetch the current value from the fritzbox and update the local state of the device before returning.
func (t *Temperature) DECTGetCelsiusNumeric(c *Client) (float64, error) {
	resp, err := dectGetter(c, "gettemperature", t)
	if err != nil {
		return 0, err
	}

	// update local device
	t.Celsius = resp
	return t.GetCelsiusNumeric(), nil
}

// DECTGetDeviceStats returns the temperatures measured from the device in the last 24 hours
func (t Temperature) DECTGetDeviceStats(c *Client) (ts TemperatureStats, err error) {
	data := url.Values{
		"sid":       {c.SID()},
		"ain":       {t.Device().Identifier},
		"switchcmd": {"getbasicdevicestats"},
	}

	code, resp, err := c.CustomRequest(http.MethodGet, "webservices/homeautoswitch.lua", data)
	if err != nil {
		return
	}

	if code != 200 {
		err = fmt.Errorf("unknown error: " + resp)
		return
	}

	mv, err := mxj.NewMapXml([]byte(resp))
	if err != nil {
		return
	}

	mv, err = mv.NewMap("devicestats.temperature.stats:stats")
	if err != nil {
		return
	}

	ets := extTemperatureStats{}
	err = mv.Struct(&ets)

	ts.AmountOfValues, _ = strconv.Atoi(ets.Stats.Count)
	ts.SecondsBetweenMeasurements, _ = strconv.Atoi(ets.Stats.Grid)
	for _, tv := range strings.Split(ets.Stats.Text, ",") {
		ts.Values = append(ts.Values, toTemp(tv))
	}
	return
}

func (ts TemperatureStats) String() string {
	return fmt.Sprintf("Temperature-Stats: {Amount of Values: %d, Amount of Seconds between Measurements: %d, Measurements: %v}", ts.AmountOfValues, ts.SecondsBetweenMeasurements, ts.Values)
}

func (t Temperature) Name() string {
	return t.CapName
}

func (t Temperature) String() string {
	return fmt.Sprintf("%s: {Celsius: %f, Offset: %f}", t.CapName, t.GetCelsiusNumeric(), t.GetOffset())
}

func (t Temperature) Device() *Device {
	return t.device
}

func (t Temperature) fromJSON(m map[string]json.RawMessage, d *Device) (Capability, error) {
	err := json.Unmarshal(m["temperature"], &t)
	if err != nil {
		return t, err
	}

	t.device = d
	return t, nil
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
