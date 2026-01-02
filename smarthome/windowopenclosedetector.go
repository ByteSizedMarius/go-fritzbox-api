package smarthome

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ByteSizedMarius/go-fritzbox-api"
)

// GetWindowOpenCloseDetectorConfig returns typed open window configuration for a unit.
func GetWindowOpenCloseDetectorConfig(c *fritzbox.Client, uid string) (*WindowOpenCloseDetectorUnit, error) {
	body, status, err := c.RestGet(configUnitPath + uid)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(body))
	}

	var unit WindowOpenCloseDetectorUnit
	if err := json.Unmarshal(body, &unit); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return &unit, nil
}
