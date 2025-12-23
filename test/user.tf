resource "gitea_user" "test_user" {
  username = "test"
  email    = "test@gitea.local"
  password = "testpassword123"
}

data "gitea_user" "test_user" {
  id = gitea_user.test_user.username
}

output "user" {
  value = data.gitea_user.test_user
}