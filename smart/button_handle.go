package smart

import (
	"fmt"
	"time"

	"github.com/ByteSizedMarius/go-fritzbox-api/v2/rest"
)

// Get fetches the current button state from the overview endpoint.
func (h *ButtonHandle) Get() (*Button, error) {
	return GetButton(h.client, h.uid)
}

// GetConfig fetches the button configuration including destinations and active period.
func (h *ButtonHandle) GetConfig() (*ButtonConfig, error) {
	config, err := rest.GetConfigurationUnitByUID(h.client, h.uid)
	if err != nil {
		return nil, err
	}

	result := &ButtonConfig{}

	if bi := config.Interfaces.ButtonInterface; bi != nil {
		if bi.ControlMode != nil {
			result.ControlMode = string(*bi.ControlMode)
		}
		if bi.DestinationMode != nil {
			result.DestinationMode = string(*bi.DestinationMode)
		}
		if bi.DestinationUids != nil {
			result.DestinationUids = *bi.DestinationUids
		}
		if bi.AvailableUnits != nil {
			result.AvailableUnits = *bi.AvailableUnits
		}
		if bi.AvailableTemplates != nil {
			result.AvailableTemplates = *bi.AvailableTemplates
		}

		if bi.SwitchDuration != nil {
			result.SwitchDuration = &SwitchDuration{
				Mode: string(bi.SwitchDuration.Mode),
			}
			if bi.SwitchDuration.ToggleBackTime != nil {
				result.SwitchDuration.ToggleBackTime = *bi.SwitchDuration.ToggleBackTime
			}
		}

		if bi.ActivePeriod != nil {
			result.ActivePeriod = &ButtonActivePeriod{}
			if bi.ActivePeriod.Mode != nil {
				result.ActivePeriod.Mode = string(*bi.ActivePeriod.Mode)
			}
			if fp := bi.ActivePeriod.FixedActivePeriod; fp != nil {
				if fp.StartDate != nil && fp.StartTimePerDay != nil {
					result.ActivePeriod.StartTime = combineDateTime(*fp.StartDate, *fp.StartTimePerDay)
				}
				if fp.EndDate != nil && fp.EndTimePerDay != nil {
					result.ActivePeriod.EndTime = combineDateTime(*fp.EndDate, *fp.EndTimePerDay)
				}
			}
		}
	}

	return result, nil
}

// combineDateTime combines a unix date timestamp with minutes-from-midnight
func combineDateTime(dateUnix, minutesFromMidnight int) time.Time {
	date := time.Unix(int64(dateUnix), 0).UTC()
	return time.Date(date.Year(), date.Month(), date.Day(),
		minutesFromMidnight/60, minutesFromMidnight%60, 0, 0, time.UTC)
}

// SetControlMode sets what happens when the button is pressed.
// mode: "on", "off", "toggle", or "unknown" (event only)
func (h *ButtonHandle) SetControlMode(mode string) error {
	m := rest.HelperControlMode(mode)
	return h.putConfig(&rest.IFButtonConfig{
		ControlMode: &m,
	})
}

// SetSwitchDuration configures how long a switch action lasts.
// mode: "permanent" or "toggleBack"
// toggleBackMinutes: time in minutes before switching back (only for toggleBack mode)
func (h *ButtonHandle) SetSwitchDuration(mode string, toggleBackMinutes int) error {
	sd := &rest.HelperSwitchDuration{
		Mode: rest.HelperSwitchDurationMode(mode),
	}
	if mode == "toggleBack" {
		sd.ToggleBackTime = &toggleBackMinutes
	}
	return h.putConfig(&rest.IFButtonConfig{
		SwitchDuration: sd,
	})
}

// SetDestinations configures which units or templates the button controls.
// mode: "disabled", "units", or "templates"
// uids: list of unit UIDs (for "units" mode) or template UIDs (for "templates" mode)
func (h *ButtonHandle) SetDestinations(mode string, uids []string) error {
	m := rest.HelperDestinationMode(mode)
	return h.putConfig(&rest.IFButtonConfig{
		DestinationMode: &m,
		DestinationUids: &uids,
	})
}

// SetActivePeriodPermanent sets button to always register presses.
func (h *ButtonHandle) SetActivePeriodPermanent() error {
	mode := rest.HelperActivePeriodAlertButtonMode("permanent")
	return h.putConfig(&rest.IFButtonConfig{
		ActivePeriod: &rest.HelperActivePeriodAlertButton{
			Mode: &mode,
		},
	})
}

// SetActivePeriodFixed sets a fixed time window when button presses are registered.
func (h *ButtonHandle) SetActivePeriodFixed(startTime, endTime time.Time) error {
	mode := rest.HelperActivePeriodAlertButtonMode("fixed")
	startDate := int(time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, time.UTC).Unix())
	endDate := int(time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 0, 0, 0, 0, time.UTC).Unix())
	startMinutes := startTime.Hour()*60 + startTime.Minute()
	endMinutes := endTime.Hour()*60 + endTime.Minute()

	return h.putConfig(&rest.IFButtonConfig{
		ActivePeriod: &rest.HelperActivePeriodAlertButton{
			Mode: &mode,
			FixedActivePeriod: &struct {
				EndDate         *int `json:"endDate,omitempty"`
				EndTimePerDay   *int `json:"endTimePerDay,omitempty"`
				StartDate       *int `json:"startDate,omitempty"`
				StartTimePerDay *int `json:"startTimePerDay,omitempty"`
			}{
				StartDate:       &startDate,
				EndDate:         &endDate,
				StartTimePerDay: &startMinutes,
				EndTimePerDay:   &endMinutes,
			},
		},
	})
}

// SetActivePeriodAstronomic sets button active period based on sunrise/sunset.
func (h *ButtonHandle) SetActivePeriodAstronomic() error {
	mode := rest.HelperActivePeriodAlertButtonMode("astronomic")
	return h.putConfig(&rest.IFButtonConfig{
		ActivePeriod: &rest.HelperActivePeriodAlertButton{
			Mode: &mode,
		},
	})
}

func (h *ButtonHandle) putConfig(cfg *rest.IFButtonConfig) error {
	data := &rest.EndpointConfigurationPutUnit{
		Interfaces: &rest.IFUnitInterfacesConfig{
			ButtonInterface: cfg,
		},
	}
	err := rest.PutConfigurationUnitByUID(h.client, h.uid, data)
	if err != nil {
		return fmt.Errorf("put button config: %w", err)
	}
	return nil
}
