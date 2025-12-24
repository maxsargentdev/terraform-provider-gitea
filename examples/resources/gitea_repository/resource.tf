resource "gitea_repository" "example" {
  name        = "my-repo"
  description = "An example repository"
  private     = false
}
