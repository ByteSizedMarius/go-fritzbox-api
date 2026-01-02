package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/ByteSizedMarius/go-fritzbox-api"
)

const (
	configDevicesPath              = "api/v0/smarthome/configuration/devices"
	configUnitsPath                = "api/v0/smarthome/configuration/units"
	configGroupsPath               = "api/v0/smarthome/configuration/groups"
	configTemplatesPath            = "api/v0/smarthome/configuration/templates"
	configTemplateCapabilitiesPath = "api/v0/smarthome/configuration/templateCapabilities"
)

// GetConfigurationDeviceByUID returns device configuration by UID.
//
// Provides extended information about physical devices including battery status,
// connection status, firmware version, and push notification settings.
// UID may be IPUI (International Portable User Identity), MACA or Zigbee Identifier.
func GetConfigurationDeviceByUID(c *fritzbox.Client, uid string) (*EndpointConfigurationDevice, error) {
	body, status, err := c.RestGet(fmt.Sprintf("%s/%s", configDevicesPath, uid))
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(body))
	}

	var result EndpointConfigurationDevice
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return &result, nil
}

// PutConfigurationDeviceByUID updates device configuration.
func PutConfigurationDeviceByUID(c *fritzbox.Client, uid string, data *EndpointConfigurationPutDevice) error {
	body, status, err := c.RestPut(fmt.Sprintf("%s/%s", configDevicesPath, uid), data)
	if err != nil {
		return err
	}
	if status != http.StatusOK && status != http.StatusNoContent {
		return fmt.Errorf("unexpected status %d: %s", status, string(body))
	}
	return nil
}

// DeleteConfigurationDeviceByUID deletes a device from smart home and unpairs it.
//
// Deletion is usually always allowed for local devices.
// Deletion of devices on remote radioBases is only allowed on the smart home master.
// Active smartmeters cannot be deleted.
func DeleteConfigurationDeviceByUID(c *fritzbox.Client, uid string) error {
	body, status, err := c.RestDelete(fmt.Sprintf("%s/%s", configDevicesPath, uid))
	if err != nil {
		return err
	}
	if status != http.StatusOK && status != http.StatusNoContent {
		return fmt.Errorf("unexpected status %d: %s", status, string(body))
	}
	return nil
}

// GetConfigurationUnitByUID returns unit configuration by UID.
//
// Provides extended configuration for units including all interface settings,
// timer configurations, holiday periods, and other detailed settings.
func GetConfigurationUnitByUID(c *fritzbox.Client, uid string) (*EndpointConfigurationUnit, error) {
	body, status, err := c.RestGet(fmt.Sprintf("%s/%s", configUnitsPath, uid))
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(body))
	}

	var result EndpointConfigurationUnit
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return &result, nil
}

// PutConfigurationUnitByUID updates unit configuration.
//
// Used to configure unit settings like thermostat schedules, holiday periods,
// temperature presets, and other detailed configuration options.
func PutConfigurationUnitByUID(c *fritzbox.Client, uid string, data *EndpointConfigurationPutUnit) error {
	body, status, err := c.RestPut(fmt.Sprintf("%s/%s", configUnitsPath, uid), data)
	if err != nil {
		return err
	}
	if status != http.StatusOK && status != http.StatusNoContent {
		return fmt.Errorf("unexpected status %d: %s", status, string(body))
	}
	return nil
}

// PostConfigurationGroup creates a new group.
//
// Groups allow a collection of units to be controlled at once.
// Controlling a group through its corresponding unit controls all member units
// and overwrites their respective status.
func PostConfigurationGroup(c *fritzbox.Client, name string, data *EndpointConfigurationPostGroup) (*CreateGroupResponse, error) {
	path := configGroupsPath + "?name=" + url.QueryEscape(name)
	body, status, err := c.RestPost(path, data)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(body))
	}

	var result CreateGroupResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return &result, nil
}

// GetConfigurationGroupByUID returns group configuration by UID.
func GetConfigurationGroupByUID(c *fritzbox.Client, uid string) (*EndpointConfigurationGroup, error) {
	body, status, err := c.RestGet(fmt.Sprintf("%s/%s", configGroupsPath, uid))
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(body))
	}

	var result EndpointConfigurationGroup
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return &result, nil
}

// PutConfigurationGroupByUID updates group configuration.
func PutConfigurationGroupByUID(c *fritzbox.Client, uid string, data *EndpointConfigurationPutGroup) error {
	body, status, err := c.RestPut(fmt.Sprintf("%s/%s", configGroupsPath, uid), data)
	if err != nil {
		return err
	}
	if status != http.StatusOK && status != http.StatusNoContent {
		return fmt.Errorf("unexpected status %d: %s", status, string(body))
	}
	return nil
}

// DeleteConfigurationGroupByUID deletes a group from smart home.
//
// Deletion is usually always allowed for local groups.
func DeleteConfigurationGroupByUID(c *fritzbox.Client, uid string) error {
	body, status, err := c.RestDelete(fmt.Sprintf("%s/%s", configGroupsPath, uid))
	if err != nil {
		return err
	}
	if status != http.StatusOK && status != http.StatusNoContent {
		return fmt.Errorf("unexpected status %d: %s", status, string(body))
	}
	return nil
}

// PostConfigurationTemplate creates a new template.
//
// Templates allow saving and recalling configuration for units.
// Either the template object or scenario object is required.
func PostConfigurationTemplate(c *fritzbox.Client, name string, data *EndpointConfigurationPostTemplate) (*CreateTemplateResponse, error) {
	path := configTemplatesPath + "?name=" + url.QueryEscape(name)
	body, status, err := c.RestPost(path, data)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(body))
	}

	var result CreateTemplateResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return &result, nil
}

// GetConfigurationTemplateByUID returns template configuration by UID.
//
// Templates allow saving and recalling configuration for units.
// Scenarios are collections of templates that can apply multiple templates at once.
func GetConfigurationTemplateByUID(c *fritzbox.Client, uid string) (*EndpointConfigurationGetTemplate, error) {
	body, status, err := c.RestGet(fmt.Sprintf("%s/%s", configTemplatesPath, uid))
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(body))
	}

	var result EndpointConfigurationGetTemplate
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return &result, nil
}

// PutConfigurationTemplateByUID updates template configuration.
func PutConfigurationTemplateByUID(c *fritzbox.Client, uid string, data *EndpointConfigurationPutTemplate) error {
	body, status, err := c.RestPut(fmt.Sprintf("%s/%s", configTemplatesPath, uid), data)
	if err != nil {
		return err
	}
	if status != http.StatusOK && status != http.StatusNoContent {
		return fmt.Errorf("unexpected status %d: %s", status, string(body))
	}
	return nil
}

// DeleteConfigurationTemplateByUID deletes a template from smart home.
func DeleteConfigurationTemplateByUID(c *fritzbox.Client, uid string) error {
	body, status, err := c.RestDelete(fmt.Sprintf("%s/%s", configTemplatesPath, uid))
	if err != nil {
		return err
	}
	if status != http.StatusOK && status != http.StatusNoContent {
		return fmt.Errorf("unexpected status %d: %s", status, string(body))
	}
	return nil
}

// GetConfigurationTemplateCapabilities returns possible template configuration capabilities.
//
// Lists available template types and which units support them.
func GetConfigurationTemplateCapabilities(c *fritzbox.Client) (*EndpointConfigurationGetTemplateCapabilities, error) {
	body, status, err := c.RestGet(configTemplateCapabilitiesPath)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(body))
	}

	var result EndpointConfigurationGetTemplateCapabilities
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return &result, nil
}
