resource "gitea_repository" "example" {
  username = "root"
  name     = "protected-repo"
}

resource "gitea_repository_branch_protection" "main" {
  username  = gitea_repository.example.username
  name      = gitea_repository.example.name
  rule_name = "main"

  enable_push             = true
  push_whitelist_users    = ["root", "admin"]
  push_whitelist_teams    = ["maintainers"]

  merge_whitelist_users   = ["root"]
  merge_whitelist_teams   = ["maintainers"]

  required_approvals      = 2
  approval_whitelist_users = ["root", "reviewer1", "reviewer2"]

  status_check_patterns = ["ci/test", "ci/lint"]

  block_merge_on_rejected_reviews         = true
  block_merge_on_official_review_requests = false
  block_merge_on_outdated_branch          = true
  dismiss_stale_approvals                 = true

  require_signed_commits    = false
  protected_file_patterns   = "*.lock;go.sum"
  unprotected_file_patterns = "docs/*;*.md"
}

# Protect release branches with a pattern
resource "gitea_repository_branch_protection" "releases" {
  username  = gitea_repository.example.username
  name      = gitea_repository.example.name
  rule_name = "release/*"

  enable_push          = true
  push_whitelist_users = ["root"]

  required_approvals      = 1
  require_signed_commits  = true
}
