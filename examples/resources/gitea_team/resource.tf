resource "gitea_team" "example" {
  org                        = "my-organization"
  name                       = "developers"
  description                = "Development team"
  permission                 = "none"  # Using unit-based permissions
  can_create_org_repo        = true
  includes_all_repositories  = false
  
  # Fine-grained permissions using units_map
  units_map = {
    "repo.code"     = "write"
    "repo.issues"   = "write"
    "repo.pulls"    = "write"
    "repo.wiki"     = "read"
    "repo.releases" = "write"
  }
}
