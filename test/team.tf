output "team_id" {
  value = gitea_team.test_team_none.id
}

resource "gitea_team" "test_team_none" {
  org                        = gitea_org.test_org.username
  name                       = "test-team"
  description                = "A test team"
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

resource "gitea_team" "test_team_write" {
  org                        = gitea_org.test_org.username
  name                       = "test-team-write"
  description                = "A test team"
  can_create_org_repo        = false
  includes_all_repositories  = false
  
  units_map = {
    "repo.code"     = "write"
    "repo.issues"   = "write"
    "repo.pulls"    = "write"
  }
}
resource "gitea_team" "test_team_read" {
  org                        = gitea_org.test_org.username
  name                       = "test-team-read"
  description                = "A test team"
  can_create_org_repo        = false
  includes_all_repositories  = false
  
  units_map = {
    "repo.code"     = "read"
    "repo.issues"   = "read"
    "repo.pulls"    = "read"
  }
}
