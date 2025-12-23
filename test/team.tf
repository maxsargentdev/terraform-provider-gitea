resource "gitea_team" "test_team" {
  org                        = gitea_org.test_org.name
  name                       = "test-team"
  description                = "A test team"
  permission                 = "none"  # Unit-based permissions
  can_create_org_repo        = false
  includes_all_repositories  = false
  
  units_map = {
    "repo.code"     = "write"
    "repo.issues"   = "read"
    "repo.pulls"    = "write"
    "repo.wiki"     = "none"
    "repo.releases" = "read"
  }
}

output "team_id" {
  value = gitea_team.test_team.id
}
