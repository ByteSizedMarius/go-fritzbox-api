package main

// quickly test most of the endpoints used

import (
	"fmt"
	fb "github.com/ByteSizedMarius/go-fritzbox-api"
	"os"
	"strings"
	"time"
)

var cl fb.Client

func sep(in string) string {
	return "—————————————————————————————————" + in + "—————————————————————————————————"
}

func main() {
	initialize()
	fritzClientlist()
	fritzClientlistSetters()
	fritzProfiles()
	fritzLogs()
	fritzStatistics()
	smarthomeDevicelist()
}

func initialize() {
	data, err := os.ReadFile("./credentials")
	panicIfError(err)

	creds := strings.Split(string(data), "\r\n")
	cl = fb.Client{
		Username: creds[0],
		Password: creds[1],
		BaseUrl:  "http://192.168.178.1",
	}
	err = cl.Initialize()
	panicIfError(err)
}

func fritzClientlist() {
	fmt.Println(sep("Client-List"))

	clientList, err := cl.GetCLientList()
	panicIfError(err)

	fmt.Printf("[+] Got %d clients: [%s (%s)", len(clientList.Devices), clientList.Devices[0].Nameinfo.Name, clientList.Devices[0].Ipinfo)
	for _, v := range clientList.Devices[1:] {
		fmt.Printf(", %s (%s, %s)", v.Nameinfo.Name, v.Ipinfo, v.UID)
	}
	fmt.Print("]\n[+] Getting Profiles (this may take a while)...\n")

	cl.AddProfiles(&clientList)
	var str string
	var ct int
	for _, v := range clientList.Devices {
		if v.Profile == (fb.Profile{}) {
			continue
		}
		ct += 1
		str += fmt.Sprintf("%s (%s, %s), ", v.Nameinfo.Name, v.Profile.Name, v.Profile.UID)
	}
	fmt.Printf("[+] Found Profiles for %d clients: [%s]\n", ct, str[:len(str)-2])
}

func fritzClientlistSetters() {
	fmt.Println(sep("Client Setters"))

	// hardcoded because I can't just set the IPs of random devices in my network
	prevIP := "192.168.178.20"
	newIP := "192.168.178.248"

	prevName := "Android-4"
	newName := "Android-5"

	deviceUID := "landevice6511"

	err := cl.SetIP(deviceUID, newIP, true)
	panicIfError(err)
	fmt.Println("[+] Set IP successfully. Resetting...")
	err = cl.SetIP(deviceUID, prevIP, false)
	panicIfError(err)

	err = cl.SetName(deviceUID, newName)
	panicIfError(err)
	fmt.Println("[+] Set Name successfully. Resetting...")
	err = cl.SetName(deviceUID, prevName)
	panicIfError(err)
}

func fritzProfiles() {
	deviceUID := "landevice6511"
	prevFilter := "filtprof1"
	newFilter := "filtprof3"
	fmt.Println(sep("Profiles"))

	profiles, err := cl.GetAvailableProfiles()
	panicIfError(err)

	fmt.Printf("[+] Got %d profiles: [", len(profiles))
	var str string
	for _, v := range profiles {
		str += fmt.Sprintf("%s (%s, %s), ", v.Name, v.UID, v.Filter)
	}
	fmt.Println(str[:len(str)-2] + "]")

	uid, err := cl.GetProfileUIDFromDevice(deviceUID)
	panicIfError(err)
	fmt.Printf("[+] Got UID %s from device %s\n", uid, deviceUID)

	err = cl.SetProfileForDevice(deviceUID, newFilter)
	panicIfError(err)

	fmt.Println("[+] Set Profile successfully. Resetting...")
	err = cl.SetProfileForDevice(deviceUID, prevFilter)
	panicIfError(err)
}

func fritzLogs() {
	fmt.Println(sep("Logs"))

	logs, err := cl.GetEventLog()
	panicIfError(err)
	fmt.Printf("[+] Got %d log messages\n", len(logs))

	logsUntil, err := cl.GetEventLogUntil(logs[len(logs)/2].ID)
	fmt.Printf("[+] Logs Until: Got %d log messages\n", len(logsUntil))
}

func fritzStatistics() {
	fmt.Println(sep("Statistics"))

	stats, err := cl.GetTrafficStats()
	panicIfError(err)
	fmt.Printf("[+] Got Statistics. MB sent today: %d, MB received today: %d\n", stats.Today.MBSent, stats.Today.MBReceived)
}

func smarthomeDevicelist() {
	fmt.Println(sep("Smarthome-Devicelist"))
	dl, err := cl.GetSmarthomeDevices()
	panicIfError(err)

	fmt.Printf("[+] Got %d devices: [", len(dl.Devices))
	var str string
	for _, v := range dl.Devices {
		str += fmt.Sprintf("%s (%s, %s), ", v.Name, v.Identifier, v.Capabilities)
	}
	fmt.Println(str[:len(str)-2] + "]")

	// Test HKRs
	hkr()
}

func hkr() {
	fmt.Println(sep("HKRs"))
	hkrs, err := cl.GetSmarthomeDevicesFilter([]string{fb.CHKR})
	panicIfError(err)

	h := hkrs.Devices[0]

	prevName := h.Name
	err = h.DECTSetName(&cl, "Test")
	panicIfError(err)
	err = h.Reload(&cl)
	panicIfError(err)
	if h.Name != "Test" {
		panic("SetName failed")
	} else {
		fmt.Println("[+] SetName successful")
	}
	fmt.Println("[+] Resetting Name...")
	err = h.DECTSetName(&cl, prevName)
	err = h.Reload(&cl)
	panicIfError(err)

	c := fb.GetCapability[*fb.Hkr](h)

	v, err := c.DECTGetSoll(&cl)
	fmt.Println("[+] Soll:", v)
	panicIfError(err)
	v, err = c.DECTGetKomfort(&cl)
	fmt.Println("[+] Komfort:", v)
	panicIfError(err)
	v, err = c.DECTGetAbsenk(&cl)
	panicIfError(err)
	fmt.Println("[+] Absenk:", v)

	err = c.DECTSetSollMax(&cl)
	panicIfError(err)
	soll, err := c.DECTGetSoll(&cl)
	panicIfError(err)
	if soll != "MAX" {
		panic("SetSollMax failed")
	} else {
		fmt.Println("[+] SetSollMax successful")
	}

	err = c.DECTSetSollOff(&cl)
	panicIfError(err)
	soll, err = c.DECTGetSoll(&cl)
	panicIfError(err)
	if soll != "OFF" {
		panic("SetSollOff failed")
	} else {
		fmt.Println("[+] SetSollOff successful")
	}

	sollPrev := c.GetSoll()
	err = c.DECTSetSoll(&cl, "26.0")
	panicIfError(err)
	soll, err = c.DECTGetSoll(&cl)
	panicIfError(err)
	if soll != "26.0" {
		panic("SetSoll failed")
	} else {
		fmt.Println("[+] SetSoll successful")
	}
	fmt.Println("[+] Resetting Soll...")
	err = c.DECTSetSoll(&cl, sollPrev)
	panicIfError(err)

	tm, err := c.SetBoost(&cl, 2*time.Minute)
	panicIfError(err)
	err = c.Reload(&cl)
	panicIfError(err)
	if tm != c.GetBoostEndtime() || !c.IsBoostActive() {
		panic("SetBoost failed")
	} else {
		fmt.Println("[+] SetBoost successful")
	}

	err = c.DECTDeactivateBoost(&cl)
	panicIfError(err)
	err = c.Reload(&cl)
	panicIfError(err)
	if c.IsBoostActive() {
		panic("DeactivateBoost failed")
	} else {
		fmt.Println("[+] DeactivateBoost successful")
	}

	tm, err = c.DECTSetWindowOpen(&cl, 2*time.Minute)
	panicIfError(err)
	err = c.Reload(&cl)
	panicIfError(err)
	if tm != c.GetWindowOpenEndtime() || !c.IsWindowOpen() {
		panic("SetWindowOpen failed")
	} else {
		fmt.Println("[+] SetWindowOpen successful")
	}

	err = c.DECTDeactivateWindowOpen(&cl)
	panicIfError(err)
	err = c.Reload(&cl)
	panicIfError(err)
	if c.IsWindowOpen() {
		panic("DeactivateWindowOpen failed")
	} else {
		fmt.Println("[+] DeactivateWindowOpen successful")
	}
}

func panicIfError(err error) {
	if err != nil {
		panic(err)
	}
}
