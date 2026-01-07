resource "icegitea_team_membership" "test_membership" {
  org       = "testorg"
  team_name = "test-team"
  username  = "test"
}
