package unsafe

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/ByteSizedMarius/go-fritzbox-api"
)

// FlexString handles JSON fields that may be either string or number.
type FlexString string

func (f *FlexString) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		*f = FlexString(s)
		return nil
	}
	var n json.Number
	if err := json.Unmarshal(b, &n); err == nil {
		*f = FlexString(n.String())
		return nil
	}
	return fmt.Errorf("cannot unmarshal %s into FlexString", string(b))
}

// Device represents a network device connected to the Fritz!Box.
// This is a simplified view returned by GetDeviceList() using query.lua.
type Device struct {
	UID          string `json:"UID"`
	IP           string `json:"ip"`
	MAC          string `json:"mac"`
	Name         string `json:"name"`
	FriendlyName string `json:"friendly_name"`
	Active       bool
	Flags        string `json:"flags"`
	Profile      Profile
}

// GetDevices returns all active network devices known to the Fritz!Box.
// Uses query.lua to fetch device data including MAC addresses and flags.
// To add profile information, call FillProfiles() on the result.
// For inactive devices (e.g., DHCP reservations), use GetAllDevices() instead.
func GetDevices(c *fritzbox.Client) (devices []Device, err error) {
	data := fritzbox.Values{
		"sid":        c.SID(),
		"mq_devices": "landevice:settings/landevice/list(UID,ip,mac,name,friendly_name,active,flags)",
	}

	_, resp, err := c.AhaRequestString(http.MethodGet, "query.lua", data)
	if err != nil {
		return
	}

	var r struct {
		Devices []struct {
			UID          string `json:"UID"`
			IP           string `json:"ip"`
			MAC          string `json:"mac"`
			Name         string `json:"name"`
			FriendlyName string `json:"friendly_name"`
			Active       string `json:"active"`
			Flags        string `json:"flags"`
		} `json:"mq_devices"`
	}
	if err = json.Unmarshal([]byte(resp), &r); err != nil {
		return
	}

	devices = make([]Device, 0, len(r.Devices))
	for _, d := range r.Devices {
		if d.UID == "" || d.Name == "" {
			continue
		}
		devices = append(devices, Device{
			UID:          d.UID,
			IP:           d.IP,
			MAC:          d.MAC,
			Name:         d.Name,
			FriendlyName: d.FriendlyName,
			Active:       d.Active == "1",
			Flags:        d.Flags,
		})
	}

	return
}

// FillProfiles gets the profile for every device and adds it to the device list.
// Warning: This sends one request per device, which may take several seconds for many devices.
// Consider calling GetProfileUIDFromDevice only for specific devices if you don't need all profiles.
func FillProfiles(c *fritzbox.Client, devices []Device) (err error) {
	profiles, err := GetAvailableProfiles(c)
	if err != nil {
		return
	}

	for i, d := range devices {
		var uid string
		uid, err = GetProfileUIDFromDevice(c, d.UID)
		if err != nil {
			return
		}
		devices[i].Profile = profiles[uid]
	}

	return
}

// MeshDevice is the full device representation from the homeNet mesh topology endpoint.
// Contains detailed connection info, mesh hierarchy, and state information.
type MeshDevice struct {
	UID             FlexString `json:"UID"`
	Parent          string `json:"parent"`
	Category        string `json:"category"`
	Profile         Profile
	OwnClientDevice bool   `json:"own_client_device"`
	Dist            int    `json:"dist"`
	Switch          bool   `json:"switch"`
	Devtype         string `json:"devtype"`
	MAC             string `json:"mac"`
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

// MeshTopology contains the mesh network structure with detailed device information.
type MeshTopology struct {
	Rootuid string       `json:"rootuid"`
	Devices []MeshDevice `json:"devices"`
}

// GetMeshTopology returns the mesh network topology with full device details.
// Use this when you need connection speeds, mesh hierarchy, or detailed state info.
// For a simpler device list with MAC addresses, use GetDeviceList() instead.
func GetMeshTopology(c *fritzbox.Client) (mt MeshTopology, err error) {
	data := fritzbox.Values{
		"sid":  c.SID(),
		"page": "homeNet",
	}

	_, resp, err := c.AhaRequestString(http.MethodPost, "data.lua", data)
	if err != nil {
		return
	}

	// remove empty conninfo array, otherwise json.Unmarshal will fail (inconsistent types)
	body := strings.ReplaceAll(resp, ",\"conninfo\":[]", "")

	r := struct {
		Data struct {
			Topology struct {
				Rootuid string                     `json:"rootuid"`
				Devices map[string]json.RawMessage `json:"devices"`
			}
		}
	}{}
	err = json.Unmarshal([]byte(body), &r)
	if err != nil {
		return
	}

	mt.Rootuid = r.Data.Topology.Rootuid
	mt.Devices = []MeshDevice{}
	for _, v := range r.Data.Topology.Devices {
		var d MeshDevice
		if err = json.Unmarshal(v, &d); err != nil {
			return
		}
		if d.Nameinfo.Name != "" && d.UID != "" {
			mt.Devices = append(mt.Devices, d)
		}
	}

	return
}
