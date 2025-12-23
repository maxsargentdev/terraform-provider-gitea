resource "gitea_team_membership" "test_membership" {
  team_id  = gitea_team.test_team.id
  username = gitea_user.test_user.username
}
