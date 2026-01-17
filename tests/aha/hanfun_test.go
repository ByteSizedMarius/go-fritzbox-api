package aha

import (
	"testing"

	"github.com/ByteSizedMarius/go-fritzbox-api/v2/aha"
)

func TestHanFun(t *testing.T) {
	if skipHanfun {
		t.Skip("HanFun tests disabled")
	}

	t.Run("String", HanfunString)
	t.Run("HasInterface", HanfunHasInterface)
	t.Run("GetInterface", HanfunGetInterface)
	t.Run("Units", HanfunUnits)
	t.Run("UnitETSIInfo", HanfunUnitETSIInfo)
	t.Run("Reload", HanfunReload)
}

func HanfunString(t *testing.T) {
	s := hanfun.String()
	if s == "" {
		t.Fatal("HanFun.String() returned empty")
	}
	t.Logf("HanFun: %s", s)
}

func HanfunHasInterface(t *testing.T) {
	if !hanfun.HasInterface(aha.SimpleButton{}) {
		t.Error("Expected device to have SimpleButton interface")
	}
	t.Logf("HasInterface(SimpleButton): true")
}

func HanfunGetInterface(t *testing.T) {
	iface := hanfun.GetInterface(aha.SimpleButton{})
	if iface == nil {
		t.Fatal("GetInterface(SimpleButton) returned nil")
	}

	_, ok := iface.(aha.SimpleButton)
	if !ok {
		t.Errorf("GetInterface returned wrong type: %T", iface)
	}
	t.Logf("GetInterface(SimpleButton): %v", iface)
}

func HanfunUnits(t *testing.T) {
	if len(hanfun.Units) == 0 {
		t.Fatal("No units found on HanFun device")
	}

	t.Logf("Device has %d units", len(hanfun.Units))
	for i, unit := range hanfun.Units {
		s := unit.String()
		if s == "" {
			t.Errorf("Unit %d String() returned empty", i)
		}
		t.Logf("Unit %d: %s", i, s)
	}
}

func HanfunUnitETSIInfo(t *testing.T) {
	if len(hanfun.Units) == 0 {
		t.Skip("No units to test")
	}

	for i, unit := range hanfun.Units {
		info := unit.ETSIUnitInfo
		ifaceStr := info.GetInterfaceString()
		unitStr := info.GetUnitString()

		t.Logf("Unit %d ETSI: DeviceID=%s, Interface=%s (%s), UnitType=%s (%s)",
			i, info.ETSIDeviceID, ifaceStr, info.Interface, unitStr, info.UnitType)

		if info.ETSIDeviceID == "" {
			t.Errorf("Unit %d has empty ETSIDeviceID", i)
		}
	}
}

func HanfunReload(t *testing.T) {
	err := hanfun.Reload(cl)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Reload successful")
}
