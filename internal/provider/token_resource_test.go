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
				Config: testAccTokenResourceConfig("test-token"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gitea_token.test", "name", "test-token"),
					resource.TestCheckResourceAttrSet("gitea_token.test", "id"),
					resource.TestCheckResourceAttrSet("gitea_token.test", "token"),
					resource.TestCheckResourceAttrSet("gitea_token.test", "last_eight"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "gitea_token.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token", "scopes"},
			},
			// Tokens cannot be updated - any change requires replacement
			// So we don't include an update test
		},
	})
}

func testAccTokenResourceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "gitea_token" "test" {
  name   = %[1]q
  scopes = ["read:user"]
}
`, name)
}
