# go fritzbox api

Quick & dirty API implementation for the AVM FRITZ!Box using Golang and REST. 

**Warning: The parsing is currently rather unstable and may panic if there is an unexpected response. There is no official API for most these functions (except any [smarthome-related-stuff](https://avm.de/fileadmin/user_upload/Global/Service/Schnittstellen/AHA-HTTP-Interface.pdf)). This means, consistency across software-versions or different routers/repeaters is never guaranteed.**

## Attribution

This repository (structure, authentication) is based on [Philipp Franke](https://github.com/philippfranke)'s implementation ([go-fritzbox](https://github.com/philippfranke/go-fritzbox)).

I added parsing for some endpoints and made some of his code worse in the process ;)

## Compatiblity

Compatibility is currently unknown. This library was tested with FRITZ!OS 07.29 with the 6690 Cable. It uses REST-Methods to call the same endpoints the frontend uses. These endpoints may neither be consistent nor backwards-compatible. They may also break unexpectedly: For example, getting the full devicelist broke, when I added a FRITZ!OS-compatible smarthome-device because it uses a datastructure different from other devices.

## Features

**Smarthome:** Using the official http-API, some types of smarthome-devices have been implemented. See the [separate readme](SMARTHOME.md).

**Clientlist:** 
`GetCLientList`
Get all currently connected devices in a list.

<details>
  <summary>[Expand] Clientlist datatype</summary>

```go
type Clientlist struct {
	Rootuid string
	Devices []struct {
		Devid     string
		Stateinfo struct {
			GuestOwe        bool
			Active          bool
			Guest           bool
			Online          bool
			Blocked         bool
			Realtime        bool
			Notallowed      bool
			InternetBlocked bool
		}
		Profile    Profile
		Devtype    string
		Dist       int
		Parent     string
		Category   string
		Ownentry   bool
		UID        string
		Conn       string
		Master     bool
		Ipinfo     []string
		Updateinfo struct {
			State string
		}
		Gateway  bool
		Nameinfo struct {
			Name string
		} 
		Children []interface{}
		Conninfo []struct {
			Speed   string
			SpeedTx int
			SpeedRx int
			Desc    string
		}
	}
}
```
</details>

`Profile` is empty by default (see Docs for more info)

**Device:** `SetIP`, `RenameDevice`

**Profiles:** `GetAvailableProfiles`, `GetProfileUIDFromDevice`, `SetProfileForDevice`

**Logs:** `GetEventLog`, `GetEventLogUntil`

**Statistics:** `GetTrafficStats`

<details>
  <summary>[Expand] Trafficstats datatype</summary>

```go
type TrafficStatistics struct {
    LastMonth TrafficForDuration
    ThisWeek  TrafficForDuration 
    Today     TrafficForDuration 
    Yesterday TrafficForDuration 
    ThisMonth TrafficForDuration 
}

type TrafficForDuration struct {
    BytesSentHigh     string 
    BytesSentLow      string 
    BytesReceivedHigh string
    BytesReceivedLow  string 
}
```
</details>

**Custom Requests:** (See [examples](/examples/main.go))

## Usage

See [examples](/examples/main.go).

### Access remote fritzbox over ssl (untested)

See Philipp's example [here](https://github.com/philippfranke/go-fritzbox#access-remote-fritzbox-over-ssl)

## Contributing

Issues, Pull Requests and [E-Mails](mailto:fritz@marius.codes) are always welcome.

I accept suggestions for any new endpoints.

### Adding endpoints

I used burpsuites proxy with the inbuilt browser to look at the request, then sent it to the repeater and played around with the parameters until I found the parameters that were mandatory for the request to work (xhr, lang, oldpage, etc. can usually be omitted).

If you would like more info, don't hesitate to shoot me an email, I don't feel like writing an essay that no one's going to read.

## Roadmap

- Tests
- Better documentation
- Implementing more smarthome-capabilities/HanFun-interfaces
- Implementing endpoints for smarthome-templates (`gettemplatelistinfos`, `applytemplate`)