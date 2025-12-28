package smarthome

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ByteSizedMarius/go-fritzbox-api"
)

// GetThermostatConfig returns typed thermostat configuration for a unit.
func GetThermostatConfig(c *fritzbox.Client, uid string) (*ThermostatUnit, error) {
	body, status, err := c.RestGet(configUnitPath + uid)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(body))
	}

	var unit ThermostatUnit
	if err := json.Unmarshal(body, &unit); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return &unit, nil
}

func putConfig(c *fritzbox.Client, uid string, body map[string]any) error {
	respBody, status, err := c.RestPut(configUnitPath+uid, body)
	if err != nil {
		return err
	}
	if status != http.StatusOK && status != http.StatusNoContent {
		return fmt.Errorf("PUT failed with status %d: %s", status, string(respBody))
	}
	return nil
}

func putOverview(c *fritzbox.Client, uid string, body map[string]any) error {
	respBody, status, err := c.RestPut(overviewUnitPath+uid, body)
	if err != nil {
		return err
	}
	if status != http.StatusOK && status != http.StatusNoContent {
		return fmt.Errorf("PUT failed with status %d: %s", status, string(respBody))
	}
	return nil
}

// SetTargetTemperature sets the current target temperature.
// Valid range: 8-28°C. Values outside this range are rejected by the API.
func SetTargetTemperature(c *fritzbox.Client, uid string, celsius float64) error {
	body := map[string]any{
		"interfaces": map[string]any{
			"thermostatInterface": map[string]any{
				"setPointTemperature": map[string]any{
					"celsius": celsius,
					"mode":    "temperature",
				},
			},
		},
	}
	return putOverview(c, uid, body)
}

// SetTargetTemperatureOff turns off heating for this thermostat.
func SetTargetTemperatureOff(c *fritzbox.Client, uid string) error {
	body := map[string]any{
		"interfaces": map[string]any{
			"thermostatInterface": map[string]any{
				"setPointTemperature": map[string]any{
					"mode": "off",
				},
			},
		},
	}
	return putOverview(c, uid, body)
}

// SetTargetTemperatureOn turns heating to maximum for this thermostat.
func SetTargetTemperatureOn(c *fritzbox.Client, uid string) error {
	body := map[string]any{
		"interfaces": map[string]any{
			"thermostatInterface": map[string]any{
				"setPointTemperature": map[string]any{
					"mode": "on",
				},
			},
		},
	}
	return putOverview(c, uid, body)
}

// SetComfortTemperature sets the "comfort" preset (heating-on temperature).
// The weekly timer switches between comfort and reduced presets.
func SetComfortTemperature(c *fritzbox.Client, uid string, celsius float64) error {
	body := map[string]any{
		"interfaces": map[string]any{
			"thermostatInterface": map[string]any{
				"comfortTemperature": map[string]any{
					"celsius": celsius,
					"mode":    "temperature",
				},
			},
		},
	}
	return putConfig(c, uid, body)
}

// SetReducedTemperature sets the "reduced" preset (economy/heating-off temperature).
// The weekly timer switches between comfort and reduced presets.
func SetReducedTemperature(c *fritzbox.Client, uid string, celsius float64) error {
	body := map[string]any{
		"interfaces": map[string]any{
			"thermostatInterface": map[string]any{
				"reducedTemperature": map[string]any{
					"celsius": celsius,
					"mode":    "temperature",
				},
			},
		},
	}
	return putConfig(c, uid, body)
}

// SetBoost activates boost mode for the given duration (max 24 hours).
func SetBoost(c *fritzbox.Client, uid string, durationMinutes int) error {
	endTime := time.Now().Add(time.Duration(durationMinutes) * time.Minute).Unix()
	body := map[string]any{
		"interfaces": map[string]any{
			"thermostatInterface": map[string]any{
				"boost": map[string]any{
					"enabled": true,
					"endTime": endTime,
				},
			},
		},
	}
	return putOverview(c, uid, body)
}

// DeactivateBoost deactivates boost mode.
func DeactivateBoost(c *fritzbox.Client, uid string) error {
	body := map[string]any{
		"interfaces": map[string]any{
			"thermostatInterface": map[string]any{
				"boost": map[string]any{
					"enabled": false,
				},
			},
		},
	}
	return putOverview(c, uid, body)
}

// SetWindowOpen activates window-open mode for the given duration.
func SetWindowOpen(c *fritzbox.Client, uid string, durationMinutes int) error {
	endTime := time.Now().Add(time.Duration(durationMinutes) * time.Minute).Unix()
	body := map[string]any{
		"interfaces": map[string]any{
			"thermostatInterface": map[string]any{
				"windowOpenMode": map[string]any{
					"enabled": true,
					"endTime": endTime,
				},
			},
		},
	}
	return putOverview(c, uid, body)
}

// DeactivateWindowOpen deactivates window-open mode.
func DeactivateWindowOpen(c *fritzbox.Client, uid string) error {
	body := map[string]any{
		"interfaces": map[string]any{
			"thermostatInterface": map[string]any{
				"windowOpenMode": map[string]any{
					"enabled": false,
				},
			},
		},
	}
	return putOverview(c, uid, body)
}

// SetWindowOpenDetection configures automatic window detection.
// sensitivity: "low", "medium", "high". duration: 1-120 minutes.
// Invalid values are rejected by the API.
func SetWindowOpenDetection(c *fritzbox.Client, uid string, duration int, sensitivity string) error {
	body := map[string]any{
		"interfaces": map[string]any{
			"thermostatInterface": map[string]any{
				"windowOpenMode": map[string]any{
					"internalDuration":    duration,
					"internalSensitivity": sensitivity,
					"sensorMode":          "internal",
				},
			},
		},
	}
	return putConfig(c, uid, body)
}

// SetAdaptiveHeating enables or disables adaptive heating mode.
func SetAdaptiveHeating(c *fritzbox.Client, uid string, enabled bool) error {
	body := map[string]any{
		"interfaces": map[string]any{
			"thermostatInterface": map[string]any{
				"adaptiveHeatingModeEnabled": enabled,
			},
		},
	}
	return putConfig(c, uid, body)
}

// SetLocks sets device button lock and API lock.
func SetLocks(c *fritzbox.Client, uid string, localLock, apiLock bool) error {
	body := map[string]any{
		"interfaces": map[string]any{
			"thermostatInterface": map[string]any{
				"lockedDeviceLocalEnabled": localLock,
				"lockedDeviceApiEnabled":   apiLock,
			},
		},
	}
	return putConfig(c, uid, body)
}

// SetTemperatureOffset sets the temperature sensor offset.
// Valid range: -10 to +10°C. Values outside this range are rejected by the API.
func SetTemperatureOffset(c *fritzbox.Client, uid string, offset float64) error {
	body := map[string]any{
		"interfaces": map[string]any{
			"thermostatInterface": map[string]any{
				"temperatureOffset": map[string]any{
					"internalOffset": offset,
					"sensorMode":     "internal",
				},
			},
		},
	}
	return putConfig(c, uid, body)
}

// SetSummerPeriod configures summer mode dates.
// Dates are specified as month (1-12) and day (1-31).
func SetSummerPeriod(c *fritzbox.Client, uid string, enabled bool, startMonth, startDay, endMonth, endDay int) error {
	startMinutes := dateToMinutesFromYearStart(startMonth, startDay)
	endMinutes := dateToMinutesFromYearStart(endMonth, endDay)

	body := map[string]any{
		"interfaces": map[string]any{
			"thermostatInterface": map[string]any{
				"summerPeriod": map[string]any{
					"enabled":   enabled,
					"startTime": startMinutes,
					"endTime":   endMinutes,
				},
			},
		},
	}
	return putConfig(c, uid, body)
}

// SetWeeklyTimer sets the complete weekly schedule.
func SetWeeklyTimer(c *fritzbox.Client, uid string, entries []TimerEntry) error {
	body := map[string]any{
		"timer": map[string]any{
			"timerMode": "weekly",
			"weekly":    entries,
		},
	}
	return putConfig(c, uid, body)
}

// AddHoliday adds a holiday period with specified temperature.
func AddHoliday(c *fritzbox.Client, uid string, start, end time.Time, celsius float64) error {
	startMinutes := timeToMinutesFromYearStart(start)
	endMinutes := timeToMinutesFromYearStart(end)

	config, err := GetThermostatConfig(c, uid)
	if err != nil {
		return fmt.Errorf("get current config: %w", err)
	}

	periods := config.Interfaces.Thermostat.HolidayPeriods.Periods
	periods = append(periods, HolidayPeriod{
		StartTime:            startMinutes,
		EndTime:              endMinutes,
		DeleteAfterEndActive: true,
	})

	body := map[string]any{
		"interfaces": map[string]any{
			"thermostatInterface": map[string]any{
				"holidayPeriods": map[string]any{
					"periods": periods,
					"temperature": map[string]any{
						"celsius": celsius,
						"mode":    "temperature",
					},
				},
			},
		},
	}
	return putConfig(c, uid, body)
}

// RemoveHoliday removes a holiday period by index.
func RemoveHoliday(c *fritzbox.Client, uid string, index int) error {
	config, err := GetThermostatConfig(c, uid)
	if err != nil {
		return fmt.Errorf("get current config: %w", err)
	}

	periods := config.Interfaces.Thermostat.HolidayPeriods.Periods
	if index < 0 || index >= len(periods) {
		return fmt.Errorf("index out of range: %d (have %d holidays)", index, len(periods))
	}

	periods = append(periods[:index], periods[index+1:]...)

	body := map[string]any{
		"interfaces": map[string]any{
			"thermostatInterface": map[string]any{
				"holidayPeriods": map[string]any{
					"periods": periods,
				},
			},
		},
	}
	return putConfig(c, uid, body)
}

// ClearHolidays removes all holiday periods.
func ClearHolidays(c *fritzbox.Client, uid string) error {
	body := map[string]any{
		"interfaces": map[string]any{
			"thermostatInterface": map[string]any{
				"holidayPeriods": map[string]any{
					"periods": []any{},
				},
			},
		},
	}
	return putConfig(c, uid, body)
}

func dateToMinutesFromYearStart(month, day int) int64 {
	now := time.Now()
	date := time.Date(now.Year(), time.Month(month), day, 0, 0, 0, 0, time.UTC)
	yearStart := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	return int64(date.Sub(yearStart).Minutes())
}

func timeToMinutesFromYearStart(t time.Time) int64 {
	utc := t.UTC()
	truncated := time.Date(utc.Year(), utc.Month(), utc.Day(), utc.Hour(), 0, 0, 0, time.UTC)
	yearStart := time.Date(utc.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	return int64(truncated.Sub(yearStart).Minutes())
}
