resource "gitea_team_membership" "test_membership" {
  org       = "testorg"
  team_name = gitea_team.test_team.name
  username  = gitea_user.test_user.username
}

# Use the team_membership datasource to verify it exists
data "gitea_team_membership" "test_membership" {
  org       = gitea_org.test_org.name
  team_name = gitea_team.test_team.name
  username  = gitea_user.test_user.username
  
  depends_on = [gitea_team_membership.test_membership]
}

# Output to verify the datasource works
output "team_membership_verified" {
  value = {
    org       = data.gitea_team_membership.test_membership.org
    team_name = data.gitea_team_membership.test_membership.team_name
    username  = data.gitea_team_membership.test_membership.username
  }
}
