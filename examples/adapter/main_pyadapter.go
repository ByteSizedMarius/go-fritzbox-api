package main

import (
	"fmt"
	fb "github.com/ByteSizedMarius/go-fritzbox-api"
	"log"
)

func main() {
	// Connection to fritzbox
	cl := fb.Client{
		Username: "username",
		Password: "password",
	}

	err := cl.Initialize()
	checkError(err)

	// Start the Adapter, which starts a ChromeDriver and logs in
	// This takes a few secs...
	pya := fb.PyAdapter{Client: &cl}

	// Setting Debug to true will show the ChromeDriver window, otherwise its started headless
	// pya := fb.PyAdapter{Client: &cl, Debug: true}

	err = pya.StartAdapter()
	checkError(err)
	fmt.Println("Started Adapter")

	// Using the Adapter, for example for a HKR-Device

	// Get all HKR-Devices
	shDevices, err := cl.GetSmarthomeDevicesFilter([]string{fb.CHKR})
	checkError(err)

	if len(shDevices.Devices) == 0 {
		log.Fatal("No HKR-Devices found")
	}

	// Get the first HKR-Device
	hkrD := fb.GetCapability[*fb.Hkr](shDevices.Devices[0])

	// Gets the Timeframe of the Summertime and sets it into the HKRs Struct.
	err = hkrD.PyaFetchSummertime(&pya)
	checkError(err)
	fmt.Printf("Summertime: %s.%s - %s.%s", hkrD.SummerTimeFrame.StartDay, hkrD.SummerTimeFrame.StartMonth, hkrD.SummerTimeFrame.EndDay, hkrD.SummerTimeFrame.EndMonth)
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
