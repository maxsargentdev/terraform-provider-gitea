resource "gitea_org_actions_secret" "example" {
  org         = "myorg"
  name        = "MY_SECRET"
  data        = "secret-value"
  description = "Example organization secret for GitHub Actions"
}
