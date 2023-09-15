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
	device  *SmarthomeDevice
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
	th := tt.(*Temperature)
	t.CapName = th.CapName
	t.Celsius = th.Celsius
	t.Offset = th.Offset
	t.device = th.device
	return nil
}

// GetCelsiusNumeric returns the temperature reading in float converted to the usual format (eg. 21.5)
func (t *Temperature) GetCelsiusNumeric() float64 {
	return toTemp(t.Celsius)
}

// GetOffsetNumeric returns the temperature offset set for the device in float converted to the usual format (eg. 21.5)
func (t *Temperature) GetOffsetNumeric() float64 {
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

// PyaSetOffset sets the temperature offset for the device using the PyAdapter
func (t *Temperature) PyaSetOffset(pya *PyAdapter, offset float64) (err error) {
	// Setting the Offset is done via the HKR, despite the Offset being part of the Temperature capability.
	// Get the HKR via the Device
	if !t.device.HasCapability(CHKR) {
		return fmt.Errorf("device does not have capability %s", CHKR)
	}

	hkr := GetCapability[*Hkr](*t.Device())
	data, err := hkr.pyaPrepare(pya)
	data["Offset"] = ToUrlValue(offset)

	_, err = pya.Client.doRequest(http.MethodPost, "data.lua", data, true)
	if err != nil {
		return
	}

	err = t.Reload(pya.Client)
	if t.GetOffsetNumeric() != offset {
		err = fmt.Errorf("could not set offset. If offset is set multiple time in a short period of time, the fritzbox will block the requests.")
	}

	return
}

// DECTGetDeviceStats returns the temperatures measured from the device in the last 24 hours
func (t *Temperature) DECTGetDeviceStats(c *Client) (ts TemperatureStats, err error) {
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

func (t *Temperature) Name() string {
	return t.CapName
}

func (t *Temperature) String() string {
	return fmt.Sprintf("%s: {Celsius: %f, Offset: %f}", t.CapName, t.GetCelsiusNumeric(), t.GetOffsetNumeric())
}

func (t *Temperature) Device() *SmarthomeDevice {
	return t.device
}

func (t *Temperature) fromJSON(m map[string]json.RawMessage, d *SmarthomeDevice) (Capability, error) {
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
