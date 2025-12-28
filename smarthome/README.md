# smarthome

FRITZ!Box Smart Home REST API client. Requires FRITZ!OS 8.20+.

## API Background

Smart Home REST API (`/api/v0/smarthome/...`)
- JSON-based, requires FRITZ!OS 8.20+
- [OpenAPI spec](https://avm.de/fileadmin/user_upload/Global/Service/Schnittstellen/smart_home_open_api_v0.yaml)
- More comprehensive (full thermostat config, timer schedules)
- AVM-recommended for new development

## Usage

```go
import (
    "github.com/ByteSizedMarius/go-fritzbox-api"
    "github.com/ByteSizedMarius/go-fritzbox-api/smarthome"
)

client := fritzbox.New("username", "password")
client.Connect()

// Get all devices
overview, _ := smarthome.GetOverview(client)
for _, device := range overview.Devices {
    fmt.Printf("%s (%s)\n", device.Name, device.ProductName)
}

// Find thermostat and set temperature
uid, _ := smarthome.FindThermostatUnitUID(client, "Living Room")
smarthome.SetTargetTemperature(client, uid, 21.5)
```

## Functions

### Overview
- `GetOverview(c)` - Full smart home state (devices, groups, units)
- `GetUnit(c, uid)` - Single unit overview
- `GetUnitConfig(c, uid)` - Raw unit configuration
- `FindDeviceByName(c, name)` - Search device by name

### Thermostat Control
- `GetThermostatConfig(c, uid)` - Full thermostat configuration
- `SetTargetTemperature(c, uid, celsius)` - Set current target (8-28°C)
- `SetTargetTemperatureOff(c, uid)` - Turn off heating
- `SetTargetTemperatureOn(c, uid)` - Set to maximum

**Temperature Presets:**
- `SetComfortTemperature(c, uid, celsius)` - Set "comfort" (heating on) preset
- `SetReducedTemperature(c, uid, celsius)` - Set "reduced" (heating off/economy) preset

The weekly timer switches between these two presets automatically.

**Modes:**
- `SetBoost(c, uid, minutes)` / `DeactivateBoost(c, uid)`
- `SetWindowOpen(c, uid, minutes)` / `DeactivateWindowOpen(c, uid)`
- `SetWindowOpenDetection(c, uid, duration, sensitivity)`
- `SetAdaptiveHeating(c, uid, enabled)`

**Settings:**
- `SetLocks(c, uid, localLock, apiLock)`
- `SetTemperatureOffset(c, uid, offset)` - Sensor calibration (-10 to +10°C)
- `SetSummerPeriod(c, uid, enabled, startMonth, startDay, endMonth, endDay)`
- `SetWeeklyTimer(c, uid, entries)` - Schedule comfort/reduced switching
- `AddHoliday(c, uid, start, end, celsius)` / `RemoveHoliday(c, uid, index)` / `ClearHolidays(c, uid)`

### Detailed HKR Documentation

See [HKR Documentation](../docs/hkr.md) for:
- Feature explanations (summer mode, holidays, window detection)
- AHA vs REST API comparison
- Temperature encoding details

## API Reference

Based on the [AVM Smart Home REST API](https://fritz.support/resources/SmarthomeRestApiFRITZOS82.html).
