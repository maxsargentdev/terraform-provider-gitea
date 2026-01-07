# terraform-provider-icegitea

A small terraform provider for the [Gitea](https://about.gitea.com/) project.

The [existing terraform provider](https://registry.terraform.io/providers/go-gitea/gitea/latest) doesnt receive frequent updates and has some issues with missing features etc.

This project also moves to the new TF plugin framework instead of the old sdk.

I plan on extending this in my spare time to support Giteas configuration via IaC in a more complete fashion than what is offered currently.

## Original Usecase

The original usecase for this project is to fix issues seen with the current Gitea provider around granular unit level permissions where there would be persistent drift.

Thus the `icegitea_team` resource provided by this provider has a different API to the existing one, exposing the `units_map` part of the Gitea API:

~~~hcl
resource "icegitea_team" "test_team" {
  org  = "testorg"
  name = "test-team"

  description = "A test team with all attributes configured"

  can_create_org_repo       = true
  includes_all_repositories = false

  units_map = {
    "repo.code"       = "write"  # Source code access (none, read, write, admin)
    "repo.issues"     = "write"  # Issue tracker access
    "repo.pulls"      = "write"  # Pull requests access
    "repo.releases"   = "none"   # Releases access
    "repo.ext_wiki"   = "none"   # External wiki access
    "repo.ext_issues" = "read"   # External issue tracker access
    "repo.actions"    = "write"  # Actions access
  }
}
~~~

Additionally the assignment of repos to teams is done using the `icegitea_team_repository` resource rather than a list of repositories:

~~~hcl
resource "icegitea_team_repository" "test_team_repo_association" {
  org             = "testorg"
  team_name       = "test-team"
  repository_name = "test-repo-for-org"
}
~~~

## Name

The provider is named this way so that it can be used alongside the existing provider within the same module to make my usage easier, its not a good name.

## Contributing

All PRs will be welcome here :)