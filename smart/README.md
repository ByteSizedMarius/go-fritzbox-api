# smart

[![Go Reference](https://pkg.go.dev/badge/github.com/ByteSizedMarius/go-fritzbox-api/v2.svg)](https://pkg.go.dev/github.com/ByteSizedMarius/go-fritzbox-api/v2)

High-level helpers for FRITZ!Box smart home devices. Wraps the `rest` package with clean Go types and a fluent Handle API.

## Architecture

### Patterns

**Reading:** Get functions fetch from the overview endpoint:
- `GetThermostat(client, uid)` / `GetAllThermostats(client)`
- `GetButton(client, uid)` / `GetAllButtons(client)`
- `GetWindowDetector(client, uid)` / `GetAllWindowDetectors(client)`

Some data (schedules, periods) isn't returned by the overview endpoint; use a handle's `GetConfig()` for that.

**Writing:** Handles provide write access:
```go
handle := smart.NewThermostatHandle(client, uid)
handle.SetTargetTemperature(21.5)
```

---

## Thermostat

See [docs/hkr.md](../docs/hkr.md) for thermostat concepts and [examples/](examples/) for code.

### Types

```go
type Thermostat struct {
    UID, AIN, Name string
    IsConnected    bool

    TargetTemp, CurrentTemp float64  // °C
    ComfortTemp, ReducedTemp float64
    TempOffset float64

    IsLocked, IsSummerActive, IsHolidayActive bool
    Boost, WindowOpen SpecialMode

    BatteryLevel int   // 0-100
    IsBatteryLow bool
    NextChange   *NextChange
}

type ThermostatConfig struct {
    SummerPeriod   SummerPeriod
    HolidayPeriods []HolidayPeriod
    WeeklySchedule []ScheduleEntry
}
```

### Functions

```go
GetThermostat(client, uid) (*Thermostat, error)
GetAllThermostats(client) ([]Thermostat, error)
NewThermostatHandle(client, uid) *ThermostatHandle
```

### ThermostatHandle Methods

**Reading:**
- `Get() (*Thermostat, error)`
- `GetConfig() (*ThermostatConfig, error)`

**Immediate Control:**
- `SetTargetTemperature(celsius float64) error` - set current target (8-28°C)
- `ApplyComfortPreset() error` - set target to comfort preset value
- `ApplyReducedPreset() error` - set target to reduced preset value
- `TurnOff() error` - frost protection mode
- `TurnOn() error` - maximum heat mode

**Preset Configuration** (values used by weekly timer):
- `SetComfortPreset(celsius float64) error` - configure comfort preset
- `SetReducedPreset(celsius float64) error` - configure reduced preset
- `SetTemperatureOffset(offset float64) error` - sensor offset (-10 to +10°C)

**Special Modes:**
- `SetBoost(minutes int) error`
- `DeactivateBoost() error`
- `SetWindowOpen(minutes int) error`
- `DeactivateWindowOpen() error`

**Configuration:**
- `SetAdaptiveHeating(enabled bool) error`
- `SetLocks(localLock, apiLock bool) error`
- `SetWindowOpenDetection(duration int, sensitivity string) error` - sensitivity: low/medium/high
- `SetWeeklyTimer(entries []rest.HelperBaseTimerWeeklyObj) error`
- `SetSummerPeriod(enabled bool, startMonth, startDay, endMonth, endDay int) error`

**Holidays:**
- `AddHoliday(start, end time.Time, celsius float64) error`
- `RemoveHoliday(index int) error`
- `ClearHolidays() error`

---

## Button

### Types

```go
type Button struct {
    UID, AIN, Name string
    IsConnected    bool
    LastPressTime  time.Time
    BatteryLevel   int
    IsBatteryLow   bool
}

type ButtonConfig struct {
    ControlMode     string   // on/off/toggle/event
    DestinationMode string   // disabled/templates/units
    DestinationUids []string
    SwitchDuration  *SwitchDuration
    ActivePeriod    *ButtonActivePeriod
    AvailableUnits, AvailableTemplates []string
}
```

### Functions

```go
GetButton(client, uid) (*Button, error)
GetAllButtons(client) ([]Button, error)
NewButtonHandle(client, uid) *ButtonHandle
```

### ButtonHandle Methods

**Reading:**
- `Get() (*Button, error)`
- `GetConfig() (*ButtonConfig, error)`

**Configuration:**
- `SetControlMode(mode string) error` - on/off/toggle/unknown
- `SetSwitchDuration(mode string, toggleBackMinutes int) error` - permanent/toggleBack
- `SetDestinations(mode string, uids []string) error` - disabled/units/templates
- `SetActivePeriodPermanent() error`
- `SetActivePeriodFixed(startTime, endTime time.Time) error`
- `SetActivePeriodAstronomic() error`

---

## WindowDetector

### Types

```go
type WindowDetector struct {
    UID, AIN, Name string
    IsConnected    bool
    IsOpen         bool
    LastAlertTime  time.Time
    BatteryLevel   int
    IsBatteryLow   bool
}

type WindowDetectorConfig struct {
    ThermostatDestinationUIDs []string
    AvailableThermostats      []string
    ControlMode               string
}
```

### Functions

```go
GetWindowDetector(client, uid) (*WindowDetector, error)
GetAllWindowDetectors(client) ([]WindowDetector, error)
NewWindowDetectorHandle(client, uid) *WindowDetectorHandle
```

### WindowDetectorHandle Methods

**Reading:**
- `Get() (*WindowDetector, error)`
- `GetConfig() (*WindowDetectorConfig, error)`

**Thermostat Links:**
- `AddThermostat(thermostatUID string) error` - link to thermostat
- `RemoveThermostat(thermostatUID string) error`

---

## Generic Helpers

```go
// Check if a unit has a specific interface
smart.HasInterface[*rest.IFThermostatOverview](unit) bool

// Get interface from unit
therm, ok := smart.GetInterface[*rest.IFThermostatOverview](unit)
```

