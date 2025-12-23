resource "gitea_token" "test_token" {
  username = gitea_user.test_user.username
  name     = "test-token"
  scopes   = ["read:user", "read:repository"]
}

output "token_id" {
  value = gitea_token.test_token.id
}

output "token_sha1" {
  value     = gitea_token.test_token.sha1
  sensitive = true
}
