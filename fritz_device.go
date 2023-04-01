package go_fritzbox_api

import (
	"fmt"
	"net/http"
	"net/url"
)

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
		Name string `json:"Name"`
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

// SetIP sets the IP address of a device like you would using the interface.
// Static indicated whether the given IP should be static and should always be true when changing the IP.
// Having it set to false is the same as un-ticking the checkbox in the edit dialog.
func (c *Client) SetIP(deviceUID string, ip string, static bool) (err error) {
	staticStr := "on"
	if !static {
		staticStr = "off"
	}

	name, err := c.getCurrentName(deviceUID)
	if err != nil {
		return
	}

	data := url.Values{
		"dev_name":    {name},
		"dev_ip":      {ip},
		"static_dhcp": {staticStr},
		"dev":         {deviceUID},
		"apply":       {""},
		"sid":         {c.session.Sid},
		"page":        {"edit_device"},
	}

	_, err = c.doRequest(http.MethodPost, "data.lua", data)
	if err != nil {
		return
	}

	if !c.isIp(deviceUID, ip) {
		return fmt.Errorf("ip change unsuccessful")
	}

	return nil
}

// SetName sets the name of a device. The name may only contain letters (a-z, A-Z), numerals (0-9), and hyphens (-).
func (c *Client) SetName(deviceUID string, newName string) (err error) {
	data := url.Values{
		"dev_name": {newName},
		"dev":      {deviceUID},
		"apply":    {""},
		"sid":      {c.session.Sid},
		"page":     {"edit_device"},
	}

	_, err = c.doRequest(http.MethodPost, "data.lua", data)
	if err != nil {
		return
	}

	if !c.isName(deviceUID, newName) {
		return fmt.Errorf("name change not successful for unknown reason")
	}
	return nil
}

func (c *Client) getEditInfos(deviceUID string) (body string, err error) {
	data := url.Values{
		"dev":  {deviceUID},
		"sid":  {c.session.Sid},
		"page": {"edit_device"},
	}

	resp, err := c.doRequest(http.MethodPost, "data.lua", data)
	if err != nil {
		return
	}

	body, err = getBody(resp)
	if err != nil {
		return
	}

	return
}

func (c *Client) getCurrentName(deviceUID string) (name string, err error) {
	body, err := c.getEditInfos(deviceUID)
	if err != nil {
		return
	}

	n, err := valueFromJson(body, []string{"data", "vars", "dev", "name"})
	name = n["displayName"].(string)
	return
}

func (c *Client) getCurrentIP(deviceUID string) (ip string, err error) {
	body, err := c.getEditInfos(deviceUID)
	if err != nil {
		return
	}

	v, err := valueFromJson(body, []string{"data", "vars", "dev", "ipv4", "current"})
	ip = v["ip"].(string)
	return
}

func (c *Client) isIp(deviceUID string, ip string) bool {
	cip, err := c.getCurrentIP(deviceUID)
	if err != nil {
		return false
	}
	return cip == ip
}

func (c *Client) isName(deviceUID string, name string) bool {
	cname, err := c.getCurrentName(deviceUID)
	if err != nil {
		return false
	}
	return cname == name
}
