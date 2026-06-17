package provider

import (
	"context"
	"fmt"
	"testing"

	"code.gitea.io/sdk/gitea"
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

// Regression test: Gitea rejects archive/un-archive on mirror repos.
// buildEditRepoOption must not set Archived when the plan marks the repo as a mirror.
func TestBuildEditRepoOption_SkipsArchivedForMirrorRepo(t *testing.T) {
	archived := true
	plan := &repositoryResourceModel{
		Name:     types.StringValue("example"),
		Mirror:   types.BoolValue(true),
		Archived: types.BoolValue(archived),
	}
	opts := buildEditRepoOption(context.Background(), plan)
	if opts.Archived != nil {
		t.Fatalf("expected Archived to be nil for mirror repo, got %v", *opts.Archived)
	}
}

// Archived must still be forwarded for non-mirror repos.
func TestBuildEditRepoOption_SetsArchivedForNonMirrorRepo(t *testing.T) {
	plan := &repositoryResourceModel{
		Name:     types.StringValue("example"),
		Mirror:   types.BoolValue(false),
		Archived: types.BoolValue(true),
	}
	opts := buildEditRepoOption(context.Background(), plan)
	if opts.Archived == nil {
		t.Fatal("expected Archived to be set for non-mirror repo")
	}
	if !*opts.Archived {
		t.Fatal("expected Archived to be true for non-mirror repo")
	}
}

func TestBuildEditRepoOption_PreservesExplicitFeatureFlagsFromPlan(t *testing.T) {
	planned := repositoryResourceModel{
		Name:      types.StringValue("example"),
		HasIssues: types.BoolValue(false),
	}

	state := planned
	mapRepositoryToModel(context.Background(), &gitea.Repository{
		ID:        1,
		Name:      "example",
		HasIssues: true,
	}, &state)

	opts := buildEditRepoOption(context.Background(), &planned)
	if opts.HasIssues == nil {
		t.Fatal("expected HasIssues to be forwarded to EditRepo")
	}
	if *opts.HasIssues {
		t.Fatal("expected EditRepo to preserve planned has_issues=false")
	}
}

func TestBuildEditRepoOption_PreservesExplicitHasWikiFromPlan(t *testing.T) {
	planned := repositoryResourceModel{
		Name:    types.StringValue("example"),
		HasWiki: types.BoolValue(false),
	}

	state := planned
	mapRepositoryToModel(context.Background(), &gitea.Repository{
		ID:      1,
		Name:    "example",
		HasWiki: true,
	}, &state)

	opts := buildEditRepoOption(context.Background(), &planned)
	if opts.HasWiki == nil {
		t.Fatal("expected HasWiki to be forwarded to EditRepo")
	}
	if *opts.HasWiki {
		t.Fatal("expected EditRepo to preserve planned has_wiki=false")
	}
}

func TestBuildEditRepoOption_PreservesExplicitHasProjectsFromPlan(t *testing.T) {
	planned := repositoryResourceModel{
		Name:        types.StringValue("example"),
		HasProjects: types.BoolValue(false),
	}

	state := planned
	mapRepositoryToModel(context.Background(), &gitea.Repository{
		ID:          1,
		Name:        "example",
		HasProjects: true,
	}, &state)

	opts := buildEditRepoOption(context.Background(), &planned)
	if opts.HasProjects == nil {
		t.Fatal("expected HasProjects to be forwarded to EditRepo")
	}
	if *opts.HasProjects {
		t.Fatal("expected EditRepo to preserve planned has_projects=false")
	}
}

func TestBuildEditRepoOption_SetsDefaultMergeStyle(t *testing.T) {
	plan := &repositoryResourceModel{
		Name:               types.StringValue("example"),
		DefaultMergeStyle:  types.StringValue("rebase"),
	}
	opts := buildEditRepoOption(context.Background(), plan)
	if opts.DefaultMergeStyle == nil || string(*opts.DefaultMergeStyle) != "rebase" {
		got := "<nil>"
		if opts.DefaultMergeStyle != nil {
			got = string(*opts.DefaultMergeStyle)
		}
		t.Fatalf("expected DefaultMergeStyle %q, got %s", "rebase", got)
	}
}

func TestBuildEditRepoOption_SetsDefaultMergeStyleRebaseFF(t *testing.T) {
	plan := &repositoryResourceModel{
		Name:              types.StringValue("example"),
		DefaultMergeStyle: types.StringValue("rebase-ff"),
	}

	opts := buildEditRepoOption(context.Background(), plan)
	if opts.DefaultMergeStyle == nil || string(*opts.DefaultMergeStyle) != "rebase-ff" {
		got := "<nil>"
		if opts.DefaultMergeStyle != nil {
			got = string(*opts.DefaultMergeStyle)
		}
		t.Fatalf("expected DefaultMergeStyle %q, got %s", "rebase-ff", got)
	}
}

func TestBuildEditRepoOption_SkipsDefaultMergeStyleWhenNull(t *testing.T) {
	plan := &repositoryResourceModel{
		Name:              types.StringValue("example"),
		DefaultMergeStyle: types.StringNull(),
	}
	opts := buildEditRepoOption(context.Background(), plan)
	if opts.DefaultMergeStyle != nil {
		t.Fatalf("expected DefaultMergeStyle to be nil for null value, got %q", string(*opts.DefaultMergeStyle))
	}
}

func TestBuildEditRepoOption_SetsMergeOptions(t *testing.T) {
	plan := &repositoryResourceModel{
		Name:               types.StringValue("example"),
		AllowMergeCommits:  types.BoolValue(false),
		AllowRebase:        types.BoolValue(true),
		AllowRebaseExplicit: types.BoolValue(false),
		AllowSquashMerge:   types.BoolValue(false),
		DefaultMergeStyle:  types.StringValue("rebase-merge"),
	}

	opts := buildEditRepoOption(context.Background(), plan)

	if opts.AllowMerge == nil || *opts.AllowMerge != false {
		t.Fatal("expected AllowMerge to be set to false")
	}
	if opts.AllowRebase == nil || *opts.AllowRebase != true {
		t.Fatal("expected AllowRebase to be set to true")
	}
	if opts.AllowRebaseMerge == nil || *opts.AllowRebaseMerge != false {
		t.Fatal("expected AllowRebaseMerge to be set to false")
	}
	if opts.AllowSquash == nil || *opts.AllowSquash != false {
		t.Fatal("expected AllowSquash to be set to false")
	}
	if opts.DefaultMergeStyle == nil || string(*opts.DefaultMergeStyle) != "rebase-merge" {
		got := "<nil>"
		if opts.DefaultMergeStyle != nil {
			got = string(*opts.DefaultMergeStyle)
		}
		t.Fatalf("expected DefaultMergeStyle %q, got %s", "rebase-merge", got)
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
//
//	"repo is not a mirror, can not change mirror interval"
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

func TestAccRepositoryResource_ExplicitlyDisablesIssues(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryResourceConfigIssuesDisabled("test-repo-issues-disabled"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gitea_repository.test", "name", "test-repo-issues-disabled"),
					resource.TestCheckResourceAttr("gitea_repository.test", "has_issues", "false"),
				),
			},
		},
	})
}

func TestAccRepositoryResource_ExplicitlyDisablesWiki(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryResourceConfigWikiDisabled("test-repo-wiki-disabled"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gitea_repository.test", "name", "test-repo-wiki-disabled"),
					resource.TestCheckResourceAttr("gitea_repository.test", "has_wiki", "false"),
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

func testAccRepositoryResourceConfigIssuesDisabled(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "gitea_repository" "test" {
	username    = "root"
	name        = %[1]q
	description = "repo with issues disabled"
	private     = true
	has_issues  = false
}
`, name)
}

func testAccRepositoryResourceConfigWikiDisabled(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "gitea_repository" "test" {
	username    = "root"
	name        = %[1]q
	description = "repo with wiki disabled"
	private     = true
	has_wiki    = false
}
`, name)
}
func TestAccRepositoryResource_ExplicitlyDisablesProjects(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryResourceConfigProjectsDisabled("test-repo-projects-disabled"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gitea_repository.test", "name", "test-repo-projects-disabled"),
					resource.TestCheckResourceAttr("gitea_repository.test", "has_projects", "false"),
				),
			},
		},
	})
}

func testAccRepositoryResourceConfigProjectsDisabled(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "gitea_repository" "test" {
	username     = "root"
	name         = %[1]q
	description  = "repo with projects disabled"
	private      = true
	has_projects = false
}
`, name)
}

// Regression test: creating a mirror repo with archived=false must not fail with
// "repo is a mirror, cannot archive/un-archive". Prior to the fix, the post-create
// EditRepo call unconditionally forwarded the Archived field, which Gitea rejects
// for mirror repositories.
func TestAccRepositoryResource_MirrorRepoWithArchivedFalse(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryResourceConfigMirror("test-mirror-archived-false"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gitea_repository.test", "name", "test-mirror-archived-false"),
					resource.TestCheckResourceAttr("gitea_repository.test", "mirror", "true"),
					resource.TestCheckResourceAttr("gitea_repository.test", "archived", "false"),
				),
			},
		},
	})
}

func testAccRepositoryResourceConfigMirror(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "gitea_repository" "test" {
  username                = "root"
  name                    = %[1]q
  description             = "mirror repo regression test"
  private                 = true
  migration_clone_address = "https://github.com/octocat/Hello-World.git"
  migration_service       = "git"
  mirror                  = true
  archived                = false
  default_branch          = "master"
}
`, name)
}

func TestAccRepositoryResource_DefaultMergeStyle(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with default_merge_style set to the server default value.
			{
				Config: testAccRepositoryResourceConfigWithDefaultMergeStyle("test-repo-merge-style", "merge"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gitea_repository.test", "name", "test-repo-merge-style"),
					resource.TestCheckResourceAttr("gitea_repository.test", "default_merge_style", "merge"),
				),
			},
			// Update to the same value to verify no unexpected drift.
			{
				Config: testAccRepositoryResourceConfigWithDefaultMergeStyle("test-repo-merge-style", "merge"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gitea_repository.test", "name", "test-repo-merge-style"),
					resource.TestCheckResourceAttr("gitea_repository.test", "default_merge_style", "merge"),
				),
			},
		},
	})
}

func TestAccRepositoryResource_MergeOptionsCreateAndUpdate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Set merge options explicitly on create.
			{
				Config: testAccRepositoryResourceConfigWithMergeOptions("test-repo-merge-options", "merge"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gitea_repository.test", "name", "test-repo-merge-options"),
					resource.TestCheckResourceAttr("gitea_repository.test", "has_pull_requests", "true"),
					resource.TestCheckResourceAttr("gitea_repository.test", "allow_merge_commits", "false"),
					resource.TestCheckResourceAttr("gitea_repository.test", "allow_rebase", "true"),
					resource.TestCheckResourceAttr("gitea_repository.test", "allow_rebase_explicit", "false"),
					resource.TestCheckResourceAttr("gitea_repository.test", "allow_squash_merge", "false"),
					resource.TestCheckResourceAttr("gitea_repository.test", "default_merge_style", "merge"),
				),
			},
			// Change merge style after creation while keeping the other merge options explicit.
			{
				Config: testAccRepositoryResourceConfigWithMergeOptions("test-repo-merge-options", "rebase-merge"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gitea_repository.test", "name", "test-repo-merge-options"),
					resource.TestCheckResourceAttr("gitea_repository.test", "has_pull_requests", "true"),
					resource.TestCheckResourceAttr("gitea_repository.test", "allow_merge_commits", "false"),
					resource.TestCheckResourceAttr("gitea_repository.test", "allow_rebase", "true"),
					resource.TestCheckResourceAttr("gitea_repository.test", "allow_rebase_explicit", "false"),
					resource.TestCheckResourceAttr("gitea_repository.test", "allow_squash_merge", "false"),
					resource.TestCheckResourceAttr("gitea_repository.test", "default_merge_style", "rebase-merge"),
				),
			},
			// Change merge style to rebase-ff and verify it persists.
			{
				Config: testAccRepositoryResourceConfigWithMergeOptions("test-repo-merge-options", "rebase-ff"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gitea_repository.test", "name", "test-repo-merge-options"),
					resource.TestCheckResourceAttr("gitea_repository.test", "has_pull_requests", "true"),
					resource.TestCheckResourceAttr("gitea_repository.test", "allow_merge_commits", "false"),
					resource.TestCheckResourceAttr("gitea_repository.test", "allow_rebase", "true"),
					resource.TestCheckResourceAttr("gitea_repository.test", "allow_rebase_explicit", "false"),
					resource.TestCheckResourceAttr("gitea_repository.test", "allow_squash_merge", "false"),
					resource.TestCheckResourceAttr("gitea_repository.test", "default_merge_style", "rebase-ff"),
				),
			},
		},
	})
}

func testAccRepositoryResourceConfigWithDefaultMergeStyle(name, mergeStyle string) string {
	return providerConfig() + fmt.Sprintf(`
resource "gitea_repository" "test" {
  username            = "root"
  name                = %[1]q
  description         = "Test repository with default merge style"
  private             = false
  default_merge_style = %[2]q
}
`, name, mergeStyle)
}

func testAccRepositoryResourceConfigWithMergeOptions(name, mergeStyle string) string {
	return providerConfig() + fmt.Sprintf(`
resource "gitea_repository" "test" {
	username              = "root"
	name                  = %[1]q
	description           = "Test repository with merge options"
	private               = false
	has_pull_requests     = true
	allow_merge_commits   = false
	allow_rebase          = true
	allow_rebase_explicit = false
	allow_squash_merge    = false
	default_merge_style   = %[2]q
}
`, name, mergeStyle)
}
