resource "gitea_org" "test_org" {
  name        = "testorg"
  full_name   = "Test Organization"
  description = "A test organization"
  visibility  = "public"
}
