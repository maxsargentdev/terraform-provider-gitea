resource "gitea_user" "example" {
  username = "testuser"
  email    = "test@example.com"
  password = "password123"
}

# Add a public SSH key for a user (requires admin privileges)
resource "gitea_public_key" "example" {
  username  = gitea_user.example.username
  title     = "My SSH Key"
  key       = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC... user@host"
  read_only = false
}

# Add a read-only deploy key
resource "gitea_public_key" "deploy" {
  username  = gitea_user.example.username
  title     = "Deploy Key"
  key       = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD... deploy@server"
  read_only = true
}
