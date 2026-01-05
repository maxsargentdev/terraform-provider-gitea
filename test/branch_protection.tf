resource "gitea_branch_protection" "main_protection" {
  owner       = "root"
  repo        = gitea_repository.test_repo.name
  branch_name = "main"
  rule_name   = "Protect main branch"

  enable_push              = true
  enable_push_whitelist    = true
  push_whitelist_usernames = ["root"]

  enable_merge_whitelist   = false
  enable_status_check      = false
  require_signed_commits   = true
}

output "branch_protection_rule_name" {
  value = gitea_branch_protection.main_protection.rule_name
}
