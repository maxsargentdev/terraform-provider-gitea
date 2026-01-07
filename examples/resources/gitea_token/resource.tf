# Create a token for the authenticated user
resource "gitea_token" "example" {
  name = "terraform-token"
}

# The token value is available as an output
output "token_value" {
  value     = gitea_token.example.token
  sensitive = true
}
