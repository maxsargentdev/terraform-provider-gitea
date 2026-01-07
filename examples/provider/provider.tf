terraform {
  required_providers {
    icegitea = {
      source = "hashicorp.com/maxsargentdev/icegitea"
    }
  }
}

provider "icegitea" {
  gitea_username = "your-username"
  gitea_password = "your-password"
  gitea_hostname = "https://gitea.example.com"
}
