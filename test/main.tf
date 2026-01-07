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

# Create an organization
resource "gitea_org" "test_org" {
  name       = "testorg"
  full_name  = "Test Organization"
  description = "A test organization"
  visibility  = "public"
}

# Create a team in the organization
resource "gitea_team" "test_team" {
  org  = gitea_org.test_org.name
  name = "test-team"

  description = "A test team with all attributes configured"

  can_create_org_repo       = true
  includes_all_repositories = false

  units_map = {
    "repo.code"       = "write" # Source code access (none, read, write, admin)
    "repo.issues"     = "write" # Issue tracker access
    "repo.pulls"      = "write" # Pull requests access
    "repo.releases"   = "none"  # Releases access
    "repo.ext_wiki"   = "none"  # External wiki access
    "repo.ext_issues" = "read"  # External issue tracker access
  }
}

# Create a user
resource "gitea_user" "test_user" {
  username  = "testuser"
  email     = "testuser@gitea.local"
  password  = "testpassword123"
  full_name = "Test User"
}

# Create a repository owned by root user
resource "gitea_repository" "test_repo" {
  username    = "root"
  name        = "test-repo"
  description = "A test repository created with Terraform, using a user as the owner"
  private     = true
}

# Create a repository owned by the organization
resource "gitea_repository" "test_repo_for_org" {
  username    = gitea_org.test_org.name
  name        = "test-repo-for-org"
  description = "A test repository owned by an organization"
  private     = true
}

# Add the organization's repository to the team
resource "gitea_team_repository" "test_team_repo_association" {
  org             = gitea_org.test_org.name
  team_name       = gitea_team.test_team.name
  repository_name = gitea_repository.test_repo_for_org.name
}

# Add the user to the team
resource "gitea_team_membership" "test_membership" {
  team_id  = gitea_team.test_team.id
  username = gitea_user.test_user.username
}

# Create a fork of the test repository
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

  enable_push             = true
  push_whitelist_users    = ["root"]
  require_signed_commits  = false
  required_approvals      = 1
  block_merge_on_outdated_branch = true
}

# Create a token for the authenticated user
resource "gitea_token" "test_token" {
  name = "test-token"
}

# Create a public key for a user
resource "gitea_public_key" "test_key" {
  username = gitea_user.test_user.username
  title    = "Test SSH Key"
  key      = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC... test@example.com"
  read_only = false
}
