package go_fritzbox_api

import (
	"encoding/json"
	"fmt"
	"github.com/anaskhan96/soup"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Hkr is the struct for HeizungsKörperRegler. Note: current temperature can be accessed via the temperature-capability.
type Hkr struct {
	CapName                 string
	Tsoll                   string   `json:"tsoll"`
	Absenk                  string   `json:"absenk"`
	Komfort                 string   `json:"komfort"`
	Lock                    string   `json:"lock"`                    // Keylock (Tastensperre) configurated via Web-UI/API, activated automatically if summeractive or holdidayactive
	Devicelock              string   `json:"devicelock"`              // Same as lock, configurated manually on the device itself
	Errorcode               string   `json:"errorcode"`               // 0 = no error
	Windowopenactiv         string   `json:"windowopenactiv"`         // 1 if window currently detected as open
	Windowopenactiveendtime string   `json:"windowopenactiveendtime"` // time in seconds until radiator turns back on
	Boostactive             string   `json:"boostactive"`             // same as window
	Boostactiveendtime      string   `json:"boostactiveendtime"`      // same as window
	Batterylow              string   `json:"batterylow"`              // 1 if battery low
	Battery                 string   `json:"battery"`                 // battery %
	Nextchange              struct { // next change in temperature
		Endperiod string `json:"endperiod"`
		Tchange   string `json:"tchange"`
	} `json:"nextchange"`
	Summeractive    string     `json:"summeractive"`  // 1 if summer is currently active
	SummerTimeSet   bool       `json:"-"`             // true if summer-time is set, call FetchSummerTime on Device to fill it. If false, SummerTimeFrame does not return real values.
	SummerTimeFrame SummerTime `json:"-"`             // value is empty by default, call FetchSummerTime on Device to fill it
	Holidayactive   string     `json:"holidayactive"` // same as summer
	device          *SmarthomeDevice
}

type SummerTime struct {
	StartDay   string
	StartMonth string
	EndDay     string
	EndMonth   string
}

// errorcodes taken from docs, 29.09.22
var errorMap = map[string]string{"0": "no error", "1": "Keine Adaptierung möglich. Gerät korrekt am Heizkörper montiert?", "2": "Ventilhub zu kurz oder Batterieleistung zu schwach. Ventilstößel per Hand mehrmals öffnen und schließen oder neue Batterien einsetzen.", "3": "Keine Ventilbewegung möglich. Ventilstößel frei?", "4": "Die Installation wird gerade vorbereitet.", "5": "Der Heizkörperregler ist im Installationsmodus und kann auf das Heizungsventil montiert werden", "6": "Der Heizkörperregler passt sich nun an den Hub des Heizungsventils an"}

func (h *Hkr) Name() string {
	return h.CapName
}

func (h *Hkr) String() string {
	s := fmt.Sprintf("%s: {Temp-Soll: %s, Temp-Absenk: %s, Temp-Komfort: %s, Tastensperre: %s, Tastensperre (Gerät): %s, Error: %s, Fenster offen: %t, ", h.CapName, h.GetSoll(), h.GetAbsenk(), h.GetKomfort(), h.Lock, h.Devicelock, h.GetErrorString(), h.IsWindowOpen())
	if h.IsWindowOpen() {
		s += fmt.Sprintf("Fenster offen Ende: %s", h.GetWindowOpenEndtime())
	}

	s += fmt.Sprintf("Boost aktiv: %t, ", h.IsBoostActive())
	if h.IsBoostActive() {
		s += fmt.Sprintf("Boost Ende: %s, ", h.GetBoostEndtime())
	}

	s += fmt.Sprintf("Batterie niedrig: %t, Batteriestand: %s, nächste Temperaturanpassung: auf %s zu Zeitpunkt %s, Sommer aktiv: %t, Urlaub aktiv: %t, Gerät: %s}", h.IsBatteryLow(), h.Battery, h.GetNextChangeTemperature(), h.GetNextchangeEndtime(), h.IsSummerActive(), h.IsHolidayActive(), h.device)
	return s
}

func (h *Hkr) Device() *SmarthomeDevice {
	return h.device
}

// -------------------------------------------
//
// SOLL
//
// -------------------------------------------

// GetSollNumeric returns the current locally saved soll-temperature. It returns temperatures in Celsius from 8-28, as well as -1 (MAX) -2 (OFF).
func (h *Hkr) GetSollNumeric() float64 {
	return temperatureHelper(h.Tsoll)
}

// GetSoll returns the same values as GetSollNumeric, but as a string. Instead of -1 and -2, it returns "OFF" or "MAX"
func (h *Hkr) GetSoll() (r string) {
	s, _ := strconv.ParseFloat(h.Tsoll, 64)
	if s == 253 {
		r = "OFF"
	} else if s == 254 {
		r = "MAX"
	} else if s == 0 {
		r = "0"
	} else {
		r = fmt.Sprintf("%.1f", s/2)
	}

	return r
}

// DECTGetSoll sends an API-Request to get the current soll-temperature from the fritzbox/device itself.
// It will then update the current device locally and return the same output as GetSoll.
func (h *Hkr) DECTGetSoll(c *Client) (r string, err error) {
	resp, err := dectGetter(c, "gethkrtsoll", h)
	if err != nil {
		return "", err
	}

	// update local device
	h.Tsoll = resp
	return h.GetSoll(), nil
}

// DECTGetSollNumeric does the same as DECTGetSoll but returns the result like GetSollNumeric
func (h *Hkr) DECTGetSollNumeric(c *Client) (float64, error) {
	_, err := h.DECTGetSoll(c)
	if err != nil {
		return -1, err
	}

	return h.GetSollNumeric(), nil
}

// DECTSetSollMax turns the soll-temperature on. Allegedly, it should use the last known temperature. However, for me, it just sets the radiator to MAX.
func (h *Hkr) DECTSetSollMax(c *Client) error {
	return h.setSollHelper(c, 254)
}

// DECTSetSollOff turns the soll-temperature off. The Hkr will show the snowflake in its display.
func (h *Hkr) DECTSetSollOff(c *Client) error {
	return h.setSollHelper(c, 253)
}

// DECTSetSoll sets the soll temperature to the given temperature (meaning 21.5 = 21.5 C).
// This method accepts float64/32, int and string (XX,X/ XX.X).
// Values with additional decimal places will be rounded to XX.0/XX.5 respectively.
// Only values from 8-28 are valid.
func (h *Hkr) DECTSetSoll(c *Client, sollTemp interface{}) error {
	var i float64 = 0
	var err error

	switch sollTemp.(type) {
	case int:
		i = float64(sollTemp.(int))
	case string:
		if sollTemp == "OFF" {
			return h.DECTSetSollOff(c)
		} else if sollTemp == "MAX" {
			return h.DECTSetSollMax(c)
		}

		if strings.Contains(sollTemp.(string), ",") {
			sollTemp = strings.Replace(sollTemp.(string), ",", ".", 1)
		}

		i, err = strconv.ParseFloat(sollTemp.(string), 64)
		if err != nil {
			return err
		}
	case float32:
		i = float64(sollTemp.(float32))
	}

	if i < 8 || i > 28 {
		return fmt.Errorf("invalid temperature for soll-temp")
	}

	return h.setSollHelper(c, i)
}

// -------------------------------------------
//
// KOMFORT
//
// -------------------------------------------

// GetKomfort is similar to GetSoll
func (h *Hkr) GetKomfort() (r string) {
	return temperatureStringHelper(h.Komfort)
}

// GetKomfortNumeric is similar to GetSollNumeric
func (h *Hkr) GetKomfortNumeric() float64 {
	return temperatureHelper(h.Komfort)
}

// DECTGetKomfort is similar to DECTGetSoll
func (h *Hkr) DECTGetKomfort(c *Client) (string, error) {
	resp, err := dectGetter(c, "gethkrkomfort", h)
	if err != nil {
		return "", err
	}

	// update local device
	h.Komfort = resp
	return h.GetKomfort(), nil
}

// DECTGetKomfortNumeric is similar to DECTGetSollNumeric
func (h *Hkr) DECTGetKomfortNumeric(c *Client) (float64, error) {
	_, err := h.DECTGetKomfort(c)
	if err != nil {
		return -1, err
	}

	return h.GetKomfortNumeric(), nil
}

// -------------------------------------------
//
// ABSENK
//
// -------------------------------------------

// GetAbsenk is similar to GetSoll
func (h *Hkr) GetAbsenk() (r string) {
	return temperatureStringHelper(h.Absenk)
}

// GetAbsenkNumeric is similar to GetSollNumeric
func (h *Hkr) GetAbsenkNumeric() float64 {
	return temperatureHelper(h.Absenk)
}

// DECTGetAbsenk is similar to DECTGetSoll
func (h *Hkr) DECTGetAbsenk(c *Client) (string, error) {
	resp, err := dectGetter(c, "gethkrabsenk", h)
	if err != nil {
		return "", err
	}

	// update local device
	h.Absenk = resp
	return h.GetKomfort(), nil
}

// DECTGetAbsenkNumeric is similar to DECTGetSollNumeric
func (h *Hkr) DECTGetAbsenkNumeric(c *Client) (float64, error) {
	_, err := h.DECTGetKomfort(c)
	if err != nil {
		return -1, err
	}

	return h.GetAbsenkNumeric(), nil
}

// -------------------------------------------
//
// BOOST
//
// -------------------------------------------

// IsBoostActive returns true, if the boost is currently active
// No API-requests are made. Instead, all this does is convert the local value.
// Only functions with the DECT-Prefix communicate with the fritzbox.
func (h *Hkr) IsBoostActive() bool { return h.Boostactive == "1" }

// GetBoostEndtime converts the endtime to a time-struct
func (h *Hkr) GetBoostEndtime() (t time.Time) {
	return unixStringToTime(h.Boostactiveendtime)
}

// DECTDeactivateBoost turns boost off if currently enabled
func (h *Hkr) DECTDeactivateBoost(c *Client) (err error) {
	return h.deactivateHelper(c, "sethkrboost")
}

// SetBoost activates the radiators boost-function for the specified duration (max. 24hrs). Returns the end-time of the boost.
func (h *Hkr) SetBoost(c *Client, d time.Duration) (tm time.Time, err error) {
	return h.setEndpointHelper(c, d, "sethkrboost")
}

// -------------------------------------------
//
// WINDOW
//
// -------------------------------------------

func (h *Hkr) IsWindowOpen() bool { return h.Windowopenactiv == "1" }

func (h *Hkr) GetWindowOpenEndtime() (t time.Time) {
	return unixStringToTime(h.Windowopenactiveendtime)
}

func (h *Hkr) DECTDeactivateWindowOpen(c *Client) (err error) {
	return h.deactivateHelper(c, "sethkrwindowopen")
}

func (h *Hkr) DECTSetWindowOpen(c *Client, d time.Duration) (tm time.Time, err error) {
	return h.setEndpointHelper(c, d, "sethkrwindowopen")
}

// -------------------------------------------
//
// ERROR
//
// -------------------------------------------

// GetErrorString returns the error-message for the respective error-code. Errorcode 0 means no error. The error-messages originate from the official docs.
func (h *Hkr) GetErrorString() string {
	s, ok := errorMap[h.Errorcode]
	if !ok {
		return "unknown error: " + h.Errorcode
	}
	return s
}

// -------------------------------------------
//
// NextChange
//
// -------------------------------------------

// GetNextChangeTemperatureNumeric returns the temperature, that the next change will set it to (numeric)
func (h *Hkr) GetNextChangeTemperatureNumeric() float64 {
	return temperatureHelper(h.Nextchange.Tchange)
}

// GetNextChangeTemperature returns the temperature, that the next change will set it to (as string)
func (h *Hkr) GetNextChangeTemperature() string {
	return temperatureStringHelper(h.Nextchange.Tchange)
}

// GetNextchangeEndtime returns the time of the next temperature-change
func (h *Hkr) GetNextchangeEndtime() (t time.Time) {
	return unixStringToTime(h.Nextchange.Endperiod)
}

// -------------------------------------------
//
// Unstable
//
// -------------------------------------------

func (h *Hkr) FetchSummerTime(c *Client) (err error) {
	if err = c.checkExpiry(); err != nil {
		return
	}

	data := url.Values{
		"sid":    {c.session.Sid},
		"page":   {"home_auto_hkr_edit"},
		"device": {h.device.ID},
	}

	resp, err := c.doRequest(http.MethodPost, "data.lua", data)
	if err != nil {
		return
	}

	body, err := getBody(resp)
	if err != nil {
		return
	}

	doc := soup.HTMLParse(body)
	row := doc.Find("tr", "id", "uiSummerEnabledRow")
	if row.Error != nil || row.Attrs()["style"] == "display:none;" {
		h.SummerTimeSet = false
		return
	}
	h.SummerTimeSet = true

	ssd := row.Find("input", "id", "uiSummerStartDay")
	ssm := row.Find("input", "id", "uiSummerStartMonth")

	sed := row.Find("input", "id", "uiSummerEndDay")
	sem := row.Find("input", "id", "uiSummerEndMonth")

	if ssd.Error != nil || ssm.Error != nil || sed.Error != nil || sem.Error != nil {
		return fmt.Errorf("%s", "Some required Inputs not found")
	}

	h.SummerTimeFrame = SummerTime{
		StartDay:   ssd.Attrs()["value"],
		StartMonth: ssm.Attrs()["value"],
		EndDay:     sed.Attrs()["value"],
		EndMonth:   sem.Attrs()["value"],
	}

	return nil
}

func dateHelper(month string, day string) time.Time {
	monthNr, _ := strconv.Atoi(month)
	dayNr, _ := strconv.Atoi(day)

	now := time.Now()
	return time.Date(now.Year(), time.Month(monthNr), dayNr, 0, 0, 0, 0, time.UTC)
}

func (stf SummerTime) GetStartDate() time.Time {
	return dateHelper(stf.StartMonth, stf.StartDay)
}

func (stf SummerTime) GetEndDate() time.Time {
	return dateHelper(stf.EndMonth, stf.EndDay)
}

// -------------------------------------------
//
// HELPERS
//
// -------------------------------------------

// IsBatteryLow returns true if the device reports battery low
func (h *Hkr) IsBatteryLow() bool {
	return h.Batterylow == "1"
}

// IsSummerActive returns true, if summer-mode is currently active
func (h *Hkr) IsSummerActive() bool {
	return h.Summeractive == "1"
}

// IsHolidayActive returns true, if holiday-mode is currently active
func (h *Hkr) IsHolidayActive() bool {
	return h.Holidayactive == "1"
}

// Reload reloads all client values
func (h *Hkr) Reload(c *Client) error {
	tt, err := getDeviceInfosFromCapability(c, h)
	if err != nil {
		return err
	}

	// update current capability
	th := tt.(*Hkr)
	h.CapName = th.CapName
	h.Tsoll = th.Tsoll
	h.Absenk = th.Absenk
	h.Komfort = th.Komfort
	h.Lock = th.Lock
	h.Devicelock = th.Devicelock
	h.Errorcode = th.Errorcode
	h.Windowopenactiv = th.Windowopenactiv
	h.Windowopenactiveendtime = th.Windowopenactiveendtime
	h.Boostactive = th.Boostactive
	h.Boostactiveendtime = th.Boostactiveendtime
	h.Batterylow = th.Batterylow
	h.Battery = th.Battery
	h.Nextchange = th.Nextchange
	h.Summeractive = th.Summeractive
	h.Holidayactive = th.Holidayactive
	h.device = th.device
	return nil
}

func (h *Hkr) fromJSON(m map[string]json.RawMessage, d *SmarthomeDevice) (Capability, error) {
	err := json.Unmarshal(m["hkr"], &h)
	if err != nil {
		return h, err
	}

	h.device = d
	return h, nil
}

func (h *Hkr) setSollHelper(c *Client, sollTemp float64) error {
	if sollTemp != 253 && sollTemp != 254 {
		sollTemp *= 2
	}

	data := url.Values{
		"sid":       {c.SID()},
		"ain":       {h.Device().Identifier},
		"switchcmd": {"sethkrtsoll"},
		"param":     {fmt.Sprintf("%.1f", math.Round(sollTemp/0.5)*0.5)},
	}

	_, resp, err := c.CustomRequest(http.MethodGet, "webservices/homeautoswitch.lua", data)
	if err != nil {
		return err
	}

	h.Tsoll = resp
	return nil
}

func temperatureHelper(r string) (s float64) {
	s, _ = strconv.ParseFloat(r, 64)
	if s == 0 {
		return 0
	} else if s == 254 {
		return -1
	} else if s == 253 {
		return -2
	} else {
		return s / 2
	}
}

func temperatureStringHelper(r string) string {
	s, _ := strconv.ParseFloat(r, 64)
	if s == 253 {
		r = "OFF"
	} else if s == 254 {
		r = "ON"
	} else if s == 0 {
		r = "0"
	} else {
		r = fmt.Sprintf("%.1f", s/2)
	}

	return r
}

func unixStringToTime(unixString string) (t time.Time) {
	i, err := strconv.ParseInt(unixString, 10, 64)
	if err != nil {
		fmt.Println(err)
		return
	}
	t = time.Unix(i, 0)
	return
}

// DECTDeactivateBoost turns boost off if currently enabled
func (h *Hkr) deactivateHelper(c *Client, endpoint string) (err error) {
	_, err = h.setEndpointHelper(c, 0, endpoint)
	if err != nil {
		return
	}

	return h.Reload(c)
}

func (h *Hkr) setEndpointHelper(c *Client, d time.Duration, endpoint string) (tm time.Time, err error) {
	ts := "0"
	if d != 0 {
		if d > time.Hour*24 {
			err = fmt.Errorf("duration cannot be longer than 24 hours")
			return
		}

		t := time.Now().Add(d).Unix()
		ts = strconv.FormatInt(t, 10)
	}

	data := url.Values{
		"sid":          {c.SID()},
		"ain":          {h.Device().Identifier},
		"switchcmd":    {endpoint},
		"endtimestamp": {ts},
	}

	_, resp, err := c.CustomRequest(http.MethodGet, "webservices/homeautoswitch.lua", data)
	if err != nil {
		return
	}

	i, err := strconv.ParseInt(resp, 10, 64)
	if err != nil {
		return
	}
	tm = time.Unix(i, 0)
	err = h.Reload(c)
	return tm, err
}
