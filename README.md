# go fritzbox api

Quick & dirty API implementation for the AVM FRITZ!Box using Golang and REST.

**Warning:** There is no official API for most these functions (except
any [smarthome-related stuff](https://avm.de/fileadmin/user_upload/Global/Service/Schnittstellen/AHA-HTTP-Interface.pdf)).
This means, consistency across software-versions or different routers/repeaters is not guaranteed.

## Attribution

This repository (structure, authentication) is based on [Philipp Franke](https://github.com/philippfranke)'s
implementation ([go-fritzbox](https://github.com/philippfranke/go-fritzbox)).

I added parsing for some endpoints and made some of his code worse in the process ;)

## Compatiblity

This library works with FRITZ!OS 07.50 with the 6690 Cable. The smarthome-implementations (DECT) are stable and
compatible with all versions and routers, except for the endpoints explicitely marked as unstable.

All other endpoints mostly use REST-Methods to call the same endpoints the frontend does. These endpoints are neither
consistent across versions nor backwards-compatible. They may also break unexpectedly.

## Features

**Smarthome:** Using the official http-API, some types of smarthome-devices have been implemented. See
the [separate readme](SMARTHOME.md).

**Clientlist:**

- `GetCLientList` Get all currently connected devices in a list.

- `Profile` is empty by default (see Docs for more info)

**Device:** `SetIP`, `SetName`

**Profiles:** `GetAvailableProfiles`, `GetProfileUIDFromDevice`, `SetProfileForDevice`

**Logs:** `GetEventLog`, `GetEventLogUntil`

**Statistics:** `GetTrafficStats`

**Custom Requests:** (See [examples](/examples/main.go))

## Usage

See [examples](/examples/).

### Access remote fritzbox over ssl (untested)

See Philipp's example [here](https://github.com/philippfranke/go-fritzbox#access-remote-fritzbox-over-ssl)

## Contributing

Issues, Pull Requests and [E-Mails](mailto:fritz@marius.codes) are always welcome.

I accept suggestions for any new endpoints.

### Adding endpoints

I used burpsuites proxy with the inbuilt browser to look at the request, then sent it to the repeater and played around
with the parameters until I found the parameters that were mandatory for the request to work (xhr, lang, oldpage, etc.
can usually be omitted).

If you would like more info, don't hesitate to shoot me an email, I don't feel like writing an essay that no one's going
to read.

## Roadmap

- Better documentation
- Implementing more smarthome-capabilities/HanFun-interfaces
- Implementing endpoints for smarthome-templates (`gettemplatelistinfos`, `applytemplate`)