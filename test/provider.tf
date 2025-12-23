terraform {
  required_providers {
    gitea = {
      source = "hashicorp.com/maxsargentdev/gitea"
    }
  }
}

provider "gitea" {
  gitea_username = "root"
  gitea_password = "admin1234"
  gitea_hostname = "http://localhost:3000"
}