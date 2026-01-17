package aha

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/ByteSizedMarius/go-fritzbox-api/v2"
	"github.com/ByteSizedMarius/go-fritzbox-api/v2/aha"
)

var testCfg struct {
	User                            string `json:"user"`
	Pass                            string `json:"pass"`
	DoHkrTests                      bool   `json:"do_hkr_tests"`
	HkrTestDeviceIdentifier         string `json:"hkr_test_device_identifier"`
	DoTemperatureTests              bool   `json:"do_temperature_tests"`
	TemperatureTestDeviceIdentifier string `json:"temperature_test_device_identifier"`
	DoButtonTests                   bool   `json:"do_button_tests"`
	ButtonTestDeviceIdentifier      string `json:"button_test_device_identifier"`
	DoHanfunTests                   bool   `json:"do_hanfun_tests"`
	HanfunTestDeviceIdentifier      string `json:"hanfun_test_device_identifier"`
}

var cl *fritzbox.Client
var skipHkr bool
var hkr *aha.Hkr
var hkrDevice *aha.Device
var skipTmp bool
var tmp *aha.Temperature
var skipBtn bool
var btn *aha.ButtonDevice
var skipHanfun bool
var hanfun *aha.HanFun

func TestMain(m *testing.M) {
	cfg, err := os.ReadFile("config.json")
	if err != nil {
		fmt.Println("Skipping: create config.json")
		os.Exit(0)
	}

	if err := json.Unmarshal(cfg, &testCfg); err != nil {
		fmt.Println("Invalid config.json:", err)
		os.Exit(1)
	}

	skipHkr = !testCfg.DoHkrTests
	skipTmp = !testCfg.DoTemperatureTests
	skipBtn = !testCfg.DoButtonTests
	skipHanfun = !testCfg.DoHanfunTests

	cl = fritzbox.New(testCfg.User, testCfg.Pass)
	if err := cl.Connect(); err != nil {
		fmt.Println("Error connecting:", err)
		os.Exit(1)
	}

	if !skipHkr {
		hkr = new(aha.Hkr)
		if err := aha.GetDeviceInfos(cl, testCfg.HkrTestDeviceIdentifier, hkr); err != nil {
			fmt.Println("Error getting HKR device:", err)
			os.Exit(1)
		}
		hkrDevice = hkr.Device()
	}

	if !skipTmp {
		tmp = new(aha.Temperature)
		if err := aha.GetDeviceInfos(cl, testCfg.TemperatureTestDeviceIdentifier, tmp); err != nil {
			fmt.Println("Error getting Temperature device:", err)
			os.Exit(1)
		}
	}

	if !skipBtn {
		btn = new(aha.ButtonDevice)
		if err := aha.GetDeviceInfos(cl, testCfg.ButtonTestDeviceIdentifier, btn); err != nil {
			fmt.Println("Error getting Button device:", err)
			os.Exit(1)
		}
	}

	if !skipHanfun {
		dl, err := aha.GetDeviceList(cl)
		if err != nil {
			fmt.Println("Error getting device list for HanFun:", err)
			os.Exit(1)
		}
		for _, d := range dl.Devices {
			if d.Identifier == testCfg.HanfunTestDeviceIdentifier {
				if hf, ok := d.Capabilities[aha.CHanfun].(*aha.HanFun); ok {
					hanfun = hf
					break
				}
			}
		}
		if hanfun == nil {
			fmt.Printf("HanFun device %s not found\n", testCfg.HanfunTestDeviceIdentifier)
			os.Exit(1)
		}
	}

	code := m.Run()
	cl.Close()
	os.Exit(code)
}
