# terraform-provider-gestioip

Terraform provider for GestioIP built with Terraform Plugin Framework.

This repository currently supports two GestioIP variants that have been validated against real instances:

- GestioIP 3.5 free container image (`gestioip/gestioip:3570`)
- GestioIP 3.2 legacy deployment protected with Basic Auth

> Warning
> This provider has only been tested against GestioIP 3.2 and GestioIP 3.5. It may also work with other GestioIP versions, but that compatibility is not currently guaranteed. The provider is published in its current state for testing purposes.

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

Additional behavior validated against a GestioIP 3.2 instance on March 23, 2026:

- the installation required Basic Auth on frontend and API-related routes
- `client_name` had to match the real client name exposed by the UI rather than assuming `DEFAULT`
- network reads needed a fallback to parse the frontend HTML when no usable JSON response was available
- `host`, `network`, and `vlan` all completed create, read, update, delete, and import successfully using temporary resources

The provider should therefore be treated as supporting two integration modes:

- GestioIP 3.5 free image:
  hybrid reads and writes across `intapi.cgi` and frontend CGI routes
- GestioIP 3.2 legacy deployments:
  Basic Auth plus frontend-backed fallbacks where the older installation does not expose the same JSON surface

## Installation

Install the provider from the Terraform Registry:

```hcl
terraform {
  required_providers {
    gestioip = {
      source  = "alberto-rodriguez-zumi/gestioip"
      version = "0.3.0"
    }
  }
}
```

## Provider configuration

Minimal provider block:

```hcl
terraform {
  required_providers {
    gestioip = {
      source  = "alberto-rodriguez-zumi/gestioip"
      version = "0.3.0"
    }
  }
}

provider "gestioip" {
  base_url        = "http://localhost"
  client_name     = "DEFAULT"
  username        = "gipadmin"
  password        = var.gestioip_password
  allow_overwrite = false
}
```

`gipadmin` is the default administrative username in GestioIP, so it is used here as an example. Use your own password or pass it through a variable or environment-backed workflow.

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

## Terraform Registry publishing

This repository includes the baseline files needed for Terraform Registry releases:

- [terraform-registry-manifest.json](terraform-registry-manifest.json)
- [.goreleaser.yml](.goreleaser.yml)
- [.github/workflows/release.yml](.github/workflows/release.yml)

The release workflow expects these GitHub repository secrets:

- `GPG_PRIVATE_KEY`
- `PASSPHRASE`

The manifest declares protocol version `6.0`, which matches Terraform Plugin Framework providers.

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

Client integration tests are env-gated and do not run by default. Examples:

```bash
GESTIOIP_BASE_URL=http://localhost \
GESTIOIP_USERNAME=gipadmin \
GESTIOIP_PASSWORD=your-password \
GESTIOIP_CLIENT_NAME=DEFAULT \
GOCACHE=$(pwd)/.cache/go-build \
GOMODCACHE=$(pwd)/.cache/gomod \
go test ./internal/client -run TestClientNetworkLifecycleIntegration -count=1 -v
```

```bash
GESTIOIP_BASE_URL=http://localhost \
GESTIOIP_USERNAME=gipadmin \
GESTIOIP_PASSWORD=your-password \
GESTIOIP_CLIENT_NAME=DEFAULT \
GOCACHE=$(pwd)/.cache/go-build \
GOMODCACHE=$(pwd)/.cache/gomod \
go test ./internal/client -run TestClientHostLifecycleIntegration -count=1 -v
```

```bash
GESTIOIP_BASE_URL=http://localhost \
GESTIOIP_USERNAME=gipadmin \
GESTIOIP_PASSWORD=your-password \
GESTIOIP_CLIENT_NAME=DEFAULT \
GOCACHE=$(pwd)/.cache/go-build \
GOMODCACHE=$(pwd)/.cache/gomod \
go test ./internal/client -run TestClientVLANLifecycleIntegration -count=1 -v
```

Provider acceptance tests are also env-gated and run full resource lifecycle coverage for both supported installation variants.

Available `make` targets:

```bash
make testacc-3.2
make testacc-3.5
make testacc
```

GestioIP 3.2 acceptance variables:

- `TF_ACC=1`
- `GESTIOIP32_BASE_URL`
- `GESTIOIP32_USERNAME`
- `GESTIOIP32_PASSWORD`
- `GESTIOIP32_CLIENT_NAME`
- optional: `GESTIOIP32_NETWORK_SITE`, `GESTIOIP32_NETWORK_CATEGORY`, `GESTIOIP32_HOST_SITE`, `GESTIOIP32_HOST_CATEGORY`, `GESTIOIP32_NETWORK_PREFIX`, `GESTIOIP32_NETWORK_START`, `GESTIOIP32_NETWORK_END`, `GESTIOIP32_VLAN_START`, `GESTIOIP32_VLAN_END`

GestioIP 3.5 acceptance variables:

- `TF_ACC=1`
- `GESTIOIP35_BASE_URL`
- `GESTIOIP35_USERNAME`
- `GESTIOIP35_PASSWORD`
- `GESTIOIP35_CLIENT_NAME`
- optional: `GESTIOIP35_NETWORK_SITE`, `GESTIOIP35_NETWORK_CATEGORY`, `GESTIOIP35_HOST_SITE`, `GESTIOIP35_HOST_CATEGORY`, `GESTIOIP35_NETWORK_PREFIX`, `GESTIOIP35_NETWORK_START`, `GESTIOIP35_NETWORK_END`, `GESTIOIP35_VLAN_START`, `GESTIOIP35_VLAN_END`

Examples:

```bash
TF_ACC=1 \
GESTIOIP32_BASE_URL=https://gestioip32.example.com \
GESTIOIP32_USERNAME=gipadmin \
GESTIOIP32_PASSWORD=your-password \
GESTIOIP32_CLIENT_NAME="Voxel Group" \
GESTIOIP32_NETWORK_SITE="ALL-DCs" \
GESTIOIP32_NETWORK_CATEGORY="DEV_TEST" \
GESTIOIP32_HOST_SITE="ALL-DCs" \
GOCACHE=$(pwd)/.cache/go-build \
GOMODCACHE=$(pwd)/.cache/go-mod \
go test ./internal/provider -run TestAccGestioIP32Lifecycle -count=1 -v
```

```bash
TF_ACC=1 \
GESTIOIP35_BASE_URL=http://localhost \
GESTIOIP35_USERNAME=gipadmin \
GESTIOIP35_PASSWORD=your-password \
GESTIOIP35_CLIENT_NAME=DEFAULT \
GESTIOIP35_NETWORK_SITE=Lon \
GESTIOIP35_NETWORK_CATEGORY=test \
GESTIOIP35_HOST_SITE=Lon \
GESTIOIP35_HOST_CATEGORY=server \
GOCACHE=$(pwd)/.cache/go-build \
GOMODCACHE=$(pwd)/.cache/go-mod \
go test ./internal/provider -run TestAccGestioIP35Lifecycle -count=1 -v
```

A manual GitHub Actions workflow is also available in [.github/workflows/acceptance.yml](.github/workflows/acceptance.yml). It expects lab connection details in repository variables and credentials in repository secrets.
