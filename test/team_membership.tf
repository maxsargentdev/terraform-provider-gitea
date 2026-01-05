resource "gitea_team_membership" "test_membership" {
  team_id  = gitea_team.test_team_none.id
  username = gitea_user.test_user.username
}

# Use the team_membership datasource to verify it exists
data "gitea_team_membership" "test_membership" {
  team_id  = gitea_team.test_team_none.id
  username = gitea_user.test_user.username
  
  depends_on = [gitea_team_membership.test_membership]
}

# Output to verify the datasource works
output "team_membership_verified" {
  value = {
    team_id  = data.gitea_team_membership.test_membership.team_id
    username = data.gitea_team_membership.test_membership.username
  }
}
