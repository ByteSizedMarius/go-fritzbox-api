package go_fritzbox_api

// todo: FD440 has humidity sensor
// note: only tested for model 400

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Button struct {
	CapName    string
	Buttons    []ButtonPress
	Batterylow string
	Battery    string
	device     *SmarthomeDevice
}

type ButtonPress struct {
	ID                   string `json:"-id"`
	Type                 string
	Name                 string `json:"Name"`
	Identifier           string `json:"-identifier"`
	LastPressedTimeStamp string `json:"lastpressedtimestamp"`
}

// GetLastPressTime converts the last-press-time string to a time-struct
func (bp ButtonPress) GetLastPressTime() (t time.Time) {
	return unixStringToTime(bp.LastPressedTimeStamp)
}

// Reload reloads all client values
func (b *Button) Reload(c *Client) error {
	tt, err := getDeviceInfosFromCapability(c, b)
	if err != nil {
		return err
	}

	// update current capability
	th := tt.(*Button)
	b.CapName = th.CapName
	b.Buttons = th.Buttons
	b.Batterylow = th.Batterylow
	b.Battery = th.Battery
	b.device = th.device
	return nil
}

func (b *Button) String() string {
	s := fmt.Sprintf("%s: {[", b.CapName)
	for _, bp := range b.Buttons {
		s += bp.String() + ", "
	}
	return s[:len(s)-2] + "]}"
}

func (bp ButtonPress) String() string {
	return fmt.Sprintf("{ID: %s, Name: %s, Button-Typ: %s, Identifier: %s, zuletzt betätigt: %s}", bp.ID, bp.Name, bp.Type, bp.Identifier, bp.GetLastPressTime())
}

func (b *Button) Name() string {
	return b.CapName
}

func (b *Button) Device() *SmarthomeDevice {
	return b.device
}

func (b *Button) fromJSON(m map[string]json.RawMessage, d *SmarthomeDevice) (Capability, error) {
	b.Batterylow = string(m["batterylow"])
	b.Batterylow = string(m["battery"])

	err := json.Unmarshal(m["button"], &b.Buttons)
	if err != nil {
		return nil, err
	}

	for i, bs := range b.Buttons {
		if strings.Contains(bs.Name, EvTastendruckKurz) || strings.Contains(bs.Name, EvTastendruckLang) {
			t := bs.Name[strings.LastIndex(bs.Name, ": ")+2:]
			if t == EvTastendruckLang {
				bs.Type = EvTastendruckLang
			} else if t == EvTastendruckKurz {
				bs.Type = EvTastendruckKurz
			}
			b.Buttons[i] = bs
		} else {
			fmt.Println("Unknown Button-Type: " + bs.Name)
		}
	}
	b.device = d
	return b, nil
}
