package go_fritzbox_api

import (
	"encoding/json"
	"net/http"
	"net/url"
)

type response struct {
	Data struct {
		Topology struct {
			Rootuid string                     `json:"rootuid"`
			Devices map[string]json.RawMessage `json:"devices"`
		}
	}
}

type DeviceList struct {
	Rootuid string `json:"rootuid"`
	Devices []Device
}

func (c *Client) GetCLientList() (clients DeviceList, err error) {
	if err = c.checkExpiry(); err != nil {
		return
	}

	data := url.Values{
		"sid":  {c.session.Sid},
		"page": {"homeNet"},
	}

	resp, err := c.doRequest(http.MethodPost, "data.lua", data)
	if err != nil {
		return
	}

	body, err := getBody(resp)
	if err != nil {
		return
	}

	r := response{}
	err = json.Unmarshal([]byte(body), &r)
	if err != nil {
		return
	}

	clients.Devices = []Device{}
	for _, v := range r.Data.Topology.Devices {
		var d Device

		err = json.Unmarshal(v, &d)
		if err != nil {
			// ignore errors because they will happen for devices not relevant here (dect)
			// fmt.Println(string(v))
			continue
		}

		if d.Nameinfo.Name != "" && d.UID != "" {
			clients.Devices = append(clients.Devices, d)
		}
	}
	err = nil

	return
}

// AddProfiles gets the profile for every device and adds it into an already existing Client-list.
// Warning: This will send one request per device, meaning it will take multiple seconds to complete.
// Alternatively, call GetProfileUIDFromDevice only for the devices you need
func (c *Client) AddProfiles(cl *DeviceList) {
	for dI, d := range cl.Devices {
		profiles, err := c.GetAvailableProfiles()
		if err != nil {
			return
		}

		var uid string
		uid, err = c.GetProfileUIDFromDevice(d.UID)
		if err != nil {
			return
		}

		d.Profile = profiles[uid]
		cl.Devices[dI] = d
	}
}
