# unsafe

Access to undocumented Fritz!Box internals via `data.lua` and `query.lua`.

## API Background

Internal Lua endpoints (`data.lua`, `query.lua`)
- Undocumented, may change between FRITZ!OS versions
- Used by the Fritz!Box web interface
- Provides access to features not exposed via REST or AHA APIs

## Usage

```go
import (
    "github.com/ByteSizedMarius/go-fritzbox-api/v2"
    "github.com/ByteSizedMarius/go-fritzbox-api/v2/unsafe"
)

client := fritzbox.New("username", "password")
client.Connect()

// Get all network devices
devices, _ := unsafe.GetDeviceList(client)

// Get mesh topology with connection speeds
mesh, _ := unsafe.GetMeshTopology(client)

// Get event log
logs, _ := unsafe.GetEventLog(client)

// Get traffic statistics
stats, _ := unsafe.GetTrafficStats(client)
```

## Functions

### Network Devices
- `GetDevices(c)` - Active devices with MAC, IP, name, flags
- `GetAllDevices(c)` - All devices including inactive (for DHCP reservations)
- `GetMeshTopology(c)` - Full mesh topology with connection speeds
- `FillProfiles(c, devices)` - Add profile info to device list

### Device Management
- `SetName(c, uid, name)` - Set device friendly name
- `SetIP(c, uid, ip, static)` - Set DHCP reservation
- `GetDeviceName(c, uid)` - Get device name
- `GetDeviceIP(c, uid)` - Get device IP

### Profiles
- `GetAvailableProfiles(c)` - List all access profiles
- `GetProfileUIDFromDevice(c, uid)` - Get profile assigned to device
- `SetProfileForDevice(c, deviceUID, profileUID)` - Assign profile

### Logs & Stats
- `GetEventLog(c)` - System event log
- `GetEventLogUntil(c, id)` - Events newer than ID
- `GetTrafficStats(c)` - Bandwidth statistics (today, week, month)
