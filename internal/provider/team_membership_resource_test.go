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
					resource.TestCheckResourceAttrSet("gitea_team_membership.test", "team_id"),
					resource.TestCheckResourceAttr("gitea_team_membership.test", "username", "testuser"),
				),
			},
			// ImportState testing
			{
				ResourceName:                         "gitea_team_membership.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "team_id",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["gitea_team_membership.test"]
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
resource "gitea_org" "test_org" {
  username   = "testorg"
  full_name  = "Test Organization"
  visibility = "public"
}

resource "gitea_team" "test" {
  org                       = gitea_org.test_org.name
  name                      = "testteam"
  description               = "Test Team"
  permission                = "read"
  can_create_org_repo       = false
  includes_all_repositories = false
}

resource "gitea_user" "test" {
  username  = "testuser"
  email     = "testuser@example.com"
  password  = "testpass123"
  full_name = "Test User"
}

resource "gitea_team_membership" "test" {
  team_id  = gitea_team.test.id
  username = gitea_user.test.username
}
`
}
