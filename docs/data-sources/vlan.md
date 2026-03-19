# gestioip_vlan data source

Reads a GestioIP VLAN by number.

## Example

```hcl
data "gestioip_vlan" "example" {
  number = "200"
}
```

## Arguments

- `number`:
  required VLAN number
- `client_name`:
  optional override for the provider-level client

## Returned attributes

- `id`
- `client_name`
- `number`
- `name`
- `description`
- `bg_color`
- `font_color`
