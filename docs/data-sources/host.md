# gestioip_host data source

Reads a GestioIP host by IP.

## Example

```hcl
data "gestioip_host" "example" {
  ip = "192.168.50.10"
}
```

## Arguments

- `ip`:
  required host IP
- `client_name`:
  optional override for the provider-level client

## Returned attributes

- `id`
- `ip_int`
- `network_id`
- `client_name`
- `ip`
- `hostname`
- `description`
- `site`
- `category`
- `comment`
- `ip_version`
