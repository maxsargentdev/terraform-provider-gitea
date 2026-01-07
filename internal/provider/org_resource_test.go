package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccOrgResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccOrgResourceConfig("testorg", "Test Org", "public"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gitea_org.test", "username", "testorg"),
					resource.TestCheckResourceAttr("gitea_org.test", "full_name", "Test Org"),
					resource.TestCheckResourceAttr("gitea_org.test", "visibility", "public"),
					resource.TestCheckResourceAttrSet("gitea_org.test", "id"),
					resource.TestCheckResourceAttrSet("gitea_org.test", "avatar_url"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "gitea_org.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return "testorg", nil
				},
			},
			// Update and Read testing
			{
				Config: testAccOrgResourceConfig("testorg", "Updated Test Org", "public"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gitea_org.test", "username", "testorg"),
					resource.TestCheckResourceAttr("gitea_org.test", "full_name", "Updated Test Org"),
				),
			},
		},
	})
}

func TestAccOrgResource_WithDescription(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOrgResourceConfigWithDescription("testorg2", "Test Org 2", "A test organization", "private"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gitea_org.test", "username", "testorg2"),
					resource.TestCheckResourceAttr("gitea_org.test", "description", "A test organization"),
					resource.TestCheckResourceAttr("gitea_org.test", "visibility", "private"),
				),
			},
		},
	})
}

func testAccOrgResourceConfig(username, fullName, visibility string) string {
	return providerConfig() + fmt.Sprintf(`
resource "gitea_org" "test" {
  username   = %[1]q
  full_name  = %[2]q
  visibility = %[3]q
}
`, username, fullName, visibility)
}

func testAccOrgResourceConfigWithDescription(username, fullName, description, visibility string) string {
	return providerConfig() + fmt.Sprintf(`
resource "gitea_org" "test" {
  username    = %[1]q
  full_name   = %[2]q
  description = %[3]q
  visibility  = %[4]q
}
`, username, fullName, description, visibility)
}
