resource "gitea_user" "test_user" {
  username = "test"
  email    = "test@gitea.local"
  password = "testpassword123"
}
