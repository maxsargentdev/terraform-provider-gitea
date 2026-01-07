package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccUserResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccUserResourceConfig("testuser1", "test1@example.com", "testpass123", "Test User"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("icegitea_user.test", "username", "testuser1"),
					resource.TestCheckResourceAttr("icegitea_user.test", "email", "test1@example.com"),
					resource.TestCheckResourceAttr("icegitea_user.test", "full_name", "Test User"),
					resource.TestCheckResourceAttrSet("icegitea_user.test", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "icegitea_user.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password", "must_change_password"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return "testuser1", nil
				},
			},
			// Update and Read testing
			{
				Config: testAccUserResourceConfig("testuser1", "test1@example.com", "testpass123", "Updated Test User"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("icegitea_user.test", "full_name", "Updated Test User"),
				),
			},
		},
	})
}

func testAccUserResourceConfig(username, email, password, fullName string) string {
	return providerConfig() + fmt.Sprintf(`
resource "icegitea_user" "test" {
  username  = %[1]q
  email     = %[2]q
  password  = %[3]q
  full_name = %[4]q
}
`, username, email, password, fullName)
}
