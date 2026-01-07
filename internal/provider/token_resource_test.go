package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTokenResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccTokenResourceConfig("test-token", "testuser"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("icegitea_token.test", "name", "test-token"),
					resource.TestCheckResourceAttr("icegitea_token.test", "username", "testuser"),
					resource.TestCheckResourceAttrSet("icegitea_token.test", "id"),
					resource.TestCheckResourceAttrSet("icegitea_token.test", "sha1"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "icegitea_token.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"sha1", "username", "scopes", "created_at", "last_used_at"},
			},
			// Tokens cannot be updated - any change requires replacement
			// So we don't include an update test
		},
	})
}

func testAccTokenResourceConfig(name, username string) string {
	return providerConfig() + fmt.Sprintf(`
resource "icegitea_user" "test" {
  username  = %[2]q
  email     = "%[2]s@example.com"
  password  = "testpass123"
  full_name = "Test User"
}

resource "icegitea_token" "test" {
  name     = %[1]q
  username = icegitea_user.test.username
  scopes   = ["read:user"]
}
`, name, username)
}
