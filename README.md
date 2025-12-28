# go-fritzbox-api

Go client library for AVM FRITZ!Box routers and smart home devices.

## Attribution

Authentication based on [Philipp Franke](https://github.com/philippfranke)'s [go-fritzbox](https://github.com/philippfranke/go-fritzbox).

## Installation

```bash
go get github.com/ByteSizedMarius/go-fritzbox-api
```

## Quick Start

```go
package main

import (
    "fmt"
    fritzbox "github.com/ByteSizedMarius/go-fritzbox-api"
    "github.com/ByteSizedMarius/go-fritzbox-api/smarthome"
)

func main() {
    client := fritzbox.New("username", "password")
    if err := client.Connect(); err != nil {
        panic(err)
    }
    defer client.Close()

    overview, _ := smarthome.GetOverview(client)
    for _, device := range overview.Devices {
        fmt.Printf("%s (%s)\n", device.Name, device.ProductName)
    }
}
```

## Packages

| Package | API | Stability | Description |
|---------|-----|-----------|-------------|
| [`aha/`](aha/) | AHA HTTP | Stable | Smart home (DECT devices, thermostats) |
| [`smarthome/`](smarthome/) | REST | Stable | Full config, JSON-based (FRITZ!OS 8.20+) |
| [`unsafe/`](unsafe/) | data.lua | Unstable | Router internals, network devices |

## Scope

Only a subset of device types is implemented for both APIs. Main focus is on **HKR (radiator thermostats)** - see [HKR Documentation](docs/hkr.md) for details. Other device types (buttons, temperature sensors, HAN-FUN) have partial support.

See the package READMEs linked above for detailed usage and available functions.

## API Landscape

FRITZ!Box has two official smart home APIs:
- **AHA HTTP Interface** (`/webservices/homeautoswitch.lua`): XML-based, available since FRITZ!OS 5.53; [Docs](https://avm.de/fileadmin/user_upload/Global/Service/Schnittstellen/AHA-HTTP-Interface.pdf)
- **Smart Home REST API** (`/api/v0/smarthome/...`): JSON-based, requires FRITZ!OS 8.20+. More comprehensive; [OpenAPI spec](https://fritz.support/resources/SmarthomeRestApiFRITZOS82.yaml)

## Compatibility

Tested with FRITZ!OS 8.21 on the 6690 Cable. Smart home implementations (DECT) are stable across versions and routers. Endpoints in the `unsafe/` package may break between firmware versions.

## Contributing

Issues, Pull Requests and [E-Mails](mailto:fritz@marius.codes) are welcome.
