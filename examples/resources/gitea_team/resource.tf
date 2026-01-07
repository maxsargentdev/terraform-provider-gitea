resource "icegitea_team" "test_team" {
  org  = "testorg"
  name = "test-team"

  description = "A test team with all attributes configured"

  can_create_org_repo       = true
  includes_all_repositories = false

  units_map = {
    "repo.code"       = "write"  # Source code access (none, read, write, admin)
    "repo.issues"     = "write"  # Issue tracker access
    "repo.pulls"      = "write"  # Pull requests access
    "repo.releases"   = "none"   # Releases access
    "repo.ext_wiki"   = "none"   # External wiki access
    "repo.ext_issues" = "read"   # External issue tracker access
  }
}
