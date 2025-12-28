# HKR (Heizkörperregler) - Radiator Thermostat

FRITZ!Box smart radiator thermostats for automated heating control.

**Supported devices:** FRITZ!DECT 301, FRITZ!DECT 302, Eurotronic COMET DECT

## Temperature Management

### Target Temperature (Soll)

The current heating target. Range: 8.0-28.0°C in 0.5°C steps.

**Special values:**
- `OFF` - Heating disabled, shows snowflake on device
- `MAX` - Maximum heating

```go
// AHA API
hkr.DECTGetSoll(client)           // Fetch current target
hkr.DECTSetSoll(client, 21.5)     // Set target (accepts float, int, string)
hkr.DECTSetSollOff(client)        // Turn off
hkr.DECTSetSollMax(client)        // Set to max

// REST API
smarthome.SetTargetTemperature(client, uid, 21.5)
smarthome.SetTargetTemperatureOff(client, uid)
smarthome.SetTargetTemperatureOn(client, uid)
```

### Comfort & Reduced Presets

The weekly timer alternates between two temperature presets:

- **Comfort (Komfort)** - "Heating on" temperature
- **Reduced (Absenk)** - "Economy" temperature

```go
// AHA API (read only)
hkr.DECTGetKomfort(client)  // Fetch comfort temperature
hkr.DECTGetAbsenk(client)   // Fetch reduced temperature

// REST API
cfg, _ := smarthome.GetThermostatConfig(client, uid)
cfg.Interfaces.Thermostat.ComfortTemperature  // Read comfort temp
cfg.Interfaces.Thermostat.ReducedTemperature  // Read reduced temp
smarthome.SetComfortTemperature(client, uid, 22.0)
smarthome.SetReducedTemperature(client, uid, 17.0)
```

## Operating Modes

### Boost Mode

Rapidly heats the radiator for a set duration (max 24 hours), then reverts to schedule.

```go
// AHA API
hkr.IsBoostActive()                        // Check status (local)
hkr.GetBoostEndtime()                      // Get end time (local)
hkr.SetBoost(client, 30*time.Minute)       // Activate for 30 min
hkr.DECTDeactivateBoost(client)            // Turn off

// REST API
cfg, _ := smarthome.GetThermostatConfig(client, uid)
cfg.Interfaces.Thermostat.Boost.Enabled    // Check status
cfg.Interfaces.Thermostat.Boost.EndTime    // Get end time
smarthome.SetBoost(client, uid, 30)        // Activate (minutes)
smarthome.DeactivateBoost(client, uid)     // Turn off
```

### Window Open Detection

Reduces heating when a window is detected open. Can be triggered automatically (via internal sensor) or manually via API.

```go
// AHA API
hkr.IsWindowOpen()                              // Check status (local)
hkr.GetWindowOpenEndtime()                      // Get end time (local)
hkr.DECTSetWindowOpen(client, 15*time.Minute)   // Simulate open
hkr.DECTDeactivateWindowOpen(client)            // Clear

// REST API
cfg, _ := smarthome.GetThermostatConfig(client, uid)
cfg.Interfaces.Thermostat.WindowOpenMode.Enabled   // Check status
cfg.Interfaces.Thermostat.WindowOpenMode.EndTime   // Get end time
smarthome.SetWindowOpen(client, uid, 15)           // Simulate open (minutes)
smarthome.DeactivateWindowOpen(client, uid)        // Clear
smarthome.SetWindowOpenDetection(client, uid, duration, sensitivity)  // Configure
```

### Summer Mode

Disables heating during warm months. When active, the device is automatically locked.

```go
// AHA API
hkr.IsSummerActive()  // Check status (local)

// REST API
cfg, _ := smarthome.GetThermostatConfig(client, uid)
cfg.Interfaces.Thermostat.IsSummertimeActive           // Check status
cfg.Interfaces.Thermostat.SummerPeriod                 // Get period config
smarthome.SetSummerPeriod(client, uid, true, 5, 1, 9, 30)  // Configure (May 1 - Sep 30)
```

### Holiday Mode

Sets a fixed temperature during vacation periods. Multiple periods can be configured. Device is automatically locked during holidays.

```go
// AHA API
hkr.IsHolidayActive()  // Check status (local)

// REST API
cfg, _ := smarthome.GetThermostatConfig(client, uid)
cfg.Interfaces.Thermostat.IsHolidayActive              // Check status
cfg.Interfaces.Thermostat.HolidayPeriods               // Get configured periods
smarthome.AddHoliday(client, uid, startTime, endTime, 18.0)
smarthome.RemoveHoliday(client, uid, 0)  // by index
smarthome.ClearHolidays(client, uid)
```

## Timer Schedule

Weekly schedule that switches between comfort and reduced temperatures.

- AHA API only shows the next scheduled change (`GetNextChangeTemperature()`, `GetNextchangeEndtime()`)
- REST API provides full schedule configuration via `SetWeeklyTimer()`

## Device Locks

| Lock Type | Description |
|-----------|-------------|
| API Lock | Prevents remote temperature changes |
| Local Lock | Disables physical buttons on device |

Locks are automatically enabled during summer and holiday modes.

```go
// AHA API
hkr.Lock        // API lock status (local)
hkr.Devicelock  // Local lock status (local)

// REST API
cfg, _ := smarthome.GetThermostatConfig(client, uid)
cfg.Interfaces.Thermostat.LockedDeviceAPIEnabled     // API lock status
cfg.Interfaces.Thermostat.LockedDeviceLocalEnabled   // Local lock status
smarthome.SetLocks(client, uid, localLock, apiLock)  // Configure
```

## Battery

```go
// AHA API
hkr.IsBatteryLow()  // true if low
hkr.Battery         // percentage as string

// REST API
overview, _ := smarthome.GetOverview(client)
device := overview.Devices[0]
device.IsBatteryLow  // true if low
device.BatteryValue  // percentage as int
```

## Error Codes (AHA only)

```go
hkr.Errorcode        // raw code
hkr.GetErrorString() // human-readable message
```

| Code | Meaning |
|------|---------|
| 0 | No error |
| 1 | Adaptation failed - device mounted correctly? |
| 2 | Valve stroke too short or battery weak |
| 3 | No valve movement - is valve plunger free? |
| 4 | Installation in progress |
| 5 | Installation mode - ready to mount |
| 6 | Auto-adapting to valve stroke |

## Temperature Sensors

The current room temperature reading:

```go
// AHA API (via Temperature capability, not HKR)
if device.HasCapability(aha.CTempSensor) {
    temp := aha.GetCapability[*aha.Temperature](device)
    celsius := temp.GetCelsius()
}

// REST API
cfg, _ := smarthome.GetThermostatConfig(client, uid)
celsius := cfg.Interfaces.Temperature.Celsius
```

**External sensor assignment** (REST only): Use `SetExternalSensor()` to use another device's temperature reading for heating control.

**Calibration offset** (REST only): Adjust sensor accuracy with `SetTemperatureOffset()` (-10 to +10°C).

## API Comparison

| Feature | AHA HTTP | REST API |
|---------|----------|----------|
| **Read** |||
| Target/comfort/economy temp | ✅ | ✅ |
| Boost/window status | ✅ | ✅ |
| Summer/holiday active | ✅ | ✅ |
| Timer schedule | next change only | ✅ full |
| **Write** |||
| Set target temperature | ✅ | ✅ |
| Set boost/window open | ✅ | ✅ |
| Set comfort/economy presets | - | ✅ |
| Configure timer schedule | - | ✅ |
| Configure summer period | - | ✅ |
| Configure holidays | - | ✅ |
| Set window sensitivity | - | ✅ |
| Set locks | - | ✅ |
| Set temp offset | - | ✅ |
| Assign external sensor | - | ✅ |