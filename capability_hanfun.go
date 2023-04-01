package go_fritzbox_api

import (
	"encoding/json"
	"fmt"
)

type HanFun struct {
	CapName string
	Units   HanFunUnits // A single HanFun-Device can potentially consist of multiple units
	device  *SmarthomeDevice
}

type HanFunUnits []*HanFunUnit

// Reload reloads the device itself and all its units.
func (hf *HanFun) Reload(c *Client) error {
	tt, err := getDeviceInfosFromCapability(c, hf)
	if err != nil {
		return err
	}

	// update current capability
	th := tt.(*HanFun)

	hf.CapName = th.CapName
	hf.Units = th.Units
	hf.device = th.device

	for i := range hf.Units {
		err = hf.Units[i].Reload(c)
		if err != nil {
			return err
		}
	}

	return nil
}

// HasInterface returns true, if the given HanFun-Interface is present in the current devices' units
// For example: [...].HasInterface(HFAlert{})
func (hf *HanFun) HasInterface(i interface{}) bool {
	for _, u := range hf.Units {
		if u.IsUnitOfType(i) {
			return true
		}
	}
	return false
}

// GetInterface returns a pointer to the requested interface if present, nil if not.
// For example: [...].GetInterface(HFAlert{}).(HFAlert)
func (hf *HanFun) GetInterface(i interface{}) interface{} {
	for in := range hf.Units {
		if hf.Units[in].IsUnitOfType(i) {
			return hf.Units[in].Interface
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

func (hf *HanFun) Name() string {
	return hf.CapName
}

func (hf *HanFun) String() string {
	return fmt.Sprintf("%s: {Units: %s}", hf.CapName, hf.Units)
}

func (hf *HanFun) Device() *SmarthomeDevice {
	return hf.device
}

func (hf *HanFun) unitFromJSON(m map[string]json.RawMessage, d *SmarthomeDevice) (hfu *HanFunUnit, err error) {
	hfu, err = hfu.fromJSON(m, d)
	if err != nil {
		return
	}

	return hfu, nil
}

func (hf *HanFun) fromJSON(_ map[string]json.RawMessage, d *SmarthomeDevice) (Capability, error) {
	hf.device = d
	return hf, nil
}
