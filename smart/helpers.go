package smart

import "github.com/ByteSizedMarius/go-fritzbox-api/v2/rest"

// findDevice finds a device by UID in the overview devices list.
func findDevice(devices []rest.HelperOverviewDevice, uid string) *rest.HelperOverviewDevice {
	for i := range devices {
		if devices[i].UID == uid {
			return &devices[i]
		}
	}
	return nil
}

// derefBool safely dereferences a bool pointer, returning false if nil.
func derefBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

// derefInt safely dereferences an int pointer, returning 0 if nil.
func derefInt(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

// derefFloat32 safely dereferences a float32 pointer, returning 0 if nil.
func derefFloat32(f *float32) float32 {
	if f == nil {
		return 0
	}
	return *f
}
