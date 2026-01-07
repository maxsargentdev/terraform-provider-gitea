resource "gitea_repository" "source" {
  username = "root"
  name     = "source-repo"
  private  = false
}

# Fork to the authenticated user's account
resource "gitea_fork" "user_fork" {
  owner = gitea_repository.source.username
  repo  = gitea_repository.source.name
}

# Fork to an organization
resource "gitea_fork" "org_fork" {
  owner        = gitea_repository.source.username
  repo         = gitea_repository.source.name
  organization = "my-org"
}
