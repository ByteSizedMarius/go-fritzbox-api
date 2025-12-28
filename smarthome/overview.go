package smarthome

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ByteSizedMarius/go-fritzbox-api"
)

const (
	overviewPath     = "api/v0/smarthome/overview"
	overviewUnitPath = "api/v0/smarthome/overview/units/"
	configUnitPath   = "api/v0/smarthome/configuration/units/"
)

// GetOverview fetches the full smart home state.
func GetOverview(c *fritzbox.Client) (*SmarthomeOverview, error) {
	body, status, err := c.RestGet(overviewPath)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(body))
	}

	var overview SmarthomeOverview
	if err := json.Unmarshal(body, &overview); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return &overview, nil
}

// GetUnit fetches a specific unit's overview by UID.
func GetUnit(c *fritzbox.Client, uid string) (*SmarthomeUnit, error) {
	body, status, err := c.RestGet(overviewUnitPath + uid)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(body))
	}

	var unit SmarthomeUnit
	if err := json.Unmarshal(body, &unit); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return &unit, nil
}

// GetUnitConfig fetches unit configuration as raw map.
func GetUnitConfig(c *fritzbox.Client, uid string) (map[string]any, error) {
	body, status, err := c.RestGet(configUnitPath + uid)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(body))
	}

	var config map[string]any
	if err := json.Unmarshal(body, &config); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return config, nil
}

// FindDeviceByName searches for a device by name.
func FindDeviceByName(c *fritzbox.Client, name string) (*SmarthomeDevice, error) {
	overview, err := GetOverview(c)
	if err != nil {
		return nil, err
	}

	for i := range overview.Devices {
		if overview.Devices[i].Name == name {
			return &overview.Devices[i], nil
		}
	}
	return nil, fmt.Errorf("device not found: %s", name)
}