package go_fritzbox_api

import (
	"encoding/json"
	"net/http"
	"net/url"
)

// Note: I have since found out about landevice:settings/landevice/list but I do not currently want to implement it
// It would break alot of things, have little benefit for me over the current method and I'm short on time currently
// PRs welcome, but please don't break API (meaning it should be implemented additional to current method, current method can be marked as deprecated however)

type response struct {
	Data struct {
		Topology struct {
			Rootuid string                     `json:"rootuid"`
			Devices map[string]json.RawMessage `json:"devices"`
		}
	}
}

type Clientlist struct {
	Rootuid string `json:"rootuid"`
	Devices []Device
}

type Device struct {
	UID             string `json:"UID"`
	Parent          string `json:"parent"`
	Category        string `json:"category"`
	Profile         Profile
	OwnClientDevice bool   `json:"own_client_device"`
	Dist            int    `json:"dist"`
	Switch          bool   `json:"switch"`
	Devtype         string `json:"devtype"`
	Ownentry        bool   `json:"ownentry"`
	Stateinfo       struct {
		GuestOwe        bool `json:"guest_owe"`
		Active          bool `json:"active"`
		Meshable        bool `json:"meshable"`
		Guest           bool `json:"guest"`
		Online          bool `json:"online"`
		Blocked         bool `json:"blocked"`
		Realtime        bool `json:"realtime"`
		NotalloWed      bool `json:"notallowed"`
		InternetBlocked bool `json:"internetBlocked"`
	} `json:"stateinfo"`
	Conn       string `json:"conn"`
	Master     bool   `json:"master"`
	Detailinfo struct {
		Edit struct {
			Pid    string `json:"pid"`
			Params struct {
				Dev        string `json:"dev"`
				BackToPage string `json:"back_to_page"`
			} `json:"params"`
		} `json:"edit"`
		Portrelease bool `json:"portrelease"`
	} `json:"detailinfo"`
	Updateinfo struct {
		State string `json:"state"`
	} `json:"updateinfo"`
	Gateway  bool `json:"gateway"`
	Nameinfo struct {
		Name string `json:"name"`
	} `json:"nameinfo"`
	Children []interface{} `json:"children"`
	Conninfo struct {
		Kind     string `json:"kind"`
		Speed    string `json:"speed"`
		Bandinfo []struct {
			Band    int    `json:"band"`
			SpeedTx int    `json:"speed_tx"`
			SpeedRx int    `json:"speed_rx"`
			Speed   string `json:"speed"`
			Desc    string `json:"desc"`
		} `json:"bandinfo"`
		Usedbands int    `json:"usedbands"`
		Desc      string `json:"desc"`
	} `json:"conninfo"`
	Ipinfo string `json:"ipinfo"`
}

func (c *Client) GetCLientList() (clients Clientlist, err error) {
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
