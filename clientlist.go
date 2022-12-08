package go_fritzbox_api

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
)

// Note: I have since found out about landevice:settings/landevice/list but I do not currently want to implement it
// It would break alot of things, have little benefit for me over the current method and I'm short on time currently
// PRs welcome, but please don't break API (meaning it should be implemented additional to current method, current method can be marked as deprecated however)

type Clientlist struct {
	Rootuid string `json:"rootuid"`
	Devices []Device
}

type Device struct {
	Devid     string `json:"devid"`
	Stateinfo struct {
		GuestOwe        bool `json:"guest_owe"`
		Active          bool `json:"active"`
		Guest           bool `json:"guest"`
		Online          bool `json:"online"`
		Blocked         bool `json:"blocked"`
		Realtime        bool `json:"realtime"`
		Notallowed      bool `json:"notallowed"`
		InternetBlocked bool `json:"internetBlocked"`
	} `json:"stateinfo,omitempty"`
	Profile    Profile
	Devtype    string   `json:"devtype"`
	Dist       int      `json:"dist"`
	Parent     string   `json:"parent"`
	Category   string   `json:"category"`
	Ownentry   bool     `json:"ownentry"`
	UID        string   `json:"UID"`
	Conn       string   `json:"conn"`
	Master     bool     `json:"master"`
	Ipinfo     []string `json:"ipinfo"`
	Updateinfo struct {
		State string `json:"state"`
	} `json:"updateinfo"`
	Gateway  bool `json:"gateway"`
	Nameinfo struct {
		Name string `json:"name"`
	} `json:"nameinfo,omitempty"`
	Children []interface{} `json:"children"`
	Conninfo []struct {
		Speed   string `json:"speed"`
		SpeedTx int    `json:"speed_tx"`
		SpeedRx int    `json:"speed_rx"`
		Desc    string `json:"desc"`
	} `json:"conninfo"`
}

func (c *Client) GetCLientList() (clients Clientlist, err error) {
	if err = c.checkExpiry(); err != nil {
		return
	}

	data := url.Values{
		"sid":         {c.session.Sid},
		"updatecheck": {""},
	}

	resp, err := c.doRequest(http.MethodGet, "net/network.lua", data)
	if err != nil {
		return
	}

	body, err := getBody(resp)
	if err != nil {
		return
	}

	// conninfo is [[]] when empty, just [...] when full (bug)
	body = strings.ReplaceAll(body, "conninfo\":[[]]", "conninfo\":[]")

	// get json from response
	body = strings.Split("{\"rootuid\""+strings.Split(body, "{\"rootuid\"")[1], "}},\"nexus\"")[0] + "}}"

	// get rid of top level
	tmp := map[string]json.RawMessage{}
	err = json.Unmarshal([]byte(body), &tmp)
	if err != nil {
		return
	}

	// rootuid
	err = json.Unmarshal(tmp["rootuid"], &clients.Rootuid)
	if err != nil {
		return
	}

	// parse devices
	clients.Devices = []Device{}
	err = json.Unmarshal(tmp["devices"], &tmp)
	if err != nil {
		return
	}

	for _, v := range tmp {
		var d Device

		err = json.Unmarshal(v, &d)
		if err != nil {
			// ignore errors because they will happen for devices not relevant here (dect)
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
func (c *Client) AddProfiles(cl *Clientlist) {
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
