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
  name         = "testorg"
  display_name = "Test Organization"
  description  = "A test organization"
  visibility   = "public"
}

resource "gitea_team" "test_team" {
  org  = gitea_org.test_org.name
  name = "test-team"

  description = "A  test team with all attributes configured"

  can_create_org_repo       = true
  includes_all_repositories = false

  units_map = {
    "repo.code"     = "write"  # Source code access (none, read, write, admin)
    "repo.issues"   = "write"  # Issue tracker access
    "repo.pulls"    = "write"  # Pull requests access
    "repo.releases" = "none"  # Releases access
    "repo.ext_wiki" = "none"   # External wiki access
    "repo.ext_issues" = "read" # External issue tracker access
  }

}

resource "gitea_repository" "test_repo" {
  owner       = "root"
  name        = "test-repo"
  description = "A test repository created with Terraform"
  private     = true
}

resource "gitea_repository" "test_repo_for_org" {
  owner       = "testorg"
  name        = "test-repo-for-org"
  description = "A test repository created with Terraform"
  private     = true
}

resource "gitea_team_repository" "test_team_repo_association" {
  org             = gitea_org.test_org.name
  team_name       = gitea_team.test_team.name
  repository_name = gitea_repository.test_repo_for_org.name
}

resource "gitea_team_membership" "test_membership" {
  org       = "testorg"
  team_name = gitea_team.test_team.name
  username  = gitea_user.test_user.username
}

resource "gitea_user" "test_user" {
  username = "test"
  email    = "test@gitea.local"
  password = "testpassword123"
}