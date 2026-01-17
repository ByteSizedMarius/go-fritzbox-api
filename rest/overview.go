// Package rest provides access to FRITZ!Box Smart Home features via the REST API.
// Requires FRITZ!OS 8.20 or later.
// API documentation: https://fritz.support/resources/SmarthomeRestApiFRITZOS82.html
package rest

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ByteSizedMarius/go-fritzbox-api/v2"
)

const (
	overviewPath          = "api/v0/smarthome/overview"
	overviewDevicesPath   = "api/v0/smarthome/overview/devices"
	overviewGroupsPath    = "api/v0/smarthome/overview/groups"
	overviewUnitsPath     = "api/v0/smarthome/overview/units"
	overviewTemplatesPath = "api/v0/smarthome/overview/templates"
	overviewTriggersPath  = "api/v0/smarthome/overview/triggers"
	overviewGlobalsPath   = "api/v0/smarthome/overview/globals"
)

// GetOverview returns all overview infos and lists.
//
// This is a collection of all components of the FRITZ!Box Smart Home.
// The overview provides a collection of basic information and control possibilities of all smart home entities:
//   - devices: physical device management (battery info, connection status, etc.)
//   - units: actuators & sensors for tracking and control of device functions
//   - groups: grouping of units for bulk control
//   - templates: saved configurations that can be applied to units
//   - triggers: if-then automations based on sensor values
//   - globals: values used throughout the smart home (colorPalettes, location)
func GetOverview(c *fritzbox.Client) (*EndpointOverview, error) {
	body, status, err := c.RestGet(overviewPath)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(body))
	}

	var result EndpointOverview
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return &result, nil
}

// GetOverviewDevicesList returns list of devices.
//
// Devices are physical devices (e.g. battery or connection status, etc.).
// Each device has at least one unit.
func GetOverviewDevicesList(c *fritzbox.Client) ([]HelperOverviewDevice, error) {
	body, status, err := c.RestGet(overviewDevicesPath)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(body))
	}

	var result []HelperOverviewDevice
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return result, nil
}

// GetOverviewDeviceByUID returns a device by UID.
//
// UID may be IPUI (International Portable User Identity), MACA or Zigbee Identifier.
func GetOverviewDeviceByUID(c *fritzbox.Client, uid string) (*HelperOverviewDevice, error) {
	body, status, err := c.RestGet(fmt.Sprintf("%s/%s", overviewDevicesPath, uid))
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(body))
	}

	var result HelperOverviewDevice
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return &result, nil
}

// GetOverviewGroupsList returns list of groups.
//
// Groups allow a collection of units to be controlled at once.
// Controlling a group through its corresponding unit controls all its member units
// at once and overwrites their respective status.
func GetOverviewGroupsList(c *fritzbox.Client) ([]EndpointOverviewGroup, error) {
	body, status, err := c.RestGet(overviewGroupsPath)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(body))
	}

	var result []EndpointOverviewGroup
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return result, nil
}

// GetOverviewGroupByUID returns a group by UID.
func GetOverviewGroupByUID(c *fritzbox.Client, uid string) (*EndpointOverviewGroup, error) {
	body, status, err := c.RestGet(fmt.Sprintf("%s/%s", overviewGroupsPath, uid))
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(body))
	}

	var result EndpointOverviewGroup
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return &result, nil
}

// GetOverviewUnitsList returns list of units.
//
// Units are actuators and sensors with their interfaces.
// These allow you to control device functions (e.g. turning a lightbulb on/off, change its color/level).
// UnitType indicates which interfaces are to be expected and helps classify the unit.
func GetOverviewUnitsList(c *fritzbox.Client) ([]HelperOverviewUnit, error) {
	body, status, err := c.RestGet(overviewUnitsPath)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(body))
	}

	var result []HelperOverviewUnit
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return result, nil
}

// GetOverviewUnitByUID returns a unit by UID.
func GetOverviewUnitByUID(c *fritzbox.Client, uid string) (*HelperOverviewUnit, error) {
	body, status, err := c.RestGet(fmt.Sprintf("%s/%s", overviewUnitsPath, uid))
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(body))
	}

	var result HelperOverviewUnit
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return &result, nil
}

// PutOverviewUnit updates a unit's interfaces.
//
// Used to control unit functions like turning a socket on/off, changing thermostat temperature, etc.
func PutOverviewUnit(c *fritzbox.Client, uid string, data *EndpointOverviewPutUnit) error {
	body, status, err := c.RestPut(fmt.Sprintf("%s/%s", overviewUnitsPath, uid), data)
	if err != nil {
		return err
	}
	if status != http.StatusOK && status != http.StatusNoContent {
		return fmt.Errorf("unexpected status %d: %s", status, string(body))
	}
	return nil
}

// GetOverviewTemplatesList returns list of templates.
//
// Templates store configuration snapshots that can be applied to units.
func GetOverviewTemplatesList(c *fritzbox.Client) ([]EndpointOverviewGetTemplate, error) {
	body, status, err := c.RestGet(overviewTemplatesPath)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(body))
	}

	var result []EndpointOverviewGetTemplate
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return result, nil
}

// GetOverviewTemplateByUID returns a template by UID.
func GetOverviewTemplateByUID(c *fritzbox.Client, uid string) (*EndpointOverviewGetTemplate, error) {
	body, status, err := c.RestGet(fmt.Sprintf("%s/%s", overviewTemplatesPath, uid))
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(body))
	}

	var result EndpointOverviewGetTemplate
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return &result, nil
}

// PostOverviewTemplate applies a template.
//
// Applies the template's stored configuration to its member units.
func PostOverviewTemplate(c *fritzbox.Client, uid string, data *EndpointOverviewPostTemplate) error {
	body, status, err := c.RestPost(fmt.Sprintf("%s/%s", overviewTemplatesPath, uid), data)
	if err != nil {
		return err
	}
	if status != http.StatusOK && status != http.StatusNoContent {
		return fmt.Errorf("unexpected status %d: %s", status, string(body))
	}
	return nil
}

// GetOverviewTriggersList returns list of triggers.
//
// Triggers are if-then automations based on sensor values.
func GetOverviewTriggersList(c *fritzbox.Client) ([]EndpointOverviewTrigger, error) {
	body, status, err := c.RestGet(overviewTriggersPath)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(body))
	}

	var result []EndpointOverviewTrigger
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return result, nil
}

// GetOverviewTriggerByUID returns a trigger by UID.
func GetOverviewTriggerByUID(c *fritzbox.Client, uid string) (*EndpointOverviewTrigger, error) {
	body, status, err := c.RestGet(fmt.Sprintf("%s/%s", overviewTriggersPath, uid))
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(body))
	}

	var result EndpointOverviewTrigger
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return &result, nil
}

// PutOverviewTrigger updates a trigger's enabled state.
func PutOverviewTrigger(c *fritzbox.Client, uid string, enabled bool) error {
	data := map[string]bool{"enabled": enabled}
	body, status, err := c.RestPut(fmt.Sprintf("%s/%s", overviewTriggersPath, uid), data)
	if err != nil {
		return err
	}
	if status != http.StatusOK && status != http.StatusNoContent {
		return fmt.Errorf("unexpected status %d: %s", status, string(body))
	}
	return nil
}

// Thermostats returns units with UnitType "avmThermostat" (physical thermostats only, excludes groups).
func (o *EndpointOverview) Thermostats() []HelperOverviewUnit {
	var result []HelperOverviewUnit
	for _, u := range o.Units {
		if u.UnitType == AvmThermostat {
			result = append(result, u)
		}
	}
	return result
}

// GetThermostats fetches overview and returns thermostat units.
func GetThermostats(c *fritzbox.Client) ([]HelperOverviewUnit, error) {
	overview, err := GetOverview(c)
	if err != nil {
		return nil, err
	}
	return overview.Thermostats(), nil
}

// Buttons returns units with button interface (simpleButton, avmButton, avmWidgetButton).
func (o *EndpointOverview) Buttons() []HelperOverviewUnit {
	var result []HelperOverviewUnit
	for _, u := range o.Units {
		if u.UnitType == SimpleButton || u.UnitType == AvmButton || u.UnitType == AvmWidgetButton {
			result = append(result, u)
		}
	}
	return result
}

// GetButtons fetches overview and returns button units.
func GetButtons(c *fritzbox.Client) ([]HelperOverviewUnit, error) {
	overview, err := GetOverview(c)
	if err != nil {
		return nil, err
	}
	return overview.Buttons(), nil
}

// WindowDetectors returns units with UnitType "windowOpenCloseDetector".
func (o *EndpointOverview) WindowDetectors() []HelperOverviewUnit {
	var result []HelperOverviewUnit
	for _, u := range o.Units {
		if u.UnitType == WindowOpenCloseDetector {
			result = append(result, u)
		}
	}
	return result
}

// GetWindowDetectors fetches overview and returns window detector units.
func GetWindowDetectors(c *fritzbox.Client) ([]HelperOverviewUnit, error) {
	overview, err := GetOverview(c)
	if err != nil {
		return nil, err
	}
	return overview.WindowDetectors(), nil
}

// GetOverviewGlobals returns global smart home settings.
//
// Includes color palettes, location coordinates, and energy key figures.
func GetOverviewGlobals(c *fritzbox.Client) (*HelperOverviewGlobals, error) {
	body, status, err := c.RestGet(overviewGlobalsPath)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(body))
	}

	var result HelperOverviewGlobals
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return &result, nil
}
