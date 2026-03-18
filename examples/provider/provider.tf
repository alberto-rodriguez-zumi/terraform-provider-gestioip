terraform {
  required_providers {
    gestioip = {
      source = "alberto-rodriguez-zumi/gestioip"
    }
  }
}

provider "gestioip" {
  base_url    = "https://gestioip.example.com"
  client_name = "DEFAULT"
  username    = "admin"
  password    = "change-me"
}

resource "gestioip_network" "example" {
  ip          = "192.168.50.0"
  bitmask     = 24
  description = "Terraform managed network"
  site        = "MAD"
  category    = "prod"
  comment     = "Created by Terraform"
  sync        = false
}
