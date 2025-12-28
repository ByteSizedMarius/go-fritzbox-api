# AHA HTTP Interface

Go wrapper for the AVM Home Automation HTTP Interface.

## API Background

AHA HTTP Interface (`/webservices/homeautoswitch.lua`)
- XML-based, available since FRITZ!OS 5.53
- [AVM Documentation](https://avm.de/fileadmin/user_upload/Global/Service/Schnittstellen/AHA-HTTP-Interface.pdf)
- Somewhat difficult to work with, missing some functionality

## Usage

```go
import (
    "github.com/ByteSizedMarius/go-fritzbox-api"
    "github.com/ByteSizedMarius/go-fritzbox-api/aha"
)

client := fritzbox.New("username", "password")
client.Connect()

// Get all devices
deviceList, err := aha.GetDeviceList(client)

// Find device with specific capability
for _, device := range deviceList.Devices {
    if device.HasCapability(aha.CHKR) {
        hkr := aha.GetCapability[*aha.Hkr](device)

        // Read temperature
        temp, _ := hkr.DECTGetSoll(client)

        // Set temperature
        hkr.DECTSetSoll(client, 21.5)
    }
}
```

## Capabilities

| Constant | Type | Description |
|----------|------|-------------|
| `CHanfun` | `*HanFun` | HAN-FUN device |
| `CButton` | `*ButtonDevice` | Button device (e.g., FRITZ!DECT 440) |
| `CHKR` | `*Hkr` | Radiator thermostat (Heizkörperregler) |
| `CTempSensor` | `*Temperature` | Temperature sensor |

Many capability types (lights, sockets, blinds, etc.) are not yet implemented - only devices I actually own are supported for now. PRs are welcome.

## HKR (Radiator Thermostat)

See [HKR Documentation](../docs/hkr.md) for comprehensive coverage of:
- Temperature presets (comfort/reduced)
- Operating modes (boost, window open, summer, holiday)
- Timer schedules and locks
- API comparison (AHA vs REST)

## Key Types

- `DeviceList` - List of all AHA devices
- `Device` - Individual device with capabilities
- `Capability` - Interface for device features

## API Pattern

Methods prefixed with `DECT` make network requests:
```go
temp, err := hkr.DECTGetSoll(client)  // network request
```

Methods without prefix read cached values from the struct:
```go
temp := hkr.GetSoll()  // local read, no network
```

Call `hkr.Reload(client)` to refresh cached values after external changes.
