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
					resource.TestCheckResourceAttr("gitea_token.test", "name", "test-token"),
					resource.TestCheckResourceAttr("gitea_token.test", "username", "testuser"),
					resource.TestCheckResourceAttrSet("gitea_token.test", "id"),
					resource.TestCheckResourceAttrSet("gitea_token.test", "sha1"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "gitea_token.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"sha1", "username"},
			},
			// Tokens cannot be updated - any change requires replacement
			// So we don't include an update test
		},
	})
}

func testAccTokenResourceConfig(name, username string) string {
	return providerConfig() + fmt.Sprintf(`
resource "gitea_user" "test" {
  username  = %[2]q
  email     = "%[2]s@example.com"
  password  = "testpass123"
  full_name = "Test User"
}

resource "gitea_token" "test" {
  name     = %[1]q
  username = gitea_user.test.username
  scopes   = ["read:user", "read:repository"]
}
`, name, username)
}
