resource "gitea_team" "example" {
  org  = "testorg"
  name = "test-team"
}

resource "gitea_user" "example" {
  username = "testuser"
  email    = "test@example.com"
  password = "password123"
}

resource "gitea_team_membership" "example" {
  team_id  = gitea_team.example.id
  username = gitea_user.example.username
}
