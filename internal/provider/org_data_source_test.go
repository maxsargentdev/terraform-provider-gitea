package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOrgDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccOrgDataSourceConfig("testorg", "Test Org Data"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gitea_org.test", "org", "testorg"),
					resource.TestCheckResourceAttr("data.gitea_org.test", "full_name", "Test Org Data"),
					resource.TestCheckResourceAttrSet("data.gitea_org.test", "id"),
					resource.TestCheckResourceAttrSet("data.gitea_org.test", "avatar_url"),
					resource.TestCheckResourceAttr("data.gitea_org.test", "visibility", "public"),
				),
			},
		},
	})
}

func testAccOrgDataSourceConfig(username, fullName string) string {
	return providerConfig() + fmt.Sprintf(`
resource "gitea_org" "test" {
  username   = %[1]q
  full_name  = %[2]q
  visibility = "public"
}

data "gitea_org" "test" {
  org = gitea_org.test.username
}
`, username, fullName)
}
