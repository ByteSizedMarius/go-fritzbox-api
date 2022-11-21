package go_fritzbox_api

import (
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

type Clientlist struct {
	Rootuid string `json:"rootuid"`
	Devices []struct {
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
	} `json:"devices"`
}

// GetCLientList returns a Clientlist-Object containing a devicelist
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

	if len(body) > 10 {
		body = strings.Split(body[strings.Index(body, "{"):], "\n")[0]
		body = body[strings.Index(body, "{\"rootuid") : strings.Index(body, "}}},")+3]

		// Devices zum Array machen
		body = strings.Replace(body, "\"devices\":{", "\"devices\":[", 1)
		body = body[:len(body)-3] + "]}"

		var re = regexp.MustCompile(`("landevice[0-9]{3,4}":{)|("[0-9]{1,4}":{)`)
		x := re.FindStringIndex(body)

		for x != nil {
			preIndex := x[0]
			sufIndex := x[1]

			// Bis zum lan-device unverändert
			pre := body[:preIndex]

			// Den landevice-string zur Variable machen
			pre += "{\"devid\":" + body[preIndex:sufIndex]
			pre = pre[:len(pre)-2] + ","

			// Den Rest des Bodys wieder anhängen
			body = pre + body[sufIndex:]

			x = re.FindStringIndex(body)
		}

		// sometimes the uids are an integer for some reason
		re = regexp.MustCompile(`"UID":[0-9]*,`)
		uid := re.FindAllStringIndex(body, -5)
		for i, uidInt := range uid {
			currUid := body[uidInt[0]+(i*2) : uidInt[1]+(i*2)]
			body = strings.Replace(body, currUid, "\"UID\":\""+strings.Split(strings.Split(currUid, ":")[1], ",")[0]+"\",", 1)
		}

		// Conninfo macht kein Sinn wenns leer ist, anderes Format als wenns voll ists
		body = strings.ReplaceAll(body, "conninfo\":[[]]", "conninfo\":[]")

		// Ende des Arrays setzen
		body = body[:len(body)-2] + "}]}"

		clients = Clientlist{}
		err = json.Unmarshal([]byte(body), &clients)
	}

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
