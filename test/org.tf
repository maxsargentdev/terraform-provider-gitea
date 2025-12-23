resource "gitea_org" "test_org" {
  username    = "testorg"
  full_name   = "Test Organization"
  description = "A test organization"
  visibility  = "public"
}

data "gitea_org" "test_org" {
  org = gitea_org.test_org.username
}

output "org" {
  value = data.gitea_org.test_org
}