package unsafe

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ByteSizedMarius/go-fritzbox-api"
)

// Profile is the internal representation of a profile in the Fritz!Box API.
type Profile struct {
	Name   string
	UID    string
	Filter string
}

// GetAvailableProfiles returns a map, where the Profile-Object is accessible via the profile-UID
func GetAvailableProfiles(c *fritzbox.Client) (profiles map[string]Profile, err error) {
	data := fritzbox.Values{
		"sid":         c.SID(),
		"mq_profiles": "filter_profile:settings/profile/list(UID,name,blacklist_enabled,whitelist_enabled,bpjm_filter_enabled)",
	}

	_, resp, err := c.AhaRequestString(http.MethodGet, "query.lua", data)
	if err != nil {
		return
	}

	var raw struct {
		Profiles []struct {
			UID              string `json:"UID"`
			Name             string `json:"name"`
			BlacklistEnabled string `json:"blacklist_enabled"`
			WhitelistEnabled string `json:"whitelist_enabled"`
			BpjmEnabled      string `json:"bpjm_filter_enabled"`
		} `json:"mq_profiles"`
	}
	if err = json.Unmarshal([]byte(resp), &raw); err != nil {
		return
	}

	profiles = make(map[string]Profile)
	for _, p := range raw.Profiles {
		filter := buildFilterString(p.BlacklistEnabled, p.WhitelistEnabled, p.BpjmEnabled)
		profiles[p.UID] = Profile{
			UID:    p.UID,
			Name:   p.Name,
			Filter: filter,
		}
	}

	return
}

func buildFilterString(blacklist, whitelist, bpjm string) string {
	var filters []string
	if blacklist == "1" {
		filters = append(filters, "Blacklist")
	}
	if whitelist == "1" {
		filters = append(filters, "Whitelist")
	}
	if bpjm == "1" {
		filters = append(filters, "BPjM")
	}
	if len(filters) == 0 {
		return "None"
	}
	return strings.Join(filters, ", ")
}

// GetProfileUIDFromDevice returns the UID of the profile that is assigned to the given device.
// Returns empty string for devices without profiles (e.g., the Fritz!Box itself).
func GetProfileUIDFromDevice(c *fritzbox.Client, deviceUID string) (profileUID string, err error) {
	body, err := getEditInfos(c, deviceUID)
	if err != nil {
		return
	}

	v, err := fritzbox.ValueFromJsonPath(body, []string{"data", "vars", "dev", "netAccess", "kisi", "profiles"})
	if err != nil {
		// Device doesn't have profile info (e.g., the router itself)
		return "", nil
	}

	if selected, ok := v["selected"].(string); ok {
		profileUID = selected
	}
	return
}

// SetProfileForDevice sets the profile from the profileUID to the device with the given deviceUID.
// Note: Assigning the guest profile does not work when guest WiFi is off.
func SetProfileForDevice(c *fritzbox.Client, deviceUID string, profileUID string) (err error) {
	data := fritzbox.Values{
		"sid":                   c.SID(),
		"edit":                  profileUID,
		"page":                  "kids_profileedit",
		"checkbox_" + deviceUID: "on",
		"apply":                 "",
	}

	_, _, err = c.AhaRequestString(http.MethodPost, "data.lua", data)
	return
}

// getEditInfos returns the JSON response body from querying the edit_device page for the given deviceUID.
func getEditInfos(c *fritzbox.Client, deviceUID string) (body string, err error) {
	data := fritzbox.Values{
		"dev":  deviceUID,
		"sid":  c.SID(),
		"page": "edit_device",
	}

	_, body, err = c.AhaRequestString(http.MethodPost, "data.lua", data)
	if err != nil {
		return
	}

	return
}
