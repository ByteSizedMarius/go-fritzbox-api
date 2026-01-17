package smart

import (
	"time"

	"github.com/ByteSizedMarius/go-fritzbox-api/v2"
	"github.com/ByteSizedMarius/go-fritzbox-api/v2/rest"
)

// Button represents a FRITZ!Box button with clean Go types.
type Button struct {
	UID         string
	AIN         string
	Name        string
	IsConnected bool

	// Last press timestamp
	LastPressTime time.Time

	// Battery (from parent device)
	BatteryLevel int
	IsBatteryLow bool
}

// ButtonConfig contains configuration data from the configuration endpoint.
type ButtonConfig struct {
	ControlMode    string // on/off/toggle/event
	DestinationMode string // disabled/templates/units
	DestinationUids []string

	SwitchDuration *SwitchDuration
	ActivePeriod   *ButtonActivePeriod

	AvailableUnits     []string
	AvailableTemplates []string
}

// SwitchDuration configures how long a switch action lasts.
type SwitchDuration struct {
	Mode           string // permanent/toggleBack
	ToggleBackTime int    // minutes, only for toggleBack mode
}

// ButtonActivePeriod configures when button presses are registered.
type ButtonActivePeriod struct {
	Mode      string    // permanent/fixed/astronomic
	StartTime time.Time // for fixed mode
	EndTime   time.Time
}

// GetAllButtons returns all buttons with clean Go types.
func GetAllButtons(c *fritzbox.Client) ([]Button, error) {
	overview, err := rest.GetOverview(c)
	if err != nil {
		return nil, err
	}

	units := overview.Buttons()
	buttons := make([]Button, 0, len(units))

	for _, unit := range units {
		device := findDevice(overview.Devices, unit.ParentUid)
		buttons = append(buttons, buttonFromOverview(unit, device))
	}

	return buttons, nil
}

// GetButton returns a single button by UID/AIN.
func GetButton(c *fritzbox.Client, uid string) (*Button, error) {
	overview, err := rest.GetOverview(c)
	if err != nil {
		return nil, err
	}

	for _, unit := range overview.Buttons() {
		if unit.UID == uid || unit.Ain == uid {
			device := findDevice(overview.Devices, unit.ParentUid)
			b := buttonFromOverview(unit, device)
			return &b, nil
		}
	}

	return nil, ErrNotFound
}

func buttonFromOverview(unit rest.HelperOverviewUnit, device *rest.HelperOverviewDevice) Button {
	b := Button{
		UID:         unit.UID,
		AIN:         unit.Ain,
		Name:        string(unit.Name),
		IsConnected: derefBool(unit.IsConnected),
	}

	// Button interface data
	if btn := unit.Interfaces.ButtonInterface; btn != nil {
		if btn.LastEventTime != nil {
			b.LastPressTime = time.Unix(int64(*btn.LastEventTime), 0)
		}
	}

	// Battery info from device
	if device != nil {
		b.BatteryLevel = derefInt(device.BatteryValue)
		b.IsBatteryLow = derefBool(device.IsBatteryLow)
	}

	return b
}

// ButtonHandle provides a fluent API for button operations.
// Handles are stateless and safe to reuse or recreate as needed.
type ButtonHandle struct {
	client *fritzbox.Client
	uid    string
}

// NewButtonHandle creates a ButtonHandle for the given button UID.
func NewButtonHandle(c *fritzbox.Client, uid string) *ButtonHandle {
	return &ButtonHandle{client: c, uid: uid}
}
