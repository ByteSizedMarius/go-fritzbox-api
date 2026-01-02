package smart

import (
	"fmt"
	"time"

	"github.com/ByteSizedMarius/go-fritzbox-api/rest"
)

// Get fetches the current thermostat state from the overview endpoint.
func (h *ThermostatHandle) Get() (*Thermostat, error) {
	return GetThermostat(h.client, h.uid)
}

// GetConfig fetches the thermostat configuration including schedules and periods.
func (h *ThermostatHandle) GetConfig() (*ThermostatConfig, error) {
	config, err := rest.GetConfigurationUnitByUID(h.client, h.uid)
	if err != nil {
		return nil, err
	}

	result := &ThermostatConfig{}

	if ti := config.Interfaces.ThermostatInterface; ti != nil {
		if ti.SummerPeriod != nil {
			result.SummerPeriod = SummerPeriod{
				Enabled:   ti.SummerPeriod.Enabled,
				StartTime: intPtrToTime(ti.SummerPeriod.StartTime),
				EndTime:   intPtrToTime(ti.SummerPeriod.EndTime),
			}
		}

		if ti.HolidayPeriods != nil && ti.HolidayPeriods.Periods != nil {
			periods := *ti.HolidayPeriods.Periods
			result.HolidayPeriods = make([]HolidayPeriod, len(periods))

			var holidayTemp float64
			if ti.HolidayPeriods.Temperature != nil && ti.HolidayPeriods.Temperature.Celsius != nil {
				holidayTemp = float64(*ti.HolidayPeriods.Temperature.Celsius)
			}

			for i, p := range periods {
				result.HolidayPeriods[i] = HolidayPeriod{
					StartTime:      intPtrToTime(&p.StartTime),
					EndTime:        intPtrToTime(&p.EndTime),
					Temperature:    holidayTemp,
					DeleteAfterEnd: p.DeleteAfterEndActive != nil && *p.DeleteAfterEndActive,
				}
			}
		}
	}

	if config.Timer != nil && config.Timer.Weekly != nil {
		weekly := *config.Timer.Weekly
		result.WeeklySchedule = make([]ScheduleEntry, 0, len(weekly))
		for _, entry := range weekly {
			if entry.Time != nil && entry.TemperaturePreset != nil {
				result.WeeklySchedule = append(result.WeeklySchedule, ScheduleEntry{
					Time:              *entry.Time,
					TemperaturePreset: string(*entry.TemperaturePreset),
				})
			}
		}
	}

	return result, nil
}

func intPtrToTime(minutes *int) time.Time {
	if minutes == nil {
		return time.Time{}
	}
	return TimeFromYearMinutes(time.Now().UTC().Year(), int64(*minutes))
}

// SetTargetTemperature sets the current target temperature (8-28°C).
func (h *ThermostatHandle) SetTargetTemperature(celsius float64) error {
	cel := float32(celsius)
	data := &rest.EndpointOverviewPutUnit{
		Interfaces: rest.IFPutUnitInterfaces{
			ThermostatInterface: &rest.IFThermostatOverview{
				SetPointTemperature: &rest.HelperTemperature{
					Celsius: &cel,
					Mode:    rest.HelperTemperatureModeTemperature,
				},
			},
		},
	}
	return rest.PutOverviewUnit(h.client, h.uid, data)
}

// TurnOff turns off heating for this thermostat.
func (h *ThermostatHandle) TurnOff() error {
	data := &rest.EndpointOverviewPutUnit{
		Interfaces: rest.IFPutUnitInterfaces{
			ThermostatInterface: &rest.IFThermostatOverview{
				SetPointTemperature: &rest.HelperTemperature{
					Mode: rest.HelperTemperatureModeOff,
				},
			},
		},
	}
	return rest.PutOverviewUnit(h.client, h.uid, data)
}

// TurnOn turns heating to maximum for this thermostat.
func (h *ThermostatHandle) TurnOn() error {
	data := &rest.EndpointOverviewPutUnit{
		Interfaces: rest.IFPutUnitInterfaces{
			ThermostatInterface: &rest.IFThermostatOverview{
				SetPointTemperature: &rest.HelperTemperature{
					Mode: rest.HelperTemperatureModeOn,
				},
			},
		},
	}
	return rest.PutOverviewUnit(h.client, h.uid, data)
}

// SetBoost activates boost mode for the given duration in minutes.
func (h *ThermostatHandle) SetBoost(minutes int) error {
	endTime := int(time.Now().Add(time.Duration(minutes) * time.Minute).Unix())
	enabled := true
	data := &rest.EndpointOverviewPutUnit{
		Interfaces: rest.IFPutUnitInterfaces{
			ThermostatInterface: &rest.IFThermostatOverview{
				Boost: &rest.HelperSpecialModeThermostat{
					Enabled: &enabled,
					EndTime: &endTime,
				},
			},
		},
	}
	return rest.PutOverviewUnit(h.client, h.uid, data)
}

// DeactivateBoost deactivates boost mode.
func (h *ThermostatHandle) DeactivateBoost() error {
	enabled := false
	data := &rest.EndpointOverviewPutUnit{
		Interfaces: rest.IFPutUnitInterfaces{
			ThermostatInterface: &rest.IFThermostatOverview{
				Boost: &rest.HelperSpecialModeThermostat{
					Enabled: &enabled,
				},
			},
		},
	}
	return rest.PutOverviewUnit(h.client, h.uid, data)
}

// SetWindowOpen activates window-open mode for the given duration in minutes.
func (h *ThermostatHandle) SetWindowOpen(minutes int) error {
	endTime := int(time.Now().Add(time.Duration(minutes) * time.Minute).Unix())
	enabled := true
	data := &rest.EndpointOverviewPutUnit{
		Interfaces: rest.IFPutUnitInterfaces{
			ThermostatInterface: &rest.IFThermostatOverview{
				WindowOpenMode: &rest.HelperSpecialModeThermostat{
					Enabled: &enabled,
					EndTime: &endTime,
				},
			},
		},
	}
	return rest.PutOverviewUnit(h.client, h.uid, data)
}

// DeactivateWindowOpen deactivates window-open mode.
func (h *ThermostatHandle) DeactivateWindowOpen() error {
	enabled := false
	data := &rest.EndpointOverviewPutUnit{
		Interfaces: rest.IFPutUnitInterfaces{
			ThermostatInterface: &rest.IFThermostatOverview{
				WindowOpenMode: &rest.HelperSpecialModeThermostat{
					Enabled: &enabled,
				},
			},
		},
	}
	return rest.PutOverviewUnit(h.client, h.uid, data)
}

// SetComfortTemperature sets the "comfort" preset temperature.
func (h *ThermostatHandle) SetComfortTemperature(celsius float64) error {
	cel := float32(celsius)
	data := &rest.EndpointConfigurationPutUnit{
		Interfaces: &rest.IFUnitInterfacesConfig{
			ThermostatInterface: &rest.IFThermostatConfig{
				ComfortTemperature: &rest.HelperTemperature{
					Celsius: &cel,
					Mode:    rest.HelperTemperatureModeTemperature,
				},
			},
		},
	}
	return rest.PutConfigurationUnitByUID(h.client, h.uid, data)
}

// SetReducedTemperature sets the "reduced" preset temperature.
func (h *ThermostatHandle) SetReducedTemperature(celsius float64) error {
	cel := float32(celsius)
	data := &rest.EndpointConfigurationPutUnit{
		Interfaces: &rest.IFUnitInterfacesConfig{
			ThermostatInterface: &rest.IFThermostatConfig{
				ReducedTemperature: &rest.HelperTemperature{
					Celsius: &cel,
					Mode:    rest.HelperTemperatureModeTemperature,
				},
			},
		},
	}
	return rest.PutConfigurationUnitByUID(h.client, h.uid, data)
}

// SetTemperatureOffset sets the temperature sensor offset (-10 to +10°C).
func (h *ThermostatHandle) SetTemperatureOffset(offset float64) error {
	off := float32(offset)
	data := &rest.EndpointConfigurationPutUnit{
		Interfaces: &rest.IFUnitInterfacesConfig{
			ThermostatInterface: &rest.IFThermostatConfig{
				TemperatureOffset: &rest.HelperTemperatureOffset{
					InternalOffset: &off,
					SensorMode:     rest.HelperTemperatureOffsetSensorModeInternal,
				},
			},
		},
	}
	return rest.PutConfigurationUnitByUID(h.client, h.uid, data)
}

// SetAdaptiveHeating enables or disables adaptive heating mode.
func (h *ThermostatHandle) SetAdaptiveHeating(enabled bool) error {
	data := &rest.EndpointConfigurationPutUnit{
		Interfaces: &rest.IFUnitInterfacesConfig{
			ThermostatInterface: &rest.IFThermostatConfig{
				AdaptiveHeatingModeEnabled: &enabled,
			},
		},
	}
	return rest.PutConfigurationUnitByUID(h.client, h.uid, data)
}

// SetLocks sets device button lock and API lock.
func (h *ThermostatHandle) SetLocks(localLock, apiLock bool) error {
	data := &rest.EndpointConfigurationPutUnit{
		Interfaces: &rest.IFUnitInterfacesConfig{
			ThermostatInterface: &rest.IFThermostatConfig{
				LockedDeviceLocalEnabled: &localLock,
				LockedDeviceApiEnabled:   &apiLock,
			},
		},
	}
	return rest.PutConfigurationUnitByUID(h.client, h.uid, data)
}

// SetWindowOpenDetection configures automatic window detection.
// sensitivity: "low", "medium", "high". duration: 1-120 minutes.
func (h *ThermostatHandle) SetWindowOpenDetection(duration int, sensitivity string) error {
	sensorMode := rest.HelperWindowOpenModeConfigSensorModeInternal
	sens := rest.HelperWindowOpenModeConfigInternalSensitivity(sensitivity)
	data := &rest.EndpointConfigurationPutUnit{
		Interfaces: &rest.IFUnitInterfacesConfig{
			ThermostatInterface: &rest.IFThermostatConfig{
				WindowOpenMode: &rest.HelperWindowOpenModeConfig{
					InternalDuration:    &duration,
					InternalSensitivity: &sens,
					SensorMode:          &sensorMode,
				},
			},
		},
	}
	return rest.PutConfigurationUnitByUID(h.client, h.uid, data)
}

// SetWeeklyTimer sets the complete weekly schedule.
func (h *ThermostatHandle) SetWeeklyTimer(entries []rest.HelperBaseTimerWeeklyObj) error {
	timerMode := rest.HelperUnitTimerTimerModeWeekly
	weekly := rest.HelperBaseTimerWeekly(entries)
	data := &rest.EndpointConfigurationPutUnit{
		Timer: &rest.HelperUnitTimer{
			TimerMode: &timerMode,
			Weekly:    &weekly,
		},
	}
	return rest.PutConfigurationUnitByUID(h.client, h.uid, data)
}

// SetSummerPeriod configures summer mode dates using month/day values.
func (h *ThermostatHandle) SetSummerPeriod(enabled bool, startMonth, startDay, endMonth, endDay int) error {
	startMinutes := int(DateToYearMinutes(startMonth, startDay))
	endMinutes := int(DateToYearMinutes(endMonth, endDay))
	data := &rest.EndpointConfigurationPutUnit{
		Interfaces: &rest.IFUnitInterfacesConfig{
			ThermostatInterface: &rest.IFThermostatConfig{
				SummerPeriod: &rest.HelperSummerPeriod{
					Enabled:   enabled,
					StartTime: &startMinutes,
					EndTime:   &endMinutes,
				},
			},
		},
	}
	return rest.PutConfigurationUnitByUID(h.client, h.uid, data)
}

// AddHoliday adds a holiday period with the specified temperature.
func (h *ThermostatHandle) AddHoliday(start, end time.Time, celsius float64) error {
	startMinutes := int(MinutesFromYearStart(start))
	endMinutes := int(MinutesFromYearStart(end))

	config, err := rest.GetConfigurationUnitByUID(h.client, h.uid)
	if err != nil {
		return fmt.Errorf("get current config: %w", err)
	}

	if config.Interfaces.ThermostatInterface == nil {
		return fmt.Errorf("unit %s does not have a thermostat interface", h.uid)
	}

	var periods []rest.HelperPeriodHolidayRange
	if config.Interfaces.ThermostatInterface.HolidayPeriods != nil &&
		config.Interfaces.ThermostatInterface.HolidayPeriods.Periods != nil {
		periods = *config.Interfaces.ThermostatInterface.HolidayPeriods.Periods
	}

	deleteAfter := true
	periods = append(periods, rest.HelperPeriodHolidayRange{
		StartTime:            startMinutes,
		EndTime:              endMinutes,
		DeleteAfterEndActive: &deleteAfter,
	})

	return h.setHolidayPeriods(periods, &celsius)
}

// RemoveHoliday removes a holiday period by index.
func (h *ThermostatHandle) RemoveHoliday(index int) error {
	config, err := rest.GetConfigurationUnitByUID(h.client, h.uid)
	if err != nil {
		return fmt.Errorf("get current config: %w", err)
	}

	if config.Interfaces.ThermostatInterface == nil ||
		config.Interfaces.ThermostatInterface.HolidayPeriods == nil ||
		config.Interfaces.ThermostatInterface.HolidayPeriods.Periods == nil {
		return fmt.Errorf("no holiday periods configured")
	}

	periods := *config.Interfaces.ThermostatInterface.HolidayPeriods.Periods
	if index < 0 || index >= len(periods) {
		return fmt.Errorf("index out of range: %d (have %d holidays)", index, len(periods))
	}

	periods = append(periods[:index], periods[index+1:]...)

	var temp *float64
	if config.Interfaces.ThermostatInterface.HolidayPeriods.Temperature != nil &&
		config.Interfaces.ThermostatInterface.HolidayPeriods.Temperature.Celsius != nil {
		t := float64(*config.Interfaces.ThermostatInterface.HolidayPeriods.Temperature.Celsius)
		temp = &t
	}
	return h.setHolidayPeriods(periods, temp)
}

// ClearHolidays removes all holiday periods.
func (h *ThermostatHandle) ClearHolidays() error {
	return h.setHolidayPeriods(nil, nil)
}

func (h *ThermostatHandle) setHolidayPeriods(periods []rest.HelperPeriodHolidayRange, celsius *float64) error {
	if periods == nil {
		periods = []rest.HelperPeriodHolidayRange{}
	}

	holidayPeriods := &rest.HelperHolidayPeriods{
		Periods: &periods,
	}

	// Only set temperature if provided and there are periods
	if len(periods) > 0 && celsius != nil {
		cel := float32(*celsius)
		holidayPeriods.Temperature = &rest.HelperTemperature{
			Celsius: &cel,
			Mode:    rest.HelperTemperatureModeTemperature,
		}
	}

	data := &rest.EndpointConfigurationPutUnit{
		Interfaces: &rest.IFUnitInterfacesConfig{
			ThermostatInterface: &rest.IFThermostatConfig{
				HolidayPeriods: holidayPeriods,
			},
		},
	}
	return rest.PutConfigurationUnitByUID(h.client, h.uid, data)
}
