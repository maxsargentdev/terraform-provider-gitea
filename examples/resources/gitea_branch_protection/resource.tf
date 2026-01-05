resource "gitea_branch_protection" "example" {
  owner       = "my-username"
  repo        = "my-repo"
  branch_name = "main"
  rule_name   = "Protect main branch"

  enable_push              = true
  enable_push_whitelist    = true
  push_whitelist_usernames = ["maintainer1", "maintainer2"]

  enable_merge_whitelist   = true
  merge_whitelist_usernames = ["maintainer1"]
  
  enable_status_check      = true
  status_check_contexts    = ["ci/build", "ci/test"]
  
  require_signed_commits   = true
}
