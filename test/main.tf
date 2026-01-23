terraform {
  required_providers {
    gitea = {
      source = "maxsargentdev/gitea"
    }
  }
}

provider "gitea" {
  gitea_username = "root"
  gitea_password = "admin1234"
  gitea_hostname = "http://localhost:3000"
}

# Create an organization
resource "gitea_org" "test_org" {
  name        = "testorg"
  full_name   = "Test Organization"
  description = "A test organization created by Terraform"
  visibility  = "public"
}

# Create a team in the organization
resource "gitea_team" "test_team" {
  org  = gitea_org.test_org.name
  name = "test-team"

  description = "A test team with custom permissions"

  can_create_org_repo       = true
  includes_all_repositories = false

  units_map = {
    "repo.code"       = "write" # Source code access (none, read, write, admin)
    "repo.issues"     = "write" # Issue tracker access
    "repo.pulls"      = "write" # Pull requests access
    "repo.releases"   = "read"  # Releases access
    "repo.ext_wiki"   = "none"  # External wiki access
    "repo.ext_issues" = "none"  # External issue tracker access
  }
}

# Create a user
resource "gitea_user" "test_user" {
  username   = "testuser"
  login_name = "login_name"
  email      = "testuser@gitea.local"
  password   = "testpassword123"
  full_name  = "Test User"
  active     = true
}

resource "gitea_user" "foo_user" {
  username   = "foobar"
  login_name = "foobar"
  email      = "foo.bar@zot.com"
  password   = "testpassword123"
  full_name  = "Foo Bar"
  active     = true
}

# Create a repository owned by root user
resource "gitea_repository" "test_repo" {
  username    = "root"
  name        = "test-repo"
  description = "A test repository created with Terraform"
  private     = true
  auto_init   = true
}

# Create a repository owned by the organization
resource "gitea_repository" "test_repo_for_org" {
  username    = gitea_org.test_org.name
  name        = "test-repo-for-org"
  description = "A test repository owned by an organization"
  private     = false
  auto_init   = true
}

# Add the user to the team
resource "gitea_team_membership" "test_membership" {
  team_id  = gitea_team.test_team.id
  username = gitea_user.test_user.username
}

# Add the organization's repository to the team
resource "gitea_team_repository" "test_team_repo_association" {
  org             = gitea_org.test_org.name
  team_name       = gitea_team.test_team.name
  repository_name = gitea_repository.test_repo_for_org.name
}

# Create a fork of the test repository to the organization
resource "gitea_fork" "test_fork" {
  owner        = "root"
  repo         = gitea_repository.test_repo.name
  organization = gitea_org.test_org.name
}

# Create branch protection on the test repository
resource "gitea_repository_branch_protection" "test_protection" {
  username  = "root"
  name      = gitea_repository.test_repo.name
  rule_name = "main"

  enable_push                    = true
  push_whitelist_users           = ["root"]
  required_approvals             = 1
  block_merge_on_outdated_branch = true
}

# Create a token for the authenticated user
resource "gitea_token" "test_token" {
  name = "terraform-test-token"
  scopes = [
    "read:issue",
  ]
}

# Create an organization actions secret
resource "gitea_org_actions_secret" "test_org_secret" {
  org         = gitea_org.test_org.name
  name        = "TEST_ORG_SECRET"
  data        = "my-secret-value"
  description = "A test organization secret for GitHub Actions"
}

# Create a public key for the user
resource "gitea_public_key" "test_public_key" {
  username  = "root"
  title     = "test-ssh-key"
  key       = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIAE1lceAANCx/Um5t/vJg3ydifXXEspgWhxUJ3Cl11/M SargentM@BLUE-00401"
  read_only = false
}

# # Create a GPG key for the user (note: requires a valid GPG key)
# resource "gitea_gpg_key" "test_gpg_key" {
#   armored_public_key = <<EOF
# -----BEGIN PGP PUBLIC KEY BLOCK-----

# mQINBGlf5ekBEAC7gilN/SJt7Sf8OZL9iv8rPm2Sk0mOZOVcBw5HCot2kd8DbreM
# 3fzCTEUOaQBEHz2ES2X15ZW33kHLVsN8RHlYrEdZOUQ6eYZI55jLDairQdQd9oMH
# HNKSC0lNbHra/oqvcaiJqVmrDsq5BCDG/ESw1wVozCGDawBVgsQUt2HegBnWik4I
# TxXSFhsQMkUu2ulyOeYYCt2Dk1Jr2661y0jWo6GTER0h4w5Z3IVHAeF2J7nTZ2v7
# n3f5w3eCaX56sHEYJnCtQuGg32f3kdewj4YtG1WRSE34DzEYXhjc5mT6VyXhSUho
# /IJ2DCZmizZ6kHY53ZNQmJCr3b66Sq0sCE5yvHKcFWjlO61/c2nbqqy4ETWtUf8x
# BTaTapNIqUWs5MfAOOfC/C4//1x061RmIvKOw3ic14/qtCzPE8Qe7P5iU/q6dtkI
# YnY1eKKXwqMp6374lXgs1KyB0AhpQgg0yZW3eqevuXTpMyIFZ9IQcEWOtB8w1vVC
# 4SafRBFNqxzE41Is8QX8B0tv4WNFj4ZD+b5+88y5Hz5f+LyRto1vb9xf7XjQz4W6
# M03GPHiz+cn9UsXyqSt5LomAddrkNVYZ1aI6x4jYSgz+FPf7ff7L04pyZy1eB+Zq
# tcjXVUuINMT6+lO/kIZA0t65dwN5KQsX90+j3uzeHvIkbI+8jE6MJM9NpQARAQAB
# tBlGb28gQmFyIDxmb28uYmFyQHpvdC5jb20+iQJRBBMBCAA7FiEEX7rwIHPmNJdL
# Twe5Vj0hhF6iowUFAmlf5ekCGwMFCwkIBwICIgIGFQoJCAsCBBYCAwECHgcCF4AA
# CgkQVj0hhF6iowVjuw//ZNn7pWV0jRAqWqkCAWTKVGJd44xr7LwHh6zICi+OtbtB
# K/LIrMcyjB1HWQVCLeQliyVHCYTgpwfymY5XFsFAtTAY2eojnFk2L9dzdyNyRFfL
# 8p5gna2Y588acvrYhMiorKuXnyk3sQ+6cwovWg8POCSeSO1chjtV6f3kR2BLyegj
# p9liSCmG/wJaR2UNibTFRzi3hPf5a//i3L+eZTjnIutdYSW8oWmzsDsgD44BfIzr
# 6NhmRm4egkHlC8UAmSd8A7WvaSojuE+jDjiCYroGaDTP/vC5EEIre9WwTS/JJpn0
# kX0iW1hqyhwMiwT68Xzme8ty8g+m1Pr8DdexJ++9Jkj2xTJUySe2MCOYKD1+9XhY
# FVrqI9LLcMA2m24e8VQDHULc1HszyjvlpAgc9s/owHzcUvJmz4FGwUNAp3RDeAG6
# WdM/GuxAJDbbBNz/zGYzWVJd0jyoScmjeXeeTpRKlw4jIvd4kxCIh6DlPTY+islP
# ikE76x1uTAiTwLgxzJPwobJSF9mMiEF8SXSp+CxLMPEIHmTs+8qpuRzPfvwT0KWk
# 28qYYIKjpEwGDWZ8Z79msbho6YmeS11zipS5ryVA73KU1d4KrVwT9ENwNgMHWLbM
# vfxVJ5YhUU52uh3+dnEQW3koiAPCOYVryKdw6VUKMXgHkg84P2Ft8odlraBkwRG5
# Ag0EaV/l6QEQAMPrsNoIe2ArCU+Y48mCu/tVYjhiBgqS2d7T0PMihUq643S1QeHC
# IDRRu65EseFzJ2/ZJ9fPXNe1P9xdSozp+5gVnGGQjtXZGJEnF1nPqJaaFnrzUlZI
# mq3iuvhaMrQHF4GQXkp7qEe7euVPHIbxuucLtHfISWPQ6SsWt2hxeuEUgHKjE36L
# gIPhnDXNWxuG7s0H/+gcX9w9DfPyTdEZ94d9U8PT/ZMuBo9y0HkhMd3jwoAhSf6W
# GBCxqPDpk9dspbPTCidW42bYSVoTDQIN8LkHt6JAgjGhSIibnqRz2ax03I2Oxfwn
# f1madzStQsWokqSOe4K6/hgrA8LlR4+8jFZzg4ID4Em+UTZ4KZI0Yh+6aM3mTNV/
# SnH9T2vs/viaOt3aC+w8AqNijl2XF87VIZj/cpbK2YOjU55rMBDouXwdEB8vMGKG
# f1FammJVSkMNr5+eOEIjp/sDXDVP8d0G9usGMS0kf5PTyysM5uSgLFM9Td8IUfN9
# A/jNRg5lccwOI/ZIinH9U5Tj6lo65REkzrdiZAMwiRnAo4CtsnHPDjkBW5qqGd9r
# 87ohevUxxTPXdcYrnD7QW5JsMYuIcPf/Kzw23diKjEpMQiREXSlm8MP1tO841Uz4
# onPrkuHOvrHo9U2O7+n+G7cMizp57sXIirL+qymUJDavic7cvtIDDqY/ABEBAAGJ
# AjYEGAEIACAWIQRfuvAgc+Y0l0tPB7lWPSGEXqKjBQUCaV/l6QIbDAAKCRBWPSGE
# XqKjBR3oEACGqKpO08yZfL+8c0Vmt1KApsFeAdiqMJRSt93WyY8tJkLoswyGCjKg
# X1g8dO5u9J8aCbmQBjosBEfff+/JN7uulXK91pzuZUJPpDBJwyZrsKXGSXwK8Er/
# 3ZXswblfxqxseRbfXE73GzNm76ts4V8293fjif2kWce21gDYQOynX8EW7JI/9c/7
# ZSbt+uB1CCBdcN75RJFApbLX8b9VYlSxEpdm7YZ9Oe3WBRb7QPeS56sVhKOJI4JN
# 6zJF4KVmkTwZHuCsULUxTHt4/zz6Im6kB9hDLnfSrr1WwY94Doo0yJdT7atAONaP
# EFQroMeg3XgB1/E4MvmXVDF5ldAdq8++nO7p4wqBgc0FZ2qCyUz9+XJ776Dcdy50
# Y03xCMDGPahYN+yq+rZ3cO2w2d4RmDI4cxjtdWHV8l7uNp1p5VlWOfoCz2pqU9Uc
# /A9ZedQC4r+kGVLuponfQgpldAuMMxyunAmXovwWHSrkxu2TQXW3SAurJr1lT3VW
# 0zhpQy4/4Y77rerqXek1VYZUeLygF01zR3Q9gMt603m0dcLuBFZYgyUv/DCeS2q/
# YcVNtdqfQ4xCIu6B+YeHap33DqKYwYnhTrSZzTcMp3x3+sHTdujbaBUANTXr1sAD
# otR8OPBPYtVcYd5xAkq03phktIXyZE7Z2Yg15yebHOap9p81cZ2yGw==
# =Jbwr
# -----END PGP PUBLIC KEY BLOCK-----
# EOF
# depends_on = [ gitea_user.foo_user ]
# }

# Create a repository key (deploy key)
resource "gitea_repository_key" "test_deploy_key" {
  owner      = "root"
  repository = gitea_repository.test_repo.name
  title      = "test-deploy-key"
  key        = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBAhragMvqcZvLxtOxVpqIuWmvPHW+VzIwKaLLmq1eyg SargentM@BLUE-00401"
  read_only  = true
}

# Create an OAuth2 application
resource "gitea_oauth2_app" "test_oauth2_app" {
  name                = "test-oauth2-app"
  redirect_uris       = ["https://example.com/callback"]
  confidential_client = true
}

# Create a repository webhook (commented out due to provider bug with timestamps)
# resource "gitea_repository_webhook" "test_webhook" {
#   owner      = "root"
#   repository = gitea_repository.test_repo.name
#   type       = "gitea"
#   active     = true
# 
#   config = {
#     url          = "https://example.com/webhook"
#     content_type = "json"
#     secret       = "webhook-secret"
#   }
# 
#   events = ["push", "pull_request"]
# }

# Create a repository actions secret (commented out due to provider bug with timestamps)
# resource "gitea_repository_actions_secret" "test_repo_secret" {
#   owner      = "root"
#   repository = gitea_repository.test_repo.name
#   name       = "TEST_REPO_SECRET"
#   data       = "my-repo-secret-value"
# }

# Create a repository actions variable (commented out due to potential provider bug)
# resource "gitea_repository_actions_variable" "test_repo_variable" {
#   owner      = "root"
#   repository = gitea_repository.test_repo.name
#   name       = "TEST_REPO_VARIABLE"
#   value      = "my-repo-variable-value"
# }

# Create a git hook for the repository (requires special Git hooks permissions)
# resource "gitea_git_hook" "test_git_hook" {
#   owner      = "root"
#   repository = gitea_repository.test_repo.name
#   name       = "pre-receive"
#   content    = <<EOF
# #!/bin/bash
# echo "Pre-receive hook executed"
# exit 0
# EOF
# }

# Data source: Read user information
data "gitea_user" "root_user" {
  username = "root"
}

# Data source: Read organization information
data "gitea_org" "test_org_data" {
  name = gitea_org.test_org.name
}

# Data source: Read repository information
data "gitea_repository" "test_repo_data" {
  owner = "root"
  name  = gitea_repository.test_repo.name
}

# Data source: Read branch protection information
data "gitea_branch_protection" "test_protection_data" {
  owner = "root"
  repo  = gitea_repository.test_repo.name
  name  = gitea_repository_branch_protection.test_protection.rule_name
}

# Data source: Read team information
data "gitea_team" "test_team_data" {
  org  = gitea_org.test_org.name
  name = gitea_team.test_team.name
}

# Data source: Read team membership information
data "gitea_team_membership" "test_membership_data" {
  org       = gitea_org.test_org.name
  team_name = gitea_team.test_team.name
  username  = gitea_user.test_user.username
}

# Data source: List all repositories for a user
data "gitea_repositories" "root_repos" {
  username = "root"

  depends_on = [
    gitea_repository.test_repo,
    gitea_repository.test_repo_for_org
  ]
}

# Data source: List all teams in an organization
data "gitea_teams" "org_teams" {
  org = gitea_org.test_org.name
}

# Output the token (sensitive)
output "token_value" {
  value     = gitea_token.test_token.token
  sensitive = true
}

# Output user data source information
output "user_data" {
  value = {
    username = data.gitea_user.root_user.username
    email    = data.gitea_user.root_user.email
  }
}

# Output organization data source information
output "org_data" {
  value = {
    name        = data.gitea_org.test_org_data.name
    description = data.gitea_org.test_org_data.description
  }
}

# Output repositories count
output "repositories_count" {
  value = length(data.gitea_repositories.root_repos.repositories)
}

# Output teams count
output "teams_count" {
  value = length(data.gitea_teams.org_teams.teams)
}

