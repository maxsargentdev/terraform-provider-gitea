resource "gitea_team_repository" "test_team_repo_association" {
  org             = "testorg"
  team_name       = "test-team"
  repository_name = "test-repo-for-org"
}
