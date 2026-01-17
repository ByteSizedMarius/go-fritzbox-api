package aha

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ByteSizedMarius/go-fritzbox-api/v2"
)

// Device is the main type for aha-devices. It holds the values all devices have;
// all other properties are handled via the respective capabilities.
type Device struct {
	Identifier   string
	ID           string
	Fwversion    string
	Manufacturer string
	Productname  string
	Txbusy       string
	Name         string
	Present      string
	Capabilities Capabilities
}

// Capabilities is a map of the capabilities available for the device.
// They can be access using the Capability-Constants (starting with C, for example CHKR -> HeizungsKörperRegler, etc.)
// HasCapability can be used to check whether a device has a certain capability, without checking the map-keys.
type Capabilities map[string]Capability

// GetCapability returns the capability of the given type for the device.
func GetCapability[C Capability](d Device) (r C) {
	for _, capab := range d.Capabilities {
		if c, ok := capab.(C); ok {
			return c
		}
	}
	return
}

// String returns a string representation of the capabilities
func (c Capabilities) String() string {
	var builder strings.Builder
	builder.WriteString("[")
	for _, cp := range c {
		if cp != nil {
			_, _ = fmt.Fprintf(&builder, "%s, ", cp.Name())
		}
	}
	result := strings.TrimSuffix(builder.String(), ", ")
	return result + "]"
}

// HasCapability returns true, if device has given capability. Use capability-constants.
func (d *Device) HasCapability(cap string) bool {
	_, ok := d.Capabilities[cap]
	return ok
}

func (d *Device) String() string {
	var sb strings.Builder

	_, _ = fmt.Fprintf(
		&sb,
		"{Devicename: %s, Identifier: %s, ID: %s, Productname: %s, Manufacturer: %s, Firmware-Version: %s, Present: %s, TX busy: %s",
		d.Name, d.Identifier, d.ID, d.Productname, d.Manufacturer, d.Fwversion, d.Present, d.Txbusy,
	)
	if len(d.Capabilities) > 0 {
		_, _ = fmt.Fprintf(&sb, ", Capabilities: %v", d.Capabilities)
	}
	sb.WriteString("}")

	return sb.String()
}

// DECTGetName fetches device-Name from the fritzbox and updates internally stored value.
func (d *Device) DECTGetName(c *fritzbox.Client) (response string, err error) {
	data := fritzbox.Values{
		"sid":       c.SID(),
		"ain":       d.Identifier,
		"switchcmd": "getswitchname",
	}

	_, response, err = c.AhaRequestString(http.MethodGet, "webservices/homeautoswitch.lua", data)
	if err != nil {
		return
	}

	d.Name = response
	return
}

// DECTSetName updates device Name based on identifier and updates internal values if successful.
func (d *Device) DECTSetName(c *fritzbox.Client, name string) error {
	data := fritzbox.Values{
		"sid":       c.SID(),
		"ain":       d.Identifier,
		"switchcmd": "setname",
		"name":      name,
	}

	_, resp, err := c.AhaRequestString(http.MethodGet, "webservices/homeautoswitch.lua", data)
	if err != nil {
		return err
	}

	d.Name = resp
	return nil
}

// DECTIsSwitchPresent fetches the connection status from the fritzbox and updates the objects internal value.
// Note: According to the documentation, it may take multiple minutes for the status to update after a device disconnects.
func (d *Device) DECTIsSwitchPresent(c *fritzbox.Client) (bool, error) {
	data := fritzbox.Values{
		"sid":       c.SID(),
		"ain":       d.Identifier,
		"switchcmd": "getswitchpresent",
	}

	_, resp, err := c.AhaRequestString(http.MethodGet, "webservices/homeautoswitch.lua", data)
	if err != nil {
		return false, err
	}

	if resp == "1" || resp == "0" {
		d.Present = resp
		return resp == "1", nil
	}
	return false, fmt.Errorf("invalid response: %s", resp)
}

// GetDeviceInfosRaw fetches the device information from the fritzbox and returns the raw response converted to json.
func GetDeviceInfosRaw(c *fritzbox.Client, identifier string) (res json.RawMessage, err error) {
	data := fritzbox.Values{
		"sid":       c.SID(),
		"switchcmd": "getdeviceinfos",
		"ain":       identifier,
	}

	resp, err := c.AhaRequest(http.MethodGet, "webservices/homeautoswitch.lua", data)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = unmarshalKey(bytes, "device", &res)
	return
}

// GetDeviceInfos fetches the device information from the fritzbox and unmarshals the response into the given Capability.
// Note: For HanFun devices, this only fetches the parent device. HanFun units (child devices) are not populated.
// Use GetDeviceList to get HanFun devices with their units.
func GetDeviceInfos(c *fritzbox.Client, identifier string, dest Capability) (err error) {
	raw, err := GetDeviceInfosRaw(c, identifier)
	if err != nil {
		return
	}

	var dRe dRepr
	err = json.Unmarshal(raw, &dRe)
	if err != nil {
		return
	}
	d := toDevice(dRe)

	var rawMap map[string]json.RawMessage
	err = json.Unmarshal(raw, &rawMap)
	if err != nil {
		return
	}
	_, err = dest.fromJSON(rawMap, &d)
	return err
}

// IsSwitchPresent returns true if device is present.
func (d *Device) IsSwitchPresent() bool {
	return d.Present == "1"
}
