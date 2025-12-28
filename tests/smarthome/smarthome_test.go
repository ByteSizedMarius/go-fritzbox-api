package smarthome

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/ByteSizedMarius/go-fritzbox-api"
	"github.com/ByteSizedMarius/go-fritzbox-api/smarthome"
)

var testCfg struct {
	User                  string `json:"user"`
	Pass                  string `json:"pass"`
	DoThermostatTests     bool   `json:"do_thermostat_tests"`
	DoDisruptiveTests     bool   `json:"do_disruptive_tests"`
	DoLockTests           bool   `json:"do_lock_tests"`
	DoScheduleTests       bool   `json:"do_schedule_tests"`
	ThermostatTestDevice  string `json:"thermostat_test_device"`
	ThermostatTestUnitUID string `json:"thermostat_test_unit_uid"`
}

var cl *fritzbox.Client
var thermostatUID string

func TestMain(m *testing.M) {
	cfg, err := os.ReadFile("config.json")
	if err != nil {
		fmt.Println("Skipping: create config.json with {\"user\":\"x\",\"pass\":\"y\"}")
		os.Exit(0)
	}

	if err := json.Unmarshal(cfg, &testCfg); err != nil {
		fmt.Println("Invalid config.json:", err)
		os.Exit(1)
	}

	cl = fritzbox.New(testCfg.User, testCfg.Pass)
	if err := cl.Connect(); err != nil {
		fmt.Println("Error connecting:", err)
		os.Exit(1)
	}

	if testCfg.ThermostatTestDevice != "" && testCfg.ThermostatTestUnitUID == "" {
		overview, err := smarthome.GetOverview(cl)
		if err != nil {
			fmt.Printf("Warning: could not get overview: %v\n", err)
		} else {
			for _, device := range overview.Devices {
				if device.Name == testCfg.ThermostatTestDevice && device.ProductCategory == "thermostat" && len(device.UnitUIDs) > 0 {
					thermostatUID = device.UnitUIDs[0]
					break
				}
			}
			if thermostatUID == "" {
				fmt.Printf("Warning: thermostat %q not found\n", testCfg.ThermostatTestDevice)
			}
		}
	} else {
		thermostatUID = testCfg.ThermostatTestUnitUID
	}

	code := m.Run()
	cl.Close()
	os.Exit(code)
}

// --- Read-only tests ---

func TestReadOnly(t *testing.T) {
	t.Run("GetOverview", getOverview)
	t.Run("GetUnit", getUnit)
	t.Run("GetThermostatConfig", getThermostatConfig)
	t.Run("FindDeviceByName", findDeviceByName)
	t.Run("GetUnitConfig", getUnitConfig)
}

func getOverview(t *testing.T) {
	overview, err := smarthome.GetOverview(cl)
	if err != nil {
		t.Fatalf("GetOverview failed: %v", err)
	}

	t.Logf("Found %d devices, %d units, %d groups", len(overview.Devices), len(overview.Units), len(overview.Groups))

	if len(overview.Devices) == 0 {
		t.Skip("no devices found")
	}

	d := overview.Devices[0]
	t.Logf("First device: %s (%s) - %s", d.Name, d.UID, d.ProductCategory)
}

func getUnit(t *testing.T) {
	overview, err := smarthome.GetOverview(cl)
	if err != nil {
		t.Fatalf("GetOverview failed: %v", err)
	}

	var unitUID string
	for _, device := range overview.Devices {
		if device.ProductCategory == "thermostat" && len(device.UnitUIDs) > 0 {
			unitUID = device.UnitUIDs[0]
			t.Logf("Found thermostat device: %s, using unit: %s", device.Name, unitUID)
			break
		}
	}

	if unitUID == "" {
		t.Skip("no thermostat unit found")
	}

	unit, err := smarthome.GetUnit(cl, unitUID)
	if err != nil {
		t.Fatalf("GetUnit failed: %v", err)
	}

	t.Logf("Unit: %s (type: %s)", unit.Name, unit.Type)
}

func getThermostatConfig(t *testing.T) {
	if thermostatUID == "" {
		t.Skip("no thermostat unit found")
	}

	config, err := smarthome.GetThermostatConfig(cl, thermostatUID)
	if err != nil {
		t.Fatalf("GetThermostatConfig failed: %v", err)
	}

	t.Logf("Thermostat: %s", config.Name)
	t.Logf("  Current temp: %.1f°C", config.Interfaces.Temperature.Celsius)
	t.Logf("  Target temp: %.1f°C (mode: %s)", config.Interfaces.Thermostat.SetPointTemperature.Celsius, config.Interfaces.Thermostat.SetPointTemperature.Mode)
	t.Logf("  Comfort: %.1f°C, Reduced: %.1f°C", config.Interfaces.Thermostat.ComfortTemperature.Celsius, config.Interfaces.Thermostat.ReducedTemperature.Celsius)
	t.Logf("  Boost: %v, WindowOpen: %v", config.Interfaces.Thermostat.Boost.Enabled, config.Interfaces.Thermostat.WindowOpenMode.Enabled)
}

func findDeviceByName(t *testing.T) {
	if testCfg.ThermostatTestDevice == "" {
		t.Skip("No thermostat_test_device configured")
	}

	device, err := smarthome.FindDeviceByName(cl, testCfg.ThermostatTestDevice)
	if err != nil {
		t.Fatalf("FindDeviceByName failed: %v", err)
	}

	t.Logf("Found device: %s (%s) - %s", device.Name, device.UID, device.ProductCategory)
}

func getUnitConfig(t *testing.T) {
	if thermostatUID == "" {
		t.Skip("No thermostat configured")
	}

	config, err := smarthome.GetUnitConfig(cl, thermostatUID)
	if err != nil {
		t.Fatalf("GetUnitConfig failed: %v", err)
	}

	if _, ok := config["interfaces"]; !ok {
		t.Error("Expected 'interfaces' key in config")
	}

	t.Logf("Config has %d top-level keys", len(config))
}

// --- Config tests (do_thermostat_tests) ---

func TestThermostatConfig(t *testing.T) {
	if !testCfg.DoThermostatTests {
		t.Skip("Thermostat tests disabled")
	}
	if thermostatUID == "" {
		t.Skip("No thermostat configured")
	}

	t.Run("WindowOpenDetection", setWindowOpenDetection)
	t.Run("ComfortTemperature", setComfortTemperature)
	t.Run("ReducedTemperature", setReducedTemperature)
	t.Run("AdaptiveHeating", setAdaptiveHeating)
	t.Run("TemperatureOffset", setTemperatureOffset)
}

func setWindowOpenDetection(t *testing.T) {

	config, err := smarthome.GetThermostatConfig(cl, thermostatUID)
	if err != nil {
		t.Fatalf("GetThermostatConfig failed: %v", err)
	}

	originalDuration := config.Interfaces.Thermostat.WindowOpenMode.InternalDuration
	originalSensitivity := config.Interfaces.Thermostat.WindowOpenMode.InternalSensitivity

	defer func() {
		if err := smarthome.SetWindowOpenDetection(cl, thermostatUID, originalDuration, originalSensitivity); err != nil {
			t.Logf("Restore failed: %v", err)
		}
	}()

	t.Logf("Original: duration=%d, sensitivity=%s", originalDuration, originalSensitivity)

	newDuration := 11
	if originalDuration == 11 {
		newDuration = 10
	}

	if err := smarthome.SetWindowOpenDetection(cl, thermostatUID, newDuration, originalSensitivity); err != nil {
		t.Fatalf("SetWindowOpenDetection failed: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	config, err = smarthome.GetThermostatConfig(cl, thermostatUID)
	if err != nil {
		t.Fatalf("GetThermostatConfig (verify) failed: %v", err)
	}

	if config.Interfaces.Thermostat.WindowOpenMode.InternalDuration != newDuration {
		t.Errorf("Duration not changed: got %d, want %d", config.Interfaces.Thermostat.WindowOpenMode.InternalDuration, newDuration)
	}
}

func setComfortTemperature(t *testing.T) {
	config, err := smarthome.GetThermostatConfig(cl, thermostatUID)
	if err != nil {
		t.Fatalf("GetThermostatConfig failed: %v", err)
	}

	original := config.Interfaces.Thermostat.ComfortTemperature.Celsius

	defer func() {
		if err := smarthome.SetComfortTemperature(cl, thermostatUID, original); err != nil {
			t.Logf("Restore failed: %v", err)
		}
	}()

	t.Logf("Original comfort: %.1f°C", original)

	newTemp := original + 0.5
	if newTemp > 28 {
		newTemp = original - 0.5
	}

	if err := smarthome.SetComfortTemperature(cl, thermostatUID, newTemp); err != nil {
		t.Fatalf("SetComfortTemperature failed: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	config, err = smarthome.GetThermostatConfig(cl, thermostatUID)
	if err != nil {
		t.Fatalf("GetThermostatConfig (verify) failed: %v", err)
	}

	if config.Interfaces.Thermostat.ComfortTemperature.Celsius != newTemp {
		t.Errorf("Comfort temp not changed: got %.1f, want %.1f", config.Interfaces.Thermostat.ComfortTemperature.Celsius, newTemp)
	}
}

func setReducedTemperature(t *testing.T) {
	config, err := smarthome.GetThermostatConfig(cl, thermostatUID)
	if err != nil {
		t.Fatalf("GetThermostatConfig failed: %v", err)
	}

	original := config.Interfaces.Thermostat.ReducedTemperature.Celsius

	defer func() {
		if err := smarthome.SetReducedTemperature(cl, thermostatUID, original); err != nil {
			t.Logf("Restore failed: %v", err)
		}
	}()

	t.Logf("Original reduced: %.1f°C", original)

	newTemp := original + 0.5
	if newTemp > 28 {
		newTemp = original - 0.5
	}

	if err := smarthome.SetReducedTemperature(cl, thermostatUID, newTemp); err != nil {
		t.Fatalf("SetReducedTemperature failed: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	config, err = smarthome.GetThermostatConfig(cl, thermostatUID)
	if err != nil {
		t.Fatalf("GetThermostatConfig (verify) failed: %v", err)
	}

	if config.Interfaces.Thermostat.ReducedTemperature.Celsius != newTemp {
		t.Errorf("Reduced temp not changed: got %.1f, want %.1f", config.Interfaces.Thermostat.ReducedTemperature.Celsius, newTemp)
	}
}

func setAdaptiveHeating(t *testing.T) {
	config, err := smarthome.GetThermostatConfig(cl, thermostatUID)
	if err != nil {
		t.Fatalf("GetThermostatConfig failed: %v", err)
	}

	original := config.Interfaces.Thermostat.AdaptiveHeatingModeEnabled

	defer func() {
		if err := smarthome.SetAdaptiveHeating(cl, thermostatUID, original); err != nil {
			t.Logf("Restore failed: %v", err)
		}
	}()

	t.Logf("Original adaptive heating: %v", original)

	if err := smarthome.SetAdaptiveHeating(cl, thermostatUID, !original); err != nil {
		t.Fatalf("SetAdaptiveHeating failed: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	config, err = smarthome.GetThermostatConfig(cl, thermostatUID)
	if err != nil {
		t.Fatalf("GetThermostatConfig (verify) failed: %v", err)
	}

	if config.Interfaces.Thermostat.AdaptiveHeatingModeEnabled == original {
		t.Error("AdaptiveHeating not changed")
	}
}

func setTemperatureOffset(t *testing.T) {
	config, err := smarthome.GetThermostatConfig(cl, thermostatUID)
	if err != nil {
		t.Fatalf("GetThermostatConfig failed: %v", err)
	}

	original := config.Interfaces.Thermostat.TemperatureOffset.InternalOffset

	defer func() {
		if err := smarthome.SetTemperatureOffset(cl, thermostatUID, original); err != nil {
			t.Logf("Restore failed: %v", err)
		}
	}()

	t.Logf("Original offset: %.1f", original)

	newOffset := original + 0.5
	if newOffset > 10 {
		newOffset = original - 0.5
	}

	if err := smarthome.SetTemperatureOffset(cl, thermostatUID, newOffset); err != nil {
		t.Fatalf("SetTemperatureOffset failed: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	config, err = smarthome.GetThermostatConfig(cl, thermostatUID)
	if err != nil {
		t.Fatalf("GetThermostatConfig (verify) failed: %v", err)
	}

	if config.Interfaces.Thermostat.TemperatureOffset.InternalOffset != newOffset {
		t.Errorf("Offset not changed: got %.1f, want %.1f", config.Interfaces.Thermostat.TemperatureOffset.InternalOffset, newOffset)
	}
}

// --- Disruptive tests (do_disruptive_tests) ---

func TestDisruptive(t *testing.T) {
	if !testCfg.DoDisruptiveTests {
		t.Skip("Disruptive tests disabled")
	}
	if thermostatUID == "" {
		t.Skip("No thermostat configured")
	}

	t.Run("TargetTemperature", setTargetTemperature)
	t.Run("TargetTemperatureOffOn", setTargetTemperatureOffOn)
	t.Run("Boost", setBoost)
	t.Run("WindowOpen", setWindowOpen)
}

func setTargetTemperature(t *testing.T) {
	config, err := smarthome.GetThermostatConfig(cl, thermostatUID)
	if err != nil {
		t.Fatalf("GetThermostatConfig failed: %v", err)
	}

	original := config.Interfaces.Thermostat.SetPointTemperature.Celsius

	defer func() {
		if err := smarthome.SetTargetTemperature(cl, thermostatUID, original); err != nil {
			t.Logf("Restore failed: %v", err)
		}
	}()

	t.Logf("Original target: %.1f°C", original)

	newTemp := original + 0.5
	if newTemp > 28 {
		newTemp = original - 0.5
	}

	if err := smarthome.SetTargetTemperature(cl, thermostatUID, newTemp); err != nil {
		t.Fatalf("SetTargetTemperature failed: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	config, err = smarthome.GetThermostatConfig(cl, thermostatUID)
	if err != nil {
		t.Fatalf("GetThermostatConfig (verify) failed: %v", err)
	}

	if config.Interfaces.Thermostat.SetPointTemperature.Celsius != newTemp {
		t.Errorf("Target temp not changed: got %.1f, want %.1f", config.Interfaces.Thermostat.SetPointTemperature.Celsius, newTemp)
	}
}

func setTargetTemperatureOffOn(t *testing.T) {
	config, err := smarthome.GetThermostatConfig(cl, thermostatUID)
	if err != nil {
		t.Fatalf("GetThermostatConfig failed: %v", err)
	}

	original := config.Interfaces.Thermostat.SetPointTemperature.Celsius

	defer func() {
		if err := smarthome.SetTargetTemperature(cl, thermostatUID, original); err != nil {
			t.Logf("Restore failed: %v", err)
		}
	}()

	t.Logf("Original: %.1f°C", original)

	// Test OFF
	if err := smarthome.SetTargetTemperatureOff(cl, thermostatUID); err != nil {
		t.Fatalf("SetTargetTemperatureOff failed: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	config, err = smarthome.GetThermostatConfig(cl, thermostatUID)
	if err != nil {
		t.Fatalf("GetThermostatConfig (off) failed: %v", err)
	}

	if config.Interfaces.Thermostat.SetPointTemperature.Mode != "off" {
		t.Errorf("Mode not 'off': got %s", config.Interfaces.Thermostat.SetPointTemperature.Mode)
	}
	t.Log("OFF mode verified")

	// Test ON
	if err := smarthome.SetTargetTemperatureOn(cl, thermostatUID); err != nil {
		t.Fatalf("SetTargetTemperatureOn failed: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	config, err = smarthome.GetThermostatConfig(cl, thermostatUID)
	if err != nil {
		t.Fatalf("GetThermostatConfig (on) failed: %v", err)
	}

	if config.Interfaces.Thermostat.SetPointTemperature.Mode != "on" {
		t.Errorf("Mode not 'on': got %s", config.Interfaces.Thermostat.SetPointTemperature.Mode)
	}
	t.Log("ON mode verified")
}

func setBoost(t *testing.T) {
	defer func() {
		if err := smarthome.DeactivateBoost(cl, thermostatUID); err != nil {
			t.Logf("Deactivate boost failed: %v", err)
		}
	}()

	config, err := smarthome.GetThermostatConfig(cl, thermostatUID)
	if err != nil {
		t.Fatalf("GetThermostatConfig failed: %v", err)
	}

	if config.Interfaces.Thermostat.Boost.Enabled {
		t.Log("Boost already active, will deactivate at end")
		return
	}

	if err := smarthome.SetBoost(cl, thermostatUID, 1); err != nil {
		t.Fatalf("SetBoost failed: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	config, err = smarthome.GetThermostatConfig(cl, thermostatUID)
	if err != nil {
		t.Fatalf("GetThermostatConfig (verify) failed: %v", err)
	}

	if !config.Interfaces.Thermostat.Boost.Enabled {
		t.Error("Boost not enabled")
	} else {
		t.Log("Boost enabled successfully")
	}
}

func setWindowOpen(t *testing.T) {
	defer func() {
		if err := smarthome.DeactivateWindowOpen(cl, thermostatUID); err != nil {
			t.Logf("Deactivate window-open failed: %v", err)
		}
	}()

	config, err := smarthome.GetThermostatConfig(cl, thermostatUID)
	if err != nil {
		t.Fatalf("GetThermostatConfig failed: %v", err)
	}

	if config.Interfaces.Thermostat.WindowOpenMode.Enabled {
		t.Log("WindowOpen already active, will deactivate at end")
		return
	}

	if err := smarthome.SetWindowOpen(cl, thermostatUID, 1); err != nil {
		t.Fatalf("SetWindowOpen failed: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	config, err = smarthome.GetThermostatConfig(cl, thermostatUID)
	if err != nil {
		t.Fatalf("GetThermostatConfig (verify) failed: %v", err)
	}

	if !config.Interfaces.Thermostat.WindowOpenMode.Enabled {
		t.Error("WindowOpen not enabled")
	} else {
		t.Log("WindowOpen enabled successfully")
	}
}

// --- Lock tests (do_lock_tests) ---

func TestLocks(t *testing.T) {
	if !testCfg.DoLockTests {
		t.Skip("Lock tests disabled")
	}
	if thermostatUID == "" {
		t.Skip("No thermostat configured")
	}

	t.Run("SetLocks", setLocks)
}

func setLocks(t *testing.T) {
	config, err := smarthome.GetThermostatConfig(cl, thermostatUID)
	if err != nil {
		t.Fatalf("GetThermostatConfig failed: %v", err)
	}

	originalLocal := config.Interfaces.Thermostat.LockedDeviceLocalEnabled
	originalAPI := config.Interfaces.Thermostat.LockedDeviceAPIEnabled

	defer func() {
		if err := smarthome.SetLocks(cl, thermostatUID, originalLocal, originalAPI); err != nil {
			t.Logf("Restore locks failed: %v", err)
		}
	}()

	t.Logf("Original locks: local=%v, api=%v", originalLocal, originalAPI)

	// Toggle local lock only (don't lock API or we can't restore!)
	if err := smarthome.SetLocks(cl, thermostatUID, !originalLocal, originalAPI); err != nil {
		t.Fatalf("SetLocks failed: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	config, err = smarthome.GetThermostatConfig(cl, thermostatUID)
	if err != nil {
		t.Fatalf("GetThermostatConfig (verify) failed: %v", err)
	}

	if config.Interfaces.Thermostat.LockedDeviceLocalEnabled == originalLocal {
		t.Error("Local lock not changed")
	} else {
		t.Logf("Local lock changed to %v", config.Interfaces.Thermostat.LockedDeviceLocalEnabled)
	}
}

// --- Schedule tests (do_schedule_tests) ---

func TestSchedule(t *testing.T) {
	if !testCfg.DoScheduleTests {
		t.Skip("Schedule tests disabled")
	}
	if thermostatUID == "" {
		t.Skip("No thermostat configured")
	}

	t.Run("SummerPeriod", setSummerPeriod)
	t.Run("WeeklyTimer", setWeeklyTimer)
	t.Run("AddAndClearHolidays", addAndClearHolidays)
}

func setSummerPeriod(t *testing.T) {
	config, err := smarthome.GetThermostatConfig(cl, thermostatUID)
	if err != nil {
		t.Fatalf("GetThermostatConfig failed: %v", err)
	}

	originalEnabled := config.Interfaces.Thermostat.SummerPeriod.Enabled
	originalStart := config.Interfaces.Thermostat.SummerPeriod.StartTime
	originalEnd := config.Interfaces.Thermostat.SummerPeriod.EndTime

	defer func() {
		body := map[string]interface{}{
			"interfaces": map[string]interface{}{
				"thermostatInterface": map[string]interface{}{
					"summerPeriod": map[string]interface{}{
						"enabled":   originalEnabled,
						"startTime": originalStart,
						"endTime":   originalEnd,
					},
				},
			},
		}
		cl.RestPut("api/v0/smarthome/configuration/units/"+thermostatUID, body)
	}()

	t.Logf("Original summer period: enabled=%v", originalEnabled)

	if err := smarthome.SetSummerPeriod(cl, thermostatUID, true, 6, 1, 8, 31); err != nil {
		t.Fatalf("SetSummerPeriod failed: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	config, err = smarthome.GetThermostatConfig(cl, thermostatUID)
	if err != nil {
		t.Fatalf("GetThermostatConfig (verify) failed: %v", err)
	}

	if !config.Interfaces.Thermostat.SummerPeriod.Enabled {
		t.Error("Summer period not enabled")
	} else {
		t.Log("Summer period set successfully")
	}
}

func setWeeklyTimer(t *testing.T) {
	config, err := smarthome.GetThermostatConfig(cl, thermostatUID)
	if err != nil {
		t.Fatalf("GetThermostatConfig failed: %v", err)
	}

	if config.Timer == nil {
		t.Skip("No timer configured on this thermostat")
	}

	original := config.Timer.Weekly

	defer func() {
		if err := smarthome.SetWeeklyTimer(cl, thermostatUID, original); err != nil {
			t.Logf("Restore timer failed: %v", err)
		}
	}()

	t.Logf("Original timer has %d entries", len(original))

	var testEntries []smarthome.TimerEntry
	for day := 0; day < 7; day++ {
		dayOffset := day * 1440
		testEntries = append(testEntries,
			smarthome.TimerEntry{TemperaturePreset: "comfort", Time: dayOffset + 360},
			smarthome.TimerEntry{TemperaturePreset: "reduced", Time: dayOffset + 1320},
		)
	}

	if err := smarthome.SetWeeklyTimer(cl, thermostatUID, testEntries); err != nil {
		t.Fatalf("SetWeeklyTimer failed: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	config, err = smarthome.GetThermostatConfig(cl, thermostatUID)
	if err != nil {
		t.Fatalf("GetThermostatConfig (verify) failed: %v", err)
	}

	if config.Timer == nil || len(config.Timer.Weekly) != len(testEntries) {
		t.Errorf("Timer entries mismatch: got %d, want %d", len(config.Timer.Weekly), len(testEntries))
	} else {
		t.Logf("Timer set with %d entries", len(config.Timer.Weekly))
	}
}

func addAndClearHolidays(t *testing.T) {
	config, err := smarthome.GetThermostatConfig(cl, thermostatUID)
	if err != nil {
		t.Fatalf("GetThermostatConfig failed: %v", err)
	}

	originalPeriods := config.Interfaces.Thermostat.HolidayPeriods.Periods
	originalTemp := config.Interfaces.Thermostat.HolidayPeriods.Temperature

	defer func() {
		body := map[string]interface{}{
			"interfaces": map[string]interface{}{
				"thermostatInterface": map[string]interface{}{
					"holidayPeriods": map[string]interface{}{
						"periods":     originalPeriods,
						"temperature": originalTemp,
					},
				},
			},
		}
		cl.RestPut("api/v0/smarthome/configuration/units/"+thermostatUID, body)
	}()

	t.Logf("Original holiday count: %d", len(originalPeriods))

	start := time.Now().AddDate(0, 0, 7)
	end := start.AddDate(0, 0, 7)

	if err := smarthome.AddHoliday(cl, thermostatUID, start, end, 16.0); err != nil {
		t.Fatalf("AddHoliday failed: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	config, err = smarthome.GetThermostatConfig(cl, thermostatUID)
	if err != nil {
		t.Fatalf("GetThermostatConfig (verify) failed: %v", err)
	}

	newCount := len(config.Interfaces.Thermostat.HolidayPeriods.Periods)
	if newCount != len(originalPeriods)+1 {
		t.Errorf("Holiday not added: got %d, want %d", newCount, len(originalPeriods)+1)
	} else {
		t.Logf("Holiday added, now %d periods", newCount)
	}

	if err := smarthome.ClearHolidays(cl, thermostatUID); err != nil {
		t.Fatalf("ClearHolidays failed: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	config, err = smarthome.GetThermostatConfig(cl, thermostatUID)
	if err != nil {
		t.Fatalf("GetThermostatConfig (clear) failed: %v", err)
	}

	if len(config.Interfaces.Thermostat.HolidayPeriods.Periods) != 0 {
		t.Errorf("Holidays not cleared: still have %d", len(config.Interfaces.Thermostat.HolidayPeriods.Periods))
	} else {
		t.Log("All holidays cleared")
	}
}
