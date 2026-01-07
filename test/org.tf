resource "gitea_org" "test_org" {
  name         = "testorg"
  display_name = "Test Organization"
  description  = "A test organization"
  visibility   = "public"
}

data "gitea_org" "test_org" {
  org = gitea_org.test_org.name
}

output "org" {
  value = data.gitea_org.test_org
}
