package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccBranchProtectionResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccBranchProtectionResourceConfig("main", "Protect main"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gitea_branch_protection.test", "branch_name", "main"),
					resource.TestCheckResourceAttr("gitea_branch_protection.test", "rule_name", "Protect main"),
					resource.TestCheckResourceAttr("gitea_branch_protection.test", "enable_push", "true"),
					resource.TestCheckResourceAttr("gitea_branch_protection.test", "require_signed_commits", "true"),
				),
			},
			// ImportState testing
			{
				ResourceName:                         "gitea_branch_protection.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "rule_name",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return "root/test-repo/Protect main", nil
				},
				ImportStateVerifyIgnore: []string{"owner", "repo", "branch_name"},
			},
			// Update and Read testing
			{
				Config: testAccBranchProtectionResourceConfig("main", "Updated Protection"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gitea_branch_protection.test", "rule_name", "Updated Protection"),
				),
			},
		},
	})
}

func testAccBranchProtectionResourceConfig(branchName, ruleName string) string {
	return providerConfig() + fmt.Sprintf(`
resource "gitea_repository" "test" {
  name        = "test-repo"
  description = "Test repository"
  private     = true
}

resource "gitea_branch_protection" "test" {
  owner       = "root"
  repo        = gitea_repository.test.name
  branch_name = %[1]q
  rule_name   = %[2]q

  enable_push              = true
  enable_push_whitelist    = true
  push_whitelist_usernames = ["root"]

  enable_merge_whitelist   = false
  enable_status_check      = false
  require_signed_commits   = true
}
`, branchName, ruleName)
}
