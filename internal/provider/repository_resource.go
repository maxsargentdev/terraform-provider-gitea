package provider

import (
	"context"
	"fmt"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
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

type repositoryResourceModel struct {
	// Required
	Username types.String `tfsdk:"username"`
	Name     types.String `tfsdk:"name"`

	// Optional - Basic settings
	Description   types.String `tfsdk:"description"`
	Private       types.Bool   `tfsdk:"private"`
	DefaultBranch types.String `tfsdk:"default_branch"`
	Website       types.String `tfsdk:"website"`
	Archived      types.Bool   `tfsdk:"archived"`

	// Optional - Creation only settings
	AutoInit    types.Bool   `tfsdk:"auto_init"`
	Gitignores  types.String `tfsdk:"gitignores"`
	IssueLabels types.String `tfsdk:"issue_labels"`
	License     types.String `tfsdk:"license"`
	Readme      types.String `tfsdk:"readme"`
	RepoTemplate types.Bool   `tfsdk:"repo_template"`

	// Optional - Feature flags
	HasIssues       types.Bool `tfsdk:"has_issues"`
	HasWiki         types.Bool `tfsdk:"has_wiki"`
	HasPullRequests types.Bool `tfsdk:"has_pull_requests"`
	HasProjects     types.Bool `tfsdk:"has_projects"`

	// Optional - Merge settings
	AllowMergeCommits         types.Bool `tfsdk:"allow_merge_commits"`
	AllowRebase               types.Bool `tfsdk:"allow_rebase"`
	AllowRebaseExplicit       types.Bool `tfsdk:"allow_rebase_explicit"`
	AllowSquashMerge          types.Bool `tfsdk:"allow_squash_merge"`
	AllowManualMerge          types.Bool `tfsdk:"allow_manual_merge"`
	AutodetectManualMerge     types.Bool `tfsdk:"autodetect_manual_merge"`
	IgnoreWhitespaceConflicts types.Bool `tfsdk:"ignore_whitespace_conflicts"`

	// Optional - Migration settings
	MigrationCloneAddress         types.String `tfsdk:"migration_clone_address"`
	MigrationCloneAddresse        types.String `tfsdk:"migration_clone_addresse"` // Deprecated
	MigrationService              types.String `tfsdk:"migration_service"`
	MigrationServiceAuthUsername  types.String `tfsdk:"migration_service_auth_username"`
	MigrationServiceAuthPassword  types.String `tfsdk:"migration_service_auth_password"`
	MigrationServiceAuthToken     types.String `tfsdk:"migration_service_auth_token"`
	MigrationIssueLabels          types.Bool   `tfsdk:"migration_issue_labels"`
	MigrationLfs                  types.Bool   `tfsdk:"migration_lfs"`
	MigrationLfsEndpoint          types.String `tfsdk:"migration_lfs_endpoint"`
	MigrationMilestones           types.Bool   `tfsdk:"migration_milestones"`
	MigrationMirrorInterval       types.String `tfsdk:"migration_mirror_interval"`
	MigrationReleases             types.Bool   `tfsdk:"migration_releases"`
	Mirror                        types.Bool   `tfsdk:"mirror"`

	// Optional - Destroy behavior
	ArchiveOnDestroy types.Bool `tfsdk:"archive_on_destroy"`

	// Computed
	Id              types.String `tfsdk:"id"`
	CloneUrl        types.String `tfsdk:"clone_url"`
	Created         types.String `tfsdk:"created"`
	HtmlUrl         types.String `tfsdk:"html_url"`
	PermissionAdmin types.Bool   `tfsdk:"permission_admin"`
	PermissionPull  types.Bool   `tfsdk:"permission_pull"`
	PermissionPush  types.Bool   `tfsdk:"permission_push"`
	SshUrl          types.String `tfsdk:"ssh_url"`
	Updated         types.String `tfsdk:"updated"`
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
			"username": schema.StringAttribute{
				Required:            true,
				Description:         "The owner of the repository.",
				MarkdownDescription: "The owner of the repository.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				Description:         "The name of the repository.",
				MarkdownDescription: "The name of the repository.",
			},

			// ==================== OPTIONAL - Basic ====================
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
			"default_branch": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("main"),
				Description:         "Default branch of the repository.",
				MarkdownDescription: "Default branch of the repository.",
			},
			"website": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "A URL with more information about the repository.",
				MarkdownDescription: "A URL with more information about the repository.",
			},
			"archived": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether the repository is archived.",
				MarkdownDescription: "Whether the repository is archived.",
			},

			// ==================== OPTIONAL - Creation Only ====================
			"auto_init": schema.BoolAttribute{
				Optional:            true,
				Description:         "Flag if the repository should be initiated with the configured values.",
				MarkdownDescription: "Flag if the repository should be initiated with the configured values.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"gitignores": schema.StringAttribute{
				Optional:            true,
				Description:         "A specific gitignore that should be committed to the repository on creation if `auto_init` is set to `true`.",
				MarkdownDescription: "A specific gitignore that should be committed to the repository on creation if `auto_init` is set to `true`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"issue_labels": schema.StringAttribute{
				Optional:            true,
				Description:         "Issue label set to use when initializing the repository.",
				MarkdownDescription: "Issue label set to use when initializing the repository.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"license": schema.StringAttribute{
				Optional:            true,
				Description:         "License to use when initializing the repository.",
				MarkdownDescription: "License to use when initializing the repository.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"readme": schema.StringAttribute{
				Optional:            true,
				Description:         "Readme template to use when initializing the repository.",
				MarkdownDescription: "Readme template to use when initializing the repository.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"repo_template": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether the repository is a template repository.",
				MarkdownDescription: "Whether the repository is a template repository.",
			},

			// ==================== OPTIONAL - Feature Flags ====================
			"has_issues": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether the repository has issues enabled.",
				MarkdownDescription: "Whether the repository has issues enabled.",
			},
			"has_wiki": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether the repository has wiki enabled.",
				MarkdownDescription: "Whether the repository has wiki enabled.",
			},
			"has_pull_requests": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether the repository allows pull requests.",
				MarkdownDescription: "Whether the repository allows pull requests.",
			},
			"has_projects": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether the repository has projects enabled.",
				MarkdownDescription: "Whether the repository has projects enabled.",
			},

			// ==================== OPTIONAL - Merge Settings ====================
			"allow_merge_commits": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether to allow merge commits.",
				MarkdownDescription: "Whether to allow merge commits.",
			},
			"allow_rebase": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether to allow rebase merges.",
				MarkdownDescription: "Whether to allow rebase merges.",
			},
			"allow_rebase_explicit": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether to allow explicit rebase merges.",
				MarkdownDescription: "Whether to allow explicit rebase merges.",
			},
			"allow_squash_merge": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether to allow squash merges.",
				MarkdownDescription: "Whether to allow squash merges.",
			},
			"allow_manual_merge": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether to allow manual merge.",
				MarkdownDescription: "Whether to allow manual merge.",
			},
			"autodetect_manual_merge": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether to autodetect manual merge.",
				MarkdownDescription: "Whether to autodetect manual merge.",
			},
			"ignore_whitespace_conflicts": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether to ignore whitespace conflicts.",
				MarkdownDescription: "Whether to ignore whitespace conflicts.",
			},

			// ==================== OPTIONAL - Migration Settings ====================
			"migration_clone_address": schema.StringAttribute{
				Optional:            true,
				Description:         "The URL to clone the repository from during migration.",
				MarkdownDescription: "The URL to clone the repository from during migration.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"migration_clone_addresse": schema.StringAttribute{
				Optional:            true,
				Description:         "Deprecated: use migration_clone_address instead.",
				MarkdownDescription: "**Deprecated:** use `migration_clone_address` instead.",
				DeprecationMessage:  "Use migration_clone_address instead.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"migration_service": schema.StringAttribute{
				Optional:            true,
				Description:         "The service type for migration (git, github, gitlab, gitea, gogs).",
				MarkdownDescription: "The service type for migration (`git`, `github`, `gitlab`, `gitea`, `gogs`).",
				Validators: []validator.String{
					stringvalidator.OneOf("git", "github", "gitlab", "gitea", "gogs"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"migration_service_auth_username": schema.StringAttribute{
				Optional:            true,
				Description:         "The username for authentication during migration.",
				MarkdownDescription: "The username for authentication during migration.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"migration_service_auth_password": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				Description:         "The password for authentication during migration.",
				MarkdownDescription: "The password for authentication during migration.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"migration_service_auth_token": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				Description:         "The token for authentication during migration.",
				MarkdownDescription: "The token for authentication during migration.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"migration_issue_labels": schema.BoolAttribute{
				Optional:            true,
				Description:         "Whether to migrate issue labels.",
				MarkdownDescription: "Whether to migrate issue labels.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"migration_lfs": schema.BoolAttribute{
				Optional:            true,
				Description:         "Whether to migrate LFS objects.",
				MarkdownDescription: "Whether to migrate LFS objects.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"migration_lfs_endpoint": schema.StringAttribute{
				Optional:            true,
				Description:         "The LFS endpoint URL for migration.",
				MarkdownDescription: "The LFS endpoint URL for migration.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"migration_milestones": schema.BoolAttribute{
				Optional:            true,
				Description:         "Whether to migrate milestones.",
				MarkdownDescription: "Whether to migrate milestones.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"migration_mirror_interval": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The mirror interval for auto-sync (e.g., '8h0m0s'). Set to 0 to disable.",
				MarkdownDescription: "The mirror interval for auto-sync (e.g., `8h0m0s`). Set to `0` to disable.",
			},
			"migration_releases": schema.BoolAttribute{
				Optional:            true,
				Description:         "Whether to migrate releases.",
				MarkdownDescription: "Whether to migrate releases.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"mirror": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether the repository is a mirror.",
				MarkdownDescription: "Whether the repository is a mirror.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},

			// ==================== OPTIONAL - Destroy Behavior ====================
			"archive_on_destroy": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				Description:         "Set to `true` to archive the repository instead of deleting on destroy.",
				MarkdownDescription: "Set to `true` to archive the repository instead of deleting on destroy.",
			},

			// ==================== COMPUTED ====================
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "The ID of the repository.",
				MarkdownDescription: "The ID of the repository.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"clone_url": schema.StringAttribute{
				Computed:            true,
				Description:         "The HTTPS clone URL of the repository.",
				MarkdownDescription: "The HTTPS clone URL of the repository.",
			},
			"created": schema.StringAttribute{
				Computed:            true,
				Description:         "Timestamp when the repository was created.",
				MarkdownDescription: "Timestamp when the repository was created.",
			},
			"html_url": schema.StringAttribute{
				Computed:            true,
				Description:         "The URL to view the repository in the web UI.",
				MarkdownDescription: "The URL to view the repository in the web UI.",
			},
			"permission_admin": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether the current user has admin permission.",
				MarkdownDescription: "Whether the current user has admin permission.",
			},
			"permission_pull": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether the current user has pull permission.",
				MarkdownDescription: "Whether the current user has pull permission.",
			},
			"permission_push": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether the current user has push permission.",
				MarkdownDescription: "Whether the current user has push permission.",
			},
			"ssh_url": schema.StringAttribute{
				Computed:            true,
				Description:         "The SSH clone URL of the repository.",
				MarkdownDescription: "The SSH clone URL of the repository.",
			},
			"updated": schema.StringAttribute{
				Computed:            true,
				Description:         "Timestamp when the repository was last updated.",
				MarkdownDescription: "Timestamp when the repository was last updated.",
			},
		},
	}
}

// Helper function to map Gitea Repository to Terraform model
func mapRepositoryToModel(ctx context.Context, repo *gitea.Repository, model *repositoryResourceModel) {
	// Computed fields
	model.Id = types.StringValue(fmt.Sprintf("%d", repo.ID))
	model.HtmlUrl = types.StringValue(repo.HTMLURL)
	model.CloneUrl = types.StringValue(repo.CloneURL)
	model.SshUrl = types.StringValue(repo.SSHURL)
	model.Created = types.StringValue(repo.Created.String())
	model.Updated = types.StringValue(repo.Updated.String())

	// Permission fields
	if repo.Permissions != nil {
		model.PermissionAdmin = types.BoolValue(repo.Permissions.Admin)
		model.PermissionPull = types.BoolValue(repo.Permissions.Pull)
		model.PermissionPush = types.BoolValue(repo.Permissions.Push)
	} else {
		model.PermissionAdmin = types.BoolValue(false)
		model.PermissionPull = types.BoolValue(false)
		model.PermissionPush = types.BoolValue(false)
	}

	// Basic fields that can be updated
	model.Name = types.StringValue(repo.Name)
	model.Description = types.StringValue(repo.Description)
	model.Private = types.BoolValue(repo.Private)
	model.DefaultBranch = types.StringValue(repo.DefaultBranch)
	model.Website = types.StringValue(repo.Website)
	model.Archived = types.BoolValue(repo.Archived)
	model.RepoTemplate = types.BoolValue(repo.Template)
	model.Mirror = types.BoolValue(repo.Mirror)
	model.MigrationMirrorInterval = types.StringValue(repo.MirrorInterval)

	// Feature flags
	model.HasIssues = types.BoolValue(repo.HasIssues)
	model.HasWiki = types.BoolValue(repo.HasWiki)
	model.HasPullRequests = types.BoolValue(repo.HasPullRequests)
	model.HasProjects = types.BoolValue(repo.HasProjects)

	// Pull request merge settings
	model.IgnoreWhitespaceConflicts = types.BoolValue(repo.IgnoreWhitespaceConflicts)
	model.AllowMergeCommits = types.BoolValue(repo.AllowMerge)
	model.AllowRebase = types.BoolValue(repo.AllowRebase)
	model.AllowRebaseExplicit = types.BoolValue(repo.AllowRebaseMerge)
	model.AllowSquashMerge = types.BoolValue(repo.AllowSquash)

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

	username := plan.Username.ValueString()

	// Check if this is a migration
	migrationAddress := plan.MigrationCloneAddress.ValueString()
	if migrationAddress == "" {
		// Check deprecated field
		migrationAddress = plan.MigrationCloneAddresse.ValueString()
	}

	var repo *gitea.Repository
	var err error

	if migrationAddress != "" {
		// Create via migration
		migrateOpts := gitea.MigrateRepoOption{
			CloneAddr:    migrationAddress,
			RepoName:     plan.Name.ValueString(),
			RepoOwner:    username,
			Service:      gitea.GitServiceType(plan.MigrationService.ValueString()),
			AuthUsername: plan.MigrationServiceAuthUsername.ValueString(),
			AuthPassword: plan.MigrationServiceAuthPassword.ValueString(),
			AuthToken:    plan.MigrationServiceAuthToken.ValueString(),
			Mirror:       plan.Mirror.ValueBool(),
			Private:      plan.Private.ValueBool(),
			Description:  plan.Description.ValueString(),
			Labels:       plan.MigrationIssueLabels.ValueBool(),
			LFS:          plan.MigrationLfs.ValueBool(),
			LFSEndpoint:  plan.MigrationLfsEndpoint.ValueString(),
			Milestones:   plan.MigrationMilestones.ValueBool(),
			Releases:     plan.MigrationReleases.ValueBool(),
		}
		if plan.MigrationMirrorInterval.ValueString() != "" {
			migrateOpts.MirrorInterval = plan.MigrationMirrorInterval.ValueString()
		}

		repo, _, err = r.client.MigrateRepo(migrateOpts)
	} else {
		// Create repository
		createOpts := gitea.CreateRepoOption{
			Name:          plan.Name.ValueString(),
			Description:   plan.Description.ValueString(),
			Private:       plan.Private.ValueBool(),
			IssueLabels:   plan.IssueLabels.ValueString(),
			AutoInit:      plan.AutoInit.ValueBool(),
			Template:      plan.RepoTemplate.ValueBool(),
			Gitignores:    plan.Gitignores.ValueString(),
			License:       plan.License.ValueString(),
			Readme:        plan.Readme.ValueString(),
			DefaultBranch: plan.DefaultBranch.ValueString(),
		}

		// Check if owner is an org or user and use appropriate API
		org, _, orgErr := r.client.GetOrg(username)
		if orgErr == nil && org != nil {
			// Owner is an organization
			repo, _, err = r.client.CreateOrgRepo(username, createOpts)
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

			if currentUser.UserName == username {
				// Owner is the current user
				repo, _, err = r.client.CreateRepo(createOpts)
			} else {
				// Try to create for another user (admin only)
				repo, _, err = r.client.AdminCreateRepo(username, createOpts)
			}
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
	plan.Username = types.StringValue(username)

	// If additional edit-only settings were specified, apply them now
	if needsPostCreateUpdate(&plan) {
		editOpts := buildEditRepoOption(ctx, &plan)
		repo, _, err = r.client.EditRepo(username, repo.Name, editOpts)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Updating Repository After Creation",
				"Repository was created but failed to apply additional settings: "+err.Error(),
			)
			return
		}
		mapRepositoryToModel(ctx, repo, &plan)
		plan.Username = types.StringValue(username)
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
		!plan.IgnoreWhitespaceConflicts.IsNull() ||
		!plan.AllowMergeCommits.IsNull() ||
		!plan.AllowRebase.IsNull() ||
		!plan.AllowRebaseExplicit.IsNull() ||
		!plan.AllowSquashMerge.IsNull() ||
		!plan.Archived.IsNull() ||
		!plan.AllowManualMerge.IsNull() ||
		!plan.AutodetectManualMerge.IsNull()
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
	if !plan.RepoTemplate.IsNull() {
		editOpts.Template = plan.RepoTemplate.ValueBoolPointer()
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
	if !plan.MigrationMirrorInterval.IsNull() {
		editOpts.MirrorInterval = plan.MigrationMirrorInterval.ValueStringPointer()
	}

	return editOpts
}

func (r *repositoryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state repositoryResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	username := state.Username.ValueString()
	repoName := state.Name.ValueString()

	repo, httpResp, err := r.client.GetRepo(username, repoName)
	if err != nil {
		// Handle 404 gracefully - remove from state
		if httpResp != nil && httpResp.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Repository",
			"Could not read repository "+username+"/"+repoName+": "+err.Error(),
		)
		return
	}

	// Preserve creation-only fields from state before mapping
	autoInit := state.AutoInit
	gitignores := state.Gitignores
	issueLabels := state.IssueLabels
	license := state.License
	readme := state.Readme
	allowManualMerge := state.AllowManualMerge
	autodetectManualMerge := state.AutodetectManualMerge
	archiveOnDestroy := state.ArchiveOnDestroy

	// Preserve migration fields
	migrationCloneAddress := state.MigrationCloneAddress
	migrationCloneAddresse := state.MigrationCloneAddresse
	migrationService := state.MigrationService
	migrationServiceAuthUsername := state.MigrationServiceAuthUsername
	migrationServiceAuthPassword := state.MigrationServiceAuthPassword
	migrationServiceAuthToken := state.MigrationServiceAuthToken
	migrationIssueLabels := state.MigrationIssueLabels
	migrationLfs := state.MigrationLfs
	migrationLfsEndpoint := state.MigrationLfsEndpoint
	migrationMilestones := state.MigrationMilestones
	migrationReleases := state.MigrationReleases

	// Map response to state
	mapRepositoryToModel(ctx, repo, &state)

	// Restore owner from state since it's not in the API response
	state.Username = types.StringValue(username)

	// Restore creation-only fields
	state.AutoInit = autoInit
	state.Gitignores = gitignores
	state.IssueLabels = issueLabels
	state.License = license
	state.Readme = readme
	state.AllowManualMerge = allowManualMerge
	state.AutodetectManualMerge = autodetectManualMerge
	state.ArchiveOnDestroy = archiveOnDestroy

	// Restore migration fields
	state.MigrationCloneAddress = migrationCloneAddress
	state.MigrationCloneAddresse = migrationCloneAddresse
	state.MigrationService = migrationService
	state.MigrationServiceAuthUsername = migrationServiceAuthUsername
	state.MigrationServiceAuthPassword = migrationServiceAuthPassword
	state.MigrationServiceAuthToken = migrationServiceAuthToken
	state.MigrationIssueLabels = migrationIssueLabels
	state.MigrationLfs = migrationLfs
	state.MigrationLfsEndpoint = migrationLfsEndpoint
	state.MigrationMilestones = migrationMilestones
	state.MigrationReleases = migrationReleases

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

	username := state.Username.ValueString()
	repoName := state.Name.ValueString()

	editOpts := buildEditRepoOption(ctx, &plan)

	repo, _, err := r.client.EditRepo(username, repoName, editOpts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Repository",
			"Could not update repository "+username+"/"+repoName+": "+err.Error(),
		)
		return
	}

	// Preserve creation-only fields from state
	autoInit := state.AutoInit
	gitignores := state.Gitignores
	issueLabels := state.IssueLabels
	license := state.License
	readme := state.Readme

	// Preserve migration fields from state
	migrationCloneAddress := state.MigrationCloneAddress
	migrationCloneAddresse := state.MigrationCloneAddresse
	migrationService := state.MigrationService
	migrationServiceAuthUsername := state.MigrationServiceAuthUsername
	migrationServiceAuthPassword := state.MigrationServiceAuthPassword
	migrationServiceAuthToken := state.MigrationServiceAuthToken
	migrationIssueLabels := state.MigrationIssueLabels
	migrationLfs := state.MigrationLfs
	migrationLfsEndpoint := state.MigrationLfsEndpoint
	migrationMilestones := state.MigrationMilestones
	migrationReleases := state.MigrationReleases

	// Map response to state
	mapRepositoryToModel(ctx, repo, &plan)
	plan.Username = types.StringValue(username)

	// Restore creation-only fields
	plan.AutoInit = autoInit
	plan.Gitignores = gitignores
	plan.IssueLabels = issueLabels
	plan.License = license
	plan.Readme = readme

	// Restore migration fields
	plan.MigrationCloneAddress = migrationCloneAddress
	plan.MigrationCloneAddresse = migrationCloneAddresse
	plan.MigrationService = migrationService
	plan.MigrationServiceAuthUsername = migrationServiceAuthUsername
	plan.MigrationServiceAuthPassword = migrationServiceAuthPassword
	plan.MigrationServiceAuthToken = migrationServiceAuthToken
	plan.MigrationIssueLabels = migrationIssueLabels
	plan.MigrationLfs = migrationLfs
	plan.MigrationLfsEndpoint = migrationLfsEndpoint
	plan.MigrationMilestones = migrationMilestones
	plan.MigrationReleases = migrationReleases

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *repositoryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state repositoryResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	username := state.Username.ValueString()
	repoName := state.Name.ValueString()

	// Check if we should archive instead of delete
	if state.ArchiveOnDestroy.ValueBool() {
		archived := true
		editOpts := gitea.EditRepoOption{
			Archived: &archived,
		}
		_, _, err := r.client.EditRepo(username, repoName, editOpts)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Archiving Repository",
				"Could not archive repository "+username+"/"+repoName+": "+err.Error(),
			)
			return
		}
		return
	}

	_, err := r.client.DeleteRepo(username, repoName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Repository",
			"Could not delete repository "+username+"/"+repoName+": "+err.Error(),
		)
		return
	}
}

func (r *repositoryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: "username/repo"
	id := req.ID

	// Parse owner/repo - handle potential slashes in repo name by only splitting on first slash
	var username, repoName string
	slashIndex := strings.Index(id, "/")
	if slashIndex == -1 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Import ID must be in format 'username/repo', got: %s", id),
		)
		return
	}

	username = id[:slashIndex]
	repoName = id[slashIndex+1:]

	if username == "" || repoName == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Import ID must be in format 'username/repo', got: %s", id),
		)
		return
	}

	// Fetch the repository
	repository, httpResp, err := r.client.GetRepo(username, repoName)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			resp.Diagnostics.AddError(
				"Repository Not Found",
				fmt.Sprintf("Repository %s/%s does not exist or is not accessible", username, repoName),
			)
			return
		}
		resp.Diagnostics.AddError(
			"Error Importing Repository",
			fmt.Sprintf("Could not import repository %s/%s: %s", username, repoName, err.Error()),
		)
		return
	}

	var data repositoryResourceModel
	mapRepositoryToModel(ctx, repository, &data)
	data.Username = types.StringValue(username)

	// Set creation-only fields to null since we don't know what they were originally
	data.AutoInit = types.BoolNull()
	data.Gitignores = types.StringNull()
	data.IssueLabels = types.StringNull()
	data.License = types.StringNull()
	data.Readme = types.StringNull()
	data.AllowManualMerge = types.BoolNull()
	data.AutodetectManualMerge = types.BoolNull()
	data.ArchiveOnDestroy = types.BoolValue(false)

	// Set migration fields to null
	data.MigrationCloneAddress = types.StringNull()
	data.MigrationCloneAddresse = types.StringNull()
	data.MigrationService = types.StringNull()
	data.MigrationServiceAuthUsername = types.StringNull()
	data.MigrationServiceAuthPassword = types.StringNull()
	data.MigrationServiceAuthToken = types.StringNull()
	data.MigrationIssueLabels = types.BoolNull()
	data.MigrationLfs = types.BoolNull()
	data.MigrationLfsEndpoint = types.StringNull()
	data.MigrationMilestones = types.BoolNull()
	data.MigrationReleases = types.BoolNull()

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
