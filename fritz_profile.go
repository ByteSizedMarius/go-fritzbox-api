package go_fritzbox_api

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type Profile struct {
	Name   string
	UID    string
	Filter string
}

// GetAvailableProfiles returns a map, where the Profile-Object is accessible via the profile-UID
func (c *Client) GetAvailableProfiles() (profiles map[string]Profile, err error) {
	if err = c.checkExpiry(); err != nil {
		return
	}

	data := url.Values{
		"sid": {c.session.Sid},
	}

	resp, err := c.doRequest(http.MethodGet, "internet/kids_profilelist.lua", data)
	if err != nil {
		return
	}

	body, err := getBody(resp)
	if err != nil {
		return
	}

	// get the table from the html
	body = body[strings.Index(body, "<table"):]
	body = body[:strings.Index(body, "<br>")-1]

	profiles = make(map[string]Profile)

	// parse the table
	// may be unstable, most effecient way I could find tho
	for strings.Contains(body, "class=\"name\"") {
		p := Profile{}

		// profile Name
		body = body[strings.Index(body, "class=\"name\"")+7:]
		body = body[strings.Index(body, "title=")+7:]
		p.Name = body[:strings.Index(body, "\"")]

		// filters
		body = body[strings.Index(body, "datalabel=\"Filter\"")+19:]
		p.Filter = strings.TrimSpace(body[:strings.Index(body, "<")])

		// profile UID
		body = body[strings.Index(body, "name=\"delete\"")+14:]
		body = body[strings.Index(body, "value=")+7:]
		p.UID = body[:strings.Index(body, "\"")]

		profiles[p.UID] = p
	}

	return
}

// GetProfileUIDFromDevice returns the UID of the profile that is assigned to the given device
// The return can be empty, for example for the fritzbox itself. This should be accounted for.
func (c *Client) GetProfileUIDFromDevice(deviceUID string) (profileUID string, err error) {
	if err = c.checkExpiry(); err != nil {
		return
	}

	data := url.Values{
		"sid":  {c.session.Sid},
		"page": {"edit_device"},
		"dev":  {deviceUID},
	}

	// todo check if theres some interesting info in this query
	resp, err := c.doRequest(http.MethodPost, "data.lua", data)
	if err != nil {
		return
	}

	body, err := getBody(resp)
	if err != nil {
		return
	}

	// the fritzbox does not have a profile
	if strings.Contains(body, "profiles") {

		body = getFrom(body, "profiles")
		body = getFromOffset(body, "selected\":\"", 11)
		body = getUntil(body, "devType")
		body = getUntil(body, "\"}}},\"")
		return body, nil
	} else {
		return "", nil
	}
}

// SetProfileForDevice (mainly untested) sets the profile from the profileUID to the device with the given deviceUID.
// Assigning the guest-Profile does not work when guest-wifi is off (makes sense), otherwise it might work
func (c *Client) SetProfileForDevice(deviceUID string, profileUID string) (err error) {
	if err = c.checkExpiry(); err != nil {
		return
	}

	data := url.Values{
		"sid":                   {c.session.Sid},
		"edit":                  {profileUID},
		"page":                  {"kids_profileedit"},
		"checkbox_" + deviceUID: {"on"},
		"apply":                 {""},
	}

	_, err = c.doRequest(http.MethodPost, "data.lua", data)
	if err != nil {
		return
	}

	body, err := c.getEditInfos(deviceUID)
	if err != nil {
		return
	}

	v, err := valueFromJson(body, []string{"data", "vars", "dev", "netAccess", "kisi", "profiles"})
	if v["selected"] != profileUID {
		err = fmt.Errorf("unknown error when applying profile")
	}
	return
}
