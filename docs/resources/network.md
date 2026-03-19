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
- import is not implemented yet
