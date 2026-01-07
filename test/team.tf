output "team_id" {
  value = gitea_team.test_team.id
}

output "team_organization" {
  value = gitea_org.test_org.name
}

# Comprehensive test team - tests ALL possible attributes
resource "gitea_team" "test_team" {
  # Required attributes
  org  = gitea_org.test_org.name
  name = "test-team"

  # Optional string attributes
  description = "A  test team with all attributes configured"

  # Optional boolean attributes
  can_create_org_repo       = true
  includes_all_repositories = false

  # Optional units_map - fine-grained permissions for each repository unit
  units_map = {
    "repo.code"     = "write"  # Source code access (none, read, write, admin)
    "repo.issues"   = "write"  # Issue tracker access
    "repo.pulls"    = "write"  # Pull requests access
    "repo.releases" = "none"  # Releases access
    "repo.ext_wiki" = "none"   # External wiki access
    "repo.ext_issues" = "read" # External issue tracker access
  }

  # Note: 'organization' is a computed read-only attribute (returned by API)
  # Note: 'id' is a computed read-only attribute (assigned by API)
}

resource "gitea_team_repository" "test_team_repo_association" {
  org             = gitea_org.test_org.name
  team_name       = "test-team"
  repository_name = gitea_repository.test_repo_3.name
}
resource "gitea_team_repository" "test_team_repo_association_2" {
  org             = gitea_org.test_org.name
  team_name       = "test-team"
  repository_name = gitea_repository.test_repo_2.name
}


data "gitea_team" "test_team" {
  org  = gitea_org.test_org.name
  name = gitea_team.test_team.name

  depends_on = [gitea_team.test_team]
}

resource "gitea_team" "foo" {
  org             = gitea_org.test_org.name
  name                     = "foo"
  description              = "foo"
  units_map = {
    "repo.code"     = "write"
    "repo.issues"   = "write"
    "repo.pulls"    = "read"
    "repo.releases" = "read"
    "repo.packages" = "read"
    "repo.actions" = "write"
  }

  can_create_org_repo       = false
  includes_all_repositories = false
}

resource "gitea_team_repository" "test_team_repo_association_3" {
  org             = gitea_org.test_org.name
  team_name       = gitea_team.foo.name
  repository_name = gitea_repository.test_repo_2.name
}