// Package smarthome provides access to FRITZ!Box Smart Home features via the REST API.
// Requires FRITZ!OS 8.20 or later.
// API documentation: https://fritz.support/resources/SmarthomeRestApiFRITZOS82.html
package smarthome

// SmarthomeOverview represents the response from GET /api/v0/smarthome/overview
type SmarthomeOverview struct {
	Devices   []SmarthomeDevice   `json:"devices"`
	Groups    []SmarthomeGroup    `json:"groups"`
	Units     []SmarthomeUnit     `json:"units"`
	Templates []SmarthomeTemplate `json:"templates"`
	Triggers  []SmarthomeTrigger  `json:"triggers"`
}

// SmarthomeDevice represents a device in the REST API response
type SmarthomeDevice struct {
	UID             string   `json:"UID"`
	AIN             string   `json:"ain"`
	Name            string   `json:"name"`
	ProductCategory string   `json:"productCategory"`
	ProductName     string   `json:"productName"`
	Manufacturer    string   `json:"manufacturer"`
	FirmwareVersion string   `json:"firmwareVersion"`
	IsConnected     bool     `json:"isConnected"`
	IsBatteryLow    bool     `json:"isBatteryLow"`
	BatteryValue    int      `json:"batteryValue"`
	UnitUIDs        []string `json:"unitUids"`
}

// SmarthomeGroup represents a group in the REST API response
type SmarthomeGroup struct {
	UID     string   `json:"UID"`
	Name    string   `json:"name"`
	Members []string `json:"members"`
}

// SmarthomeUnit represents a unit in the REST API response (overview level).
// Interfaces is untyped because different unit types have different interface schemas.
type SmarthomeUnit struct {
	UID        string         `json:"UID"`
	Name       string         `json:"name"`
	Type       string         `json:"type"`
	Interfaces map[string]any `json:"interfaces"`
}

// SmarthomeTemplate represents a template
type SmarthomeTemplate struct {
	UID  string `json:"UID"`
	Name string `json:"name"`
}

// SmarthomeTrigger represents a trigger
type SmarthomeTrigger struct {
	UID    string `json:"UID"`
	Active bool   `json:"active"`
}

// ThermostatUnit is the full unit response from GET /configuration/units/{UID}
type ThermostatUnit struct {
	UID        string               `json:"UID"`
	AIN        string               `json:"ain"`
	Name       string               `json:"name"`
	UnitType   string               `json:"unitType"`
	DeviceUID  string               `json:"deviceUid"`
	Interfaces ThermostatInterfaces `json:"interfaces"`
	Timer      *ThermostatTimer     `json:"timer,omitempty"`
}

// ThermostatInterfaces contains all interface data for a thermostat unit
type ThermostatInterfaces struct {
	Temperature TemperatureInterface `json:"temperatureInterface"`
	Thermostat  ThermostatInterface  `json:"thermostatInterface"`
}

// TemperatureInterface contains current temperature reading
type TemperatureInterface struct {
	Celsius float64 `json:"celsius"`
	State   string  `json:"state"`
}

// ThermostatInterface contains full thermostat configuration
type ThermostatInterface struct {
	SetPointTemperature Temperature `json:"setPointTemperature"`
	ThermostatState     string      `json:"thermostatState"`
	State               string      `json:"state"`

	ComfortTemperature Temperature `json:"comfortTemperature"`
	ReducedTemperature Temperature `json:"reducedTemperature"`

	Boost          SpecialMode `json:"boost"`
	WindowOpenMode WindowOpen  `json:"windowOpenMode"`

	AdaptiveHeatingModeEnabled bool `json:"adaptiveHeatingModeEnabled"`
	IsAdaptiveActive           bool `json:"isAdaptiveActive"`
	IsSummertimeActive         bool `json:"isSummertimeActive"`
	IsHolidayActive            bool `json:"isHolidayActive"`

	LockedDeviceAPIEnabled   bool `json:"lockedDeviceApiEnabled"`
	LockedDeviceLocalEnabled bool `json:"lockedDeviceLocalEnabled"`

	TemperatureOffset    TemperatureOffset    `json:"temperatureOffset"`
	TemperatureRangeLock TemperatureRangeLock `json:"temperatureRangeLock"`
	SummerPeriod         SummerPeriod         `json:"summerPeriod"`
	HolidayPeriods       HolidayPeriods       `json:"holidayPeriods"`

	NextChange *NextChange `json:"nextChange,omitempty"`
}

// Temperature represents a temperature value with optional mode
type Temperature struct {
	Celsius float64 `json:"celsius"`
	Mode    string  `json:"mode,omitempty"` // "temperature", "on", "off"
}

// SpecialMode for boost mode
type SpecialMode struct {
	Enabled bool   `json:"enabled"`
	EndTime *int64 `json:"endTime"`
}

// WindowOpen contains window-open mode configuration
type WindowOpen struct {
	Enabled                  bool     `json:"enabled"`
	EndTime                  *int64   `json:"endTime"`
	InternalDuration         int      `json:"internalDuration"`
	InternalSensitivity      string   `json:"internalSensitivity"` // "low", "medium", "high"
	SensorMode               string   `json:"sensorMode"`          // "internal", "external"
	ExternalSensorUIDs       []string `json:"externalSensorUids"`
	ExternalAvailableSensors []string `json:"externalAvailableSensors"`
}

// TemperatureOffset configuration
type TemperatureOffset struct {
	InternalOffset           float64  `json:"internalOffset"` // -10 to +10
	SensorMode               string   `json:"sensorMode"`
	ExternalSensorUID        string   `json:"externalSensorUid"`
	ExternalAvailableSensors []string `json:"externalAvailableSensors"`
}

// TemperatureRangeLock restricts the allowed temperature range
type TemperatureRangeLock struct {
	TemperatureLockEnabled bool        `json:"temperatureLockEnabled"`
	Minimum                Temperature `json:"minimum"`
	Maximum                Temperature `json:"maximum"`
}

// SummerPeriod configuration
type SummerPeriod struct {
	Enabled   bool  `json:"enabled"`
	StartTime int64 `json:"startTime"` // Minutes from year start
	EndTime   int64 `json:"endTime"`
}

// HolidayPeriods contains all configured holiday periods
type HolidayPeriods struct {
	Periods     []HolidayPeriod `json:"periods"`
	Temperature Temperature     `json:"temperature"`
}

// HolidayPeriod represents a single holiday period
type HolidayPeriod struct {
	StartTime            int64 `json:"startTime"` // Minutes from year start
	EndTime              int64 `json:"endTime"`
	DeleteAfterEndActive bool  `json:"deleteAfterEndActive"`
}

// ThermostatTimer contains the weekly schedule
type ThermostatTimer struct {
	TimerMode string       `json:"timerMode"` // "weekly"
	Weekly    []TimerEntry `json:"weekly"`
}

// TimerEntry represents a single entry in the weekly schedule
type TimerEntry struct {
	TemperaturePreset string `json:"temperaturePreset"` // "comfort", "reduced"
	Time              int    `json:"time"`              // Minutes from week start (Mon 00:00 = 0)
}

// NextChange contains info about the next scheduled temperature change
type NextChange struct {
	ChangeTime        int64       `json:"changeTime"` // Unix timestamp
	TemperatureChange Temperature `json:"temperatureChange"`
}
