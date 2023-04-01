package go_fritzbox_api

import (
	"encoding/json"
	"github.com/clbanning/mxj/v2"
	"io"
	"net/http"
	"net/url"
)

func (c *Client) GetSmarthomeDevices() (dl *SHDevicelist, err error) {
	r, err := doRequest(c)
	if err != nil {
		return
	}

	return dl.fromReader(r)
}

func (c *Client) GetSmarthomeDevicesFilter(requiredCapabilities []string) (dl *SHDevicelist, err error) {
	dl, err = c.GetSmarthomeDevices()
	if err != nil {
		return
	}

	var tmp []SmarthomeDevice
	valid := true
	for _, e := range dl.Devices {
		for _, capab := range requiredCapabilities {
			if !e.HasCapability(capab) {
				valid = false
			}
		}

		if valid {
			tmp = append(tmp, e)
		}

		valid = true
	}

	dl.Devices = tmp
	return dl, nil
}

func (c *Client) GetDeviceInfos(devIdentifier string, dest interface{}) (err error) {
	data := url.Values{
		"sid":       {c.session.Sid},
		"switchcmd": {"getdeviceinfos"},
		"ain":       {devIdentifier},
	}

	resp, err := c.doRequest(http.MethodGet, "webservices/homeautoswitch.lua", data)
	if err != nil {
		return err
	}

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	mv, err := mxj.NewMapXml(bytes)
	if err != nil {
		return
	}

	j, err := mv.Json(true)
	if err != nil {
		return
	}

	return json.Unmarshal(j, dest)
}
