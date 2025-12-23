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

resource "gitea_org" "test_org" {
  username    = "testorg"
  full_name   = "Test Organization"
  description = "A test organization"
  visibility  = "public"
}

data "gitea_org" "test_org" {
  org = gitea_org.test_org.username
}

output "org" {
  value = data.gitea_org.test_org
}