// Window detector support based on contribution by @btwotch
// https://github.com/ByteSizedMarius/go-fritzbox-api/pull/1

package smart

import (
	"fmt"

	"github.com/ByteSizedMarius/go-fritzbox-api/v2/rest"
)

// WindowDetectorConfig contains configuration data from the configuration endpoint.
type WindowDetectorConfig struct {
	ThermostatDestinationUIDs []string
	AvailableThermostats      []string
	ControlMode               string
}

// Get fetches the current window detector state from the overview endpoint.
func (h *WindowDetectorHandle) Get() (*WindowDetector, error) {
	return GetWindowDetector(h.client, h.uid)
}

// GetConfig fetches the window detector configuration including thermostat links.
func (h *WindowDetectorHandle) GetConfig() (*WindowDetectorConfig, error) {
	config, err := rest.GetConfigurationUnitByUID(h.client, h.uid)
	if err != nil {
		return nil, err
	}

	result := &WindowDetectorConfig{}

	if alert := config.Interfaces.AlertInterface; alert != nil {
		if uids, exists := (*alert)["thermostatDestinationUids"]; exists {
			if arr, ok := uids.([]interface{}); ok {
				for _, v := range arr {
					if s, ok := v.(string); ok {
						result.ThermostatDestinationUIDs = append(result.ThermostatDestinationUIDs, s)
					}
				}
			}
		}
		if uids, exists := (*alert)["availableThermostats"]; exists {
			if arr, ok := uids.([]interface{}); ok {
				for _, v := range arr {
					if s, ok := v.(string); ok {
						result.AvailableThermostats = append(result.AvailableThermostats, s)
					}
				}
			}
		}
		if mode, exists := (*alert)["controlMode"]; exists {
			if s, ok := mode.(string); ok {
				result.ControlMode = s
			}
		}
	}

	return result, nil
}

// AddThermostat links a thermostat to this window detector.
// When the window opens, the thermostat's windowOpenMode will be activated.
func (h *WindowDetectorHandle) AddThermostat(thermostatUID string) error {
	config, err := h.GetConfig()
	if err != nil {
		return fmt.Errorf("get current config: %w", err)
	}

	for _, uid := range config.ThermostatDestinationUIDs {
		if uid == thermostatUID {
			return nil
		}
	}

	uids := append(config.ThermostatDestinationUIDs, thermostatUID)
	return h.setThermostatDestinations(uids)
}

// RemoveThermostat unlinks a thermostat from this window detector.
func (h *WindowDetectorHandle) RemoveThermostat(thermostatUID string) error {
	config, err := h.GetConfig()
	if err != nil {
		return fmt.Errorf("get current config: %w", err)
	}

	var uids []string
	for _, uid := range config.ThermostatDestinationUIDs {
		if uid != thermostatUID {
			uids = append(uids, uid)
		}
	}

	return h.setThermostatDestinations(uids)
}

func (h *WindowDetectorHandle) setThermostatDestinations(uids []string) error {
	data := &rest.EndpointConfigurationPutUnit{
		Interfaces: &rest.IFUnitInterfacesConfig{
			AlertInterface: &rest.IFAlertConfig{
				"thermostatDestinationUids": uids,
			},
		},
	}
	return rest.PutConfigurationUnitByUID(h.client, h.uid, data)
}
