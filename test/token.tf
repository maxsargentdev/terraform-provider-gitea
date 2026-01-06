resource "gitea_token" "test_token" {
  name     = "test-token"
  scopes   = ["read:user", "read:repository"]
}
resource "gitea_token" "test_token_2" {
  name     = "test-token-2"
  scopes   = ["write:package"]
}

output "token_id" {
  value = gitea_token.test_token.id
}

output "token_sha1" {
  value     = gitea_token.test_token.sha1
  sensitive = true
}
