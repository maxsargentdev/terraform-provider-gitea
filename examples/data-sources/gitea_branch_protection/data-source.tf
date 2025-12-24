data "gitea_branch_protection" "example" {
  owner       = "my-username"
  repo        = "my-repo"
  branch_name = "main"
}
