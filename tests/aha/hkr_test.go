package aha

import (
	"fmt"
	"testing"
	"time"

	"github.com/ByteSizedMarius/go-fritzbox-api/aha"
)

func TestHKR(t *testing.T) {
	if skipHkr {
		t.Skip("HKR tests disabled")
	}

	t.Run("GetAbsenk", DECTGetAbsenk)
	t.Run("GetAbsenkNumeric", DECTGetAbsenkNumeric)
	t.Run("Absenk", Absenk)
	t.Run("GetKomfort", DECTGetKomfort)
	t.Run("GetKomfortNumeric", DECTGetKomfortNumeric)
	t.Run("GetSoll", DECTGetSoll)
	t.Run("GetSollNumeric", DECTGetSollNumeric)
	t.Run("Soll", Soll)
	t.Run("SetSollMax", DECTSetSollMax)
	t.Run("SetSollOff", DECTSetSollOff)
	t.Run("ActivateBoost", ActivateBoost)
	t.Run("Window", Window)
	t.Run("DeviceGetName", DeviceDECTGetName)
	t.Run("DeviceHasCapability", DeviceHasCapability)
	t.Run("NextChange", NextChange)
	t.Run("DeviceSetName", DeviceDECTSetName)
}

func DECTGetAbsenk(t *testing.T) {
	temperature, err := hkr.DECTGetAbsenk(cl)
	if err != nil {
		t.Fatal(err)
	}
	if temperature == "" {
		t.Fatal("Reading failed: Absenk is empty")
	}
}

func DECTGetAbsenkNumeric(t *testing.T) {
	temperature, err := hkr.DECTGetAbsenkNumeric(cl)
	if err != nil {
		t.Fatal(err)
	}
	if temperature == 0 {
		t.Fatal("Reading failed: Absenk is 0")
	}
}

func Absenk(t *testing.T) {
	v1, err := hkr.DECTGetAbsenk(cl)
	if err != nil {
		t.Fatal(err)
	}

	v2, err := hkr.DECTGetAbsenkNumeric(cl)
	if err != nil {
		t.Fatal(err)
	}

	if v1 != fmt.Sprintf("%.1f", v2) {
		t.Fatalf("DECTGetAbsenk and DECTGetAbsenkNumeric are not equal: %s != %.1f", v1, v2)
	}
}

func DECTGetKomfort(t *testing.T) {
	temperature, err := hkr.DECTGetKomfort(cl)
	if err != nil {
		t.Fatal(err)
	}
	if temperature == "" {
		t.Fatal("Reading failed: Komfort is empty")
	}
}

func DECTGetKomfortNumeric(t *testing.T) {
	temperature, err := hkr.DECTGetKomfortNumeric(cl)
	if err != nil {
		t.Fatal(err)
	}
	if temperature == 0 {
		t.Fatal("Reading failed: Komfort is 0")
	}
}

func DECTGetSoll(t *testing.T) {
	temperature, err := hkr.DECTGetSoll(cl)
	if err != nil {
		t.Fatal(err)
	}
	if temperature == "" {
		t.Fatal("Reading failed: Soll is empty")
	}
}

func DECTGetSollNumeric(t *testing.T) {
	_, err := hkr.DECTGetSollNumeric(cl)
	if err != nil {
		t.Fatal(err)
	}
}

func Soll(t *testing.T) {
	v1, err := hkr.DECTGetSoll(cl)
	if err != nil {
		t.Fatal(err)
	}

	v2, err := hkr.DECTGetSollNumeric(cl)
	if err != nil {
		t.Fatal(err)
	}

	vt := v1
	if v1 == "OFF" {
		vt = "-2.0"
	} else if v1 == "MAX" {
		vt = "28.0"
	}

	if vt != fmt.Sprintf("%.1f", v2) {
		t.Fatalf("DECTGetSoll and DECTGetSollNumeric are not equal: %s != %.1f", vt, v2)
	}

	// Set Soll
	if v2 < 0 || v2 > 27 {
		v2 = 18
	}

	err = hkr.DECTSetSoll(cl, v2+1)
	if err != nil {
		t.Fatal(err)
	}

	nv, err := hkr.DECTGetSollNumeric(cl)
	if err != nil {
		t.Fatal(err)
	}

	if nv != v2+1 {
		t.Fatal("DECTSetSoll failed")
	}

	// Set Soll back, test string
	err = hkr.DECTSetSoll(cl, v1)
	if err != nil {
		t.Fatal(err)
	}

	v3, err := hkr.DECTGetSoll(cl)
	if err != nil {
		t.Fatal(err)
	}

	if v3 != v1 {
		t.Fatal("DECTSetSoll failed")
	}
}

func DECTSetSollMax(t *testing.T) {
	original, err := hkr.DECTGetSoll(cl)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := hkr.DECTSetSoll(cl, original); err != nil {
			t.Logf("Failed to restore Soll: %v", err)
		}
	}()

	if err := hkr.DECTSetSollMax(cl); err != nil {
		t.Fatal(err)
	}

	numericValue, err := hkr.DECTGetSollNumeric(cl)
	if err != nil {
		t.Fatal(err)
	}

	if numericValue != -1 {
		t.Fatal("DECTSetSollMax failed:", numericValue)
	}

	value, err := hkr.DECTGetSoll(cl)
	if err != nil {
		t.Fatal(err)
	}

	if value != "MAX" {
		t.Fatal("DECTSetSollMax failed:", value)
	}
}

func DECTSetSollOff(t *testing.T) {
	original, err := hkr.DECTGetSoll(cl)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := hkr.DECTSetSoll(cl, original); err != nil {
			t.Logf("Failed to restore Soll: %v", err)
		}
	}()

	if err := hkr.DECTSetSollOff(cl); err != nil {
		t.Fatal(err)
	}

	numericValue, err := hkr.DECTGetSollNumeric(cl)
	if err != nil {
		t.Fatal(err)
	}

	if numericValue != -2 {
		t.Fatal("DECTSetSollOff failed:", numericValue)
	}

	value, err := hkr.DECTGetSoll(cl)
	if err != nil {
		t.Fatal(err)
	}

	if value != "OFF" {
		t.Fatal("DECTSetSollOff failed:", value)
	}
}

func ActivateBoost(t *testing.T) {
	if hkr.IsBoostActive() {
		t.Fatal("Boost is already active and I'm too lazy to handle this case")
	}

	end, err := hkr.SetBoost(cl, 30*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	if err := hkr.Reload(cl); err != nil {
		t.Fatal(err)
	}

	if !hkr.IsBoostActive() {
		t.Fatal("SetBoost failed")
	}

	if !hkr.GetBoostEndtime().Equal(end) {
		t.Fatalf("SetBoost endtimes do not match: %s, %s", hkr.GetBoostEndtime(), end)
	}

	if err := hkr.DECTDeactivateBoost(cl); err != nil {
		t.Fatal(err)
	}

	if err := hkr.Reload(cl); err != nil {
		t.Fatal(err)
	}

	if hkr.IsBoostActive() {
		t.Fatal("DECTDeactivateBoost failed")
	}
}

func Window(t *testing.T) {
	if hkr.IsWindowOpen() {
		t.Fatal("Window is already open and I'm too lazy to handle this case")
	}

	end, err := hkr.DECTSetWindowOpen(cl, 30*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	if err := hkr.Reload(cl); err != nil {
		t.Fatal(err)
	}

	if !hkr.IsWindowOpen() {
		t.Fatal("DECTSetWindowOpen failed")
	}

	if !hkr.GetWindowOpenEndtime().Equal(end) {
		t.Fatalf("WindowOpen endtimes do not match: %s, %s", hkr.GetWindowOpenEndtime(), end)
	}

	if err := hkr.DECTDeactivateWindowOpen(cl); err != nil {
		t.Fatal(err)
	}

	if err := hkr.Reload(cl); err != nil {
		t.Fatal(err)
	}

	if hkr.IsWindowOpen() {
		t.Fatal("DECTDeactivateWindowOpen failed")
	}
}

func DeviceDECTGetName(t *testing.T) {
	name, err := hkrDevice.DECTGetName(cl)
	if err != nil {
		t.Fatal(err)
	}
	if name == "" {
		t.Fatal("DECTGetName returned empty string")
	}
	t.Logf("Device name: %s", name)
}

func DeviceHasCapability(t *testing.T) {
	if !hkrDevice.HasCapability(aha.CHKR) {
		t.Fatal("HKR device should have CHKR capability")
	}

	if !hkrDevice.HasCapability(aha.CTempSensor) {
		t.Fatal("HKR device should have CTempSensor capability")
	}

	// Should not have button capability
	if hkrDevice.HasCapability(aha.CButton) {
		t.Log("Note: HKR device unexpectedly has CButton capability")
	}
}

func NextChange(t *testing.T) {
	// These read cached values from the timer schedule
	temp := hkr.GetNextChangeTemperature()
	tempNum := hkr.GetNextChangeTemperatureNumeric()
	endtime := hkr.GetNextchangeEndtime()

	// May be zero if no timer schedule is configured
	if endtime.IsZero() {
		t.Log("No next change scheduled (timer may not be configured)")
		return
	}

	t.Logf("Next change: %s (%.1f°C) at %s", temp, tempNum, endtime)

	if temp == "" {
		t.Error("GetNextChangeTemperature returned empty but endtime is set")
	}
}

func DeviceDECTSetName(t *testing.T) {
	originalName, err := hkrDevice.DECTGetName(cl)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Original name: %s", originalName)

	defer func() {
		if err := hkrDevice.DECTSetName(cl, originalName); err != nil {
			t.Logf("Failed to restore name: %v", err)
		}
	}()

	testName := originalName + "-test"
	if err := hkrDevice.DECTSetName(cl, testName); err != nil {
		t.Fatal(err)
	}

	newName, err := hkrDevice.DECTGetName(cl)
	if err != nil {
		t.Fatal(err)
	}

	if newName != testName {
		t.Errorf("Name not changed: got %s, want %s", newName, testName)
	}
	t.Logf("Name changed to: %s", newName)
}
