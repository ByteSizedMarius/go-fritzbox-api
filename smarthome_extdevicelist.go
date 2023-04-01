package go_fritzbox_api

import (
	"encoding/xml"
)

// extDeviceList is used for easily parsing the fritzbox's xml into a go-struct, then transforming it usable structs (Devicelist, Device)
type extDevicelist struct {
	XMLName   xml.Name    `xml:"devicelist"`
	Text      string      `xml:",chardata"`
	Version   string      `xml:"version,attr"`
	Fwversion string      `xml:"fwversion,attr"`
	Device    []extDevice `xml:"device"`
}

type extDevice struct {
	Text            string `xml:",chardata"`
	Identifier      string `xml:"identifier,attr"`
	ID              string `xml:"id,attr"`
	Functionbitmask string `xml:"functionbitmask,attr"`
	Fwversion       string `xml:"fwversion,attr"`
	Manufacturer    string `xml:"manufacturer,attr"`
	Productname     string `xml:"productname,attr"`
	Present         string `xml:"present"`
	Txbusy          string `xml:"txbusy"`
	Name            string `xml:"Name"`
}

func (extDevicelist) fromBytes(b []byte) (dlt extDevicelist, err error) {
	err = xml.Unmarshal(b, &dlt)
	return
}
