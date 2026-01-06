resource "gitea_team_repository" "example" {
  org                        = "my-organization"
  team_name                  = "developers"
  repository_name            = "my-repository"
}
