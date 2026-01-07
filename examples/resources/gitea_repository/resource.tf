resource "gitea_repository" "test_repo" {
  username    = "root"
  name        = "test-repo"
  description = "A test repository created with Terraform, using a user as the owner"
  private     = true
}

resource "gitea_repository" "test_repo_for_org" {
  username    = "testorg"
  name        = "test-repo-for-org"
  description = "A test repository created with Terraform"
  private     = true
}
