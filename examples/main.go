package main

import (
	"fmt"
	"github.com/ByteSizedMarius/go-fritzbox-api"
	"log"
	"net/http"
	"net/url"
)

func main() {
	// Connection to fritzbox
	c := go_fritzbox_api.Client{
		Username: "username",
		Password: "password",
	}

	err := c.Initialize()
	checkError(err)

	// devicelist
	fmt.Println("Currently connected devices:")
	devices, err := c.GetCLientList()
	for _, d := range devices.Devices {
		fmt.Println(d.Nameinfo.Name)
	}
	fmt.Println()

	// Profiles
	// Get all available profiles
	profiles, err := c.GetAvailableProfiles()
	checkError(err)
	fmt.Println("Available profiles:")
	for _, prof := range profiles {
		fmt.Println(prof)
	}
	fmt.Println()

	// Get the profile uid from a specific device
	dev := devices.Devices[1]
	devProfileUID, err := c.GetProfileUIDFromDevice(dev.UID)
	checkError(err)
	fmt.Println("Profile UID of device " + dev.Nameinfo.Name + ": " + devProfileUID)
	fmt.Println("Profile: " + fmt.Sprint(profiles[devProfileUID]))

	// Set the profile of a specific device
	err = c.SetProfileForDevice(dev.UID, "filtprof3")
	checkError(err)
	fmt.Println()

	// Get all currently available eventlogs
	l, err := c.GetEventLog()
	if err != nil {
		log.Fatal(err)
	}

	// get all eventlogs newer than logmessage
	l, err = c.GetEventLogUntil(l[len(l)/2].ID)

	for _, lo := range l {
		fmt.Println(lo)
	}

	// custom POST request
	// add entry to default phone book
	data := url.Values{
		"sid":            {c.SID()},
		"entryname":      {"My Test Entry"},
		"numbertypenew1": {"home"},
		"numbernew1":     {"666"},
		"prionumber":     {"none"},
		"apply":          {""},
	}

	status, body, err := c.CustomRequest(http.MethodPost, "/fon_num/fonbook_entry.lua", data)
	checkError(err)
	fmt.Println(status)
	//fmt.Println(body)

	// custom GET request
	// get missed calls
	data = url.Values{
		"sid": {c.SID()},
		"csv": {""},
	}
	status, body, err = c.CustomRequest(http.MethodGet, "/fon_num/foncalls_list.lua", data)
	checkError(err)
	fmt.Println(status)
	fmt.Println(body)

	// Open a second session in parallel
	// Connection to a Fritz!Repeater
	// FRITZ!Repeater 1200: User field is ignored when logging in (can be anything)
	clRep := go_fritzbox_api.Client{
		BaseUrl:  "http://192.168.178.28",
		Username: "",
		Password: "password",
	}

	err = clRep.Initialize()
	checkError(err)

	l, err = clRep.GetEventLog()
	for _, lo := range l {
		fmt.Println(lo)
	}
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
