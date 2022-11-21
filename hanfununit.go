package go_fritzbox_api

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type HanFunUnit struct {
	RawProperties map[string]json.RawMessage
	ETSIUnitInfo  ETSIUnitInfo
	Interface     HFInterface
	device        *Device
}

type ETSIUnitInfo struct {
	ETSIDeviceID string `json:"etsideviceid"`
	Interface    string `json:"interfaces"`
	UnitType     string `json:"unittype"`
}

func (h *HanFunUnit) Reload(c *Client) error {
	resp, err := getDeviceInfos(c, h.Device())
	if err != nil {
		return err
	}

	newCapa, err := capabilityMapFromString(resp)
	if err != nil {
		return err
	}

	// get capability struct
	tt, err := h.fromJSON(newCapa, h.device)
	if err != nil {
		return err
	}

	// update current capability
	*h = tt
	return nil
}

// UnmarshalProperty unmarshals a property into the given interface
func (h HanFunUnit) UnmarshalProperty(propertyKey string, dest interface{}) error {
	_, ok := h.RawProperties[propertyKey]
	if ok {
		err := json.Unmarshal(h.RawProperties[propertyKey], dest)
		return err
	} else {
		return errors.New("invalid key")
	}
}

// GetRawProperties returns the local interface-values as a string in json.
func (h HanFunUnit) GetRawProperties() (s map[string]string, err error) {
	s = make(map[string]string)
	for k, v := range h.RawProperties {
		s[k] = string(v)
	}

	return
}

func (h HanFunUnit) fromJSON(m map[string]json.RawMessage, d *Device) (HanFunUnit, error) {
	err := json.Unmarshal(m["etsiunitinfo"], &h.ETSIUnitInfo)
	if err != nil {
		return h, err
	}

	h.Interface = hanFunInterfaces[h.ETSIUnitInfo.Interface]
	h.Interface, err = h.Interface.fromJSON(m)
	if err != nil {
		return h, err
	}

	h.RawProperties = make(map[string]json.RawMessage)
	for k, v := range m {
		if !strings.Contains(ignoreKeywords, k) {
			h.RawProperties[k] = v
		}
	}

	h.device = d
	return h, nil
}

func (h HanFunUnit) String() string {
	return fmt.Sprintf("{ETSI Units Info: %s, Interface: %s, Device: %s}", h.ETSIUnitInfo, h.Interface, h.Device())
}

func (h HanFunUnit) Device() *Device {
	return h.device
}

// GetInterfaceString returns the units interface-type as a string (values taken from documentation)
func (e ETSIUnitInfo) GetInterfaceString() string {
	return hanFunInterfacesStr[e.Interface]
}

// GetUnitString returns the units type as a string (values taken from documentation)
func (e ETSIUnitInfo) GetUnitString() string {
	return hanFunUnitTypes[e.UnitType]
}

func (e ETSIUnitInfo) String() string {
	return fmt.Sprintf("{ETSI-Device-ID: %s, Interface-Type: %s (%s), Units-Type: %s (%s)}", e.ETSIDeviceID, e.GetInterfaceString(), e.Interface, e.GetUnitString(), e.UnitType)
}
