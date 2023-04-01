# Smarthome

AVM provides a
semi-detailed [documentation](https://avm.de/fileadmin/user_upload/Global/Service/Schnittstellen/AHA-HTTP-Interface.pdf)
for their http-interface used to control their DECT devices.

## Checklist of fully implemented capabilities

I won't implement devices I do not own. Any help in completing this list is appreciated.
See [Contributing](README.md#contributing) or just open a PR. You can take the already implemented capabilities as a
template or make them better. I've previously had someone ask if they can send me a device --
it depends, just contact me.

### List of implemented endpoints

Additional information is usually available, but does not have its own endpoint. Instead, it is included with the
general
capability-info.

- [x] HanFunGerät (partially, see below)
- [ ] Licht
- [ ] Alarm
- [ ] Button
- [x] HKR (Heizkörperregler)
    - Notes: The device takes some seconds to receive the command and then some more seconds/minutes to update the paperview
      display. The device itself delivers the current temperature (Tist), but in order to refresh the current
      temperature, the `gettemperature`-endpoint from the `Temperature`-Capability should be used. Alternatively,
      use `Reload` in order to fetch and update all capability-values. This is similar to `IsWindowOpen`
      and `GetWindowOpenEndtime`. For these, only `Reload` is available, because there is no specific endpoint available
      for any capability.
    - [x] `gethkrtsoll` (returns current soll-temp)
    - [x] `sethkrtsoll` (set new soll-temp)
    - [x] `gethkrkomfort` (returns currently set komfort-temp)
    - [x] `gethkrabsenk` (returns currently set absenk-temp)
    - [x] `sethkrboost` (activate boost and set endtime)
    - [x] `sethkrwindowopen` (set window open with endtime)
- [ ] EnergieMesser
- [x] TempSensor
    - [x] `gettemperature` (returns current temperature information)
    - [x] `getbasicdevicestats` (returns temperature-history)
- [ ] Steckdose
- [ ] Repeater
- [ ] Mikrofon
- [x] HanFunUnit (partially, see below)
- [ ] Schaltbar
- [ ] Dimmbar
- [ ] LampeMitFarbtemp
- [ ] Rollladen

#### Implementation status of HanFun-Interfaces:

- [ ] HFKeepAlive
- [x] HFAlert
- [ ] HFOnOff
- [ ] HFLevelControl
- [ ] HFColorControl
- [ ] HFOpenClose
- [ ] HFOpenCloseConfig
- [ ] HFSimpleButton
- [ ] HFSuotaUpdate

Boilerplate-Code is available for all interfaces.

## Implementation

Every device has "function classes", which I call "capabilities". Every device can have multiple of these. For
example, a radiator-controller (Heizungskörperregler/HKR) has the capabilities to control radiators and measure room
temperature.
This modular setup means, that the datatypes need to do the heavy lifting here.

Every device has a map of capabilities, that are populated if available. The capabilities are accessed using
string-constants that contain the description of the capability from the official documentation. This is what it looks
like for a HKR:

```golang
dl, _ := cl.GetSmarthomeDevices()
d := dl.Devices[0]
fmt.Println(d.Capabilities)

Output:
[Heizkörperregler, Temperatursensor]
```

The capabilities and their respective functions can be accessed (among others) via the constants:

```golang
if d.HasCapability(CTempSensor) {
    c := d.Capabilities[CTempSensor]
}
```

This will result in the capability-interface, which is not very useful before casting it to its respective type. This
will allow access to its fields and functions:

```golang
ts := d.Capabilities[CTempSensor].(Temperature)
fmt.Println(ts.Celsius)
fmt.Println(ts.Offset)

Output:
215
0
```

The fields of the capabilities contain the raw responses in strings from the API as they are parsed directly from its
XML-responses. For some of them there are getters which deliver formatted data:

```golang
fmt.Println(ts.GetCelsius()) // float64

Output:
21.5
```

**Only functions starting with `DECT` will fetch the data from the fritzbox and update the local object**, functions without it will
return the locally stored value.

Every device, irrespective of capabilities contains miscellaneous information:

```golang
fmt.Println(d)

Output:
{Devicename: XXXX, Identifier: XXXX, ID: XXXX, Productname FRITZ!DECT XXX, Manufacturer: AVM, Firmware-Version: 05.XX, Present: 1, TX busy: 0, Capabilities: [Temperatursensor, Heizkörperregler]}
```

`present` indicates whether a device is connected (`1` = connected) - takes some time to update after device disconnects

`tx busy` indicates whether there is currently a command being sent to the device (`1` = tx busy)