package aha

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/ByteSizedMarius/go-fritzbox-api"
)

var maskTranslation = []string{CHanfun, "", CLicht, "", CAlarm, CButton, CHKR, CEnergieMesser, CTempSensor, CSteckdose, CRepeater, CMikrofon, "", CHanfunUnit, "", CSchaltbar, CDimmbar, CLampeMitFarbtemp, CRollladen}

// Capability is the interface for all capabilities.
// It is self-contained in that it has a reference to the device it belongs to.
type Capability interface {
	Name() string
	String() string
	Device() *Device
	fromJSON(m map[string]json.RawMessage, d *Device) (Capability, error)
}

// fromBitmask puts empty capability-structs into the capability map where applicable.
// The bitmask is used to decide which capabilities are available for a device.
func capabilitiesFromBitmask(bitmask string) (cap Capabilities) {
	cap = make(Capabilities)

	bmI, err := strconv.Atoi(bitmask)
	if err != nil {
		return
	}

	bitRepr := strconv.FormatInt(int64(bmI), 2)

	capabilitiesMap := map[string]Capability{
		CHanfun: &HanFun{CapName: CHanfun},
		// CLicht
		// CAlarm
		CButton: &ButtonDevice{CapName: CButton},
		CHKR:    &Hkr{CapName: CHKR},
		// CEnergieMesser
		CTempSensor: &Temperature{CapName: CTempSensor},
		// CSteckdose
		// CRepeater
		// CMikrofon
		// CSchaltbar
		// CDimmbar
		// CLampeMitFarbtemp
		// CRollladen
	}

	for i := len(bitRepr) - 1; i >= 0; i-- {
		if bitRepr[i] == '1' {
			ind := len(bitRepr) - i - 1
			if ind >= len(maskTranslation) {
				continue
			}
			if capability, ok := capabilitiesMap[maskTranslation[ind]]; ok {
				cap[maskTranslation[ind]] = capability
			}
		}
	}

	return cap
}

// dectGetter is a helper function to get the state of a DECT device
func dectGetter(c *fritzbox.Client, req string, ca Capability) (string, error) {
	data := fritzbox.Values{
		"sid":       c.SID(),
		"ain":       ca.Device().Identifier,
		"switchcmd": req,
	}

	_, resp, err := c.AhaRequestString(http.MethodGet, "webservices/homeautoswitch.lua", data)
	if err != nil {
		return "", err
	}

	return resp, nil
}
