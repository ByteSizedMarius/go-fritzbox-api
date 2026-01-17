package aha

import (
	"encoding/json"
	"fmt"

	"github.com/ByteSizedMarius/go-fritzbox-api/v2"
)

// HanFun is the capability for a HanFun-Device
// HanFun is an open standard for smart home devices and is supported by the Fritz!Box
type HanFun struct {
	CapName string
	Units   Units // A single HanFun-Device can potentially consist of multiple units
	device  *Device
}

// Name returns the name of the capability
func (hf *HanFun) Name() string {
	return hf.CapName
}

// String returns a string representation of the capability
func (hf *HanFun) String() string {
	return fmt.Sprintf("%s: {Units: %s}", hf.CapName, hf.Units)
}

// Device returns the device the capability is attached to
func (hf *HanFun) Device() *Device {
	return hf.device
}

// fromJSON is the Capability-Interface-Method. It is not used here, the HF Capability is handled separately.
func (hf *HanFun) fromJSON(_ map[string]json.RawMessage, d *Device) (Capability, error) {
	hf.device = d
	return hf, nil
}

// Reload reloads the device itself and all its units.
func (hf *HanFun) Reload(c *fritzbox.Client) error {
	for i := range hf.Units {
		err := hf.Units[i].Reload(c)
		if err != nil {
			return err
		}
	}

	return nil
}

// HasInterface returns true, if the given HanFun-Interface is present in the current devices' units
// For example: [...].HasInterface(Alert{})
func (hf *HanFun) HasInterface(i Interface) bool {
	for _, u := range hf.Units {
		if u.IsUnitOfType(i) {
			return true
		}
	}
	return false
}

// GetInterface returns a pointer to the requested interface if present, nil if not.
// For example: [...].GetInterface(Alert{}).(Alert)
func (hf *HanFun) GetInterface(i Interface) interface{} {
	for in := range hf.Units {
		if hf.Units[in].IsUnitOfType(i) {
			return hf.Units[in].Interface
		}
	}
	return nil
}

// unitFromJSON unmarshals a unit from a json map
func (hf *HanFun) unitFromJSON(m map[string]json.RawMessage, d *Device) (*Unit, error) {
	return new(Unit).FromJSON(m, d)
}
