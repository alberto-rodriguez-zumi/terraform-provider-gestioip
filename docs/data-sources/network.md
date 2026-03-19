# gestioip_network data source

Reads a GestioIP network by IP and bitmask.

## Example

```hcl
data "gestioip_network" "example" {
  ip      = "192.168.50.0"
  bitmask = 24
}
```

## Arguments

- `ip`:
  required network IP
- `bitmask`:
  required network bitmask
- `client_name`:
  optional override for the provider-level client

## Returned attributes

- `id`
- `client_name`
- `ip`
- `bitmask`
- `description`
- `site`
- `category`
- `comment`
- `sync`
- `ip_version`
