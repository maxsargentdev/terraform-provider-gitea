package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccRepositoryResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccRepositoryResourceConfig("test-repo-1", "Test repository", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("icegitea_repository.test", "name", "test-repo-1"),
					resource.TestCheckResourceAttr("icegitea_repository.test", "description", "Test repository"),
					resource.TestCheckResourceAttr("icegitea_repository.test", "private", "true"),
					resource.TestCheckResourceAttrSet("icegitea_repository.test", "id"),
					resource.TestCheckResourceAttrSet("icegitea_repository.test", "full_name"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "icegitea_repository.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return "root/test-repo-1", nil
				},
			},
			// Update and Read testing
			{
				Config: testAccRepositoryResourceConfig("test-repo-1", "Updated repository", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("icegitea_repository.test", "description", "Updated repository"),
					resource.TestCheckResourceAttr("icegitea_repository.test", "private", "false"),
				),
			},
		},
	})
}

func testAccRepositoryResourceConfig(name, description string, private bool) string {
	return providerConfig() + fmt.Sprintf(`
resource "icegitea_repository" "test" {
  name        = %[1]q
  description = %[2]q
  private     = %[3]t
}
`, name, description, private)
}
