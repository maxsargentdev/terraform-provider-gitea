resource "gitea_repository" "test_repo" {
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
