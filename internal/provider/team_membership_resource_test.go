package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccTeamMembershipResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccTeamMembershipResourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("icegitea_team_membership.test", "team_id"),
					resource.TestCheckResourceAttr("icegitea_team_membership.test", "username", "testuser"),
				),
			},
			// ImportState testing
			{
				ResourceName:                         "icegitea_team_membership.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "team_id",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["icegitea_team_membership.test"]
					if !ok {
						return "", fmt.Errorf("Resource not found")
					}
					teamID := rs.Primary.Attributes["team_id"]
					username := rs.Primary.Attributes["username"]
					return fmt.Sprintf("%s:%s", teamID, username), nil
				},
			},
		},
	})
}

func testAccTeamMembershipResourceConfig() string {
	return providerConfig() + `
resource "icegitea_org" "test_org" {
  username   = "testorg"
  full_name  = "Test Organization"
  visibility = "public"
}

resource "icegitea_team" "test" {
  org                       = icegitea_org.test_org.name
  name                      = "testteam"
  description               = "Test Team"
  can_create_org_repo       = false
  includes_all_repositories = false
  units_map = {
    "repo.code"     = "write"  # Source code access (none, read, write, admin)
    "repo.issues"   = "write"  # Issue tracker access
    "repo.pulls"    = "write"  # Pull requests access
    "repo.releases" = "none"  # Releases access
    "repo.ext_wiki" = "none"   # External wiki access
    "repo.ext_issues" = "none" # External issue tracker access
  }
}

resource "icegitea_user" "test" {
  username  = "testuser"
  email     = "testuser@example.com"
  password  = "testpass123"
  full_name = "Test User"
}

resource "icegitea_team_membership" "test" {
  team_id  = icegitea_team.test.id
  username = icegitea_user.test.username
}
`
}
