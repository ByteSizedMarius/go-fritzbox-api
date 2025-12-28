package aha

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ByteSizedMarius/go-fritzbox-api"
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
	Windowopenactiv         string   `json:"windowopenactiv"`         // 1 if window currently detected as open. The typo is part of the official API.
	Windowopenactiveendtime string   `json:"windowopenactiveendtime"` // time in seconds until radiator turns back on
	Boostactive             string   `json:"boostactive"`             // same as window
	Boostactiveendtime      string   `json:"boostactiveendtime"`      // same as window
	Batterylow              string   `json:"batterylow"`              // 1 if battery low
	Battery                 string   `json:"battery"`                 // battery %
	Nextchange              struct { // next change in temperature
		Endperiod string `json:"endperiod"`
		Tchange   string `json:"tchange"`
	} `json:"nextchange"`
	Summeractive  string `json:"summeractive"`  // 1 if summer is currently active
	Holidayactive string `json:"holidayactive"` // same as summer
	device        *Device
}

// errorcodes taken from docs, 29.09.22
var errorMap = map[string]string{"0": "no error", "1": "Keine Adaptierung möglich. Gerät korrekt am Heizkörper montiert?", "2": "Ventilhub zu kurz oder Batterieleistung zu schwach. Ventilstößel per Hand mehrmals öffnen und schließen oder neue Batterien einsetzen.", "3": "Keine Ventilbewegung möglich. Ventilstößel frei?", "4": "Die Installation wird gerade vorbereitet.", "5": "Der Heizkörperregler ist im Installationsmodus und kann auf das Heizungsventil montiert werden", "6": "Der Heizkörperregler passt sich nun an den Hub des Heizungsventils an"}

func (h *Hkr) Name() string {
	return h.CapName
}

func (h *Hkr) String() string {
	var builder strings.Builder
	_, _ = fmt.Fprintf(
		&builder,
		"%s: {Temp-Soll: %s, Temp-Absenk: %s, Temp-Komfort: %s, Tastensperre: %s, Tastensperre (Gerät): %s, Error: %s, Fenster offen: %t, ",
		h.CapName, h.GetSoll(), h.GetAbsenk(), h.GetKomfort(), h.Lock, h.Devicelock, h.GetErrorString(), h.IsWindowOpen(),
	)
	if h.IsWindowOpen() {
		_, _ = fmt.Fprintf(&builder, "Fenster offen Ende: %s", h.GetWindowOpenEndtime())
	}
	_, _ = fmt.Fprintf(&builder, "Boost aktiv: %t, ", h.IsBoostActive())
	if h.IsBoostActive() {
		_, _ = fmt.Fprintf(&builder, "Boost Ende: %s, ", h.GetBoostEndtime())
	}
	_, _ = fmt.Fprintf(
		&builder,
		"Batterie niedrig: %t, Batteriestand: %s, nächste Temperaturanpassung: auf %s zu Zeitpunkt %s, Sommer aktiv: %t, Urlaub aktiv: %t, Gerät: %s}",
		h.IsBatteryLow(), h.Battery, h.GetNextChangeTemperature(), h.GetNextchangeEndtime(), h.IsSummerActive(), h.IsHolidayActive(), h.device,
	)
	return builder.String()
}

func (h *Hkr) Device() *Device {
	return h.device
}

// GetSollNumeric returns the current locally saved soll-temperature. It returns temperatures in Celsius from 8-28, as well as -1 (MAX) -2 (OFF).
func (h *Hkr) GetSollNumeric() float64 {
	return temperatureHelper(h.Tsoll)
}

// GetSoll returns the same values as GetSollNumeric, but as a string. Instead of -1 and -2, it returns "OFF" or "MAX"
func (h *Hkr) GetSoll() (r string) {
	s, _ := strconv.ParseFloat(h.Tsoll, 64)
	if s == TempOff {
		r = "OFF"
	} else if s == TempMax {
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
func (h *Hkr) DECTGetSoll(c *fritzbox.Client) (r string, err error) {
	resp, err := dectGetter(c, "gethkrtsoll", h)
	if err != nil {
		return "", err
	}

	h.Tsoll = resp
	return h.GetSoll(), nil
}

// DECTGetSollNumeric does the same as DECTGetSoll but returns the result like GetSollNumeric
func (h *Hkr) DECTGetSollNumeric(c *fritzbox.Client) (float64, error) {
	_, err := h.DECTGetSoll(c)
	if err != nil {
		return -1, err
	}

	return h.GetSollNumeric(), nil
}

// DECTSetSollMax turns the soll-temperature on. Allegedly, it should use the last known temperature. However, for me, it just sets the radiator to MAX.
func (h *Hkr) DECTSetSollMax(c *fritzbox.Client) error {
	return h.setSollHelper(c, TempMax)
}

// DECTSetSollOff turns the soll-temperature off. The Hkr will show the snowflake in its display.
func (h *Hkr) DECTSetSollOff(c *fritzbox.Client) error {
	return h.setSollHelper(c, TempOff)
}

// DECTSetSoll sets the soll temperature to the given temperature (meaning 21.5 = 21.5 C).
// This method accepts float64/32, int and string (XX,X/ XX.X).
// Values with additional decimal places will be rounded to XX.0/XX.5 respectively.
// Only values from 8-28 are valid.
func (h *Hkr) DECTSetSoll(c *fritzbox.Client, sollTemp any) error {
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
	case float64:
		i = sollTemp.(float64)
	default:
		return fmt.Errorf("invalid temperature for soll-temp")
	}

	if i < 8 || i > 28 {
		return fmt.Errorf("invalid temperature for soll-temp")
	}

	return h.setSollHelper(c, i)
}

// GetKomfort is similar to GetSoll
func (h *Hkr) GetKomfort() (r string) {
	return temperatureStringHelper(h.Komfort)
}

// GetKomfortNumeric is similar to GetSollNumeric
func (h *Hkr) GetKomfortNumeric() float64 {
	return temperatureHelper(h.Komfort)
}

// DECTGetKomfort is similar to DECTGetSoll
func (h *Hkr) DECTGetKomfort(c *fritzbox.Client) (string, error) {
	resp, err := dectGetter(c, "gethkrkomfort", h)
	if err != nil {
		return "", err
	}

	h.Komfort = resp
	return h.GetKomfort(), nil
}

// DECTGetKomfortNumeric is similar to DECTGetSollNumeric
func (h *Hkr) DECTGetKomfortNumeric(c *fritzbox.Client) (float64, error) {
	_, err := h.DECTGetKomfort(c)
	if err != nil {
		return -1, err
	}

	return h.GetKomfortNumeric(), nil
}

// GetAbsenk is similar to GetSoll
func (h *Hkr) GetAbsenk() (r string) {
	return temperatureStringHelper(h.Absenk)
}

// GetAbsenkNumeric is similar to GetSollNumeric
func (h *Hkr) GetAbsenkNumeric() float64 {
	return temperatureHelper(h.Absenk)
}

// DECTGetAbsenk is similar to DECTGetSoll
func (h *Hkr) DECTGetAbsenk(c *fritzbox.Client) (string, error) {
	resp, err := dectGetter(c, "gethkrabsenk", h)
	if err != nil {
		return "", err
	}

	h.Absenk = resp
	return h.GetAbsenk(), nil
}

// DECTGetAbsenkNumeric is similar to DECTGetSollNumeric
func (h *Hkr) DECTGetAbsenkNumeric(c *fritzbox.Client) (float64, error) {
	_, err := h.DECTGetAbsenk(c)
	if err != nil {
		return -1, err
	}

	return h.GetAbsenkNumeric(), nil
}

// IsBoostActive returns true, if the boost is currently active
func (h *Hkr) IsBoostActive() bool { return h.Boostactive == "1" }

// GetBoostEndtime converts the endtime to a time-struct
func (h *Hkr) GetBoostEndtime() time.Time {
	return unixStringToTime(h.Boostactiveendtime)
}

// DECTDeactivateBoost turns boost off if currently enabled
func (h *Hkr) DECTDeactivateBoost(c *fritzbox.Client) error {
	return h.deactivateHelper(c, "sethkrboost")
}

// SetBoost activates the radiators boost-function for the specified duration (max. 24hrs). Returns the end-time of the boost.
func (h *Hkr) SetBoost(c *fritzbox.Client, d time.Duration) (tm time.Time, err error) {
	return h.setEndpointHelper(c, d, "sethkrboost")
}

func (h *Hkr) IsWindowOpen() bool { return h.Windowopenactiv == "1" }

func (h *Hkr) GetWindowOpenEndtime() time.Time {
	return unixStringToTime(h.Windowopenactiveendtime)
}

func (h *Hkr) DECTDeactivateWindowOpen(c *fritzbox.Client) error {
	return h.deactivateHelper(c, "sethkrwindowopen")
}

func (h *Hkr) DECTSetWindowOpen(c *fritzbox.Client, d time.Duration) (tm time.Time, err error) {
	return h.setEndpointHelper(c, d, "sethkrwindowopen")
}

// GetErrorString returns the error-message for the respective error-code. Errorcode 0 means no error. The error-messages originate from the official docs.
func (h *Hkr) GetErrorString() string {
	s, ok := errorMap[h.Errorcode]
	if !ok {
		return "unknown error: " + h.Errorcode
	}
	return s
}

// GetNextChangeTemperatureNumeric returns the temperature, that the next change will set it to (numeric)
func (h *Hkr) GetNextChangeTemperatureNumeric() float64 {
	return temperatureHelper(h.Nextchange.Tchange)
}

// GetNextChangeTemperature returns the temperature, that the next change will set it to (as string)
func (h *Hkr) GetNextChangeTemperature() string {
	return temperatureStringHelper(h.Nextchange.Tchange)
}

// GetNextchangeEndtime returns the time of the next temperature-change
func (h *Hkr) GetNextchangeEndtime() time.Time {
	return unixStringToTime(h.Nextchange.Endperiod)
}

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
func (h *Hkr) Reload(c *fritzbox.Client) error {
	return GetDeviceInfos(c, h.Device().Identifier, h)
}

// fromJSON is a helper function to unmarshal the JSON-Data into the struct
func (h *Hkr) fromJSON(m map[string]json.RawMessage, d *Device) (Capability, error) {
	err := json.Unmarshal(m["hkr"], &h)
	if err != nil {
		return h, err
	}

	h.device = d
	return h, nil
}

// setSollHelper is a helper function for setting the soll-temperature
func (h *Hkr) setSollHelper(c *fritzbox.Client, sollTemp float64) error {
	if sollTemp != TempOff && sollTemp != TempMax {
		sollTemp *= 2
	}

	data := fritzbox.Values{
		"sid":       c.SID(),
		"ain":       h.Device().Identifier,
		"switchcmd": "sethkrtsoll",
		"param":     strconv.FormatInt(int64(math.Round(sollTemp)), 10),
	}

	_, resp, err := c.AhaRequestString(http.MethodGet, "webservices/homeautoswitch.lua", data)
	if err != nil {
		return err
	}

	h.Tsoll = resp
	return nil
}

// temperatureHelper is a helper function for converting the temperature from string to float64
func temperatureHelper(r string) (s float64) {
	s, _ = strconv.ParseFloat(r, 64)
	if s == 0 {
		return 0
	} else if s == TempMax {
		return -1
	} else if s == TempOff {
		return -2
	} else {
		return s / 2
	}
}

// temperatureStringHelper is a helper function for converting the temperature from float64 to string
func temperatureStringHelper(r string) string {
	s, _ := strconv.ParseFloat(r, 64)
	if s == TempOff {
		r = "OFF"
	} else if s == TempMax {
		r = "MAX"
	} else if s == 0 {
		r = "0"
	} else {
		r = fmt.Sprintf("%.1f", s/2)
	}

	return r
}

// deactivateHelper is a helper function for deactivating boost/windowopen
func (h *Hkr) deactivateHelper(c *fritzbox.Client, endpoint string) error {
	_, err := h.setEndpointHelper(c, 0, endpoint)
	if err != nil {
		return err
	}

	return h.Reload(c)
}

// setEndpointHelper is a helper function for setting the endtimestamp for boost/windowopen
func (h *Hkr) setEndpointHelper(c *fritzbox.Client, d time.Duration, endpoint string) (tm time.Time, err error) {
	ts := "0"
	if d != 0 {
		if d > time.Hour*24 {
			err = fmt.Errorf("duration cannot be longer than 24 hours")
			return
		}

		t := time.Now().Add(d).Unix()
		ts = strconv.FormatInt(t, 10)
	}

	data := fritzbox.Values{
		"sid":          c.SID(),
		"ain":          h.Device().Identifier,
		"switchcmd":    endpoint,
		"endtimestamp": ts,
	}

	_, resp, err := c.AhaRequestString(http.MethodGet, "webservices/homeautoswitch.lua", data)
	if err != nil {
		return
	}
	i, _ := strconv.ParseInt(resp, 10, 64)
	tm = time.Unix(i, 0)

	err = h.Reload(c)
	return tm, err
}
