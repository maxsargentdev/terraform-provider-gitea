resource "icegitea_repository" "test_repo" {
  owner       = "root"
  name        = "test-repo"
  description = "A test repository created with Terraform, using a user as the owner"
  private     = true
}

resource "icegitea_repository" "test_repo_for_org" {
  owner       = "testorg"
  name        = "test-repo-for-org"
  description = "A test repository created with Terraform"
  private     = true
}
