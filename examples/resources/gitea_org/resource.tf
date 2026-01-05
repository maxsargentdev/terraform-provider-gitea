resource "gitea_org" "example" {
  username    = "my-organization"
  full_name   = "My Organization"
  description = "An example organization"
  visibility  = "public"
}
