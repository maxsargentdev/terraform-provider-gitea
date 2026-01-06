resource "gitea_repository" "test_repo" {
  owner       = "root"
  name        = "test-repo"
  description = "A test repository created with Terraform"
  private     = true
}

output "repository_id" {
  value = gitea_repository.test_repo.id
}

output "repository_full_name" {
  value = gitea_repository.test_repo.full_name
}

resource "gitea_repository" "test_repo_2" {
  owner       = "testorg"
  name        = "test-repo-2"
  description = "A test repository created with Terraform"
  private     = true
}