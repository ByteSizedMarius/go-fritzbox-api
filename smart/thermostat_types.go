package smart

import (
	"errors"
	"time"

	"github.com/ByteSizedMarius/go-fritzbox-api"
	"github.com/ByteSizedMarius/go-fritzbox-api/rest"
)

// ErrNotFound is returned when a device is not found by UID/AIN.
var ErrNotFound = errors.New("device not found")

// Thermostat represents a thermostat with clean Go types.
// Fields are sourced from ThermostatInterface unless noted otherwise.
type Thermostat struct {
	UID         string
	AIN         string
	Name        string
	IsConnected bool

	// Temperatures (°C). CurrentTemp is from TemperatureInterface, others from ThermostatInterface.
	TargetTemp  float64
	CurrentTemp float64 // from TemperatureInterface
	ComfortTemp float64
	ReducedTemp float64
	TempOffset  float64

	// State flags
	IsLocked        bool
	IsSummerActive  bool
	IsHolidayActive bool

	// Special modes
	Boost      SpecialMode
	WindowOpen SpecialMode

	// Battery (from parent device)
	BatteryLevel int
	IsBatteryLow bool

	// Schedule info
	NextChange *NextChange
}

// SpecialMode represents boost or window-open mode state.
type SpecialMode struct {
	Active  bool
	EndTime time.Time // zero if not active or no end time set
}

// NextChange represents an upcoming scheduled temperature change.
type NextChange struct {
	Time        time.Time
	Temperature float64
}

// GetAllThermostats returns all thermostats with clean Go types.
func GetAllThermostats(c *fritzbox.Client) ([]Thermostat, error) {
	overview, err := rest.GetOverview(c)
	if err != nil {
		return nil, err
	}

	units := overview.Thermostats()
	thermostats := make([]Thermostat, 0, len(units))

	for _, unit := range units {
		device := findDevice(overview.Devices, unit.ParentUid)
		thermostats = append(thermostats, thermostatFromOverview(unit, device))
	}

	return thermostats, nil
}

// GetThermostat returns a single thermostat by UID/AIN.
func GetThermostat(c *fritzbox.Client, uid string) (*Thermostat, error) {
	configuration, err := rest.GetConfigurationUnitByUID(c, uid)
	if err != nil {
		return nil, err
	}
	device, err := rest.GetConfigurationDeviceByUID(c, uid)
	if err != nil {
		return nil, err
	}
	t := thermostatFromConfiguration(configuration, device)

	return &t, nil
}

func thermostatFromConfiguration(unit *rest.EndpointConfigurationUnit, device *rest.EndpointConfigurationDevice) Thermostat {
	t := Thermostat{
		UID:         unit.UID,
		AIN:         unit.Ain,
		Name:        string(unit.Name),
		IsConnected: unit.IsConnected,
	}

	// Thermostat interface data
	if thermo := unit.Interfaces.ThermostatInterface; thermo != nil {
		t.TargetTemp = tempToCelsius(thermo.SetPointTemperature)
		t.ComfortTemp = tempToCelsius(thermo.ComfortTemperature)
		t.ReducedTemp = tempToCelsius(thermo.ReducedTemperature)
		t.IsLocked = derefBool(thermo.LockedDeviceLocalEnabled)
		t.IsSummerActive = derefBool(thermo.IsSummertimeActive)
		t.IsHolidayActive = derefBool(thermo.IsHolidayActive)

		if thermo.Boost != nil {
			t.Boost = specialModeFromRest(thermo.Boost)
		}
		if thermo.WindowOpenMode != nil {
			t.WindowOpen = specialModeFromRestConfig(thermo.WindowOpenMode)
		}
		if thermo.NextChange != nil {
			t.NextChange = nextChangeFromRest(thermo.NextChange)
		}
		if thermo.TemperatureOffset != nil {
			t.TempOffset = float64(derefFloat32(thermo.TemperatureOffset.InternalOffset))
		}
	}

	// Temperature interface (current temp)
	if temp := unit.Interfaces.TemperatureInterface; temp != nil {
		t.CurrentTemp = float64(derefFloat32(temp.Celsius))
	}

	// Battery info from device
	if device != nil {
		t.BatteryLevel = derefInt(device.BatteryValue)
		t.IsBatteryLow = derefBool(device.IsBatteryLow)
	}

	return t
}

func thermostatFromOverview(unit rest.HelperOverviewUnit, device *rest.HelperOverviewDevice) Thermostat {
	t := Thermostat{
		UID:         unit.UID,
		AIN:         unit.Ain,
		Name:        string(unit.Name),
		IsConnected: derefBool(unit.IsConnected),
		TempOffset:  -8000,
	}

	// Thermostat interface data
	if thermo := unit.Interfaces.ThermostatInterface; thermo != nil {
		t.TargetTemp = tempToCelsius(thermo.SetPointTemperature)
		t.ComfortTemp = tempToCelsius(thermo.ComfortTemperature)
		t.ReducedTemp = tempToCelsius(thermo.ReducedTemperature)
		t.IsLocked = derefBool(thermo.IsLockedDeviceLocal)
		t.IsSummerActive = derefBool(thermo.IsSummertimeActive)
		t.IsHolidayActive = derefBool(thermo.IsHolidayActive)

		if thermo.Boost != nil {
			t.Boost = specialModeFromRest(thermo.Boost)
		}
		if thermo.WindowOpenMode != nil {
			t.WindowOpen = specialModeFromRest(thermo.WindowOpenMode)
		}
		if thermo.NextChange != nil {
			t.NextChange = nextChangeFromRest(thermo.NextChange)
		}
	}

	// Temperature interface (current temp)
	if temp := unit.Interfaces.TemperatureInterface; temp != nil {
		t.CurrentTemp = float64(derefFloat32(temp.Celsius))
	}

	// Battery info from device
	if device != nil {
		t.BatteryLevel = derefInt(device.BatteryValue)
		t.IsBatteryLow = derefBool(device.IsBatteryLow)
	}

	return t
}

func specialModeFromRest(sm *rest.HelperSpecialModeThermostat) SpecialMode {
	mode := SpecialMode{
		Active: derefBool(sm.Enabled),
	}
	if sm.EndTime != nil && *sm.EndTime > 0 {
		mode.EndTime = time.Unix(int64(*sm.EndTime), 0)
	}
	return mode
}

func specialModeFromRestConfig(sm *rest.HelperWindowOpenModeConfig) SpecialMode {
	mode := SpecialMode{
		Active: derefBool(sm.Enabled),
	}
	if sm.EndTime != nil && *sm.EndTime > 0 {
		mode.EndTime = time.Unix(int64(*sm.EndTime), 0)
	}
	return mode
}

func nextChangeFromRest(nc *rest.HelperNextChange) *NextChange {
	if nc.ChangeTime == nil {
		return nil
	}
	return &NextChange{
		Time:        time.Unix(int64(*nc.ChangeTime), 0),
		Temperature: tempToCelsius(nc.TemperatureChange),
	}
}

func tempToCelsius(t *rest.HelperTemperature) float64 {
	if t == nil || t.Celsius == nil {
		return 0
	}
	return float64(*t.Celsius)
}

// ThermostatHandle provides a fluent API for thermostat operations.
// It wraps a client and UID, allowing method chaining without passing these repeatedly.
// Handles are stateless and safe to reuse or recreate as needed.
type ThermostatHandle struct {
	client *fritzbox.Client
	uid    string
}

// NewThermostatHandle creates a ThermostatHandle for the given thermostat UID.
func NewThermostatHandle(c *fritzbox.Client, uid string) *ThermostatHandle {
	return &ThermostatHandle{client: c, uid: uid}
}

// ThermostatConfig contains configuration data from the configuration endpoint.
// This includes schedules and periods that aren't available in the overview.
type ThermostatConfig struct {
	SummerPeriod   SummerPeriod
	HolidayPeriods []HolidayPeriod
	WeeklySchedule []ScheduleEntry
}

// SummerPeriod represents the configured summer mode period.
type SummerPeriod struct {
	Enabled   bool
	StartTime time.Time // month/day only (year is current year)
	EndTime   time.Time
}

// HolidayPeriod represents a configured holiday period.
type HolidayPeriod struct {
	StartTime      time.Time
	EndTime        time.Time
	Temperature    float64
	DeleteAfterEnd bool
}

// ScheduleEntry represents a single entry in the weekly heating schedule.
type ScheduleEntry struct {
	Time              int    // minutes from week start (Monday 00:00)
	TemperaturePreset string // "comfort" or "reduced"
}
