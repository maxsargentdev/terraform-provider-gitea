package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// Regression test for https://github.com/maxsargentdev/terraform-provider-gitea/issues/16
// Gitea returns a non-empty mirror_interval for non-mirror repos; the Optional+Computed
// attribute is therefore never null after Read, and we must NOT forward it on EditRepo
// or the server rejects with "repo is not a mirror, can not change mirror interval".
func TestBuildEditRepoOption_SkipsMirrorIntervalForNonMirror(t *testing.T) {
	plan := &repositoryResourceModel{
		Name:                    types.StringValue("example"),
		Mirror:                  types.BoolValue(false),
		MigrationMirrorInterval: types.StringValue("8h0m0s"),
	}
	opts := buildEditRepoOption(context.Background(), plan)
	if opts.MirrorInterval != nil {
		t.Fatalf("expected MirrorInterval to be nil for non-mirror repo, got %q", *opts.MirrorInterval)
	}
}

func TestBuildEditRepoOption_SetsMirrorIntervalForMirror(t *testing.T) {
	plan := &repositoryResourceModel{
		Name:                    types.StringValue("example"),
		Mirror:                  types.BoolValue(true),
		MigrationMirrorInterval: types.StringValue("1h0m0s"),
	}
	opts := buildEditRepoOption(context.Background(), plan)
	if opts.MirrorInterval == nil || *opts.MirrorInterval != "1h0m0s" {
		got := "<nil>"
		if opts.MirrorInterval != nil {
			got = *opts.MirrorInterval
		}
		t.Fatalf("expected MirrorInterval %q for mirror repo, got %s", "1h0m0s", got)
	}
}

func TestAccRepositoryResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccRepositoryResourceConfig("test-repo-1", "Test repository", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gitea_repository.test", "name", "test-repo-1"),
					resource.TestCheckResourceAttr("gitea_repository.test", "description", "Test repository"),
					resource.TestCheckResourceAttr("gitea_repository.test", "private", "true"),
					resource.TestCheckResourceAttrSet("gitea_repository.test", "id"),
					resource.TestCheckResourceAttrSet("gitea_repository.test", "full_name"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "gitea_repository.test",
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
					resource.TestCheckResourceAttr("gitea_repository.test", "description", "Updated repository"),
					resource.TestCheckResourceAttr("gitea_repository.test", "private", "false"),
				),
			},
		},
	})
}

func testAccRepositoryResourceConfig(name, description string, private bool) string {
	return providerConfig() + fmt.Sprintf(`
resource "gitea_repository" "test" {
  username    = "root"
  name        = %[1]q
  description = %[2]q
  private     = %[3]t
}
`, name, description, private)
}

// Regression test for https://github.com/maxsargentdev/terraform-provider-gitea/issues/16
// Creates a non-mirror repo with feature flags set, which forces the post-create
// EditRepo path. Prior to the fix this failed with:
//   "repo is not a mirror, can not change mirror interval"
func TestAccRepositoryResource_NonMirrorWithEditSettings(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryResourceConfigNonMirror("test-repo-issue-16"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gitea_repository.test", "name", "test-repo-issue-16"),
					resource.TestCheckResourceAttr("gitea_repository.test", "mirror", "false"),
					resource.TestCheckResourceAttr("gitea_repository.test", "has_issues", "true"),
					resource.TestCheckResourceAttr("gitea_repository.test", "has_wiki", "false"),
					resource.TestCheckResourceAttr("gitea_repository.test", "has_pull_requests", "true"),
				),
			},
		},
	})
}

func testAccRepositoryResourceConfigNonMirror(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "gitea_repository" "test" {
  username          = "root"
  name              = %[1]q
  description       = "non-mirror repo with edit settings"
  private           = true
  mirror            = false
  has_issues        = true
  has_wiki          = false
  has_pull_requests = true
}
`, name)
}
