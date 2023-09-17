# go fritzbox api

Quick & dirty API implementation for the AVM FRITZ!Box using Golang and REST.

**Warning:** There is no official API for some of these functions (except
any [smarthome-related stuff](https://avm.de/fileadmin/user_upload/Global/Service/Schnittstellen/AHA-HTTP-Interface.pdf)).
This means, consistency across software-versions or different routers/repeaters is not guaranteed.

## Attribution

This repository (structure, authentication) is based on [Philipp Franke](https://github.com/philippfranke)'s
implementation ([go-fritzbox](https://github.com/philippfranke/go-fritzbox)).

I added parsing for some endpoints and made some of his code worse in the process ;)

## Contributing

Issues, Pull Requests and [E-Mails](mailto:fritz@marius.codes) are always welcome.

I accept suggestions for any new endpoints.

## Compatiblity

This library works with FRITZ!OS 7.56 with the 6690 Cable. The smarthome-implementations (DECT) are stable and
compatible with all versions and routers, except for any Pya endpoints and endpoints explicitely marked as unstable.

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

I will eventually come around to writing a proper documentation. For now, you can look at the [examples](/examples/) and
check the
[godoc](https://pkg.go.dev/github.com/ByteSizedMarius/go-fritzbox-api).

### Setting Up the PyAdapter

```
pip install selenium-wire
```

Try running the script using commandline-args:

```
python pyAdapter/main.py OK;DEBUG;LOGIN http://192.168.178.1/ <current SID>
```

You can get an SID by logging into the Fritzbox in your brower and right-click + copying the link from (for example) one
of the buttons in the top right (MyFritz, Fritz!Nas). The SID is contained in the link.

Remove the Debug if executing in headless (for example when connected via ssh):

```
python pyAdapter/main.py OK;LOGIN http://192.168.178.1/ <current SID>
```

Arguments can be given to the ChromeDriver like so:

```
python pyAdapter/main.py OK;DEBUG;ARGS --no-sandbox|--disable-gpu|disable-dev-shm-usage;LOGIN http://192.168.178.1/ <current SID>
```

If there are crashes, `no-sandbox`, `disable-gpu` and/or `disable-dev-shm-usage` are usually good arguments to try.

If that works and logs into the FritzBox, it should be okay to start the PyAdapter from the Go-Program.

## Implementation

This Library is a Project I maintain and develop based on my personal needs. It's public because there aren't many
maintained alternatives, not because I'm necessarily proud of its implementation. Code quality is not great, especially
older code. There are no tests, because what I would really like to test (whether what I'm sending the FritzBox is still
correct, for example after a FritzOS-Update) is very difficult to implement. For example, for Testing if changing the
Name for a Device works, I would have to get some Device, change the Name, check if it works, then change it back,
without breaking anything in my internal Network. Doing this in a way that would allow other people to also execute
these Tests is even more difficult.

### Types of API-Endpoints

There are two (three) types of API-Endpoints used in this libary:

- DECT: These are the endpoints used for smarthome-devices. They are documented in
  the [official documentation](https://avm.de/fileadmin/user_upload/Global/Service/Schnittstellen/AHA-HTTP-Interface.pdf)
  and usually do not change. These Endpoints are implemented in the `smarthome`-package and their respective Functions
  have the Prefix `DECT`.
- Frontend Endpoints: POST-Request are sent to the FritzBox mimicking the Requests the FrontEnd would send during normal
  User Interaction. Because the required Parameters change regularly, all of these Endpoints are unstable and may break
  at any time. They are usually marked as Unstable in their GoDoc. In the beginning I was planning to use
  this type of Request for most of the Functionality I need, but I quickly ran into a Roadblock with the Requests'
  Parameters. For Example when editing a Device, the Request sent by the Frontend always contains all the Devices'
  Parameters, even if they haven't changed. This means, no matter what Setting you want to change, you would have to get
  all the Parameters for a Device, change the one you want to change, then send all the Parameters back to the FritzBox.
  While sometimes, Parameters are optional in the Request, the FritzBox seems to reset these Params to their default if
  you don't send them. This proved basically impossible, because I would have to manually parse the Parameters for every
  Request from the respective Setting Pages HTML. Instead, this Type of Request is only used for Endpoints that
  fetch and parse Data, like the `GetClientList`-Endpoint (There are still older Endpoints that use this Type of
  Request, but I'm planning to change them at some point).
- PYA Endpoints: What I came up with for the Problem described above are PYA Endpoints, where PYA stands for Python
  Adapter. When wanting to use PYA Endpoints, a PyAdapter-Struct has to be created and started. This will start
  the [Python-Script](/pyAdapter/) which in turn starts a headless Chromedriver and listens to commands on stdin. The
  Chromedriver is used to send an empty Request to the FritzBox in order to retrieve the required Parameters, which are
  then returned to the Go-Program. After the Parameters have been retrieved, it is then simple to change some Parameters
  and send the correct Post-Request to the FritBox. Retrieving the Parameters takes some Time (1-3 seconds), but
  generally, it's not much slower than some of the Endpoints of the official Smarthome-API. To improve the Time it takes
  to retrieve the Parameters, the PyAdapter is started concurrently to the Go-Program, meaning the Chromedriver is
  already started and logged in to the Fritzbox, which takes about 7 seconds. This means a "cold" start, where the
  PyAdapter first has to start and log in before sending the request can take up to 10 seconds. Currently the PYA
  expects the Webdriver in the PyAdapter directory or in the Path.

## Roadmap

- Better documentation
- Pya Cache
- Implementing more smarthome-capabilities/HanFun-interfaces
- Implementing endpoints for smarthome-templates (`gettemplatelistinfos`, `applytemplate`)

### Access remote fritzbox over ssl (untested)

See Philipp's example [here](https://github.com/philippfranke/go-fritzbox#access-remote-fritzbox-over-ssl)