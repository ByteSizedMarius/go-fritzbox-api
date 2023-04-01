package go_fritzbox_api

import (
	"encoding/json"
	"fmt"
	"github.com/clbanning/mxj/v2"
	"net/http"
	"net/url"
	"strconv"
)

// Capabilities is a map of all available capabilities.
// They can be access using the Capability-Constants (starting with C, for example CHKR -> HeizungsKörperRegler, etc.)
// HasCapability can be used to check whether a device has a certain capability, without checking the map-keys.
type Capabilities map[string]Capability

type Capability interface {
	Name() string
	String() string
	Device() *SmarthomeDevice
	fromJSON(m map[string]json.RawMessage, d *SmarthomeDevice) (Capability, error)
}

func (c Capabilities) String() string {
	rt := "["
	for _, cp := range c {
		if cp != nil {
			rt += cp.Name() + ", "
		}
	}

	if len(rt) > 2 {
		rt = rt[:len(rt)-2]
	}

	return rt + "]"
}

func capabilityMapFromString(s string) (m map[string]json.RawMessage, err error) {
	// to xml
	mv, err := mxj.NewMapXml([]byte(s))
	if err != nil {
		return
	}

	// to json
	j, err := mv.Json(true)
	if err != nil {
		return
	}

	// map
	tmp := map[string]json.RawMessage{}
	err = json.Unmarshal(j, &tmp)
	if err != nil {
		return
	}

	// get rid of parent
	err = json.Unmarshal(tmp["device"], &m)
	if err != nil {
		return
	}

	return
}

func (c Capabilities) fromBitmask(bitmask string) (Capabilities, error) {
	var err error

	// bitmask to int
	bmI, err := strconv.Atoi(bitmask)
	if err != nil {
		return c, err
	}

	// int to bit representation in string
	bitRepr := strconv.FormatInt(int64(bmI), 2)

	// iterate backwards, map capabilities
	for i := len(bitRepr) - 1; i >= 0; i-- {
		if string(bitRepr[i]) == "1" {
			ind := len(bitRepr) - i - 1
			c[MaskTranslStr[ind]] = MaskTransl[ind]
		}
	}

	return c, err
}

func getDeviceInfosFromCapability(c *Client, ca Capability) (Capability, error) {
	resp, err := getDeviceInfos(c, ca.Device())
	if err != nil {
		return ca, err
	}

	newCapa, err := capabilityMapFromString(resp)
	if err != nil {
		return ca, err
	}

	// get & return capability struct
	return ca.fromJSON(newCapa, ca.Device())
}

func getDeviceInfos(c *Client, d *SmarthomeDevice) (string, error) {
	data := url.Values{
		"sid":       {c.SID()},
		"ain":       {d.Identifier},
		"switchcmd": {"getdeviceinfos"},
	}

	// get response
	code, resp, err := c.CustomRequest(http.MethodGet, "webservices/homeautoswitch.lua", data)
	if err != nil {
		return "", err
	}
	if code != 200 {
		return "", fmt.Errorf("unknown error: " + resp)
	}

	return resp, nil
}

func dectGetter(c *Client, req string, ca Capability) (string, error) {
	data := url.Values{
		"sid":       {c.SID()},
		"ain":       {ca.Device().Identifier},
		"switchcmd": {req},
	}

	code, resp, err := c.CustomRequest(http.MethodGet, "webservices/homeautoswitch.lua", data)
	if err != nil {
		return "", err
	} else if code != 200 {
		return "", fmt.Errorf("unknown error: " + resp)
	}

	return resp, nil
}
