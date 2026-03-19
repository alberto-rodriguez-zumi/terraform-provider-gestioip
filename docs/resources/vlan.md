# gestioip_vlan resource

Manages a GestioIP VLAN entry.

## Example

```hcl
resource "gestioip_vlan" "example" {
  number      = "200"
  name        = "terraform-vlan"
  description = "Terraform managed VLAN"
  bg_color    = "blue"
  font_color  = "white"
}
```

## Arguments

- `number`:
  required VLAN number
- `name`:
  required VLAN name
- `description`:
  optional VLAN description
- `bg_color`:
  optional VLAN background color, defaults to `blue`
- `font_color`:
  optional VLAN font color, defaults to `white`
- `client_name`:
  optional override for the provider-level client

## Attributes

- `id`:
  GestioIP VLAN identifier

## Notes

- `number` forces replacement
- the provider currently manages VLAN number, name, description and colors
- VLAN provider linkage is not exposed yet in this first cut
- by default, apply fails if a VLAN with the same `number` already exists
- if the provider sets `allow_overwrite = true`, apply updates the existing VLAN and adopts it into Terraform state

## Import

Import format:

```bash
terraform import gestioip_vlan.example 200
```

Or, if you prefer to pass the client inline:

```bash
terraform import gestioip_vlan.example DEFAULT|200
```
