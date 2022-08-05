# go-fritzbox-api

**Warning: The parsing may currently be unstable and may panic if there is an unexpected response.**

## Attribution

This repository (structure, authentication) is based on [Philipp Franke](https://github.com/philippfranke)'s implementation ([go-fritzbox](https://github.com/philippfranke/go-fritzbox)).

I added parsing for some endpoints and made some of his code worse in the process ;)

## Features

**Clientlist**: 
`GetCLientList`
Get all currently connected devices in a list.

<details>
  <summary>[Expand] Clientlist datatype</summary>

```go
type Clientlist struct {
	Rootuid string `json:"rootuid"`
	Devices []struct {
		Devid     string `json:"devid"`
		Stateinfo struct {
			GuestOwe        bool `json:"guest_owe"`
			Active          bool `json:"active"`
			Guest           bool `json:"guest"`
			Online          bool `json:"online"`
			Blocked         bool `json:"blocked"`
			Realtime        bool `json:"realtime"`
			Notallowed      bool `json:"notallowed"`
			InternetBlocked bool `json:"internetBlocked"`
		} `json:"stateinfo,omitempty"`
		Profile    Profile
		Devtype    string   `json:"devtype"`
		Dist       int      `json:"dist"`
		Parent     string   `json:"parent"`
		Category   string   `json:"category"`
		Ownentry   bool     `json:"ownentry"`
		UID        string   `json:"UID"`
		Conn       string   `json:"conn"`
		Master     bool     `json:"master"`
		Ipinfo     []string `json:"ipinfo"`
		Updateinfo struct {
			State string `json:"state"`
		} `json:"updateinfo"`
		Gateway  bool `json:"gateway"`
		Nameinfo struct {
			Name string `json:"name"`
		} `json:"nameinfo,omitempty"`
		Children []interface{} `json:"children"`
		Conninfo []struct {
			Speed   string `json:"speed"`
			SpeedTx int    `json:"speed_tx"`
			SpeedRx int    `json:"speed_rx"`
			Desc    string `json:"desc"`
		} `json:"conninfo"`
	} `json:"devices"`
}
```
</details>

`Profile` is empty by default (see Docs for more info)

**Device**: `SetIP`, `RenameDevice`

**Profiles**: `GetAvailableProfiles`, `GetProfileUIDFromDevice`, `SetProfileForDevice`

**Logs**: `GetEventLog`, `GetEventLogUntil`

**Custom Requests**

## Usage

See [examples](/examples/main.go).

### Access remote fritzbox over ssl (untested)

See Philipp's example [here](https://github.com/philippfranke/go-fritzbox#access-remote-fritzbox-over-ssl)

## Contributing

Issues, Pull Requests and [E-Mails](mailto:fritz@marius.codes) are always welcome.

I accept suggestions for any new endpoints.

### How to add endpoints

I used burpsuites proxy with the inbuilt browser to look at the request, then sent it to the repeater and played around with the parameters until I found the parameters that were mandatory for the request to work (xhr, lang, oldpage, etc. can usually be omitted)

If you would like more info, don't hesitate to shoot me an email, I don't feel like writing an essay that no one's going to read . 