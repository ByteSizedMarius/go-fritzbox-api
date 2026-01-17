package aha

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ByteSizedMarius/go-fritzbox-api/v2"
)

// ButtonDevice is the capability for a button
type ButtonDevice struct {
	CapName    string
	Buttons    []Button
	Batterylow string
	Battery    string
	device     *Device
}

// String returns a string representation of the capability
func (b *ButtonDevice) String() string {
	var builder strings.Builder
	_, _ = fmt.Fprintf(&builder, "%s: {[", b.CapName)
	for _, bp := range b.Buttons {
		_, _ = fmt.Fprintf(&builder, "%s, ", bp.String())
	}
	result := strings.TrimSuffix(builder.String(), ", ")
	return result + "]}"
}

// Name returns the name of the capability
func (b *ButtonDevice) Name() string {
	return b.CapName
}

// Device returns the device the capability belongs to
func (b *ButtonDevice) Device() *Device {
	return b.device
}

// fromJSON unmarshals the capability from a json-representation
func (b *ButtonDevice) fromJSON(m map[string]json.RawMessage, d *Device) (Capability, error) {
	b.Batterylow = string(m["batterylow"])
	b.Battery = string(m["battery"])

	err := json.Unmarshal(m["button"], &b.Buttons)
	if err != nil {
		return nil, err
	}

	for i, bs := range b.Buttons {
		idx := strings.LastIndex(bs.Name, ": ")
		if idx >= 0 && idx+2 < len(bs.Name) {
			t := bs.Name[idx+2:]
			if buttonTypes[t] {
				bs.Type = t
			}
		}
		b.Buttons[i] = bs
	}
	b.device = d
	return b, nil
}

// Button is a part of the ButtonDevice capability
// A ButtonDevice can have multiple buttons
type Button struct {
	ID                   string `json:"-id"`
	Type                 string
	Name                 string `json:"name"`
	Identifier           string `json:"-identifier"`
	LastPressedTimeStamp string `json:"lastpressedtimestamp"`
}

// GetLastPressTime converts the last-press-time string to a time-struct
func (bp Button) GetLastPressTime() time.Time {
	return unixStringToTime(bp.LastPressedTimeStamp)
}

// Reload reloads all client values
func (b *ButtonDevice) Reload(c *fritzbox.Client) error {
	return GetDeviceInfos(c, b.Device().Identifier, b)
}

func (bp Button) String() string {
	return fmt.Sprintf(
		"{ID: %s, Name: %s, ButtonDevice-Typ: %s, Identifier: %s, zuletzt betätigt: %s}",
		bp.ID, bp.Name, bp.Type, bp.Identifier, bp.GetLastPressTime(),
	)
}
