package unsafe

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ByteSizedMarius/go-fritzbox-api/v2"
	"github.com/ByteSizedMarius/go-fritzbox-api/v2/unsafe"
)

var testCfg struct {
	User                 string `json:"user"`
	Pass                 string `json:"pass"`
	DoProfileTests       bool   `json:"do_profile_tests"`
	ProfileTestDeviceUID string `json:"profile_test_device_uid"`
	DoDeviceTests        bool   `json:"do_device_tests"`
	DeviceTestUID        string `json:"device_test_uid"`
}

var cl *fritzbox.Client

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

	code := m.Run()
	cl.Close()
	os.Exit(code)
}

func TestReadOnly(t *testing.T) {
	t.Run("GetTrafficStats", getTrafficStats)
	t.Run("GetAvailableProfiles", getAvailableProfiles)
	t.Run("GetDevices", getDevices)
	t.Run("GetMeshTopology", getMeshTopology)
	t.Run("GetEventLog", getEventLog)
	t.Run("GetProfileUIDFromDevice", getProfileUIDFromDevice)
	t.Run("GetDeviceName", getDeviceName)
	t.Run("GetDeviceIP", getDeviceIP)
}

func getTrafficStats(t *testing.T) {
	ts, err := unsafe.GetTrafficStats(cl)
	if err != nil {
		t.Fatalf("GetTrafficStats failed: %v", err)
	}

	if ts.Today.BytesSentLow == "" {
		t.Error("Today.BytesSentLow is empty, query.lua may have failed")
	}

	if ts.Today.MBReceived < 0 {
		t.Error("MBReceived should be >= 0")
	}
}

func getAvailableProfiles(t *testing.T) {
	profiles, err := unsafe.GetAvailableProfiles(cl)
	if err != nil {
		t.Fatalf("GetAvailableProfiles failed: %v", err)
	}

	if len(profiles) == 0 {
		t.Fatal("Expected at least one profile (Fritz!Box has defaults)")
	}

	for uid := range profiles {
		if !strings.HasPrefix(uid, "filtprof") {
			t.Errorf("Unexpected profile UID format: %s", uid)
		}
	}
}

func getDevices(t *testing.T) {
	devices, err := unsafe.GetDevices(cl)
	if err != nil {
		t.Fatalf("GetDevices failed: %v", err)
	}

	if len(devices) == 0 {
		t.Error("Expected at least one device")
	}

	for _, d := range devices {
		if d.MAC == "" {
			t.Logf("Device %s has no MAC (may be expected for some devices)", d.UID)
		}
	}
}

func getMeshTopology(t *testing.T) {
	mt, err := unsafe.GetMeshTopology(cl)
	if err != nil {
		t.Fatalf("GetMeshTopology failed: %v", err)
	}

	if len(mt.Devices) == 0 {
		t.Error("Expected at least one mesh device")
	}

	if mt.Rootuid == "" {
		t.Error("Expected rootuid to be set")
	}
}

func getEventLog(t *testing.T) {
	logs, err := unsafe.GetEventLog(cl)
	if err != nil {
		t.Fatalf("GetEventLog failed: %v", err)
	}

	if len(logs) == 0 {
		t.Error("Expected at least one log entry")
	}
}

func getProfileUIDFromDevice(t *testing.T) {
	devices, err := unsafe.GetDevices(cl)
	if err != nil {
		t.Fatalf("GetDevices failed: %v", err)
	}

	if len(devices) == 0 {
		t.Skip("No devices to test")
	}

	device := devices[0]
	_, err = unsafe.GetProfileUIDFromDevice(cl, device.UID)
	if err != nil {
		t.Errorf("GetProfileUIDFromDevice failed: %v", err)
	}
}

func TestProfiles(t *testing.T) {
	if !testCfg.DoProfileTests {
		t.Skip("Profile tests disabled (set do_profile_tests: true in config.json)")
	}
	if testCfg.ProfileTestDeviceUID == "" {
		t.Skip("No profile_test_device_uid configured")
	}

	t.Run("SetProfileForDevice", setProfileForDevice)
}

func setProfileForDevice(t *testing.T) {
	deviceUID := testCfg.ProfileTestDeviceUID

	originalProfile, err := unsafe.GetProfileUIDFromDevice(cl, deviceUID)
	if err != nil {
		t.Fatalf("GetProfileUIDFromDevice failed: %v", err)
	}

	profiles, err := unsafe.GetAvailableProfiles(cl)
	if err != nil {
		t.Fatalf("GetAvailableProfiles failed: %v", err)
	}

	var newProfile string
	for uid := range profiles {
		if uid != originalProfile {
			newProfile = uid
			break
		}
	}
	if newProfile == "" {
		t.Skip("Only one profile available, cannot test switching")
	}

	err = unsafe.SetProfileForDevice(cl, deviceUID, newProfile)
	if err != nil {
		t.Fatalf("SetProfileForDevice failed: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	currentProfile, err := unsafe.GetProfileUIDFromDevice(cl, deviceUID)
	if err != nil {
		t.Fatalf("GetProfileUIDFromDevice (verify) failed: %v", err)
	}
	if currentProfile != newProfile {
		t.Errorf("Profile not changed: got %s, want %s", currentProfile, newProfile)
	}

	err = unsafe.SetProfileForDevice(cl, deviceUID, originalProfile)
	if err != nil {
		t.Errorf("Failed to restore original profile: %v", err)
	}
}

func getDeviceName(t *testing.T) {
	devices, err := unsafe.GetDevices(cl)
	if err != nil {
		t.Fatalf("GetDevices failed: %v", err)
	}

	if len(devices) == 0 {
		t.Skip("No devices to test")
	}

	device := devices[0]
	name, err := unsafe.GetDeviceName(cl, device.UID)
	if err != nil {
		t.Fatalf("GetDeviceName failed: %v", err)
	}

	if name == "" {
		t.Error("GetDeviceName returned empty string")
	}
}

func getDeviceIP(t *testing.T) {
	devices, err := unsafe.GetDevices(cl)
	if err != nil {
		t.Fatalf("GetDevices failed: %v", err)
	}

	if len(devices) == 0 {
		t.Skip("No devices to test")
	}

	device := devices[0]
	ip, err := unsafe.GetDeviceIP(cl, device.UID)
	if err != nil {
		t.Fatalf("GetDeviceIP failed: %v", err)
	}

	if ip != "" && !strings.Contains(ip, ".") {
		t.Errorf("GetDeviceIP returned invalid IP: %s", ip)
	}
}

func TestDevices(t *testing.T) {
	if !testCfg.DoDeviceTests {
		t.Skip("Device tests disabled (set do_device_tests: true in config.json)")
	}
	if testCfg.DeviceTestUID == "" {
		t.Skip("No device_test_uid configured")
	}

	t.Run("SetName", setName)
	t.Run("SetName_Validation", setNameValidation)
	t.Run("SetIP", setIP)
}

func setName(t *testing.T) {
	deviceUID := testCfg.DeviceTestUID

	originalName, err := unsafe.GetDeviceName(cl, deviceUID)
	if err != nil {
		t.Fatalf("GetDeviceName failed: %v", err)
	}
	t.Logf("Original name: %s", originalName)

	newName := "test-device-123"

	err = unsafe.SetName(cl, deviceUID, newName)
	if err != nil {
		t.Fatalf("SetName failed: %v", err)
	}

	currentName, err := unsafe.GetDeviceName(cl, deviceUID)
	if err != nil {
		t.Fatalf("GetDeviceName (verify) failed: %v", err)
	}
	if currentName != newName {
		t.Errorf("Name not changed: got %s, want %s", currentName, newName)
	}

	err = unsafe.SetName(cl, deviceUID, originalName)
	if err != nil {
		t.Logf("Could not restore original name %q: %v", originalName, err)
	}
}

func setNameValidation(t *testing.T) {
	err := unsafe.SetName(cl, "dummy-uid", "")
	if err == nil {
		t.Error("SetName should reject empty names")
	}

	longName := strings.Repeat("a", 64)
	err = unsafe.SetName(cl, "dummy-uid", longName)
	if err == nil {
		t.Error("SetName should reject names > 63 characters")
	}

	err = unsafe.SetName(cl, "dummy-uid", "invalid name")
	if err == nil {
		t.Error("SetName should reject names with spaces")
	}
}

func setIP(t *testing.T) {
	deviceUID := testCfg.DeviceTestUID

	devices, err := unsafe.GetAllDevices(cl)
	if err != nil {
		t.Fatalf("GetAllDevices failed: %v", err)
	}
	t.Logf("Found %d devices (including inactive)", len(devices))

	var originalIP, deviceMAC string
	for _, d := range devices {
		if d.UID == deviceUID {
			originalIP = d.IP
			deviceMAC = d.MAC
			break
		}
	}
	if deviceMAC == "" {
		t.Fatalf("Device %s not found", deviceUID)
	}
	t.Logf("Original IP: %s, MAC: %s", originalIP, deviceMAC)

	ipChanged := false
	defer func() {
		if !ipChanged {
			return
		}
		time.Sleep(2 * time.Second)
		if err := unsafe.SetIP(cl, deviceUID, originalIP, true); err != nil {
			t.Logf("Could not restore original IP %s: %v", originalIP, err)
		} else {
			t.Logf("Restored original IP %s", originalIP)
		}
	}()

	usedIPs := make(map[string]bool)
	for _, d := range devices {
		if d.IP != "" {
			usedIPs[d.IP] = true
		}
	}

	parts := strings.Split(originalIP, ".")
	if len(parts) != 4 {
		t.Fatalf("Invalid IP format: %s", originalIP)
	}
	prefix := fmt.Sprintf("%s.%s.%s.", parts[0], parts[1], parts[2])

	var newIP string
	for octet := 100; octet <= 200; octet++ {
		candidate := fmt.Sprintf("%s%d", prefix, octet)
		if usedIPs[candidate] {
			continue
		}
		err = unsafe.SetIP(cl, deviceUID, candidate, true)
		if err == nil {
			newIP = candidate
			ipChanged = true
			break
		}
		t.Logf("IP %s rejected, trying next...", candidate)
	}
	if newIP == "" {
		t.Skip("Could not find an available IP in range .100-.200")
	}
	t.Logf("Successfully set IP to: %s", newIP)

	time.Sleep(2 * time.Second)
	devices, err = unsafe.GetAllDevices(cl)
	if err != nil {
		t.Fatalf("GetAllDevices (verify) failed: %v", err)
	}
	var currentIP string
	for _, d := range devices {
		if d.MAC == deviceMAC {
			currentIP = d.IP
			t.Logf("Found device: UID=%s IP=%s", d.UID, d.IP)
			break
		}
	}
	if currentIP != newIP {
		t.Errorf("IP not changed: got %s, want %s", currentIP, newIP)
	} else {
		t.Logf("Verified IP changed to %s", newIP)
	}
}
