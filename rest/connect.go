package rest

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ByteSizedMarius/go-fritzbox-api/v2"
)

const (
	connectRadioBasesPath        = "api/v0/smarthome/connect/radioBases"
	connectSubscriptionStatePath = "api/v0/smarthome/connect/subscriptionState"
	connectStartSubscriptionPath = "api/v0/smarthome/connect/startSubscription"
	connectStopSubscriptionPath  = "api/v0/smarthome/connect/stopSubscription"
	connectResetCodePath         = "api/v0/smarthome/connect/resetCode"
	connectInstallCodePath       = "api/v0/smarthome/connect/installCode"
)

// GetRadioBasesList returns list of radioBases.
//
// RadioBases provide DECT, Zigbee or both - they are your smart home gateway.
// In a smart home mesh, only the master will provide a full list of available radioBases.
// All others will only provide information about themselves.
func GetRadioBasesList(c *fritzbox.Client) ([]EndpointRadioBases, error) {
	body, status, err := c.RestGet(connectRadioBasesPath)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(body))
	}

	var result []EndpointRadioBases
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return result, nil
}

// GetRadioBaseBySerial returns a radioBase by its serial number.
//
// Serial is based on MAC-address of the radioBase.
func GetRadioBaseBySerial(c *fritzbox.Client, serial string) (*EndpointRadioBases, error) {
	body, status, err := c.RestGet(fmt.Sprintf("%s/%s", connectRadioBasesPath, serial))
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(body))
	}

	var result EndpointRadioBases
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return &result, nil
}

// GetSubscriptionStateByUID returns subscription state by UID.
//
// Local subscription and deletion of smart home devices is always available.
// Remote subscriptions and deletions are only available on the smart home master.
func GetSubscriptionStateByUID(c *fritzbox.Client, uid string) (*EndpointSubscriptionState, error) {
	body, status, err := c.RestGet(fmt.Sprintf("%s/%s", connectSubscriptionStatePath, uid))
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(body))
	}

	var result EndpointSubscriptionState
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return &result, nil
}

// PostStartSubscriptionBySerial starts device subscription on a radioBase.
//
// After starting a subscription, new devices can be paired with the FRITZ!Box.
// Local subscription is always available.
// Remote subscriptions are only available on the smart home master.
func PostStartSubscriptionBySerial(c *fritzbox.Client, serial string, data *EndpointStartSubscription) (*StartSubscriptionResponse, error) {
	body, status, err := c.RestPost(fmt.Sprintf("%s/%s", connectStartSubscriptionPath, serial), data)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(body))
	}

	var result StartSubscriptionResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	return &result, nil
}

// PostStopSubscriptionBySerial stops device subscription on a radioBase.
func PostStopSubscriptionBySerial(c *fritzbox.Client, serial string) error {
	body, status, err := c.RestPost(fmt.Sprintf("%s/%s", connectStopSubscriptionPath, serial), nil)
	if err != nil {
		return err
	}
	if status != http.StatusOK && status != http.StatusNoContent {
		return fmt.Errorf("unexpected status %d: %s", status, string(body))
	}
	return nil
}

// PostResetCodeBySerial sets a reset code on a Zigbee radioBase.
//
// This is a Zigbee-specific reset code for unpairing devices.
func PostResetCodeBySerial(c *fritzbox.Client, serial string, data *EndpointResetCode) error {
	body, status, err := c.RestPost(fmt.Sprintf("%s/%s", connectResetCodePath, serial), data)
	if err != nil {
		return err
	}
	if status != http.StatusOK && status != http.StatusNoContent {
		return fmt.Errorf("unexpected status %d: %s", status, string(body))
	}
	return nil
}

// PostInstallCodeBySerial sets an install code on a Zigbee radioBase.
//
// This is a Zigbee-specific installation code for secure pairing.
func PostInstallCodeBySerial(c *fritzbox.Client, serial string, data *EndpointInstallCode) error {
	body, status, err := c.RestPost(fmt.Sprintf("%s/%s", connectInstallCodePath, serial), data)
	if err != nil {
		return err
	}
	if status != http.StatusOK && status != http.StatusNoContent {
		return fmt.Errorf("unexpected status %d: %s", status, string(body))
	}
	return nil
}
