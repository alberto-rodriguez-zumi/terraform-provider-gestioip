terraform {
  required_providers {
    gestioip = {
      source = "alberto-rodriguez-zumi/gestioip"
    }
  }
}

provider "gestioip" {
  base_url = "https://gestioip.example.com/api"
  username = "admin"
  password = "change-me"
}

