# HKR (Heizkörperregler) - Radiator Thermostat

FRITZ!Box smart radiator thermostats for automated heating control.

**Supported devices:** FRITZ!DECT 301, FRITZ!DECT 302, Eurotronic COMET DECT

**Packages:** The `smart` package provides a high-level API for common operations. For advanced features, use the `rest` package directly with generated types. Please open an issue for common use cases that should be added as helpers to `smart`.

**See also:** [smart/README.md](../smart/README.md) for API reference, [examples/](../smart/examples/) for code

## Temperature Management

### Target Temperature (Soll)

The current heating target. Range: 8.0-28.0°C in 0.5°C steps.

**Special values:**
- `OFF` - Heating disabled, shows snowflake on device
- `MAX` - Maximum heating

### Comfort & Reduced Presets

The weekly timer alternates between two temperature presets:

- **Comfort (Komfort)** - "Heating on" temperature
- **Reduced (Absenk)** - "Economy" temperature

## Operating Modes

### Boost Mode

Rapidly heats the radiator for a set duration (max 24 hours), then reverts to schedule.

### Window Open Detection

Reduces heating when a window is detected open. Can be triggered automatically (via internal sensor) or manually via API. Sensitivity can be configured to low, medium, or high.

### Summer Mode

Disables heating during warm months. When active, the device is automatically locked.

### Holiday Mode

Sets a fixed temperature during vacation periods. Multiple periods can be configured. Device is automatically locked during holidays.

## Timer Schedule

Weekly schedule that switches between comfort and reduced temperatures.

- AHA API only shows the next scheduled change
- `smart` package provides full schedule configuration

## Device Locks

| Lock Type | Description |
|-----------|-------------|
| API Lock | Prevents remote temperature changes |
| Local Lock | Disables physical buttons on device |

Locks are automatically enabled during summer and holiday modes.

## Battery

Thermostats report battery level (0-100%) and a low battery warning flag.

## Error Codes

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

Thermostats have a built-in temperature sensor. An external sensor can be assigned via the REST API for more accurate room temperature readings. A calibration offset (-10 to +10°C) can be configured.

## API Comparison

| Feature | AHA HTTP | smart package |
|---------|----------|---------------|
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
| Assign external sensor | - | via `rest` |
