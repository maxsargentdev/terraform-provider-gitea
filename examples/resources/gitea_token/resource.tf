resource "gitea_token" "example" {
  username = "johndoe"
  name     = "automation-token"
  scopes   = ["read:user", "write:repository", "read:organization"]
}
