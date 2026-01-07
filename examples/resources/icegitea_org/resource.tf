resource "icegitea_org" "test_org" {
  name         = "testorg"
  display_name = "Test Organization"
  description  = "A test organization"
  visibility   = "public"
}
