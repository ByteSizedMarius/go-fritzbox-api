package go_fritzbox_api

import (
	"encoding/json"
	"fmt"
	"github.com/clbanning/mxj/v2"
	"io"
	"net/http"
	"net/url"
	"strings"
)

var MaskTranslStr = []string{CHanfun, "", CLicht, "", CAlarm, CButton, CHKR, CEnergieMesser, CTempSensor, CSteckdose, CRepeater, CMikrofon, "", CHanfunUnit, "", CSchaltbar, CDimmbar, CLampeMitFarbtemp, CRollladen}

type SHDevicelist struct {
	Version   string
	Fwversion string
	Devices   []SmarthomeDevice
}

func (dl *SHDevicelist) Reload(c *Client) error {
	r, err := doRequest(c)
	if err != nil {
		return err
	}

	var tdl *SHDevicelist
	tdl, err = dl.fromReader(r)
	if err != nil {
		return err
	}

	dl.Devices = tdl.Devices
	dl.Version = tdl.Version
	dl.Fwversion = tdl.Fwversion
	return nil
}

func (dl *SHDevicelist) String() string {
	rt := ""
	for _, d := range dl.Devices {
		rt += d.String() + ", "
	}
	if rt != "" {
		rt = rt[:len(rt)-2]
	}
	return fmt.Sprintf("Version: %s, Firmware-Version: %s Devices: [%s]", dl.Version, dl.Fwversion, rt)
}

func (dl *SHDevicelist) populateCapabilities(b []byte) error {
	var err error

	mv, err := mxj.NewMapXml(b)
	if err != nil {
		return err
	}

	mv, err = mv.NewMap("devicelist:.")
	if err != nil {
		return err
	}

	// i *HATE* working with xml
	r, err := mv.Json(true)

	// get rid of empty parent
	var objs map[string]json.RawMessage
	err = json.Unmarshal(r, &objs)
	if err != nil {
		return err
	}

	type St struct {
		Fwversion  string            `json:"-fwversion"`
		Version    string            `json:"-version"`
		Devicelist []json.RawMessage `json:"device"`
	}

	s := St{}
	err = json.Unmarshal(objs[""], &s)
	if err != nil {
		return err
	}

	var devices []map[string]json.RawMessage
	for _, y := range s.Devicelist {
		dev := map[string]json.RawMessage{}
		err = json.Unmarshal(y, &dev)
		devices = append(devices, dev)
	}

	for i := range dl.Devices {
		var currentDevice map[string]json.RawMessage
		for _, dev := range devices {
			if strings.Trim(string(dev["-identifier"]), "\"") == dl.Devices[i].Identifier {
				currentDevice = dev
			}
		}
		if currentDevice == nil {
			fmt.Printf("device with identifier %s not found", dl.Devices[i].Identifier)
			continue
		}

		//  HanFun
		if dl.Devices[i].HasCapability(CHanfun) {
			var c Capability
			c, err = dl.Devices[i].Capabilities[CHanfun].fromJSON(currentDevice, &dl.Devices[i])
			if err != nil {
				return err
			}

			// HanFun-Gerät könnte mehrere Units haben? Idk
			for _, dev := range devices {

				// find units
				ident := strings.Trim(string(dev["-identifier"]), "\"")
				if strings.HasPrefix(ident, c.Device().Identifier) && c.Device().Identifier != strings.Trim(string(dev["-identifier"]), "\"") {

					// find matching device
					hfDevice := dl.Devices[i]
					for _, dd := range dl.Devices {
						if dd.Identifier == ident {
							hfDevice = dd
						}
					}

					// add unit
					fun := c.(*HanFun)
					var unit *HanFunUnit
					unit, err = fun.unitFromJSON(dev, &hfDevice)
					if err != nil {
						return err
					}
					fun.Units = append(fun.Units, unit)
					c = fun
				}
			}

			dl.Devices[i].Capabilities[CHanfun] = c
		}

		//	Licht

		//	Alarm

		//	Button
		if dl.Devices[i].HasCapability(CButton) {
			var c Capability
			c, err = dl.Devices[i].Capabilities[CButton].fromJSON(currentDevice, &dl.Devices[i])
			if err != nil {
				return err
			}
			dl.Devices[i].Capabilities[CButton] = c
		}

		//	HKR
		if dl.Devices[i].HasCapability(CHKR) {
			var c Capability
			c, err = dl.Devices[i].Capabilities[CHKR].fromJSON(currentDevice, &dl.Devices[i])
			if err != nil {
				return err
			}
			dl.Devices[i].Capabilities[CHKR] = c
		}

		//	EnergieMesser

		//
		//	TempSensor
		//
		if dl.Devices[i].HasCapability(CTempSensor) {
			var c Capability
			c, err = dl.Devices[i].Capabilities[CTempSensor].fromJSON(currentDevice, &dl.Devices[i])
			if err != nil {
				return err
			}
			dl.Devices[i].Capabilities[CTempSensor] = c
		}

		//	Steckdose

		//	Repeater

		//	Mikrofon

		//	Schaltbar

		//	Dimmbar

		//	LampeMitFarbtemp

		//	Rollladen

		currentDevice = nil
	}

	return nil
}

func (*SHDevicelist) fromDeviceList(dlt extDevicelist) *SHDevicelist {
	return &SHDevicelist{
		Version:   dlt.Version,
		Fwversion: dlt.Fwversion,
	}
}

// devicesFromExtDevices translates a list of extDevices to a list of SmarthomeDevices
// This requires the parsing of all available raw information into defined structs
func (*SHDevicelist) devicesFromExtDevices(devices []extDevice) (dl []SmarthomeDevice, err error) {
	for _, d := range devices {
		nd := (&SmarthomeDevice{}).fromDevice(d)

		dl = append(dl, nd)
	}

	return
}

func (*SHDevicelist) fromReader(r io.ReadCloser) (dl *SHDevicelist, err error) {
	bytes, err := io.ReadAll(r)
	if err != nil {
		return
	}

	edl, err := extDevicelist{}.fromBytes(bytes)

	// translate extDevicelist
	dl = dl.fromDeviceList(edl)
	dl.Devices, err = dl.devicesFromExtDevices(edl.Device)

	err = dl.populateCapabilities(bytes)
	if err != nil {
		return
	}

	return
}

func doRequest(c *Client) (io.ReadCloser, error) {
	data := url.Values{
		"sid":       {c.session.Sid},
		"switchcmd": {"getdevicelistinfos"},
	}

	resp, err := c.doRequest(http.MethodGet, "webservices/homeautoswitch.lua", data, true)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}
