package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccRepositoryBranchProtectionResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccRepositoryBranchProtectionResourceConfig("Protect main"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gitea_repository_branch_protection.test", "rule_name", "Protect main"),
					resource.TestCheckResourceAttr("gitea_repository_branch_protection.test", "enable_push", "true"),
					resource.TestCheckResourceAttr("gitea_repository_branch_protection.test", "require_signed_commits", "true"),
				),
			},
			// ImportState testing
			{
				ResourceName:                         "gitea_repository_branch_protection.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "rule_name",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return "root/test-repo/Protect main", nil
				},
				ImportStateVerifyIgnore: []string{"username", "name"},
			},
			// Update and Read testing
			{
				Config: testAccRepositoryBranchProtectionResourceConfig("Protect main"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gitea_repository_branch_protection.test", "require_signed_commits", "true"),
				),
			},
		},
	})
}

func testAccRepositoryBranchProtectionResourceConfig(ruleName string) string {
	return providerConfig() + fmt.Sprintf(`
resource "gitea_repository" "test" {
  username    = "root"
  name        = "test-repo"
  description = "Test repository"
  private     = true
}

resource "gitea_repository_branch_protection" "test" {
  username  = "root"
  name      = gitea_repository.test.name
  rule_name = %[1]q

  enable_push             = true
  push_whitelist_users    = ["root"]
  require_signed_commits  = true
}
`, ruleName)
}
