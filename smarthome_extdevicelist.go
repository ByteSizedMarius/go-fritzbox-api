package go_fritzbox_api

import (
	"encoding/json"
	"github.com/clbanning/mxj/v2"
)

// extDeviceList is used for easily parsing the fritzbox's xml into a go-struct, then transforming it usable structs (Devicelist, Device)
type extDevicelist struct {
	Version   string      `json:"-version"`
	Fwversion string      `json:"-fwversion"`
	Device    []extDevice `json:"device"`
}

type extDevice struct {
	Functionbitmask string `json:"-functionbitmask"`
	Fwversion       string `json:"-fwversion"`
	ID              string `json:"-id"`
	Identifier      string `json:"-identifier"`
	Manufacturer    string `json:"-manufacturer"`
	Productname     string `json:"-productname"`
	Name            string `json:"name"`
	Present         string `json:"present"`
	Txbusy          string `json:"txbusy"`
}

func (extDevicelist) fromBytes(b []byte) (dlt extDevicelist, err error) {
	mv, err := mxj.NewMapXml(b)
	if err != nil {
		return
	}

	j, err := mv.Json(true)
	if err != nil {
		return
	}

	tmp := map[string]json.RawMessage{}
	err = json.Unmarshal(j, &tmp)
	if err != nil {
		return
	}

	err = json.Unmarshal(tmp["devicelist"], &dlt)
	return
}
