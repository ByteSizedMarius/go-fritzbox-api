package unsafe

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/ByteSizedMarius/go-fritzbox-api"
)

// SetName sets the friendly name of a device.
func SetName(c *fritzbox.Client, deviceUID string, newName string) error {
	body := map[string]string{"friendly_name": newName}
	respBody, statusCode, err := c.RestPut("api/v0/generic/landevice/landevice/"+deviceUID, body)
	if err != nil {
		return err
	}
	if statusCode != http.StatusOK {
		return fmt.Errorf("failed to set name: HTTP %d: %s", statusCode, string(respBody))
	}
	return nil
}

// SetIP sets the DHCP reservation for a device's IP address.
// If static is true, the device will always get this address when it renews its DHCP lease.
//
// Note: This sets a reservation - the device's actual IP won't change until it
// renews its DHCP lease. The device UID may also change after the IP changes.
//
// SetIP may fail with HTTP 400 "bad value" for various reasons:
//   - IP is already in use by another device (including inactive/offline devices)
//   - IP is the router's own address or gateway
//   - IP is outside the DHCP pool range configured on the router
//   - IP is a broadcast address (e.g., .255)
//   - IP conflicts with internal reservations not visible via GetAllDevicesREST
func SetIP(c *fritzbox.Client, deviceUID string, ip string, static bool) error {
	staticDHCP := "0"
	if static {
		staticDHCP = "1"
	}

	body := map[string]string{
		"ip":          ip,
		"static_dhcp": staticDHCP,
	}

	respBody, statusCode, err := c.RestPut("api/v0/generic/landevice/landevice/"+deviceUID, body)
	if err != nil {
		return err
	}
	if statusCode != http.StatusOK {
		return fmt.Errorf("failed to set IP: HTTP %d: %s", statusCode, string(respBody))
	}
	return nil
}

// GetDeviceName returns the name of a device based on its UID.
func GetDeviceName(c *fritzbox.Client, deviceUID string) (name string, err error) {
	devices, err := queryDeviceFields(c, "friendly_name")
	if err != nil {
		return
	}
	for _, d := range devices {
		if d.UID == deviceUID {
			return d.FriendlyName, nil
		}
	}
	return "", errors.New("device not found")
}

// GetDeviceIP returns the IP address of a device based on its UID.
func GetDeviceIP(c *fritzbox.Client, deviceUID string) (ip string, err error) {
	devices, err := queryDeviceFields(c, "ip")
	if err != nil {
		return
	}
	for _, d := range devices {
		if d.UID == deviceUID {
			return d.IP, nil
		}
	}
	return "", errors.New("device not found")
}

// queryDeviceFields fetches all devices with the specified fields via query.lua
func queryDeviceFields(c *fritzbox.Client, fields ...string) (devices []struct {
	UID          string `json:"UID"`
	FriendlyName string `json:"friendly_name"`
	IP           string `json:"ip"`
}, err error) {
	fieldList := "UID"
	for _, f := range fields {
		fieldList += "," + f
	}

	data := fritzbox.Values{
		"sid":        c.SID(),
		"mq_devices": "landevice:settings/landevice/list(" + fieldList + ")",
	}

	_, resp, err := c.AhaRequestString(http.MethodGet, "query.lua", data)
	if err != nil {
		return
	}

	var r struct {
		Devices []struct {
			UID          string `json:"UID"`
			FriendlyName string `json:"friendly_name"`
			IP           string `json:"ip"`
		} `json:"mq_devices"`
	}
	if err = json.Unmarshal([]byte(resp), &r); err != nil {
		return
	}
	devices = r.Devices
	return
}

// GetAllDevices returns all devices including inactive ones.
// Useful for finding DHCP reservations and offline devices.
func GetAllDevices(c *fritzbox.Client) (devices []Device, err error) {
	respBody, statusCode, err := c.RestGet("api/v0/generic/landevice")
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get devices: HTTP %d", statusCode)
	}

	var r struct {
		Landevice []struct {
			UID          string `json:"UID"`
			IP           string `json:"ip"`
			MAC          string `json:"mac"`
			Name         string `json:"name"`
			FriendlyName string `json:"friendly_name"`
			Active       string `json:"active"`
		} `json:"landevice"`
	}
	if err = json.Unmarshal(respBody, &r); err != nil {
		return nil, err
	}

	devices = make([]Device, 0, len(r.Landevice))
	for _, d := range r.Landevice {
		devices = append(devices, Device{
			UID:          d.UID,
			IP:           d.IP,
			MAC:          d.MAC,
			Name:         d.Name,
			FriendlyName: d.FriendlyName,
			Active:       d.Active == "1",
		})
	}
	return devices, nil
}
