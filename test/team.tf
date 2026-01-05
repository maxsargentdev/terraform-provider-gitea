output "team_id" {
  value = gitea_team.test_team.id
}

output "team_organization" {
  value = gitea_team.test_team.organization
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
  includes_all_repositories = true

  # Optional units_map - fine-grained permissions for each repository unit
  units_map = {
    "repo.code"     = "write"  # Source code access (none, read, write, admin)
    "repo.issues"   = "write"  # Issue tracker access
    "repo.pulls"    = "write"  # Pull requests access
    "repo.wiki"     = "read"   # Wiki access
    "repo.releases" = "write"  # Releases access
    "repo.ext_wiki" = "none"   # External wiki access
    "repo.ext_issues" = "none" # External issue tracker access
  }

  # Note: 'organization' is a computed read-only attribute (returned by API)
  # Note: 'id' is a computed read-only attribute (assigned by API)
}
