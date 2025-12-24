resource "gitea_team_membership" "example" {
  team_id  = gitea_team.example.id
  username = "johndoe"
}
