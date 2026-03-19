terraform {
  required_providers {
    gestioip = {
      source = "alberto-rodriguez-zumi/gestioip"
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

resource "gestioip_network" "example" {
  ip          = "192.168.50.0"
  bitmask     = 24
  description = "Terraform managed network"
  site        = "Lon"
  category    = "prod"
  comment     = "Created by Terraform"
  sync        = false
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

data "gestioip_network" "example" {
  ip      = gestioip_network.example.ip
  bitmask = gestioip_network.example.bitmask
}

data "gestioip_host" "example" {
  ip = gestioip_host.example.ip
}

resource "gestioip_vlan" "example" {
  number      = "200"
  name        = "terraform-vlan"
  description = "Terraform managed VLAN"
  bg_color    = "blue"
  font_color  = "white"
}

data "gestioip_vlan" "example" {
  number = gestioip_vlan.example.number
}
