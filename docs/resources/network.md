# gestioip_network resource

Manages a GestioIP network inside a client.

## Example

```hcl
resource "gestioip_network" "example" {
  ip          = "192.168.50.0"
  bitmask     = 24
  description = "Terraform managed network"
  site        = "Lon"
  category    = "prod"
  comment     = "Created by Terraform"
  sync        = false
}
```

## Arguments

- `ip`:
  required network IP
- `bitmask`:
  required network bitmask
- `description`:
  required network description
- `site`:
  required GestioIP site
- `category`:
  required GestioIP category
- `comment`:
  optional network comment
- `sync`:
  optional boolean, defaults to `false`
- `client_name`:
  optional override for the provider-level client

## Attributes

- `id`:
  GestioIP internal network identifier
- `ip_version`:
  `v4` or `v6` as reported by GestioIP

## Notes

- `ip` and `bitmask` force replacement
- `site` and `category` must already exist in GestioIP
- by default, apply fails if a network with the same `ip` and `bitmask` already exists
- if the provider sets `allow_overwrite = true`, apply updates the existing network and adopts it into Terraform state

## Import

Import format:

```bash
terraform import gestioip_network.example 192.168.50.0/24
```

Or, if you prefer to pass the client inline:

```bash
terraform import gestioip_network.example DEFAULT|192.168.50.0/24
```
