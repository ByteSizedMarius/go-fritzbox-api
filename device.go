package fritzbox

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func (c *Client) SetIP(deviceUID string, ip string) (err error) {
	data := url.Values{
		"dev_ip":    {ip},
		"dev":       {deviceUID},
		"apply":     {""},
		"sid":       {c.session.Sid},
		"page":      {"edit_device"},
		"confirmed": {""},
	}

	resp, err := c.doRequest(http.MethodPost, "data.lua", data)
	if err != nil {
		return
	}

	body, err := getBody(resp)
	if err != nil {
		return
	}

	if !strings.Contains(body, "\"ip\":\""+ip+"\",") {
		return fmt.Errorf("ip change unsuccessful")
	}

	return nil
}

func (c *Client) RenameDevice(deviceUID string, newName string) (err error) {
	// 1. get current deviceip (this is required due to a quirk in fritzbox's request data validation)
	data := url.Values{
		"sid":  {c.session.Sid},
		"page": {"edit_device"},
		"dev":  {deviceUID},
	}

	resp, err := c.doRequest(http.MethodPost, "data.lua", data)
	if err != nil {
		return
	}

	body, err := getBody(resp)
	if err != nil {
		return
	}

	var result map[string]interface{}
	err = json.Unmarshal([]byte(body), &result)
	if err != nil {
		if e, ok := err.(*json.SyntaxError); ok {
			log.Printf("syntax error at byte offset %d", e.Offset)
		}

		return err
	}

	// having finished this abomination, I feel like just string splitting would have been the cleaner solution
	ip := result["data"].(map[string]interface{})["vars"].(map[string]interface{})["dev"].(map[string]interface{})["ipv4"].(map[string]interface{})["current"].(map[string]interface{})["ip"]
	if ip == nil {
		return fmt.Errorf("couldn't find IP for device %s. Probably due to changes in fritzbox datatypes", deviceUID)
	}

	// 2. gsend the actual request for changing the devicename
	data = url.Values{
		"dev_name":  {newName},
		"dev_ip":    {ip.(string)},
		"dev":       {deviceUID},
		"apply":     {""},
		"sid":       {c.session.Sid},
		"page":      {"edit_device"},
		"confirmed": {""},
	}

	resp, err = c.doRequest(http.MethodPost, "data.lua", data)
	if err != nil {
		return
	}

	body, err = getBody(resp)
	if err != nil {
		return
	}

	if !strings.Contains(body, "\"name\":\""+newName+"\",") {
		return fmt.Errorf("name change not successful for unknown reason")
	}

	return nil
}
