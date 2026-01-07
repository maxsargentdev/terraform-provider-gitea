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
					resource.TestCheckResourceAttr("icegitea_team.test", "name", "testteam1"),
					resource.TestCheckResourceAttr("icegitea_team.test", "description", "Test Team"),
					resource.TestCheckResourceAttr("icegitea_team.test", "units_map.repo.code", "read"),
					resource.TestCheckResourceAttr("icegitea_team.test", "units_map.repo.issues", "read"),
					resource.TestCheckResourceAttr("icegitea_team.test", "units_map.repo.pulls", "read"),
					resource.TestCheckResourceAttrSet("icegitea_team.test", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "icegitea_team.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"org"},
			},
			// Update and Read testing
			{
				Config: testAccTeamResourceConfig("testteam1", "Updated Test Team", "write"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("icegitea_team.test", "description", "Updated Test Team"),
					resource.TestCheckResourceAttr("icegitea_team.test", "units_map.repo.code", "write"),
					resource.TestCheckResourceAttr("icegitea_team.test", "units_map.repo.issues", "write"),
					resource.TestCheckResourceAttr("icegitea_team.test", "units_map.repo.pulls", "write"),
				),
			},
		},
	})
}

func TestAccTeamResourcePermissionsFormat(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test with 'read' permission
			{
				Config: testAccTeamResourceConfig("testteam_perms", "Team with read permission", "read"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("icegitea_team.test", "name", "testteam_perms"),
					resource.TestCheckResourceAttr("icegitea_team.test", "description", "Team with read permission"),
					resource.TestCheckResourceAttr("icegitea_team.test", "units_map.repo.code", "read"),
					resource.TestCheckResourceAttr("icegitea_team.test", "units_map.repo.issues", "read"),
					resource.TestCheckResourceAttr("icegitea_team.test", "units_map.repo.pulls", "read"),
					resource.TestCheckResourceAttrSet("icegitea_team.test", "id"),
				),
			},
			// Update to 'write' permission
			{
				Config: testAccTeamResourceConfig("testteam_perms", "Team with write permission", "write"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("icegitea_team.test", "name", "testteam_perms"),
					resource.TestCheckResourceAttr("icegitea_team.test", "description", "Team with write permission"),
					resource.TestCheckResourceAttr("icegitea_team.test", "units_map.repo.code", "write"),
					resource.TestCheckResourceAttr("icegitea_team.test", "units_map.repo.issues", "write"),
					resource.TestCheckResourceAttr("icegitea_team.test", "units_map.repo.pulls", "write"),
				),
			},
			// Update to 'admin' permission
			{
				Config: testAccTeamResourceConfig("testteam_perms", "Team with admin permission", "admin"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("icegitea_team.test", "name", "testteam_perms"),
					resource.TestCheckResourceAttr("icegitea_team.test", "description", "Team with admin permission"),
					resource.TestCheckResourceAttr("icegitea_team.test", "units_map.repo.code", "admin"),
					resource.TestCheckResourceAttr("icegitea_team.test", "units_map.repo.issues", "admin"),
					resource.TestCheckResourceAttr("icegitea_team.test", "units_map.repo.pulls", "admin"),
				),
			},
			// Update to fine-grained permissions with units_map
			{
				Config: testAccTeamResourceConfigWithUnitsMap("testteam_perms", "Team with fine-grained permissions"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("icegitea_team.test", "name", "testteam_perms"),
					resource.TestCheckResourceAttr("icegitea_team.test", "description", "Team with fine-grained permissions"),
					resource.TestCheckResourceAttr("icegitea_team.test", "units_map.repo.code", "write"),
					resource.TestCheckResourceAttr("icegitea_team.test", "units_map.repo.issues", "read"),
					resource.TestCheckResourceAttr("icegitea_team.test", "units_map.repo.pulls", "write"),
					resource.TestCheckResourceAttr("icegitea_team.test", "units_map.repo.wiki", "none"),
				),
			},
		},
	})
}

func testAccTeamResourceConfig(name, description, permission string) string {
	return providerConfig() + fmt.Sprintf(`
resource "icegitea_org" "test_org" {
  username   = "testorg"
  full_name  = "Test Organization"
  visibility = "public"
}

resource "icegitea_team" "test" {
  org                       = icegitea_org.test_org.name
  name                      = %[1]q
  description               = %[2]q
  can_create_org_repo       = false
  includes_all_repositories = false
  
  units_map = {
    "repo.code"   = %[3]q
    "repo.issues" = %[3]q
    "repo.pulls"  = %[3]q
  }
}
`, name, description, permission)
}

func testAccTeamResourceConfigWithUnitsMap(name, description string) string {
	return providerConfig() + fmt.Sprintf(`
resource "icegitea_org" "test_org" {
  username   = "testorg"
  full_name  = "Test Organization"
  visibility = "public"
}

resource "icegitea_team" "test" {
  org                       = icegitea_org.test_org.name
  name                      = %[1]q
  description               = %[2]q
  can_create_org_repo       = false
  includes_all_repositories = false
  
  units_map = {
    "repo.code"     = "write"
    "repo.issues"   = "read"
    "repo.pulls"    = "write"
    "repo.wiki"     = "none"
  }
}
`, name, description)
}
