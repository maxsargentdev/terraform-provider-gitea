package provider

import (
	"context"
	"fmt"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &repositoryResource{}
	_ resource.ResourceWithConfigure   = &repositoryResource{}
	_ resource.ResourceWithImportState = &repositoryResource{}
)

func NewRepositoryResource() resource.Resource {
	return &repositoryResource{}
}

type repositoryResource struct {
	client *gitea.Client
}

// internalTrackerModel represents the internal issue tracker settings
type internalTrackerModel struct {
	EnableTimeTracker                types.Bool `tfsdk:"enable_time_tracker"`
	AllowOnlyContributorsToTrackTime types.Bool `tfsdk:"allow_only_contributors_to_track_time"`
	EnableIssueDependencies          types.Bool `tfsdk:"enable_issue_dependencies"`
}

// externalTrackerModel represents the external issue tracker settings
type externalTrackerModel struct {
	ExternalTrackerURL    types.String `tfsdk:"external_tracker_url"`
	ExternalTrackerFormat types.String `tfsdk:"external_tracker_format"`
	ExternalTrackerStyle  types.String `tfsdk:"external_tracker_style"`
}

// externalWikiModel represents the external wiki settings
type externalWikiModel struct {
	ExternalWikiURL types.String `tfsdk:"external_wiki_url"`
}

type repositoryResourceModel struct {
	// Required
	Owner types.String `tfsdk:"owner"`
	Name  types.String `tfsdk:"name"`

	// Optional - from CreateRepoOption (creation-only fields)
	IssueLabels types.String `tfsdk:"issue_labels"`
	AutoInit    types.Bool   `tfsdk:"auto_init"`
	Gitignores  types.String `tfsdk:"gitignores"`
	License     types.String `tfsdk:"license"`
	Readme      types.String `tfsdk:"readme"`
	TrustModel  types.String `tfsdk:"trust_model"`

	// Optional - from both Create and Edit
	Description      types.String `tfsdk:"description"`
	Private          types.Bool   `tfsdk:"private"`
	Template         types.Bool   `tfsdk:"template"`
	DefaultBranch    types.String `tfsdk:"default_branch"`
	ObjectFormatName types.String `tfsdk:"object_format_name"`

	// Optional - from EditRepoOption only (can be set after creation)
	Website                       types.String `tfsdk:"website"`
	HasIssues                     types.Bool   `tfsdk:"has_issues"`
	HasWiki                       types.Bool   `tfsdk:"has_wiki"`
	HasPullRequests               types.Bool   `tfsdk:"has_pull_requests"`
	HasProjects                   types.Bool   `tfsdk:"has_projects"`
	HasReleases                   types.Bool   `tfsdk:"has_releases"`
	HasPackages                   types.Bool   `tfsdk:"has_packages"`
	HasActions                    types.Bool   `tfsdk:"has_actions"`
	IgnoreWhitespaceConflicts     types.Bool   `tfsdk:"ignore_whitespace_conflicts"`
	AllowMergeCommits             types.Bool   `tfsdk:"allow_merge_commits"`
	AllowRebase                   types.Bool   `tfsdk:"allow_rebase"`
	AllowRebaseExplicit           types.Bool   `tfsdk:"allow_rebase_explicit"`
	AllowSquashMerge              types.Bool   `tfsdk:"allow_squash_merge"`
	AllowFastForwardOnlyMerge     types.Bool   `tfsdk:"allow_fast_forward_only_merge"`
	Archived                      types.Bool   `tfsdk:"archived"`
	DefaultMergeStyle             types.String `tfsdk:"default_merge_style"`
	DefaultDeleteBranchAfterMerge types.Bool   `tfsdk:"default_delete_branch_after_merge"`
	MirrorInterval                types.String `tfsdk:"mirror_interval"`
	AllowManualMerge              types.Bool   `tfsdk:"allow_manual_merge"`
	AutodetectManualMerge         types.Bool   `tfsdk:"autodetect_manual_merge"`
	ProjectsMode                  types.String `tfsdk:"projects_mode"`

	// Nested objects
	InternalTracker types.Object `tfsdk:"internal_tracker"`
	ExternalTracker types.Object `tfsdk:"external_tracker"`
	ExternalWiki    types.Object `tfsdk:"external_wiki"`

	// Computed - from Repository response
	Id              types.Int64  `tfsdk:"id"`
	FullName        types.String `tfsdk:"full_name"`
	HtmlUrl         types.String `tfsdk:"html_url"`
	SshUrl          types.String `tfsdk:"ssh_url"`
	CloneUrl        types.String `tfsdk:"clone_url"`
	OriginalUrl     types.String `tfsdk:"original_url"`
	Empty           types.Bool   `tfsdk:"empty"`
	Fork            types.Bool   `tfsdk:"fork"`
	Mirror          types.Bool   `tfsdk:"mirror"`
	Size            types.Int64  `tfsdk:"size"`
	StarsCount      types.Int64  `tfsdk:"stars_count"`
	ForksCount      types.Int64  `tfsdk:"forks_count"`
	WatchersCount   types.Int64  `tfsdk:"watchers_count"`
	OpenIssuesCount types.Int64  `tfsdk:"open_issues_count"`
	OpenPrCounter   types.Int64  `tfsdk:"open_pr_counter"`
	ReleaseCounter  types.Int64  `tfsdk:"release_counter"`
	CreatedAt       types.String `tfsdk:"created_at"`
	UpdatedAt       types.String `tfsdk:"updated_at"`
	AvatarUrl       types.String `tfsdk:"avatar_url"`
	Internal        types.Bool   `tfsdk:"internal"`
}

// Attribute types for nested objects
var internalTrackerAttrTypes = map[string]attr.Type{
	"enable_time_tracker":                   types.BoolType,
	"allow_only_contributors_to_track_time": types.BoolType,
	"enable_issue_dependencies":             types.BoolType,
}

var externalTrackerAttrTypes = map[string]attr.Type{
	"external_tracker_url":    types.StringType,
	"external_tracker_format": types.StringType,
	"external_tracker_style":  types.StringType,
}

var externalWikiAttrTypes = map[string]attr.Type{
	"external_wiki_url": types.StringType,
}

func (r *repositoryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repository"
}

func (r *repositoryResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manages a Gitea repository.",
		MarkdownDescription: "Manages a Gitea repository. This resource allows you to create, update, and delete repositories in Gitea.",
		Attributes: map[string]schema.Attribute{
			// ==================== REQUIRED ====================
			"owner": schema.StringAttribute{
				Required:            true,
				Description:         "The owner of the repository (username or organization name).",
				MarkdownDescription: "The owner of the repository (username or organization name).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				Description:         "Name of the repository to create. Must be unique within the owner's namespace.",
				MarkdownDescription: "Name of the repository to create. Must be unique within the owner's namespace.",
			},

			// ==================== OPTIONAL - Creation Only ====================
			"issue_labels": schema.StringAttribute{
				Optional:            true,
				Description:         "Issue label set to use when initializing the repository. Only used during creation.",
				MarkdownDescription: "Issue label set to use when initializing the repository. Only used during creation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"auto_init": schema.BoolAttribute{
				Optional:            true,
				Description:         "Whether the repository should be auto-initialized with a README. Only used during creation.",
				MarkdownDescription: "Whether the repository should be auto-initialized with a README. Only used during creation.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"gitignores": schema.StringAttribute{
				Optional:            true,
				Description:         "Gitignore templates to use when initializing the repository. Comma-separated list. Only used during creation.",
				MarkdownDescription: "Gitignore templates to use when initializing the repository. Comma-separated list. Only used during creation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"license": schema.StringAttribute{
				Optional:            true,
				Description:         "License template to use when initializing the repository. Only used during creation.",
				MarkdownDescription: "License template to use when initializing the repository. Only used during creation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"readme": schema.StringAttribute{
				Optional:            true,
				Description:         "Readme template to use when initializing the repository. Only used during creation.",
				MarkdownDescription: "Readme template to use when initializing the repository. Only used during creation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"trust_model": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Trust model for verifying commits. Valid values: default, collaborator, committer, collaboratorcommitter.",
				MarkdownDescription: "Trust model for verifying commits. Valid values: `default`, `collaborator`, `committer`, `collaboratorcommitter`.",
				Validators: []validator.String{
					stringvalidator.OneOf(
						"default",
						"collaborator",
						"committer",
						"collaboratorcommitter",
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"object_format_name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Object format name of the underlying git repository (sha1 or sha256). Only used during creation. Requires Gitea 1.22.0+.",
				MarkdownDescription: "Object format name of the underlying git repository (`sha1` or `sha256`). Only used during creation. Requires Gitea 1.22.0+.",
				Validators: []validator.String{
					stringvalidator.OneOf("sha1", "sha256"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			// ==================== OPTIONAL - Create and Edit ====================
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Description of the repository.",
				MarkdownDescription: "Description of the repository.",
			},
			"private": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether the repository is private.",
				MarkdownDescription: "Whether the repository is private.",
			},
			"template": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether the repository is a template repository.",
				MarkdownDescription: "Whether the repository is a template repository.",
			},
			"default_branch": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Default branch of the repository.",
				MarkdownDescription: "Default branch of the repository.",
			},

			// ==================== OPTIONAL - Edit Only ====================
			"website": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "A URL with more information about the repository.",
				MarkdownDescription: "A URL with more information about the repository.",
			},
			"has_issues": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				Description:         "Whether the repository has issues enabled.",
				MarkdownDescription: "Whether the repository has issues enabled.",
			},
			"has_wiki": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				Description:         "Whether the repository has wiki enabled.",
				MarkdownDescription: "Whether the repository has wiki enabled.",
			},
			"has_pull_requests": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				Description:         "Whether the repository allows pull requests.",
				MarkdownDescription: "Whether the repository allows pull requests.",
			},
			"has_projects": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether the repository has projects enabled.",
				MarkdownDescription: "Whether the repository has projects enabled.",
			},
			"has_releases": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether the repository has releases enabled.",
				MarkdownDescription: "Whether the repository has releases enabled.",
			},
			"has_packages": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether the repository has packages enabled.",
				MarkdownDescription: "Whether the repository has packages enabled.",
			},
			"has_actions": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether the repository has actions enabled.",
				MarkdownDescription: "Whether the repository has actions enabled.",
			},
			"ignore_whitespace_conflicts": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether to ignore whitespace for conflicts. Requires has_pull_requests to be true.",
				MarkdownDescription: "Whether to ignore whitespace for conflicts. Requires `has_pull_requests` to be `true`.",
			},
			"allow_merge_commits": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether to allow merging pull requests with a merge commit. Requires has_pull_requests to be true.",
				MarkdownDescription: "Whether to allow merging pull requests with a merge commit. Requires `has_pull_requests` to be `true`.",
			},
			"allow_rebase": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether to allow rebase-merging pull requests. Requires has_pull_requests to be true.",
				MarkdownDescription: "Whether to allow rebase-merging pull requests. Requires `has_pull_requests` to be `true`.",
			},
			"allow_rebase_explicit": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether to allow rebase with explicit merge commits (--no-ff). Requires has_pull_requests to be true.",
				MarkdownDescription: "Whether to allow rebase with explicit merge commits (`--no-ff`). Requires `has_pull_requests` to be `true`.",
			},
			"allow_squash_merge": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether to allow squash-merging pull requests. Requires has_pull_requests to be true.",
				MarkdownDescription: "Whether to allow squash-merging pull requests. Requires `has_pull_requests` to be `true`.",
			},
			"allow_fast_forward_only_merge": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether to allow fast-forward only merging pull requests. Requires has_pull_requests to be true.",
				MarkdownDescription: "Whether to allow fast-forward only merging pull requests. Requires `has_pull_requests` to be `true`.",
			},
			"archived": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether the repository is archived.",
				MarkdownDescription: "Whether the repository is archived.",
			},
			"default_merge_style": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Default merge style for pull requests. Valid values: merge, rebase, rebase-merge, squash. Requires has_pull_requests to be true.",
				MarkdownDescription: "Default merge style for pull requests. Valid values: `merge`, `rebase`, `rebase-merge`, `squash`. Requires `has_pull_requests` to be `true`.",
				Validators: []validator.String{
					stringvalidator.OneOf("merge", "rebase", "rebase-merge", "squash"),
				},
			},
			"default_delete_branch_after_merge": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether to delete the branch after merging a pull request. Requires has_pull_requests to be true.",
				MarkdownDescription: "Whether to delete the branch after merging a pull request. Requires `has_pull_requests` to be `true`.",
			},
			"mirror_interval": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Mirror interval for mirror repositories (e.g., '8h30m0s'). Only applicable for mirror repositories.",
				MarkdownDescription: "Mirror interval for mirror repositories (e.g., `8h30m0s`). Only applicable for mirror repositories.",
			},
			"allow_manual_merge": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether to allow marking pull requests as merged manually. Requires has_pull_requests to be true.",
				MarkdownDescription: "Whether to allow marking pull requests as merged manually. Requires `has_pull_requests` to be `true`.",
			},
			"autodetect_manual_merge": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether to enable autodetection of manual merge. Requires has_pull_requests to be true. Note: May cause misjudgments in some cases.",
				MarkdownDescription: "Whether to enable autodetection of manual merge. Requires `has_pull_requests` to be `true`. Note: May cause misjudgments in some cases.",
			},
			"projects_mode": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Projects mode for the repository. Valid values: repo, owner, all. Requires has_projects to be true.",
				MarkdownDescription: "Projects mode for the repository. Valid values: `repo`, `owner`, `all`. Requires `has_projects` to be `true`.",
				Validators: []validator.String{
					stringvalidator.OneOf("repo", "owner", "all"),
				},
			},

			// ==================== NESTED OBJECTS ====================
			"internal_tracker": schema.SingleNestedAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Settings for the built-in issue tracker. Requires has_issues to be true.",
				MarkdownDescription: "Settings for the built-in issue tracker. Requires `has_issues` to be `true`.",
				Attributes: map[string]schema.Attribute{
					"enable_time_tracker": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						Description:         "Whether to enable time tracking.",
						MarkdownDescription: "Whether to enable time tracking.",
					},
					"allow_only_contributors_to_track_time": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						Description:         "Whether only contributors can track time.",
						MarkdownDescription: "Whether only contributors can track time.",
					},
					"enable_issue_dependencies": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						Description:         "Whether to enable issue dependencies.",
						MarkdownDescription: "Whether to enable issue dependencies.",
					},
				},
			},
			"external_tracker": schema.SingleNestedAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Settings for external issue tracker. Requires has_issues to be true.",
				MarkdownDescription: "Settings for external issue tracker. Requires `has_issues` to be `true`.",
				Attributes: map[string]schema.Attribute{
					"external_tracker_url": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						Description:         "URL of external issue tracker.",
						MarkdownDescription: "URL of external issue tracker.",
					},
					"external_tracker_format": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						Description:         "External issue tracker URL format. Use placeholders {user}, {repo} and {index}.",
						MarkdownDescription: "External issue tracker URL format. Use placeholders `{user}`, `{repo}` and `{index}`.",
					},
					"external_tracker_style": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						Description:         "External issue tracker number format. Valid values: numeric, alphanumeric.",
						MarkdownDescription: "External issue tracker number format. Valid values: `numeric`, `alphanumeric`.",
						Validators: []validator.String{
							stringvalidator.OneOf("numeric", "alphanumeric"),
						},
					},
				},
			},
			"external_wiki": schema.SingleNestedAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Settings for external wiki. Requires has_wiki to be true.",
				MarkdownDescription: "Settings for external wiki. Requires `has_wiki` to be `true`.",
				Attributes: map[string]schema.Attribute{
					"external_wiki_url": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						Description:         "URL of external wiki.",
						MarkdownDescription: "URL of external wiki.",
					},
				},
			},

			// ==================== COMPUTED ====================
			"id": schema.Int64Attribute{
				Computed:            true,
				Description:         "The unique identifier of the repository.",
				MarkdownDescription: "The unique identifier of the repository.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"full_name": schema.StringAttribute{
				Computed:            true,
				Description:         "Full name of the repository (owner/name).",
				MarkdownDescription: "Full name of the repository (`owner/name`).",
			},
			"html_url": schema.StringAttribute{
				Computed:            true,
				Description:         "The URL to view the repository in the web UI.",
				MarkdownDescription: "The URL to view the repository in the web UI.",
			},
			"ssh_url": schema.StringAttribute{
				Computed:            true,
				Description:         "The SSH URL to clone the repository.",
				MarkdownDescription: "The SSH URL to clone the repository.",
			},
			"clone_url": schema.StringAttribute{
				Computed:            true,
				Description:         "The HTTPS URL to clone the repository.",
				MarkdownDescription: "The HTTPS URL to clone the repository.",
			},
			"original_url": schema.StringAttribute{
				Computed:            true,
				Description:         "The original URL of the repository (for mirrors/forks).",
				MarkdownDescription: "The original URL of the repository (for mirrors/forks).",
			},
			"empty": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether the repository is empty.",
				MarkdownDescription: "Whether the repository is empty.",
			},
			"fork": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether the repository is a fork.",
				MarkdownDescription: "Whether the repository is a fork.",
			},
			"mirror": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether the repository is a mirror.",
				MarkdownDescription: "Whether the repository is a mirror.",
			},
			"size": schema.Int64Attribute{
				Computed:            true,
				Description:         "Size of the repository in kilobytes.",
				MarkdownDescription: "Size of the repository in kilobytes.",
			},
			"stars_count": schema.Int64Attribute{
				Computed:            true,
				Description:         "Number of stars the repository has.",
				MarkdownDescription: "Number of stars the repository has.",
			},
			"forks_count": schema.Int64Attribute{
				Computed:            true,
				Description:         "Number of forks of the repository.",
				MarkdownDescription: "Number of forks of the repository.",
			},
			"watchers_count": schema.Int64Attribute{
				Computed:            true,
				Description:         "Number of watchers of the repository.",
				MarkdownDescription: "Number of watchers of the repository.",
			},
			"open_issues_count": schema.Int64Attribute{
				Computed:            true,
				Description:         "Number of open issues in the repository.",
				MarkdownDescription: "Number of open issues in the repository.",
			},
			"open_pr_counter": schema.Int64Attribute{
				Computed:            true,
				Description:         "Number of open pull requests in the repository.",
				MarkdownDescription: "Number of open pull requests in the repository.",
			},
			"release_counter": schema.Int64Attribute{
				Computed:            true,
				Description:         "Number of releases in the repository.",
				MarkdownDescription: "Number of releases in the repository.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				Description:         "Timestamp when the repository was created.",
				MarkdownDescription: "Timestamp when the repository was created.",
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				Description:         "Timestamp when the repository was last updated.",
				MarkdownDescription: "Timestamp when the repository was last updated.",
			},
			"avatar_url": schema.StringAttribute{
				Computed:            true,
				Description:         "URL of the repository avatar.",
				MarkdownDescription: "URL of the repository avatar.",
			},
			"internal": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether the repository is internal (visible only to organization members).",
				MarkdownDescription: "Whether the repository is internal (visible only to organization members).",
			},
		},
	}
}

// Helper function to map Gitea Repository to Terraform model
func mapRepositoryToModel(ctx context.Context, repo *gitea.Repository, model *repositoryResourceModel) {
	// Computed fields
	model.Id = types.Int64Value(repo.ID)
	model.FullName = types.StringValue(repo.FullName)
	model.HtmlUrl = types.StringValue(repo.HTMLURL)
	model.CloneUrl = types.StringValue(repo.CloneURL)
	model.SshUrl = types.StringValue(repo.SSHURL)
	model.OriginalUrl = types.StringValue(repo.OriginalURL)
	model.Empty = types.BoolValue(repo.Empty)
	model.Fork = types.BoolValue(repo.Fork)
	model.Mirror = types.BoolValue(repo.Mirror)
	model.Size = types.Int64Value(int64(repo.Size))
	model.StarsCount = types.Int64Value(int64(repo.Stars))
	model.ForksCount = types.Int64Value(int64(repo.Forks))
	model.WatchersCount = types.Int64Value(int64(repo.Watchers))
	model.OpenIssuesCount = types.Int64Value(int64(repo.OpenIssues))
	model.OpenPrCounter = types.Int64Value(int64(repo.OpenPulls))
	model.ReleaseCounter = types.Int64Value(int64(repo.Releases))
	model.CreatedAt = types.StringValue(repo.Created.String())
	model.UpdatedAt = types.StringValue(repo.Updated.String())
	model.AvatarUrl = types.StringValue(repo.AvatarURL)
	model.Internal = types.BoolValue(repo.Internal)

	// Basic fields that can be updated
	model.Name = types.StringValue(repo.Name)
	model.Description = types.StringValue(repo.Description)
	model.Private = types.BoolValue(repo.Private)
	model.DefaultBranch = types.StringValue(repo.DefaultBranch)
	model.Template = types.BoolValue(repo.Template)
	model.ObjectFormatName = types.StringValue(repo.ObjectFormatName)
	model.Website = types.StringValue(repo.Website)
	model.Archived = types.BoolValue(repo.Archived)
	model.MirrorInterval = types.StringValue(repo.MirrorInterval)

	// Feature flags
	model.HasIssues = types.BoolValue(repo.HasIssues)
	model.HasWiki = types.BoolValue(repo.HasWiki)
	model.HasPullRequests = types.BoolValue(repo.HasPullRequests)
	model.HasProjects = types.BoolValue(repo.HasProjects)
	model.HasReleases = types.BoolValue(repo.HasReleases)
	model.HasPackages = types.BoolValue(repo.HasPackages)
	model.HasActions = types.BoolValue(repo.HasActions)

	// Pull request merge settings
	model.IgnoreWhitespaceConflicts = types.BoolValue(repo.IgnoreWhitespaceConflicts)
	model.AllowMergeCommits = types.BoolValue(repo.AllowMerge)
	model.AllowRebase = types.BoolValue(repo.AllowRebase)
	model.AllowRebaseExplicit = types.BoolValue(repo.AllowRebaseMerge)
	model.AllowSquashMerge = types.BoolValue(repo.AllowSquash)
	model.AllowFastForwardOnlyMerge = types.BoolValue(repo.AllowFastForwardOnlyMerge)
	model.DefaultMergeStyle = types.StringValue(string(repo.DefaultMergeStyle))
	model.DefaultDeleteBranchAfterMerge = types.BoolValue(repo.DefaultDeleteBranchAfterMerge)

	// Projects mode
	if repo.ProjectsMode != nil {
		model.ProjectsMode = types.StringValue(string(*repo.ProjectsMode))
	} else {
		model.ProjectsMode = types.StringNull()
	}

	// Internal tracker
	if repo.InternalTracker != nil {
		internalTrackerObj, _ := types.ObjectValue(internalTrackerAttrTypes, map[string]attr.Value{
			"enable_time_tracker":                   types.BoolValue(repo.InternalTracker.EnableTimeTracker),
			"allow_only_contributors_to_track_time": types.BoolValue(repo.InternalTracker.AllowOnlyContributorsToTrackTime),
			"enable_issue_dependencies":             types.BoolValue(repo.InternalTracker.EnableIssueDependencies),
		})
		model.InternalTracker = internalTrackerObj
	} else {
		model.InternalTracker = types.ObjectNull(internalTrackerAttrTypes)
	}

	// External tracker
	if repo.ExternalTracker != nil {
		externalTrackerObj, _ := types.ObjectValue(externalTrackerAttrTypes, map[string]attr.Value{
			"external_tracker_url":    types.StringValue(repo.ExternalTracker.ExternalTrackerURL),
			"external_tracker_format": types.StringValue(repo.ExternalTracker.ExternalTrackerFormat),
			"external_tracker_style":  types.StringValue(repo.ExternalTracker.ExternalTrackerStyle),
		})
		model.ExternalTracker = externalTrackerObj
	} else {
		model.ExternalTracker = types.ObjectNull(externalTrackerAttrTypes)
	}

	// External wiki
	if repo.ExternalWiki != nil {
		externalWikiObj, _ := types.ObjectValue(externalWikiAttrTypes, map[string]attr.Value{
			"external_wiki_url": types.StringValue(repo.ExternalWiki.ExternalWikiURL),
		})
		model.ExternalWiki = externalWikiObj
	} else {
		model.ExternalWiki = types.ObjectNull(externalWikiAttrTypes)
	}

	// Creation-only fields - preserve from existing model if unknown/null (not returned by API after creation)
	if model.AutoInit.IsUnknown() {
		model.AutoInit = types.BoolNull()
	}
	if model.Gitignores.IsUnknown() {
		model.Gitignores = types.StringNull()
	}
	if model.IssueLabels.IsUnknown() {
		model.IssueLabels = types.StringNull()
	}
	if model.License.IsUnknown() {
		model.License = types.StringNull()
	}
	if model.Readme.IsUnknown() {
		model.Readme = types.StringNull()
	}
	if model.TrustModel.IsUnknown() {
		model.TrustModel = types.StringNull()
	}
	// Note: AllowManualMerge and AutodetectManualMerge are not returned by the API
	if model.AllowManualMerge.IsUnknown() {
		model.AllowManualMerge = types.BoolNull()
	}
	if model.AutodetectManualMerge.IsUnknown() {
		model.AutodetectManualMerge = types.BoolNull()
	}
}

func (r *repositoryResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*gitea.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *gitea.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *repositoryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan repositoryResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create repository
	owner := plan.Owner.ValueString()

	createOpts := gitea.CreateRepoOption{
		Name:             plan.Name.ValueString(),
		Description:      plan.Description.ValueString(),
		Private:          plan.Private.ValueBool(),
		IssueLabels:      plan.IssueLabels.ValueString(),
		AutoInit:         plan.AutoInit.ValueBool(),
		Template:         plan.Template.ValueBool(),
		Gitignores:       plan.Gitignores.ValueString(),
		License:          plan.License.ValueString(),
		Readme:           plan.Readme.ValueString(),
		DefaultBranch:    plan.DefaultBranch.ValueString(),
		TrustModel:       gitea.TrustModel(plan.TrustModel.ValueString()),
		ObjectFormatName: plan.ObjectFormatName.ValueString(),
	}

	// Check if owner is an org or user and use appropriate API
	var repo *gitea.Repository
	var err error

	// Try to get org first to determine if it's an org
	org, _, orgErr := r.client.GetOrg(owner)
	if orgErr == nil && org != nil {
		// Owner is an organization
		repo, _, err = r.client.CreateOrgRepo(owner, createOpts)
	} else {
		// Check if the owner is the currently authenticated user
		currentUser, _, userErr := r.client.GetMyUserInfo()
		if userErr != nil {
			resp.Diagnostics.AddError(
				"Error Getting Current User",
				"Could not determine current user: "+userErr.Error(),
			)
			return
		}

		if currentUser.UserName == owner {
			// Owner is the current user
			repo, _, err = r.client.CreateRepo(createOpts)
		} else {
			// Try to create for another user (admin only)
			repo, _, err = r.client.AdminCreateRepo(owner, createOpts)
		}
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Repository",
			"Could not create repository, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response to state
	mapRepositoryToModel(ctx, repo, &plan)
	plan.Owner = types.StringValue(owner)

	// Preserve creation-only fields that aren't returned by API
	// These stay as the user configured them
	// (they're already in plan, so no action needed)

	// If additional edit-only settings were specified, apply them now
	if needsPostCreateUpdate(&plan) {
		editOpts := buildEditRepoOption(ctx, &plan)
		repo, _, err = r.client.EditRepo(owner, repo.Name, editOpts)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Updating Repository After Creation",
				"Repository was created but failed to apply additional settings: "+err.Error(),
			)
			return
		}
		mapRepositoryToModel(ctx, repo, &plan)
		plan.Owner = types.StringValue(owner)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// needsPostCreateUpdate checks if any edit-only fields were specified that need to be applied after creation
func needsPostCreateUpdate(plan *repositoryResourceModel) bool {
	return !plan.Website.IsNull() && plan.Website.ValueString() != "" ||
		!plan.HasIssues.IsNull() ||
		!plan.HasWiki.IsNull() ||
		!plan.HasPullRequests.IsNull() ||
		!plan.HasProjects.IsNull() ||
		!plan.HasReleases.IsNull() ||
		!plan.HasPackages.IsNull() ||
		!plan.HasActions.IsNull() ||
		!plan.IgnoreWhitespaceConflicts.IsNull() ||
		!plan.AllowMergeCommits.IsNull() ||
		!plan.AllowRebase.IsNull() ||
		!plan.AllowRebaseExplicit.IsNull() ||
		!plan.AllowSquashMerge.IsNull() ||
		!plan.AllowFastForwardOnlyMerge.IsNull() ||
		!plan.Archived.IsNull() ||
		!plan.DefaultMergeStyle.IsNull() ||
		!plan.DefaultDeleteBranchAfterMerge.IsNull() ||
		!plan.MirrorInterval.IsNull() ||
		!plan.AllowManualMerge.IsNull() ||
		!plan.AutodetectManualMerge.IsNull() ||
		!plan.ProjectsMode.IsNull() ||
		!plan.InternalTracker.IsNull() ||
		!plan.ExternalTracker.IsNull() ||
		!plan.ExternalWiki.IsNull()
}

// buildEditRepoOption builds an EditRepoOption from the plan
func buildEditRepoOption(ctx context.Context, plan *repositoryResourceModel) gitea.EditRepoOption {
	editOpts := gitea.EditRepoOption{}

	// Basic fields
	if !plan.Name.IsNull() {
		editOpts.Name = plan.Name.ValueStringPointer()
	}
	if !plan.Description.IsNull() {
		editOpts.Description = plan.Description.ValueStringPointer()
	}
	if !plan.Website.IsNull() {
		editOpts.Website = plan.Website.ValueStringPointer()
	}
	if !plan.Private.IsNull() {
		editOpts.Private = plan.Private.ValueBoolPointer()
	}
	if !plan.Template.IsNull() {
		editOpts.Template = plan.Template.ValueBoolPointer()
	}
	if !plan.DefaultBranch.IsNull() {
		editOpts.DefaultBranch = plan.DefaultBranch.ValueStringPointer()
	}

	// Feature flags
	if !plan.HasIssues.IsNull() {
		editOpts.HasIssues = plan.HasIssues.ValueBoolPointer()
	}
	if !plan.HasWiki.IsNull() {
		editOpts.HasWiki = plan.HasWiki.ValueBoolPointer()
	}
	if !plan.HasPullRequests.IsNull() {
		editOpts.HasPullRequests = plan.HasPullRequests.ValueBoolPointer()
	}
	if !plan.HasProjects.IsNull() {
		editOpts.HasProjects = plan.HasProjects.ValueBoolPointer()
	}
	if !plan.HasReleases.IsNull() {
		editOpts.HasReleases = plan.HasReleases.ValueBoolPointer()
	}
	if !plan.HasPackages.IsNull() {
		editOpts.HasPackages = plan.HasPackages.ValueBoolPointer()
	}
	if !plan.HasActions.IsNull() {
		editOpts.HasActions = plan.HasActions.ValueBoolPointer()
	}

	// Merge settings
	if !plan.IgnoreWhitespaceConflicts.IsNull() {
		editOpts.IgnoreWhitespaceConflicts = plan.IgnoreWhitespaceConflicts.ValueBoolPointer()
	}
	if !plan.AllowMergeCommits.IsNull() {
		editOpts.AllowMerge = plan.AllowMergeCommits.ValueBoolPointer()
	}
	if !plan.AllowRebase.IsNull() {
		editOpts.AllowRebase = plan.AllowRebase.ValueBoolPointer()
	}
	if !plan.AllowRebaseExplicit.IsNull() {
		editOpts.AllowRebaseMerge = plan.AllowRebaseExplicit.ValueBoolPointer()
	}
	if !plan.AllowSquashMerge.IsNull() {
		editOpts.AllowSquash = plan.AllowSquashMerge.ValueBoolPointer()
	}
	if !plan.AllowFastForwardOnlyMerge.IsNull() {
		editOpts.AllowFastForwardOnlyMerge = plan.AllowFastForwardOnlyMerge.ValueBoolPointer()
	}
	if !plan.DefaultMergeStyle.IsNull() {
		mergeStyle := gitea.MergeStyle(plan.DefaultMergeStyle.ValueString())
		editOpts.DefaultMergeStyle = &mergeStyle
	}
	if !plan.DefaultDeleteBranchAfterMerge.IsNull() {
		editOpts.DefaultDeleteBranchAfterMerge = plan.DefaultDeleteBranchAfterMerge.ValueBoolPointer()
	}
	if !plan.AllowManualMerge.IsNull() {
		editOpts.AllowManualMerge = plan.AllowManualMerge.ValueBoolPointer()
	}
	if !plan.AutodetectManualMerge.IsNull() {
		editOpts.AutodetectManualMerge = plan.AutodetectManualMerge.ValueBoolPointer()
	}

	// Other settings
	if !plan.Archived.IsNull() {
		editOpts.Archived = plan.Archived.ValueBoolPointer()
	}
	if !plan.MirrorInterval.IsNull() {
		editOpts.MirrorInterval = plan.MirrorInterval.ValueStringPointer()
	}
	if !plan.ProjectsMode.IsNull() {
		projectsMode := gitea.ProjectsMode(plan.ProjectsMode.ValueString())
		editOpts.ProjectsMode = &projectsMode
	}

	// Internal tracker
	if !plan.InternalTracker.IsNull() {
		var internalTracker internalTrackerModel
		plan.InternalTracker.As(ctx, &internalTracker, types.ObjectAsOptions{})
		editOpts.InternalTracker = &gitea.InternalTracker{
			EnableTimeTracker:                internalTracker.EnableTimeTracker.ValueBool(),
			AllowOnlyContributorsToTrackTime: internalTracker.AllowOnlyContributorsToTrackTime.ValueBool(),
			EnableIssueDependencies:          internalTracker.EnableIssueDependencies.ValueBool(),
		}
	}

	// External tracker
	if !plan.ExternalTracker.IsNull() {
		var externalTracker externalTrackerModel
		plan.ExternalTracker.As(ctx, &externalTracker, types.ObjectAsOptions{})
		editOpts.ExternalTracker = &gitea.ExternalTracker{
			ExternalTrackerURL:    externalTracker.ExternalTrackerURL.ValueString(),
			ExternalTrackerFormat: externalTracker.ExternalTrackerFormat.ValueString(),
			ExternalTrackerStyle:  externalTracker.ExternalTrackerStyle.ValueString(),
		}
	}

	// External wiki
	if !plan.ExternalWiki.IsNull() {
		var externalWiki externalWikiModel
		plan.ExternalWiki.As(ctx, &externalWiki, types.ObjectAsOptions{})
		editOpts.ExternalWiki = &gitea.ExternalWiki{
			ExternalWikiURL: externalWiki.ExternalWikiURL.ValueString(),
		}
	}

	return editOpts
}

func (r *repositoryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state repositoryResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	owner := state.Owner.ValueString()
	repoName := state.Name.ValueString()

	repo, httpResp, err := r.client.GetRepo(owner, repoName)
	if err != nil {
		// Handle 404 gracefully - remove from state
		if httpResp != nil && httpResp.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Repository",
			"Could not read repository "+owner+"/"+repoName+": "+err.Error(),
		)
		return
	}

	// Preserve creation-only fields from state before mapping
	autoInit := state.AutoInit
	gitignores := state.Gitignores
	issueLabels := state.IssueLabels
	license := state.License
	readme := state.Readme
	trustModel := state.TrustModel
	allowManualMerge := state.AllowManualMerge
	autodetectManualMerge := state.AutodetectManualMerge

	// Map response to state
	mapRepositoryToModel(ctx, repo, &state)

	// Restore owner from state since it's not in the API response
	state.Owner = types.StringValue(owner)

	// Restore creation-only fields
	state.AutoInit = autoInit
	state.Gitignores = gitignores
	state.IssueLabels = issueLabels
	state.License = license
	state.Readme = readme
	state.TrustModel = trustModel
	state.AllowManualMerge = allowManualMerge
	state.AutodetectManualMerge = autodetectManualMerge

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *repositoryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan repositoryResourceModel
	var state repositoryResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	owner := state.Owner.ValueString()
	repoName := state.Name.ValueString()

	editOpts := buildEditRepoOption(ctx, &plan)

	repo, _, err := r.client.EditRepo(owner, repoName, editOpts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Repository",
			"Could not update repository "+owner+"/"+repoName+": "+err.Error(),
		)
		return
	}

	// Preserve creation-only fields from state
	autoInit := state.AutoInit
	gitignores := state.Gitignores
	issueLabels := state.IssueLabels
	license := state.License
	readme := state.Readme
	trustModel := state.TrustModel

	// Map response to state
	mapRepositoryToModel(ctx, repo, &plan)
	plan.Owner = types.StringValue(owner)

	// Restore creation-only fields
	plan.AutoInit = autoInit
	plan.Gitignores = gitignores
	plan.IssueLabels = issueLabels
	plan.License = license
	plan.Readme = readme
	plan.TrustModel = trustModel

	// Preserve AllowManualMerge and AutodetectManualMerge from plan since they're not returned by API
	// (they were set by the user in the plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *repositoryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state repositoryResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	owner := state.Owner.ValueString()
	repoName := state.Name.ValueString()

	_, err := r.client.DeleteRepo(owner, repoName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Repository",
			"Could not delete repository "+owner+"/"+repoName+": "+err.Error(),
		)
		return
	}
}

func (r *repositoryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: "owner/repo"
	id := req.ID

	// Parse owner/repo - handle potential slashes in repo name by only splitting on first slash
	var owner, repoName string
	slashIndex := strings.Index(id, "/")
	if slashIndex == -1 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Import ID must be in format 'owner/repo', got: %s", id),
		)
		return
	}

	owner = id[:slashIndex]
	repoName = id[slashIndex+1:]

	if owner == "" || repoName == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Import ID must be in format 'owner/repo', got: %s", id),
		)
		return
	}

	// Fetch the repository
	repository, httpResp, err := r.client.GetRepo(owner, repoName)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			resp.Diagnostics.AddError(
				"Repository Not Found",
				fmt.Sprintf("Repository %s/%s does not exist or is not accessible", owner, repoName),
			)
			return
		}
		resp.Diagnostics.AddError(
			"Error Importing Repository",
			fmt.Sprintf("Could not import repository %s/%s: %s", owner, repoName, err.Error()),
		)
		return
	}

	var data repositoryResourceModel
	mapRepositoryToModel(ctx, repository, &data)
	data.Owner = types.StringValue(owner)

	// Set creation-only fields to null since we don't know what they were originally
	data.AutoInit = types.BoolNull()
	data.Gitignores = types.StringNull()
	data.IssueLabels = types.StringNull()
	data.License = types.StringNull()
	data.Readme = types.StringNull()
	data.TrustModel = types.StringNull()
	data.AllowManualMerge = types.BoolNull()
	data.AutodetectManualMerge = types.BoolNull()

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
