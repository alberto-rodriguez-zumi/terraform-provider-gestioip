# gestioip_host resource

Manages a GestioIP host entry for an IP contained in an existing network.

## Example

```hcl
resource "gestioip_network" "example" {
  ip          = "192.168.50.0"
  bitmask     = 24
  description = "Terraform managed network"
  site        = "Lon"
  category    = "prod"
}

resource "gestioip_host" "example" {
  depends_on = [gestioip_network.example]

  ip          = "192.168.50.10"
  hostname    = "terraform-host"
  description = "Terraform managed host"
  site        = "Lon"
  category    = "server"
  comment     = "Created by Terraform"
}
```

## Arguments

- `ip`:
  required host IP
- `hostname`:
  required host name
- `site`:
  required GestioIP site
- `description`:
  optional host description
- `category`:
  optional host category
- `comment`:
  optional host comment
- `client_name`:
  optional override for the provider-level client

## Attributes

- `id`:
  GestioIP host identifier
- `ip_int`:
  internal integer identifier used by the GestioIP frontend delete flow
- `network_id`:
  internal network identifier of the containing network
- `ip_version`:
  `v4` or `v6`

## Notes

- the provider resolves the containing network automatically from the host IP
- `ip` forces replacement
- the target network must already exist
- `site` and `category` values must exist in GestioIP
