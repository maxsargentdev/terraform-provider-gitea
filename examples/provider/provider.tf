terraform {
  required_providers {
    gitea = {
      source = "hashicorp.com/maxsargentdev/gitea"
    }
  }
}

provider "gitea" {
  gitea_username = "your-username"
  gitea_password = "your-password"
  gitea_hostname = "https://gitea.example.com"
}
