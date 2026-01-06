output "team_id" {
  value = gitea_team.test_team.id
}

output "team_organization" {
  value = gitea_org.test_org.username
}

# Comprehensive test team - tests ALL possible attributes
resource "gitea_team" "test_team" {
  # Required attributes
  org  = gitea_org.test_org.username
  name = "test-team"

  # Optional string attributes
  description = "A  test team with all attributes configured"

  # Optional boolean attributes
  can_create_org_repo       = false
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
  org             = gitea_org.test_org.username
  team_name       = "test-team"
  repository_name = gitea_repository.test_repo_2.name
}
