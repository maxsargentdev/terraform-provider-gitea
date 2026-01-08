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
  username  = "testuser"
  login_name  = "login_name"
  email     = "testuser@gitea.local"
  password  = "testpassword123"
  full_name = "Test User"
  active   = true
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

  enable_push             = true
  push_whitelist_users    = ["root"]
  required_approvals      = 1
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

# Output the token (sensitive)
output "token_value" {
  value     = gitea_token.test_token.token
  sensitive = true
}

