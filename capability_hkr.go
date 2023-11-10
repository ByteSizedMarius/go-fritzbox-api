package go_fritzbox_api

import (
	"encoding/json"
	"fmt"
	"github.com/ByteSizedMarius/go-fritzbox-api/util"
	"github.com/anaskhan96/soup"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Note:
// noinspection GoMixedReceiverTypes is required, because I use pointer receivers for classes,
// but the String() Method only works for non-pointer receivers.
// Maybe I'm missing something here? How is this supposed to work?

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
	Summeractive  string `json:"summeractive"`  // 1 if summer is currently active
	Holidayactive string `json:"holidayactive"` // same as summer

	// HkrUnstableInformation contains Information, that is not usually accessible via the API.
	// Empty by default. Call PyaFetchInformation on the Hkr to fill it using the PyAdapter.
	// SummerTime can be also fetched using FetchSummerTime (without PYA) or PyaFetchSummertime (with PYA)
	HkrUnstableInformation HkrPyaInformation `json:"-"`
	device                 *SmarthomeDevice
}

// HkrPyaInformation contains Information, that is not usually accessible via the API.
// Empty by default. Call PyaFetchInformation on the Hkr to fill it using the PyAdapter.
type HkrPyaInformation struct {
	SummerTime    SummerTime
	Zeitschaltung Zeitschaltung
	Holidays      Holidays
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
// PYA
//
// -------------------------------------------

// PyaFetchInformation fetches the SummerTimeFrame, the Zeitschaltung and the Holiday-Fields for HkrUnstableInformation.
// It does not return anything, but instead fills the HkrUnstableInformation-Field of the Hkr.
// There is only one Method to fill all three fields at once, as they are all fetched using the same initial request,
// thus making fetching all three of them separately much slower than just fetching them together.
// This Method is preferred over FetchSummerTime, as it does not require parsing the HTML-Response, thus (hopefully) making it more stable.
func (h *Hkr) PyaFetchInformation(pya *PyAdapter) (err error) {
	data, err := h.pyaPrepare(pya)
	if err != nil {
		return
	}

	h.HkrUnstableInformation.SummerTime = SummerTime{}.fromData(data)

	h.HkrUnstableInformation.Zeitschaltung, err = Zeitschaltung{}.fromData(data)
	if err != nil {
		return
	}

	h.HkrUnstableInformation.Holidays, err = Holidays{}.fromData(data)
	if err != nil {
		return
	}

	return
}

func (h *Hkr) pyaPrepare(pya *PyAdapter) (data url.Values, err error) {
	params, err := pya.GetArgsHKR(*h)
	if err != nil {
		return
	}

	// SID is invalidated after it is used with the webdriver, even if the driver has its own client
	// not sure why, but regenerating the SID doesn't hurt much
	err = pya.Client.Initialize()
	if err != nil {
		return
	}

	data = url.Values{
		"sid": {pya.Client.SID()},
	}

	for k, v := range params {
		data[k] = []string{v}
	}

	return data, nil
}

// -------------------------------------------
//
// Summertime
//
// -------------------------------------------

// SummerTime is the field that holds Information about the Start- and End-Date of the SummerTime-Frame set for the HKR.
// The struct only holds the raw Information as Strings. To get the actual Dates, use GetStartDate and GetEndDate.
// To fetch the Data, use FetchSummerTime (without PYA) or PyaFetchSummertime (with PYA).
type SummerTime struct {
	IsEnabled  bool
	StartDay   string
	StartMonth string
	EndDay     string
	EndMonth   string
}

// FromDates allows the Creation of a SummerTime struct using two dates
//
//goland:noinspection GoMixedReceiverTypes
func (stf *SummerTime) FromDates(start time.Time, end time.Time) {
	stf.StartDay = strconv.Itoa(start.Day())
	stf.StartMonth = strconv.Itoa(int(start.Month()))
	stf.EndDay = strconv.Itoa(end.Day())
	stf.EndMonth = strconv.Itoa(int(end.Month()))
	stf.IsEnabled = true
}

//goland:noinspection GoMixedReceiverTypes
func (SummerTime) fromData(data url.Values) (st SummerTime) {
	st.StartDay = data["SummerStartDay"][0]
	st.StartMonth = data["SummerStartMonth"][0]
	st.EndDay = data["SummerEndDay"][0]
	st.EndMonth = data["SummerEndMonth"][0]
	st.IsEnabled = data["SummerEnabled"][0] == "1"
	return
}

// GetStartDate returns a formatted time.Time-Struct for the Start of the SummerTime-Frame
//
//goland:noinspection GoMixedReceiverTypes
func (stf *SummerTime) GetStartDate() time.Time {
	return dateHelper(stf.StartMonth, stf.StartDay)
}

// GetEndDate returns a formatted time.Time-Struct for the End of the SummerTime-Frame
//
//goland:noinspection GoMixedReceiverTypes
func (stf *SummerTime) GetEndDate() time.Time {
	return dateHelper(stf.EndMonth, stf.EndDay)
}

// IsEmpty returns true, if the SummerTime is empty or not enabled.
//
//goland:noinspection GoMixedReceiverTypes
func (stf *SummerTime) IsEmpty() bool {
	return stf.IsEnabled == false && stf.StartDay == "" && stf.StartMonth == "" && stf.EndDay == "" && stf.EndMonth == ""

}

// Validate checks, if the SummerTime is valid.
// For this, it needs the Holidays, as SummerTime and Holidays cannot overlap.
// Validation happens automatically on the call of PyaSetSummerTime, with no additional time cost for fetching the holidays.
//
//goland:noinspection GoMixedReceiverTypes
func (stf *SummerTime) Validate(holidays Holidays) (err error) {
	if err = holidays.validateHolidays(); err != nil {
		return err
	}

	if !stf.IsEnabled || stf.IsEmpty() {
		return fmt.Errorf("summertime is empty or not enabled")
	}

	if stf.GetStartDate().After(stf.GetEndDate()) || stf.GetStartDate().Equal(stf.GetEndDate()) {
		return fmt.Errorf("summertime start must be before summertime end: %s - %s",
			stf.GetStartDate().Format("02.01"), stf.GetEndDate().Format("02.01"))
	}

	// Check if the Summertime overlaps with any holiday
	for _, hol := range holidays.Holidays {
		if !hol.IsEnabled() {
			continue
		}

		if util.DoDatesOverlap(stf.GetStartDate(), stf.GetEndDate(), hol.GetStartDate(), hol.GetEndDate()) {
			return fmt.Errorf("summertime overlaps with holiday %s - %s", hol.GetStartDate().Format("02.01."), hol.GetEndDate().Format("02.01."))
		}
	}

	return nil
}

//goland:noinspection GoMixedReceiverTypes
func (stf SummerTime) String() string {
	return fmt.Sprintf("SummerTime: {Start: %s, End: %s, Enabled: %t}",
		stf.GetStartDate().Format("02.01."), stf.GetEndDate().Format("02.01."), stf.IsEnabled)
}

// FetchSummerTime fetches the SummerTime-Frame for the HKR. It does not return anything, but instead fills the SummerTimeFrame-Field for the Hkr.
// If Pya is already initialized and logged in, using the Function PyaFetchSummertime is generally preferable.
func (h *Hkr) FetchSummerTime(c *Client) (err error) {
	data := url.Values{
		"sid":    {c.session.Sid},
		"page":   {"home_auto_hkr_edit"},
		"device": {h.device.ID},
	}

	resp, err := c.doRequest(http.MethodPost, "data.lua", data, true)
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
		h.HkrUnstableInformation.SummerTime = SummerTime{IsEnabled: false}
		return
	}

	ssd := row.Find("input", "id", "uiSummerStartDay")
	ssm := row.Find("input", "id", "uiSummerStartMonth")

	sed := row.Find("input", "id", "uiSummerEndDay")
	sem := row.Find("input", "id", "uiSummerEndMonth")

	if ssd.Error != nil || ssm.Error != nil || sed.Error != nil || sem.Error != nil {
		return fmt.Errorf("%s", "Some required Inputs not found")
	}

	h.HkrUnstableInformation.SummerTime = SummerTime{
		StartDay:   ssd.Attrs()["value"],
		StartMonth: ssm.Attrs()["value"],
		EndDay:     sed.Attrs()["value"],
		EndMonth:   sem.Attrs()["value"],
		IsEnabled:  true,
	}

	return nil
}

// PyaSetSummerTime sets the SummerTime for the HKR.
// Only Day/Month of the Time-Values is required. The Helper-Method DateFromMD can be used to create the Time-Values.
func (h *Hkr) PyaSetSummerTime(pya *PyAdapter, st SummerTime) (err error) {
	// get the data
	data, err := h.pyaPrepare(pya)
	if err != nil {
		return
	}

	// parse the holidays from data
	holidays, err := Holidays{}.fromData(data)
	if err != nil {
		return
	}

	// validate the given summertime using the holidays
	if err = st.Validate(holidays); err != nil {
		return
	}

	// update the data
	data["SummerEndDay"] = util.ToUrlValue(st.EndDay)
	data["SummerEndMonth"] = util.ToUrlValue(st.EndMonth)
	data["SummerStartDay"] = util.ToUrlValue(st.StartDay)
	data["SummerStartMonth"] = util.ToUrlValue(st.StartMonth)
	data["SummerEnabled"] = util.ToUrlValue(1)

	_, err = pya.Client.doRequest(http.MethodPost, "data.lua", data, true)
	return
}

// PyaDisableSummerTime disables the SummerTime for the HKR.
func (h *Hkr) PyaDisableSummerTime(pya *PyAdapter) (err error) {
	data, err := h.pyaPrepare(pya)
	if err != nil {
		return
	}

	data["SummerEnabled"] = util.ToUrlValue(0)
	_, err = pya.Client.doRequest(http.MethodPost, "data.lua", data, true)
	return
}

func dateHelper(month string, day string) time.Time {
	monthNr, _ := strconv.Atoi(month)
	dayNr, _ := strconv.Atoi(day)

	return util.DateFromMD(monthNr, dayNr)
}

// -------------------------------------------
//
// PyAdapter Zeitschaltung
//
// -------------------------------------------

// todo: Slots can be simplified. The third value in the timer_item-string is used to determine, to which days the timer_item applies
// if a similar start/end time is applied to multiple different days, this value can be set to the sum of the respective days
// it doesnt seem to matter from what I can tell, it's just a bit shorter. Probably not worth the effort.

// PyaSetZeitschaltung sets the Zeitschaltung for the HKR.
// Please see the Documentation for Zeitschaltung.
func (h *Hkr) PyaSetZeitschaltung(pya *PyAdapter, z Zeitschaltung) (err error) {
	err = z.Validate()
	if err != nil {
		return
	}

	data, err := h.pyaPrepare(pya)
	if err != nil {
		return
	}

	// Timer complete seems to indicate that the hkr is enabled 24/7. Remove it here.
	_, exists := data["timer_complete"]
	if exists {
		delete(data, "timer_complete")
	}

	// If there are already other timer_items set, remove them here, since we cannot be sure that all are overriden.
	for k := range data {
		if strings.HasPrefix(k, "timer_item_") {
			delete(data, k)
		}
	}

	slots := z.ToValues()
	for k, v := range slots {
		data[k] = util.ToUrlValue(v)
	}

	_, err = pya.Client.doRequest(http.MethodPost, "data.lua", data, true)
	return
}

// Zeitschaltung is a struct that holds the Values for the HKR-Timer in the Format that is expected by the Fritzbox.
// It consists of all days of the week, which in turn consist of ZeitSlots.
type Zeitschaltung struct {
	Tage []Tag
}

// Tag is a Part of Zeitschaltung. Every Tag has a Weekday and a list of ZeitSlots.
type Tag struct {
	Tag   time.Weekday
	Slots []ZeitSlot
}

// ZeitSlot is a Part of Tag. It consists of a Start and End time of the given Slot.
type ZeitSlot struct {
	Start time.Time
	End   time.Time
}

func timeString(t time.Time) string {
	return fmt.Sprintf("%02d%02d", t.Hour(), t.Minute())
}

// SlotTemplate is a helper for creating Slots.
// It consists of a Weekday, a Start and an End time.
// To create a Slot for Monday, from 11 AM - 3 PM (11-15:00), use SlotTemplate{time.Monday, "11:00", "15:00"}
// The Slots created this Way can then be added to the Zeitschaltung using Zeitschaltung.SlotFromTemplate or Zeitschaltung.SlotsFromTemplates
type SlotTemplate struct {
	Weekday time.Weekday
	Start   string
	End     string
}

// SlotFromTemplate creates a Slot from a SlotTemplate and adds it to the Zeitschaltung.
//
//goland:noinspection GoMixedReceiverTypes
func (z *Zeitschaltung) SlotFromTemplate(data SlotTemplate) (err error) {
	return z.SlotFromStrings(data.Weekday, data.Start, data.End)
}

// SlotsFromTemplates creates Slots from multiple SlotTemplates and adds them to the Zeitschaltung.
//
//goland:noinspection GoMixedReceiverTypes
func (z *Zeitschaltung) SlotsFromTemplates(data []SlotTemplate) (err error) {
	for _, d := range data {
		err = z.SlotFromTemplate(d)
		if err != nil {
			return
		}
	}
	return nil
}

// SlotFromStrings creates a Slot from a Weekday and two Strings (Start and End time) and adds it to the Zeitschaltung.
//
//goland:noinspection GoMixedReceiverTypes
func (z *Zeitschaltung) SlotFromStrings(weekday time.Weekday, start string, end string) (err error) {
	s, err := time.Parse("15:04", start)
	if err != nil {
		return
	}
	e, err := time.Parse("15:04", end)
	if err != nil {
		return
	}

	zs := ZeitSlot{Start: s, End: e}
	for i, t := range z.Tage {
		if t.Tag == weekday {
			z.Tage[i].Slots = append(z.Tage[i].Slots, zs)
			return
		}
	}

	z.Tage = append(z.Tage, Tag{Tag: weekday, Slots: []ZeitSlot{zs}})
	return
}

// Validate checks if the Zeitschaltung is valid. It returns an error if the Zeitschaltung is invalid.
// This Method is automatically called when PyaSetZeitschaltung is called, but it may be called in addition manually to check User-Input.
//
//goland:noinspection GoMixedReceiverTypes
func (z *Zeitschaltung) Validate() error {
	for _, t := range z.Tage {
		if t.Tag < 0 || t.Tag > 6 {
			return fmt.Errorf("invalid weekday: %d", t.Tag)
		}
		for _, s := range t.Slots {
			m := s.Start.Minute()
			if m >= 60 || m < 0 || m%15 != 0 {
				return fmt.Errorf("invalid minute: %d", m)
			}
		}
	}
	return nil
}

// ToValues converts the Zeitschaltung to a map[string]string, which is the Format the FritzBox expects.
//
//goland:noinspection GoMixedReceiverTypes
func (z *Zeitschaltung) ToValues() map[string]string {
	// Group the Start-Times
	var startTimes map[string]int
	var endTimes map[string]int
	startTimes = make(map[string]int)
	endTimes = make(map[string]int)

	for _, t := range z.Tage {
		if t.Tag == 0 {
			t.Tag = 7
		}

		for _, s := range t.Slots {
			dayN := util.Pow(2, int(t.Tag)-1)

			ts := timeString(s.Start)
			v, ok := startTimes[ts]
			if ok {
				startTimes[ts] = v + dayN
			} else {
				startTimes[ts] = dayN
			}

			es := timeString(s.End)
			v, ok = endTimes[es]
			if ok {
				endTimes[es] = v + dayN
			} else {
				endTimes[es] = dayN
			}
		}
	}

	var i int
	timerItems := make(map[string]string)
	for k, v := range startTimes {
		timerItems[fmt.Sprintf("timer_item_%d", i)] = fmt.Sprintf("%s;%d;%d", k, 1, v)
		i++
	}
	for k, v := range endTimes {
		timerItems[fmt.Sprintf("timer_item_%d", i)] = fmt.Sprintf("%s;%d;%d", k, 0, v)
		i++
	}

	return timerItems
}

//goland:noinspection GoMixedReceiverTypes
func (z Zeitschaltung) String() string {
	if len(z.Tage) == 0 {
		return "Zeitschaltung: Empty"
	}

	var s string
	for i := 1; i < 8; i++ {
		for _, t := range z.Tage {
			if int(t.Tag) == i%7 {
				s += fmt.Sprintf("%s: ", t.Tag)
				for _, slot := range t.Slots {
					s += fmt.Sprintf("%s - %s, ", slot.Start.Format("15:04"), slot.End.Format("15:04"))
				}
				s = strings.TrimSuffix(s, ", ") + "\n"
			}
		}
	}

	return s
}

// todo: testdata remove at some point
//data := url.Values{
//	"timer_item_0": {"0000;1;1"},
//	"timer_item_1": {"0400;0;1"},
//	"timer_item_2": {"2000;1;1"},
//
//	"timer_item_3": {"0000;0;2"},
//	"timer_item_4": {"0200;1;2"},
//	"timer_item_5": {"0600;0;2"},
//	"timer_item_6": {"1800;1;2"},
//	"timer_item_7": {"2200;0;2"},
//
//	"timer_item_10": {"1600;1;4"},
//	"timer_item_11": {"2000;0;4"},
//	"timer_item_12": {"0600;1;8"},
//	"timer_item_13": {"1000;0;8"},
//	"timer_item_14": {"1400;1;8"},
//	"timer_item_15": {"1800;0;8"},
//	"timer_item_16": {"0800;1;80"},
//	"timer_item_17": {"1600;0;80"},
//	"timer_item_18": {"1000;1;32"},
//	"timer_item_19": {"1400;0;32"},
//
//	"timer_item_8": {"0400;1;4"},
//	"timer_item_9": {"0800;0;4"},
//}

// Contains algorithm to parse the Zeitschaltung for the HKR from the Post-Parameters given, when a HKR-Device is edited.
// Algorithm can surely be improved, also there is a small Bug when Timeframes last until Midnight.
//
//goland:noinspection GoMixedReceiverTypes
func (Zeitschaltung) fromData(data url.Values) (z Zeitschaltung, err error) {
	relevantDataItems := url.Values{}
	for k, v := range data {
		if strings.HasPrefix(k, "timer_item_") {
			relevantDataItems[k] = v
		}
	}

	type dayRes struct {
		starts []time.Time
		ends   []time.Time
	}
	var res map[time.Weekday]dayRes
	res = make(map[time.Weekday]dayRes)

	for _, v := range relevantDataItems {
		// split the timer item into its values
		values := strings.Split(v[0], ";")

		// the third value is a binary representation of the days the timer item applies to
		var appliesTo int
		appliesTo, err = strconv.Atoi(values[2])
		if err != nil {
			return
		}

		// iterate over the binary representation
		for i, c := range util.Reverse(strings.ReplaceAll(fmt.Sprintf("%8s", strconv.FormatInt(int64(appliesTo), 2)), " ", "0")) {

			// if the bit is 0, the timer item does not apply to the day
			// skip
			if c == '0' {
				continue
			}

			// the first value is the time, parse it
			var t time.Time
			t, err = time.Parse("1504", values[0])
			if err != nil {
				return
			}

			// go weekdays are 0: Sunday, but the fritzbox uses 0: Monday
			// correct this
			wDay := time.Weekday(i + 1)
			if wDay == 7 {
				wDay = time.Weekday(0)
			}

			// TODO: HOTFIX (logic error)
			// Basically, what happens is that when 00:00 is put into the form, the fritzbox interprets it as 24:00
			// and puts it into the next day, while I currently interpret it as 23:59 of the same day
			// what I'm doing here is checking if the time is 00:00 and the value is 0 (meaning it's an end)
			// and then putting it into the previous day
			// maybe if its a start it should be put into the next day, but I'm not sure
			// also todo: investigate how the tovalues handles this
			if timeString(t) == "0000" && values[1] == "0" {
				wDay -= 1
				if wDay == -1 {
					wDay = time.Weekday(6)
				}
			}

			// Get the result's day
			day := res[wDay]

			// the second value is the type (0 = end, 1 = start)
			// if the value is one, search the item, where the start is not set yet
			if values[1] == "1" {
				day.starts = append(day.starts, t)
			} else
			// Value is 0, so it's an end
			{
				day.ends = append(day.ends, t)
			}

			res[wDay] = day
		}
	}

	// We now potentially have multiple starts and ends for each day
	// Map these together now by mapping the first start to the first end, the second start to the second end, etc.
	for k, v := range res {
		var day Tag
		day.Tag = k

		// First, sort both lists by time
		sort.Slice(v.starts, func(i, j int) bool {
			return v.starts[i].Before(v.starts[j])
		})
		sort.Slice(v.ends, func(i, j int) bool {
			return v.ends[i].Before(v.ends[j])
		})

		// Then, map the starts and ends together
		for i, start := range v.starts {
			if i >= len(v.ends) {
				day.Slots = append(day.Slots, ZeitSlot{Start: start, End: time.Time{}})
				break
			}

			day.Slots = append(day.Slots, ZeitSlot{Start: start, End: v.ends[i]})
		}
		if len(v.starts) < len(v.ends) {
			day.Slots = append(day.Slots, ZeitSlot{Start: time.Time{}, End: v.ends[len(v.ends)-1]})
		}

		z.Tage = append(z.Tage, day)
	}

	return
}

// -------------------------------------------
//
// PYA Holidays
//
// -------------------------------------------

// todo always keep holidays sorted by ID

// Holidays is the Struct that is used to set Holidays for the HKR.
// The FritzBox allows a maximum of 4 Holidays.
type Holidays struct {
	Holidays    [4]Holiday
	HolidayTemp float64
}

// Holiday is a single Holiday. It consists of a Start and End time, as well as an ID.
// For creating Holidays, use Holidays.AddHoliday.
type Holiday struct {
	ID         int
	StartDay   int
	StartMonth int
	StartHour  int
	EndDay     int
	EndMonth   int
	EndHour    int
	Enabled    int
}

// GetStartDate converts the Start-Values to a time.Time-Struct
//
//goland:noinspection GoMixedReceiverTypes
func (h *Holiday) GetStartDate() time.Time {
	return util.DateFromMDH(h.StartMonth, h.StartDay, h.StartHour)
}

// GetEndDate converts the End-Values to a time.Time-Struct
//
//goland:noinspection GoMixedReceiverTypes
func (h *Holiday) GetEndDate() time.Time {
	return util.DateFromMDH(h.EndMonth, h.EndDay, h.EndHour)
}

// IsEmpty returns true if the given Holiday is empty
//
//goland:noinspection GoMixedReceiverTypes
func (h *Holiday) IsEmpty() bool {
	return h.ID == 0
}

// IsEnabled returns true if the given Holiday is enabled
//
//goland:noinspection GoMixedReceiverTypes
func (h *Holiday) IsEnabled() bool {
	return h.Enabled == 1
}

//goland:noinspection GoMixedReceiverTypes
func (h Holiday) String() string {
	if !h.IsEnabled() {
		return fmt.Sprintf("Holiday %d: Not Set\n", h.ID)
	}

	return fmt.Sprintf("Holiday %d: %s - %s\n", h.ID, h.GetStartDate().Format("02.01, 15:04"), h.GetEndDate().Format("02.01, 15:04"))
}

//goland:noinspection GoMixedReceiverTypes
func (h *Holiday) fromData(data url.Values, id int) {
	h.ID, _ = strconv.Atoi(data[fmt.Sprintf("Holiday%dID", id)][0])
	h.StartDay, _ = strconv.Atoi(data[fmt.Sprintf("Holiday%dStartDay", id)][0])
	h.StartMonth, _ = strconv.Atoi(data[fmt.Sprintf("Holiday%dStartMonth", id)][0])
	h.StartHour, _ = strconv.Atoi(data[fmt.Sprintf("Holiday%dStartHour", id)][0])
	h.EndDay, _ = strconv.Atoi(data[fmt.Sprintf("Holiday%dEndDay", id)][0])
	h.EndMonth, _ = strconv.Atoi(data[fmt.Sprintf("Holiday%dEndMonth", id)][0])
	h.EndHour, _ = strconv.Atoi(data[fmt.Sprintf("Holiday%dEndHour", id)][0])
	h.Enabled, _ = strconv.Atoi(data[fmt.Sprintf("Holiday%dEnabled", id)][0])
}

// AddHoliday adds a Holiday to the Holidays-Struct. Helper Method. No Request is sent.
// It returns an error if the maximum amount of Holidays is reached.
// Holidays added via this Method are enabled by default. There can be 0-4 Holidays total.
//
//goland:noinspection GoMixedReceiverTypes
func (h *Holidays) AddHoliday(from time.Time, to time.Time) error {
	for i, hol := range h.Holidays {
		if !hol.IsEnabled() {
			h.Holidays[i] = Holiday{
				ID:         i + 1,
				StartDay:   from.Day(),
				StartMonth: int(from.Month()),
				StartHour:  from.Hour(),
				EndDay:     to.Day(),
				EndMonth:   int(to.Month()),
				EndHour:    to.Hour(),
				Enabled:    1,
			}
			return nil
		}
	}
	return fmt.Errorf("reached max. amount of holidays")
}

// Validate checks if the Holidays are valid. It returns an error if the Holidays are invalid.
// Holidays are invalid, if any Holidays overlap, or if the Start is after the End-Date of the same Holiday.
//
//goland:noinspection GoMixedReceiverTypes
func (h *Holidays) Validate(stf SummerTime) error {

	// Validate with Summertime (Dates cannot overlap)
	err := stf.Validate(*h)
	if err != nil {
		return err
	}

	return h.validateHolidays()
}

//goland:noinspection GoMixedReceiverTypes
func (h *Holidays) validateHolidays() error {
	for i, hol1 := range h.Holidays {
		if hol1.IsEmpty() || !hol1.IsEnabled() {
			continue
		}

		if hol1.GetStartDate().After(hol1.GetEndDate()) || hol1.GetStartDate().Equal(hol1.GetEndDate()) {
			return fmt.Errorf("holiday start must be before holiday end: %s - %s", hol1.GetStartDate().Format("02.01"), hol1.GetEndDate().Format("02.01"))
		}

		for j, hol2 := range h.Holidays {
			if hol2.IsEmpty() || !hol2.IsEnabled() || i == j {
				continue
			}

			if util.DoDatesOverlap(hol1.GetStartDate(), hol1.GetEndDate(), hol2.GetStartDate(), hol2.GetEndDate()) {
				return fmt.Errorf("holidays cannot overlap: %s - %s, %s - %s", hol1.GetStartDate().Format("02.01"), hol1.GetEndDate().Format("02.01"), hol2.GetStartDate().Format("02.01"), hol2.GetEndDate().Format("02.01"))
			}
		}
	}
	return nil
}

// ToValues converts the Holidays to a map[string]string, which is the Format the FritzBox expects.
//
//goland:noinspection GoMixedReceiverTypes
func (h *Holidays) ToValues() map[string]string {
	// if all empty, return nothing
	allEmpty := true
	for _, hol := range h.Holidays {
		if !hol.IsEmpty() {
			allEmpty = false
		}
	}
	if allEmpty {
		return nil
	}

	// if some empty, fill with previous and set enabled to 0 (this is the fritzbox behaviour)
	for i, hol := range h.Holidays {
		if hol.IsEmpty() {
			h.Holidays[i] = h.Holidays[0]
			h.Holidays[i].Enabled = 0
			h.Holidays[i].ID = i + 1
		}
	}

	params := make(map[string]string)
	for _, hol := range h.Holidays {
		params[fmt.Sprintf("Holiday%dID", hol.ID)] = strconv.Itoa(hol.ID)
		params[fmt.Sprintf("Holiday%dStartDay", hol.ID)] = strconv.Itoa(hol.StartDay)
		params[fmt.Sprintf("Holiday%dStartMonth", hol.ID)] = strconv.Itoa(hol.StartMonth)
		params[fmt.Sprintf("Holiday%dStartHour", hol.ID)] = strconv.Itoa(hol.StartHour)
		params[fmt.Sprintf("Holiday%dEndDay", hol.ID)] = strconv.Itoa(hol.EndDay)
		params[fmt.Sprintf("Holiday%dEndMonth", hol.ID)] = strconv.Itoa(hol.EndMonth)
		params[fmt.Sprintf("Holiday%dEndHour", hol.ID)] = strconv.Itoa(hol.EndHour)
		params[fmt.Sprintf("Holiday%dEnabled", hol.ID)] = strconv.Itoa(hol.Enabled)
	}

	params["Holidaytemp"] = fmt.Sprintf("%.1f", math.Round(h.HolidayTemp/0.5)*0.5)
	return params
}

//goland:noinspection GoMixedReceiverTypes
func (h Holidays) String() string {
	rt := fmt.Sprintf("Holiday-Temperature: %.1f\n", h.HolidayTemp)

	for _, hol := range h.Holidays {
		rt += hol.String()
	}

	return rt
}

// sort sorts the Holidays by ID.
// If this is not done, they might be in an inconsistent order.
//
//goland:noinspection GoMixedReceiverTypes
func (h *Holidays) sort() {
	sort.Slice(h.Holidays[:], func(i, j int) bool {
		return h.Holidays[i].ID < h.Holidays[j].ID
	})
}

// PyaSetHolidays sets the Holidays for the HKR.
// Please see the Documentation for Holidays.
func (h *Hkr) PyaSetHolidays(pya *PyAdapter, holidays Holidays) (err error) {
	data, err := h.pyaPrepare(pya)
	if err != nil {
		return
	}

	st := SummerTime{}.fromData(data)
	err = holidays.Validate(st)
	if err != nil {
		return
	}

	for k, v := range holidays.ToValues() {
		data[k] = util.ToUrlValue(v)
	}

	_, err = pya.Client.doRequest(http.MethodPost, "data.lua", data, true)
	return
}

// PyaDisableHoliday disables the given Holiday using the PYA.
func (h *Hkr) PyaDisableHoliday(pya *PyAdapter, hol Holiday) (err error) {
	data, err := h.pyaPrepare(pya)
	if err != nil {
		return
	}

	holStr := fmt.Sprintf("Holiday%dEnabled", hol.ID)
	if data[holStr][0] != "1" {
		err = fmt.Errorf("holiday %d is not enabled", hol.ID)
		return
	}

	data[holStr] = util.ToUrlValue(0)
	_, err = pya.Client.doRequest(http.MethodPost, "data.lua", data, true)
	return
}

// PyaDisableCurrentHoliday disables the currently active Holiday using the PYA.
// If no Holiday is currently active, an error is returned.
func (h *Hkr) PyaDisableCurrentHoliday(pya *PyAdapter) (err error) {
	data, err := h.pyaPrepare(pya)
	if err != nil {
		return
	}

	holidays, err := Holidays{}.fromData(data)
	if err != nil {
		return
	}

	for _, hol := range holidays.Holidays {
		now := time.Now()
		end := util.DateFromYMDH(now.Year(), hol.EndMonth, hol.EndDay, hol.EndHour)
		start := util.DateFromYMDH(now.Year(), hol.StartMonth, hol.StartDay, hol.StartHour)

		if hol.IsEnabled() && start.Before(time.Now()) && end.After(time.Now()) {
			data[fmt.Sprintf("Holiday%dEnabled", hol.ID)] = util.ToUrlValue(0)
			_, err = pya.Client.doRequest(http.MethodPost, "data.lua", data, true)

			// there can only be exactly one active holiday, so we can stop here
			return
		}
	}

	return fmt.Errorf("no holiday is currently enabled")
}

func (h *Hkr) PyaDisableAllHolidays(pya *PyAdapter) (err error) {
	data, err := h.pyaPrepare(pya)
	if err != nil {
		return
	}

	for i := 1; i <= 4; i++ {
		data[fmt.Sprintf("Holiday%dEnabled", i)] = util.ToUrlValue(0)
	}

	_, err = pya.Client.doRequest(http.MethodPost, "data.lua", data, true)
	return
}

//goland:noinspection GoMixedReceiverTypes
func (Holidays) fromData(data url.Values) (h Holidays, err error) {
	hTemp, ok := data["Holidaytemp"]

	// Field is only present if there is at least 1 holiday set
	if ok {
		h.HolidayTemp, _ = strconv.ParseFloat(hTemp[0], 64)
	}

	for i := 1; i <= 4; i++ {
		h.Holidays[i-1].fromData(data, i)
	}

	h.sort()
	return
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
