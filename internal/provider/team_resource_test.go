package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTeamResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccTeamResourceConfig("testteam1", "Test Team", "read"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gitea_team.test", "name", "testteam1"),
					resource.TestCheckResourceAttr("gitea_team.test", "description", "Test Team"),
					resource.TestCheckResourceAttr("gitea_team.test", "permission", "read"),
					resource.TestCheckResourceAttrSet("gitea_team.test", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "gitea_team.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"org"},
			},
			// Update and Read testing
			{
				Config: testAccTeamResourceConfig("testteam1", "Updated Test Team", "write"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gitea_team.test", "description", "Updated Test Team"),
					resource.TestCheckResourceAttr("gitea_team.test", "permission", "write"),
				),
			},
		},
	})
}

func testAccTeamResourceConfig(name, description, permission string) string {
	return providerConfig() + fmt.Sprintf(`
resource "gitea_org" "test_org" {
  username   = "testorg"
  full_name  = "Test Organization"
  visibility = "public"
}

resource "gitea_team" "test" {
  org                       = gitea_org.test_org.name
  name                      = %[1]q
  description               = %[2]q
  permission                = %[3]q
  can_create_org_repo       = false
  includes_all_repositories = false
}
`, name, description, permission)
}
