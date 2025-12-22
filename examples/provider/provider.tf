terraform {
  required_providers {
    gitea = {
      source = "hashicorp.com/maxsargentdev/gitea"
    }
  }
}

provider "gitea" {}

resource "gitea_user" "test_user" {
  username = "test"
  email    = "test@gitea.local"
}

data "gitea_user" "test_user" {
  id = "1"
}

output "user" {
  value = data.gitea_user.test_user
}