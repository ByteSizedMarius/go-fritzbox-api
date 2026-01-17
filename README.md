# go-fritzbox-api

Go client library for AVM FRITZ!Box routers and smart home devices.

## Attribution

- Authentication based on [Philipp Franke](https://github.com/philippfranke)'s [go-fritzbox](https://github.com/philippfranke/go-fritzbox)
- Window detector support based on contribution by [@btwotch](https://github.com/ByteSizedMarius/go-fritzbox-api/pull/1)

## Installation

```bash
go get github.com/ByteSizedMarius/go-fritzbox-api/v2
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/ByteSizedMarius/go-fritzbox-api/v2"
    "github.com/ByteSizedMarius/go-fritzbox-api/v2/smart"
)

func main() {
    client := fritzbox.New("username", "password")
    if err := client.Connect(); err != nil {
        panic(err)
    }

    thermostats, _ := smart.GetAllThermostats(client)
    for _, t := range thermostats {
        fmt.Printf("%s: %.1f°C\n", t.Name, t.CurrentTemp)
    }
}
```

Or run the [example CLI](smart/examples/thermostat/) directly:

```bash
go run github.com/ByteSizedMarius/go-fritzbox-api/v2/smart/examples/thermostat@latest -user=admin -pass=secret
```

See [smart/README.md](smart/README.md) for the full API, [examples](smart/examples/), and [thermostat concepts](docs/hkr.md).

## Packages

| Package | API | Description |
|---------|-----|-------------|
| [`smart/`](smart/) | REST | Wrapper for `rest/` with nicer API for selected functionality |
| [`rest/`](rest/) | REST | JSON API, generated types (FRITZ!OS 8.20+) |
| [`unsafe/`](unsafe/) | data.lua | Router internals (unstable) |
| [`aha/`](aha/) | AHA HTTP | (Legacy) XML API for DECT devices |

## Scope

Currently supported device types:

- **Thermostats** - full support (state, config, schedules, holidays)
- **Buttons** - partial support
- **Window detectors** - state and thermostat linking
- **Temperature sensors** - read-only

If you own any of the missing devices and would like to help with implementation, please let me know!

## API Landscape

FRITZ!Box has two official smart home APIs:

- **Smart Home REST API** (`/api/v0/smarthome/...`): JSON-based, requires FRITZ!OS 8.20+. More comprehensive; [OpenAPI spec](https://fritz.support/resources/SmarthomeRestApiFRITZOS82.yaml)
- **AHA HTTP Interface** (`/webservices/homeautoswitch.lua`): XML-based, available since FRITZ!OS 5.53; [Docs](https://avm.de/fileadmin/user_upload/Global/Service/Schnittstellen/AHA-HTTP-Interface.pdf)

## Compatibility

Tested with FRITZ!OS 8.21 on the 6690 Cable. Smart home implementations (DECT) are stable across versions and routers. Endpoints in the `unsafe/` package may break between firmware versions.

Breaking changes are possible in v2.X, but will always be released with a new tag.

## Contributing

Issues, Pull Requests and [E-Mails](mailto:fritz@marius.codes) are welcome.
