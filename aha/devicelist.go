package aha

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ByteSizedMarius/go-fritzbox-api"
	"github.com/clbanning/mxj"
)

type DeviceList struct {
	Version   string
	Fwversion string
	Devices   []Device
	filter    Capability
}

// GetDeviceList implements the getdevicelistinfos endpoint of the AHA-API and returns a DeviceList.
func GetDeviceList(c *fritzbox.Client) (*DeviceList, error) {
	data := fritzbox.Values{
		"sid":       c.SID(),
		"switchcmd": "getdevicelistinfos",
	}

	resp, err := c.AhaRequest(http.MethodGet, "webservices/homeautoswitch.lua", data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return deviceListFromReader(resp.Body)
}

func GetDeviceListFilter(c *fritzbox.Client, cap Capability) (dl *DeviceList, err error) {
	dl, err = GetDeviceList(c)
	if err != nil {
		return
	}
	dl.filter = cap
	dl.doFilter()
	return dl, nil
}

// String returns a string representation of the DeviceList.
func (dl *DeviceList) String() string {
	var sb strings.Builder
	for i, d := range dl.Devices {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(d.String())
	}
	return fmt.Sprintf("Version: %s, Firmware-Version: %s Devices: [%s]", dl.Version, dl.Fwversion, sb.String())
}

// deviceListFromReader creates a new DeviceList from the given reader.
func deviceListFromReader(r io.Reader) (dl *DeviceList, err error) {
	bytes, err := io.ReadAll(r)
	if err != nil {
		return
	}

	var dlR dlRepr
	err = unmarshalKey(bytes, "devicelist", &dlR)

	dl = toDeviceList(dlR)
	err = dl.populateCapabilities(bytes)
	if err != nil {
		return
	}

	return
}

func (dl *DeviceList) populateDeviceCapabilities(i int, currentDevice map[string]json.RawMessage, dsToParse []map[string]json.RawMessage) error {
	capabilities := []string{CHanfun, CButton, CHKR, CTempSensor}
	for _, cp := range capabilities {
		if dl.Devices[i].HasCapability(cp) {
			capab, err := dl.Devices[i].Capabilities[cp].fromJSON(currentDevice, &dl.Devices[i])
			if err != nil {
				return err
			}
			dl.Devices[i].Capabilities[cp] = capab
		}
	}

	if dl.Devices[i].HasCapability(CHanfun) {
		err := dl.populateHanfunUnits(i, dsToParse)
		if err != nil {
			return err
		}
	}

	return nil
}

// populateCapabilities is given the response from the getdevicelistinfos command and populates the capabilities of the devices.
func (dl *DeviceList) populateCapabilities(b []byte) (err error) {
	dlRaw, err := toDlReprRaw(b)
	if err != nil {
		return err
	}

	dsToParse, err := parseDevices(dlRaw)
	if err != nil {
		return err
	}

	for i := range dl.Devices {
		currentDevice, err := findCurrentDevice(dsToParse, dl.Devices[i].Identifier)
		if err != nil {
			continue
		}

		err = dl.populateDeviceCapabilities(i, currentDevice, dsToParse)
		if err != nil {
			return err
		}
	}

	return nil
}

// dlReprRaw is similar to dlRepr, except the Devices are of the type json.RawMessage.
type dlReprRaw struct {
	Fwversion  string            `json:"-fwversion"`
	Version    string            `json:"-version"`
	Devicelist []json.RawMessage `json:"device"`
}

// dlRepr is the Go-Representation of the XML-Response from the Fritz!Box for the getdevicelistinfos command.
type dlRepr struct {
	Version   string  `json:"-version"`
	Fwversion string  `json:"-fwversion"`
	Device    []dRepr `json:"device"`
}

// dRepr is the Go-Representation of the XML-Response from the Fritz!Box for a single device.
type dRepr struct {
	Functionbitmask string `json:"-functionbitmask"`
	Fwversion       string `json:"-fwversion"`
	ID              string `json:"-id"`
	Identifier      string `json:"-identifier"`
	Manufacturer    string `json:"-manufacturer"`
	Productname     string `json:"-productname"`
	Name            string `json:"name"`
	Present         string `json:"present"`
	Txbusy          string `json:"txbusy"`
}

// toDeviceList translates a list of dRepr-Devices to a list of Devices
func toDeviceList(dlR dlRepr) *DeviceList {
	dl := &DeviceList{
		Version:   dlR.Version,
		Fwversion: dlR.Fwversion,
	}

	for _, d := range dlR.Device {
		dl.Devices = append(dl.Devices, toDevice(d))
	}

	return dl
}

// toDevice converts a dRepr to a Device. This includes parsing the capabilities from the given Bitmask.
func toDevice(d dRepr) Device {
	return Device{
		Identifier:   d.Identifier,
		ID:           d.ID,
		Fwversion:    d.Fwversion,
		Manufacturer: d.Manufacturer,
		Productname:  d.Productname,
		Txbusy:       d.Txbusy,
		Name:         d.Name,
		Present:      d.Present,
		Capabilities: capabilitiesFromBitmask(d.Functionbitmask),
	}
}

func toDlReprRaw(bs []byte) (dl dlReprRaw, err error) {
	mv, err := mxj.NewMapXml(bs)
	if err != nil {
		return
	}

	mv, err = mv.NewMap("devicelist:.")
	if err != nil {
		return
	}

	// I *HATE* working with xml
	// So I just won't do that :)
	r, err := mv.Json(true)

	var objs map[string]json.RawMessage
	err = json.Unmarshal(r, &objs)
	if err != nil {
		return
	}

	err = json.Unmarshal(objs[""], &dl)
	return
}

// parseDevices takes a dlReprRaw and returns an unmarshaled list of devices to parse.
func parseDevices(dlRaw dlReprRaw) (dsToParse []map[string]json.RawMessage, err error) {
	for _, y := range dlRaw.Devicelist {
		dev := map[string]json.RawMessage{}
		err = json.Unmarshal(y, &dev)
		if err != nil {
			return
		}
		dsToParse = append(dsToParse, dev)
	}
	return
}

// findCurrentDevice takes a list of devices to parse and an identifier and returns the device with the given identifier.
func findCurrentDevice(dsToParse []map[string]json.RawMessage, identifier string) (map[string]json.RawMessage, error) {
	for _, dev := range dsToParse {
		if strings.Trim(string(dev["-identifier"]), "\"") == identifier {
			return dev, nil
		}
	}
	return nil, fmt.Errorf("device with identifier %s not found", identifier)
}

// populateHanfunUnits populates the HanfunUnits of the given Hanfun capability.
func (dl *DeviceList) populateHanfunUnits(i int, dsToParse []map[string]json.RawMessage) error {
	capab := dl.Devices[i].Capabilities[CHanfun]
	if capab == nil {
		return nil
	}

	for _, dev := range dsToParse {
		ident := strings.Trim(string(dev["-identifier"]), "\"")
		if strings.HasPrefix(ident, capab.Device().Identifier) && capab.Device().Identifier != ident {
			hfDevice := dl.findDeviceByIdentifier(ident)
			if hfDevice == nil {
				continue
			}

			fun, ok := capab.(*HanFun)
			if !ok {
				return fmt.Errorf("capability is not *HanFun")
			}
			unit, err := fun.unitFromJSON(dev, hfDevice)
			if err != nil {
				return err
			}
			fun.Units = append(fun.Units, unit)
			dl.Devices[i].Capabilities[CHanfun] = fun
		}
	}
	return nil
}

// findDeviceByIdentifier returns the device with the given identifier.
func (dl *DeviceList) findDeviceByIdentifier(identifier string) *Device {
	for i := range dl.Devices {
		if dl.Devices[i].Identifier == identifier {
			return &dl.Devices[i]
		}
	}
	return nil
}

// doFilter filters the devices in the list by the given capability.
func (dl *DeviceList) doFilter() {
	var tmp []Device
	for _, e := range dl.Devices {
		if e.HasCapability(dl.filter.Name()) {
			tmp = append(tmp, e)
		}
	}
	dl.Devices = tmp
}
