package go_fritzbox_api

import (
	"encoding/json"
	"fmt"
)

type HanFun struct {
	CapName string
	// A single HanFun-Device can potentially consist of multiple units (I think)?
	Units  HanFunUnits
	device *Device
}

type HanFunUnits []HanFunUnit

// Reload reloads the device itself and all its units.
func (hfu *HanFun) Reload(c *Client) error {
	tt, err := getDeviceInfosFromCapability(c, hfu)
	if err != nil {
		return err
	}

	// update current capability
	th := tt.(HanFun)
	*hfu = th

	for i := range hfu.Units {
		err = hfu.Units[i].Reload(c)
		if err != nil {
			return err
		}
	}

	return nil
}

// HasInterface returns true, if the given HanFun-Interface is present in the current devices' units
// For example: [...].HasInterface(HFAlert{})
func (hfu HanFun) HasInterface(i interface{}) bool {
	for _, u := range hfu.Units {
		if u.IsUnitOfType(i) {
			return true
		}
	}
	return false
}

// GetInterface returns a pointer to the requested interface if present, nil if not.
// For example: [...].GetInterface(HFAlert{}).(HFAlert)
func (hfu HanFun) GetInterface(i interface{}) interface{} {
	for in := range hfu.Units {
		if hfu.Units[in].IsUnitOfType(i) {
			return hfu.Units[in].Interface
		}
	}
	return nil
}

func (hfus HanFunUnits) String() string {
	var t = "["
	for _, u := range hfus {
		t += fmt.Sprint(u) + ", "
	}
	return t[:len(t)-2] + "]"
}

func (hfu HanFun) Name() string {
	return hfu.CapName
}

func (hfu HanFun) String() string {
	return fmt.Sprintf("%s: {Units: %s}", hfu.CapName, hfu.Units)
}

func (hfu HanFun) Device() *Device {
	return hfu.device
}

func (hfu *HanFun) unitFromJSON(m map[string]json.RawMessage, d *Device) (hf HanFunUnit, err error) {
	hu, err := hf.fromJSON(m, d)
	if err != nil {
		return
	}

	return hu, nil
}

func (hfu HanFun) fromJSON(_ map[string]json.RawMessage, d *Device) (Capability, error) {
	hfu.device = d
	return hfu, nil
}
