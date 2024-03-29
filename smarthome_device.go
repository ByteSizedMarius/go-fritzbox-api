package go_fritzbox_api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// SmarthomeDevice is the main type for fritz-smarthome-devices. It only holds values all devices have, all other properties are handled in their respective capabilities.
type SmarthomeDevice struct {
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

func GetCapability[C Capability](d SmarthomeDevice) (r C) {
	for _, c := range d.Capabilities {
		switch c.(type) {
		case C:
			return c.(C)
		}
	}
	return
}

func (d *SmarthomeDevice) Reload(cl *Client) error {
	var data map[string]json.RawMessage
	err := cl.GetDeviceInfos(d.Identifier, &data)
	if err != nil {
		return err
	}

	var ed extDevice
	err = json.Unmarshal(data["device"], &ed)
	if err != nil {
		return err
	}
	var nd SmarthomeDevice
	nd = nd.fromDevice(ed)

	var devicedata map[string]json.RawMessage
	err = json.Unmarshal(data["device"], &devicedata)
	if err != nil {
		return err
	}

	for k, v := range d.Capabilities {
		var c Capability
		c, err = v.fromJSON(devicedata, d)
		if err != nil {
			fmt.Println(err)
			continue
		}

		nd.Capabilities[k] = c
	}

	d.Identifier = nd.Identifier
	d.ID = nd.ID
	d.Fwversion = nd.Fwversion
	d.Manufacturer = nd.Manufacturer
	d.Productname = nd.Productname
	d.Txbusy = nd.Txbusy
	d.Name = nd.Name
	d.Present = nd.Present
	d.Capabilities = nd.Capabilities
	return nil
}

func (d *SmarthomeDevice) String() string {
	t := fmt.Sprintf("{Devicename: %s, Identifier: %s, ID: %s, Productname: %s, Manufacturer: %s, Firmware-Version: %s, Present: %s, TX busy: %s", d.Name, d.Identifier, d.ID, d.Productname, d.Manufacturer, d.Fwversion, d.Present, d.Txbusy)
	if fmt.Sprint(d.Capabilities) != "[]" {
		t += fmt.Sprintf(", Capabilities: %s", d.Capabilities)
	}

	return t + "}"
}

// HasCapability returns true, if device has given capability. Use capability-constants.
func (d *SmarthomeDevice) HasCapability(cap string) bool {
	_, ok := d.Capabilities[cap]
	return ok
}

// DECTGetName fetches device-Name from the fritzbox and updates internally stored value.
func (d *SmarthomeDevice) DECTGetName(c *Client) (string, error) {
	data := url.Values{
		"sid":       {c.SID()},
		"ain":       {d.Identifier},
		"switchcmd": {"getswitchname"},
	}

	code, resp, err := c.CustomRequest(http.MethodGet, "webservices/homeautoswitch.lua", data)
	if err != nil {
		return "", err
	} else if code != 200 {
		return "", fmt.Errorf("unknown error: " + resp)
	}

	d.Name = resp
	return resp, nil
}

// DECTSetName updates device Name based on identifier and updates internal values if successful
func (d *SmarthomeDevice) DECTSetName(c *Client, name string) error {
	data := url.Values{
		"sid":       {c.SID()},
		"ain":       {d.Identifier},
		"switchcmd": {"setname"},
		"name":      {name},
	}

	code, resp, err := c.CustomRequest(http.MethodGet, "webservices/homeautoswitch.lua", data)
	if err != nil {
		return err
	} else if code != 200 {
		return fmt.Errorf("unknown error: " + resp)
	}

	d.Name = resp
	return nil
}

// DECTIsSwitchPrsent fetches current status from the fritzbox and updates internally stored value.
// Note: According to the documentation, it may take multiple minutes for the status to update after a device disconnected.
func (d *SmarthomeDevice) DECTIsSwitchPrsent(c *Client) (bool, error) {
	data := url.Values{
		"sid":       {c.SID()},
		"ain":       {d.Identifier},
		"switchcmd": {"getswitchpresent"},
	}

	code, resp, err := c.CustomRequest(http.MethodGet, "webservices/homeautoswitch.lua", data)
	if err != nil {
		return false, err
	} else if code != 200 {
		return false, fmt.Errorf("unknown error: " + resp)
	}

	if resp == "1" || resp == "0" {
		d.Present = resp
		return resp == "1", nil
	} else {
		return false, fmt.Errorf("invalid response: " + resp)
	}
}

// IsSwitchPrsent returns true if device is present. Uses locally stored value.
func (d *SmarthomeDevice) IsSwitchPrsent() bool {
	return d.Present == "1"
}

// fromDevice converts a extDevice to a SmarthomeDevice. This includes parsing the capabilities from the given Bitmask.
func (*SmarthomeDevice) fromDevice(d extDevice) SmarthomeDevice {
	cpb, err := Capabilities{}.fromBitmask(d.Functionbitmask)
	if err != nil {
		fmt.Println(err)
	}

	return SmarthomeDevice{
		Identifier:   d.Identifier,
		ID:           d.ID,
		Fwversion:    d.Fwversion,
		Manufacturer: d.Manufacturer,
		Productname:  d.Productname,
		Txbusy:       d.Txbusy,
		Name:         d.Name,
		Present:      d.Present,
		Capabilities: cpb,
	}
}
