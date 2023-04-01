package main

import (
	"fmt"
	fb "github.com/ByteSizedMarius/go-fritzbox-api"
	"os"
	"strings"
)

var cl fb.Client

func sep(in string) string {
	return "—————————————————————————————————" + in + "—————————————————————————————————"
}

func main() {
	initialize()

	//fritzClientlist()
	//fritzClientlistSetters()
	//fritzProfiles()
	//fritzLogs()
	//fritzStatistics()
}

func initialize() {
	data, err := os.ReadFile("./credentials")
	if err != nil {
		panic(err)
	}

	creds := strings.Split(string(data), "\r\n")
	cl = fb.Client{
		Username: creds[0],
		Password: creds[1],
		BaseUrl:  "http://192.168.178.1",
	}
	err = cl.Initialize()
	if err != nil {
		panic(err)
	}
}

func fritzClientlist() {
	fmt.Println(sep("Client-List"))

	clientList, err := cl.GetCLientList()
	if err != nil {
		fmt.Println("Clientlist-Error: ", err)
		return
	}

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
	if err != nil {
		panic(err)
	}
	fmt.Println("[+] Set IP successfully. Resetting...")
	err = cl.SetIP(deviceUID, prevIP, false)
	if err != nil {
		panic(err)
	}

	err = cl.SetName(deviceUID, newName)
	if err != nil {
		panic(err)
	}
	fmt.Println("[+] Set Name successfully. Resetting...")
	err = cl.SetName(deviceUID, prevName)
	if err != nil {
		panic(err)
	}
}

func fritzProfiles() {
	deviceUID := "landevice6511"
	prevFilter := "filtprof1"
	newFilter := "filtprof3"
	fmt.Println(sep("Profiles"))

	profiles, err := cl.GetAvailableProfiles()
	if err != nil {
		panic(err)
	}

	fmt.Printf("[+] Got %d profiles: [", len(profiles))
	var str string
	for _, v := range profiles {
		str += fmt.Sprintf("%s (%s, %s), ", v.Name, v.UID, v.Filter)
	}
	fmt.Println(str[:len(str)-2] + "]")

	uid, err := cl.GetProfileUIDFromDevice(deviceUID)
	if err != nil {
		panic(err)
	}
	fmt.Printf("[+] Got UID %s from device %s\n", uid, deviceUID)

	err = cl.SetProfileForDevice(deviceUID, newFilter)
	if err != nil {
		panic(err)
	}

	fmt.Println("[+] Set Profile successfully. Resetting...")
	err = cl.SetProfileForDevice(deviceUID, prevFilter)
	if err != nil {
		panic(err)
	}
}

func fritzLogs() {
	fmt.Println(sep("Logs"))

	logs, err := cl.GetEventLog()
	if err != nil {
		panic(err)
	}
	fmt.Printf("[+] Got %d log messages\n", len(logs))

	logsUntil, err := cl.GetEventLogUntil(logs[len(logs)/2].ID)
	fmt.Printf("[+] Logs Until: Got %d log messages\n", len(logsUntil))
}

func fritzStatistics() {
	fmt.Println(sep("Statistics"))

	stats, err := cl.GetTrafficStats()
	if err != nil {
		panic(err)
	}
	fmt.Printf("[+] Got Statistics. MB sent today: %d, MB received today: %d\n", stats.Today.MBSent, stats.Today.MBReceived)
}
