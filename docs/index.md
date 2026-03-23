---
page_title: "Provider: GestioIP"
subcategory: ""
description: |-
  The GestioIP provider manages networks, hosts, and VLANs in GestioIP.
---

# GestioIP Provider

The GestioIP provider manages networks, hosts, and VLANs in GestioIP.

This provider is built with Terraform Plugin Framework and has been validated against two GestioIP variants:

- GestioIP 3.5 free container image, where some operations rely on a hybrid approach between `intapi.cgi` and the frontend CGI flows
- GestioIP 3.2 legacy deployments, where both API-related and frontend routes may require Basic Auth

!> This provider has only been tested against GestioIP 3.2 and GestioIP 3.5. It may also work with other GestioIP versions, but that compatibility is not currently guaranteed. The provider is published in its current state for testing purposes.

## Example Usage

```hcl
terraform {
  required_providers {
    gestioip = {
      source  = "alberto-rodriguez-zumi/gestioip"
      version = "0.3.0"
    }
  }
}

variable "gestioip_password" {
  description = "GestioIP password"
  type        = string
  sensitive   = true
}

provider "gestioip" {
  base_url        = "https://gestioip.example.com"
  client_name     = "DEFAULT"
  username        = "gipadmin"
  password        = var.gestioip_password
  allow_overwrite = false
}
```

## Argument Reference

- `base_url` - (Required) Base URL of the GestioIP instance, for example `https://gestioip.example.com` or `http://localhost`.
- `username` - (Required) Username used to authenticate against GestioIP. `gipadmin` is the default administrative username in GestioIP.
- `password` - (Required, Sensitive) Password used to authenticate against GestioIP.
- `client_name` - (Optional) Default GestioIP client name used by resources and data sources that operate within a client context.
- `allow_overwrite` - (Optional) Defaults to `false`. When set to `true`, create operations for supported resources update an existing object with the same identity and adopt it into Terraform state instead of returning an error.

## Supported Resources

- `gestioip_network`
- `gestioip_host`
- `gestioip_vlan`

## Supported Data Sources

- `gestioip_network`
- `gestioip_host`
- `gestioip_vlan`

## Import Support

The provider supports import for all currently implemented resources.

- `gestioip_network`: `[client_name|]<ip>/<bitmask>`
- `gestioip_host`: `[client_name|]<ip>`
- `gestioip_vlan`: `[client_name|]<number>`

If `client_name` is already configured in the provider block, the import ID can omit it.

## Notes

- In the free GestioIP image tested for this provider, the documented `api/api.cgi` endpoint was not exposed.
- Network reads are handled through `intapi.cgi` and `listNetworks`.
- Network, host, and VLAN writes are implemented using the frontend CGI flows exposed by the free image.
- In the tested GestioIP 3.2 deployment, frontend and API-related routes required Basic Auth and network reads fell back to frontend HTML parsing.
- In legacy installations, `client_name` must match the client name exposed in the UI and should not be assumed to be `DEFAULT`.
- `site` and `category` values for networks and hosts must already exist in GestioIP.
- VLAN colors must match values supported by the GestioIP UI.
