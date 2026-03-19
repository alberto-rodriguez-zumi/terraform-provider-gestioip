# terraform-provider-gestioip

Terraform provider for GestioIP built with Terraform Plugin Framework.

This repository currently targets the free GestioIP 3.5 container image and adapts to the behavior that was validated against `gestioip/gestioip:3570` on March 18, 2026.

## Status

Current provider coverage:

- `gestioip_network` resource
- `gestioip_network` data source
- `gestioip_host` resource
- `gestioip_host` data source
- `gestioip_vlan` resource
- `gestioip_vlan` data source

## Compatibility notes

The official GestioIP 3.5 API guide documents `.../gestioip/api/api.cgi` with Basic Auth, but the free container image behaves differently:

- `/gestioip/api/api.cgi` was not exposed in the tested image
- `/gestioip/intapi.cgi` was exposed
- `intapi.cgi` required a cookie-backed session flow
- networks were readable through `listNetworks`
- hosts and VLANs did not expose useful read APIs in `intapi.cgi`

Because of that, the provider uses a hybrid or frontend-backed approach depending on the entity:

- networks:
  - reads use `intapi.cgi` plus `listNetworks`
  - writes use `res/ip_insertred.cgi`, `res/ip_modred.cgi` and `res/ip_deletered.cgi`
- hosts:
  - reads parse `ip_show.cgi`
  - writes use `res/ip_modip.cgi`
  - deletes use `res/ip_deleteip.cgi`
- VLANs:
  - reads parse `show_vlans.cgi`
  - writes use `res/ip_insertvlan.cgi` and `res/ip_modvlan.cgi`
  - deletes use `res/ip_deletevlan.cgi`

Operational constraints observed in the tested image:

- `client_name` must match a GestioIP client visible in the UI
- `site` and `category` values for networks and hosts must already exist in GestioIP
- VLAN colors must match values supported by the GestioIP form
- resources support `terraform import`

Behavior validated against a local GestioIP instance on March 19, 2026:

- by default, creating a `host`, `network` or `vlan` that already exists returns an error during `apply`
- if `allow_overwrite = true` is set in the provider, the provider updates the existing object and adopts it into Terraform state
- `terraform import` works for `host`, `network` and `vlan`

## Installation

This provider is not documented here as a Terraform Registry provider yet. The current installation path is from the GitHub release artifacts.

Release page:

- [v0.1](https://github.com/alberto-rodriguez-zumi/terraform-provider-gestioip/releases/tag/v0.1)

### macOS arm64

1. Download `terraform-provider-gestioip_v0.1_darwin_arm64.zip`.
2. Create the local plugin directory:

```bash
mkdir -p ~/.terraform.d/plugins/alberto-rodriguez-zumi/gestioip/0.1/darwin_arm64
```

3. Unzip the asset into that directory.
4. Rename the binary to the Terraform local plugin naming convention:

```bash
mv ~/.terraform.d/plugins/alberto-rodriguez-zumi/gestioip/0.1/darwin_arm64/terraform-provider-gestioip_v0.1_darwin_arm64 \
  ~/.terraform.d/plugins/alberto-rodriguez-zumi/gestioip/0.1/darwin_arm64/terraform-provider-gestioip_v0.1
```

### macOS x86_64

1. Download `terraform-provider-gestioip_v0.1_darwin_amd64.zip`.
2. Create the local plugin directory:

```bash
mkdir -p ~/.terraform.d/plugins/alberto-rodriguez-zumi/gestioip/0.1/darwin_amd64
```

3. Unzip the asset into that directory.
4. Rename the binary:

```bash
mv ~/.terraform.d/plugins/alberto-rodriguez-zumi/gestioip/0.1/darwin_amd64/terraform-provider-gestioip_v0.1_darwin_amd64 \
  ~/.terraform.d/plugins/alberto-rodriguez-zumi/gestioip/0.1/darwin_amd64/terraform-provider-gestioip_v0.1
```

## Provider configuration

Minimal provider block:

```hcl
terraform {
  required_providers {
    gestioip = {
      source  = "alberto-rodriguez-zumi/gestioip"
      version = "0.1"
    }
  }
}

provider "gestioip" {
  base_url    = "http://localhost"
  client_name = "DEFAULT"
  username    = "gipadmin"
  password    = "password"
  allow_overwrite = false
}
```

Provider attributes:

- `base_url`:
  GestioIP base URL, for example `http://localhost`
- `client_name`:
  optional default client for resources and data sources
- `allow_overwrite`:
  optional boolean, defaults to `false`; if set to `true`, create operations overwrite an existing object with the same identity and adopt it into Terraform state
- `username`:
  GestioIP username
- `password`:
  GestioIP password

## Import

Supported import identifiers:

- `gestioip_network`: `[client_name|]<ip>/<bitmask>`
- `gestioip_host`: `[client_name|]<ip>`
- `gestioip_vlan`: `[client_name|]<number>`

If `client_name` is already configured in the provider, the import ID can omit it.

Examples:

```bash
terraform import gestioip_network.example 10.63.0.0/24
terraform import gestioip_host.example 10.63.0.10
terraform import gestioip_vlan.example 263
```

## Full example

The repository includes a combined example in [examples/provider/provider.tf](examples/provider/provider.tf).

## Resources

- [gestioip_network](docs/resources/network.md)
- [gestioip_host](docs/resources/host.md)
- [gestioip_vlan](docs/resources/vlan.md)

## Data sources

- [gestioip_network](docs/data-sources/network.md)
- [gestioip_host](docs/data-sources/host.md)
- [gestioip_vlan](docs/data-sources/vlan.md)

## Testing

Unit and provider tests:

```bash
GOCACHE=$(pwd)/.cache/go-build GOMODCACHE=$(pwd)/.cache/gomod go test ./...
```

Integration tests are env-gated and do not run by default. Examples:

```bash
GESTIOIP_BASE_URL=http://localhost \
GESTIOIP_USERNAME=gipadmin \
GESTIOIP_PASSWORD=password \
GESTIOIP_CLIENT_NAME=DEFAULT \
GOCACHE=$(pwd)/.cache/go-build \
GOMODCACHE=$(pwd)/.cache/gomod \
go test ./internal/client -run TestClientNetworkLifecycleIntegration -count=1 -v
```

```bash
GESTIOIP_BASE_URL=http://localhost \
GESTIOIP_USERNAME=gipadmin \
GESTIOIP_PASSWORD=password \
GESTIOIP_CLIENT_NAME=DEFAULT \
GOCACHE=$(pwd)/.cache/go-build \
GOMODCACHE=$(pwd)/.cache/gomod \
go test ./internal/client -run TestClientHostLifecycleIntegration -count=1 -v
```

```bash
GESTIOIP_BASE_URL=http://localhost \
GESTIOIP_USERNAME=gipadmin \
GESTIOIP_PASSWORD=password \
GESTIOIP_CLIENT_NAME=DEFAULT \
GOCACHE=$(pwd)/.cache/go-build \
GOMODCACHE=$(pwd)/.cache/gomod \
go test ./internal/client -run TestClientVLANLifecycleIntegration -count=1 -v
```
