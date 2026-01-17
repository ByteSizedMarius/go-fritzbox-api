package aha

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/ByteSizedMarius/go-fritzbox-api/v2"
)

type Units []*Unit

// String returns a string representation of the units
func (hfus Units) String() string {
	var builder strings.Builder
	builder.WriteString("[")
	for _, u := range hfus {
		_, _ = fmt.Fprintf(&builder, "%s, ", u.String())
	}
	result := strings.TrimSuffix(builder.String(), ", ")
	return result + "]"
}

// Unit is a generic struct for a hanfun unit
// Hanfun is an open standard for smart home devices by the ULE Alliance that seems to be mostly dead in the water.
type Unit struct {
	// Saves all properties of the unit in raw json-messages.
	// Does not include values already represented (values present in the device-struct, as well as etsi-unit-info)
	// This allows access to functionality that may not yet be implemented.
	RawProperties map[string]json.RawMessage
	ETSIUnitInfo  ETSIUnitInfo
	Interface     Interface
	device        *Device
}

// IsUnitOfType returns true if the unit's interface is of the given type
// For example [...].IsUnitOfType(Alert{})
func (h *Unit) IsUnitOfType(t Interface) bool {
	return reflect.TypeOf(h.Interface) == reflect.TypeOf(t)
}

// Reload fetches the device-information for this unit from the fritzbox and updates the structs' values
func (h *Unit) Reload(c *fritzbox.Client) error {
	data, err := GetDeviceInfosRaw(c, h.Device().Identifier)
	if err != nil {
		return err
	}

	var rawMap map[string]json.RawMessage
	err = json.Unmarshal(data, &rawMap)
	if err != nil {
		return err
	}

	tt, err := h.FromJSON(rawMap, h.device)
	if err != nil {
		return err
	}

	h.RawProperties = tt.RawProperties
	h.ETSIUnitInfo = tt.ETSIUnitInfo
	h.Interface = tt.Interface

	return nil
}

// UnmarshalProperty unmarshals a property into the given interface
func (h *Unit) UnmarshalProperty(propertyKey string, dest Interface) error {
	_, ok := h.RawProperties[propertyKey]
	if ok {
		err := json.Unmarshal(h.RawProperties[propertyKey], dest)
		return err
	} else {
		return errors.New("invalid key")
	}
}

// GetRawProperties returns the local interface-values as a string in json.
func (h *Unit) GetRawProperties() (s map[string]string, err error) {
	s = make(map[string]string)
	for k, v := range h.RawProperties {
		s[k] = string(v)
	}
	return
}

// String returns a string representation of the unit
func (h *Unit) String() string {
	return fmt.Sprintf(
		"{ETSI Units Info: %s, Interface: %s, Device: %s}",
		h.ETSIUnitInfo, h.Interface, h.Device(),
	)
}

// Device returns the device the unit is associated with
func (h *Unit) Device() *Device {
	return h.device
}

// FromJSON parses the json into a HanFun-Unit
func (*Unit) FromJSON(m map[string]json.RawMessage, d *Device) (*Unit, error) {
	hfu := &Unit{}
	eui := ETSIUnitInfo{}
	err := json.Unmarshal(m["etsiunitinfo"], &eui)
	if err != nil {
		return hfu, err
	}
	hfu.ETSIUnitInfo = eui

	// if interface is known, parse it
	// otherwise its values are still accessible via RawProperties
	if i, ok := hanFunInterfaces[hfu.ETSIUnitInfo.Interface]; ok {
		hfu.Interface = i
		hfu.Interface, err = hfu.Interface.fromJSON(m)
		if err != nil {
			return hfu, err
		}
	}

	// Filter out the ignoreKeywords
	hfu.RawProperties = make(map[string]json.RawMessage)
	for k, v := range m {
		if !strings.Contains(ignoreKeywords, k) {
			hfu.RawProperties[k] = v
		}
	}

	hfu.device = d
	return hfu, nil
}

type ETSIUnitInfo struct {
	ETSIDeviceID string `json:"etsideviceid"`
	Interface    string `json:"interfaces"`
	UnitType     string `json:"unittype"`
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
	return fmt.Sprintf(
		"{ETSI-Device-ID: %s, Interface-Type: %s (%s), Units-Type: %s (%s)}",
		e.ETSIDeviceID, e.GetInterfaceString(), e.Interface, e.GetUnitString(), e.UnitType,
	)
}
