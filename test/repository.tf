resource "gitea_repository" "test_repo" {
  owner       = "root"
  name        = "test-repo"
  description = "A test repository created with Terraform"
  private     = true
}

output "repository_id" {
  value = gitea_repository.test_repo.id
}


resource "gitea_repository" "test_repo_2" {
  owner       = "testorg"
  name        = "test-repo-2"
  description = "A test repository created with Terraform"
  private     = true
}

resource "gitea_repository" "test_repo_3" {
  owner       = "testorg"
  name        = "test-repo-3"
  description = "A test repository created with Terraform"
  private     = true
}

# resource "gitea_repository" "foo" {
#   owner             = "test-org"
#   name              = "foo"
#   auto_init         = true
#   private           = true
#   issue_labels      = "Default"
#   has_issues        = true
#   has_projects      = false
#   has_pull_requests = true
#   has_wiki          = false

#   allow_manual_merge    = false
#   allow_merge_commits   = false
#   allow_rebase          = false
#   allow_rebase_explicit = false
#   allow_squash_merge    = true

#   lifecycle {
#     prevent_destroy = true
#   }

#   description = "foo"
# }
