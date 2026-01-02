// Window detector support based on contribution by @btwotch
// https://github.com/ByteSizedMarius/go-fritzbox-api/pull/1

package smart

import (
	"time"

	"github.com/ByteSizedMarius/go-fritzbox-api"
	"github.com/ByteSizedMarius/go-fritzbox-api/rest"
)

// WindowDetector represents a window open/close detector with clean Go types.
type WindowDetector struct {
	UID         string
	AIN         string
	Name        string
	IsConnected bool

	// Alert state
	IsOpen        bool
	LastAlertTime time.Time

	// Battery (from parent device)
	BatteryLevel int
	IsBatteryLow bool
}

// GetAllWindowDetectors returns all window detectors with clean Go types.
func GetAllWindowDetectors(c *fritzbox.Client) ([]WindowDetector, error) {
	overview, err := rest.GetOverview(c)
	if err != nil {
		return nil, err
	}

	units := overview.WindowDetectors()
	detectors := make([]WindowDetector, 0, len(units))

	for _, unit := range units {
		device := findDevice(overview.Devices, unit.ParentUid)
		detectors = append(detectors, windowDetectorFromOverview(unit, device))
	}

	return detectors, nil
}

// GetWindowDetector returns a single window detector by UID/AIN.
func GetWindowDetector(c *fritzbox.Client, uid string) (*WindowDetector, error) {
	overview, err := rest.GetOverview(c)
	if err != nil {
		return nil, err
	}

	for _, unit := range overview.WindowDetectors() {
		if unit.UID == uid || unit.Ain == uid {
			device := findDevice(overview.Devices, unit.ParentUid)
			d := windowDetectorFromOverview(unit, device)
			return &d, nil
		}
	}

	return nil, ErrNotFound
}

func windowDetectorFromOverview(unit rest.HelperOverviewUnit, device *rest.HelperOverviewDevice) WindowDetector {
	d := WindowDetector{
		UID:         unit.UID,
		AIN:         unit.Ain,
		Name:        string(unit.Name),
		IsConnected: derefBool(unit.IsConnected),
	}

	if alert := unit.Interfaces.AlertInterface; alert != nil {
		d.IsOpen = parseAlertState(alert.Alerts)
		if alert.LastAlertTime != nil {
			d.LastAlertTime = time.Unix(int64(*alert.LastAlertTime), 0)
		}
	}

	if device != nil {
		d.BatteryLevel = derefInt(device.BatteryValue)
		d.IsBatteryLow = derefBool(device.IsBatteryLow)
	}

	return d
}

func parseAlertState(alerts *[]rest.IFAlertOverview_Alerts_Item) bool {
	if alerts == nil {
		return false
	}
	for _, alert := range *alerts {
		alertType, err := alert.AsTypeAlertTypeDefinitions()
		if err != nil {
			continue
		}
		if alertType == rest.Open {
			return true
		}
	}
	return false
}

// WindowDetectorHandle provides a fluent API for window detector operations.
type WindowDetectorHandle struct {
	client *fritzbox.Client
	uid    string
}

// NewWindowDetectorHandle creates a WindowDetectorHandle for the given detector UID.
func NewWindowDetectorHandle(c *fritzbox.Client, uid string) *WindowDetectorHandle {
	return &WindowDetectorHandle{client: c, uid: uid}
}
