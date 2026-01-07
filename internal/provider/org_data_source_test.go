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
					resource.TestCheckResourceAttr("data.icegitea_org.test", "org", "testorg"),
					resource.TestCheckResourceAttr("data.icegitea_org.test", "full_name", "Test Org Data"),
					resource.TestCheckResourceAttrSet("data.icegitea_org.test", "id"),
					resource.TestCheckResourceAttrSet("data.icegitea_org.test", "avatar_url"),
					resource.TestCheckResourceAttr("data.icegitea_org.test", "visibility", "public"),
				),
			},
		},
	})
}

func testAccOrgDataSourceConfig(username, fullName string) string {
	return providerConfig() + fmt.Sprintf(`
resource "icegitea_org" "test" {
  username   = %[1]q
  full_name  = %[2]q
  visibility = "public"
}

data "icegitea_org" "test" {
  org = icegitea_org.test.username
}
`, username, fullName)
}
