package provider

import (
	"context"
	"fmt"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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
	Owner                         types.String `tfsdk:"owner"`
	AllowFastForwardOnlyMerge     types.Bool   `tfsdk:"allow_fast_forward_only_merge"`
	AllowManualMerge              types.Bool   `tfsdk:"allow_manual_merge"`
	AllowMergeCommits             types.Bool   `tfsdk:"allow_merge_commits"`
	AllowRebase                   types.Bool   `tfsdk:"allow_rebase"`
	AllowRebaseExplicit           types.Bool   `tfsdk:"allow_rebase_explicit"`
	AllowRebaseUpdate             types.Bool   `tfsdk:"allow_rebase_update"`
	AllowSquashMerge              types.Bool   `tfsdk:"allow_squash_merge"`
	Archived                      types.Bool   `tfsdk:"archived"`
	ArchivedAt                    types.String `tfsdk:"archived_at"`
	AutoInit                      types.Bool   `tfsdk:"auto_init"`
	AutodetectManualMerge         types.Bool   `tfsdk:"autodetect_manual_merge"`
	AvatarUrl                     types.String `tfsdk:"avatar_url"`
	CloneUrl                      types.String `tfsdk:"clone_url"`
	CreatedAt                     types.String `tfsdk:"created_at"`
	DefaultAllowMaintainerEdit    types.Bool   `tfsdk:"default_allow_maintainer_edit"`
	DefaultBranch                 types.String `tfsdk:"default_branch"`
	DefaultDeleteBranchAfterMerge types.Bool   `tfsdk:"default_delete_branch_after_merge"`
	DefaultMergeStyle             types.String `tfsdk:"default_merge_style"`
	Description                   types.String `tfsdk:"description"`
	Empty                         types.Bool   `tfsdk:"empty"`
	Fork                          types.Bool   `tfsdk:"fork"`
	ForksCount                    types.Int64  `tfsdk:"forks_count"`
	Gitignores                    types.String `tfsdk:"gitignores"`
	HasActions                    types.Bool   `tfsdk:"has_actions"`
	HasCode                       types.Bool   `tfsdk:"has_code"`
	HasIssues                     types.Bool   `tfsdk:"has_issues"`
	HasPackages                   types.Bool   `tfsdk:"has_packages"`
	HasProjects                   types.Bool   `tfsdk:"has_projects"`
	HasPullRequests               types.Bool   `tfsdk:"has_pull_requests"`
	HasReleases                   types.Bool   `tfsdk:"has_releases"`
	HasWiki                       types.Bool   `tfsdk:"has_wiki"`
	HtmlUrl                       types.String `tfsdk:"html_url"`
	Id                            types.Int64  `tfsdk:"id"`
	IgnoreWhitespaceConflicts     types.Bool   `tfsdk:"ignore_whitespace_conflicts"`
	Internal                      types.Bool   `tfsdk:"internal"`
	IssueLabels                   types.String `tfsdk:"issue_labels"`
	Language                      types.String `tfsdk:"language"`
	LanguagesUrl                  types.String `tfsdk:"languages_url"`
	License                       types.String `tfsdk:"license"`
	Licenses                      types.List   `tfsdk:"licenses"`
	Link                          types.String `tfsdk:"link"`
	Name                          types.String `tfsdk:"name"`
	OpenIssuesCount               types.Int64  `tfsdk:"open_issues_count"`
	OpenPrCounter                 types.Int64  `tfsdk:"open_pr_counter"`
	Private                       types.Bool   `tfsdk:"private"`
	ProjectsMode                  types.String `tfsdk:"projects_mode"`
	Readme                        types.String `tfsdk:"readme"`
	ReleaseCounter                types.Int64  `tfsdk:"release_counter"`
	Size                          types.Int64  `tfsdk:"size"`
	SshUrl                        types.String `tfsdk:"ssh_url"`
	StarsCount                    types.Int64  `tfsdk:"stars_count"`
	Template                      types.Bool   `tfsdk:"template"`
	Topics                        types.List   `tfsdk:"topics"`
	UpdatedAt                     types.String `tfsdk:"updated_at"`
	Url                           types.String `tfsdk:"url"`
	WatchersCount                 types.Int64  `tfsdk:"watchers_count"`
	Website                       types.String `tfsdk:"website"`
}

func (r *repositoryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repository"
}

func (r *repositoryResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{

			// required - these are fundamental configuration options
			"owner": schema.StringAttribute{
				Required:            true,
				Description:         "The owner of the repository (username or organization name)",
				MarkdownDescription: "The owner of the repository (username or organization name)",
			},
			"name": schema.StringAttribute{
				Required:            true,
				Description:         "Name of the repository to create",
				MarkdownDescription: "Name of the repository to create",
			},

			// optional - these tweak the created resource away from its defaults
			"auto_init": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether the repository should be auto-initialized?",
				MarkdownDescription: "Whether the repository should be auto-initialized?",
			},
			"default_branch": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "DefaultBranch of the repository (used when initializes and in template)",
				MarkdownDescription: "DefaultBranch of the repository (used when initializes and in template)",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Description of the repository to create",
				MarkdownDescription: "Description of the repository to create",
			},
			"gitignores": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Gitignores to use",
				MarkdownDescription: "Gitignores to use",
			},
			"issue_labels": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Label-Set to use",
				MarkdownDescription: "Label-Set to use",
			},
			"license": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "License to use",
				MarkdownDescription: "License to use",
			},
			"readme": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Readme of the repository to create",
				MarkdownDescription: "Readme of the repository to create",
			},
			"private": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether the repository is private",
				MarkdownDescription: "Whether the repository is private",
			},
			"template": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether the repository is template",
				MarkdownDescription: "Whether the repository is template",
			},

			// computed - these are available to read back after creation but are really just metadata
			"allow_fast_forward_only_merge": schema.BoolAttribute{
				Computed: true,
			},
			"allow_manual_merge": schema.BoolAttribute{
				Computed: true,
			},
			"allow_merge_commits": schema.BoolAttribute{
				Computed: true,
			},
			"allow_rebase": schema.BoolAttribute{
				Computed: true,
			},
			"allow_rebase_explicit": schema.BoolAttribute{
				Computed: true,
			},
			"allow_rebase_update": schema.BoolAttribute{
				Computed: true,
			},
			"allow_squash_merge": schema.BoolAttribute{
				Computed: true,
			},
			"archived": schema.BoolAttribute{
				Computed: true,
			},
			"archived_at": schema.StringAttribute{
				Computed: true,
			},
			"autodetect_manual_merge": schema.BoolAttribute{
				Computed: true,
			},
			"avatar_url": schema.StringAttribute{
				Computed: true,
			},
			"clone_url": schema.StringAttribute{
				Computed: true,
			},
			"created_at": schema.StringAttribute{
				Computed: true,
			},
			"default_allow_maintainer_edit": schema.BoolAttribute{
				Computed: true,
			},
			"default_delete_branch_after_merge": schema.BoolAttribute{
				Computed: true,
			},
			"default_merge_style": schema.StringAttribute{
				Computed: true,
			},
			"empty": schema.BoolAttribute{
				Computed: true,
			},
			"fork": schema.BoolAttribute{
				Computed: true,
			},
			"forks_count": schema.Int64Attribute{
				Computed: true,
			},
			"has_actions": schema.BoolAttribute{
				Computed: true,
			},
			"has_code": schema.BoolAttribute{
				Computed: true,
			},
			"has_issues": schema.BoolAttribute{
				Computed: true,
			},
			"has_packages": schema.BoolAttribute{
				Computed: true,
			},
			"has_projects": schema.BoolAttribute{
				Computed: true,
			},
			"has_pull_requests": schema.BoolAttribute{
				Computed: true,
			},
			"has_releases": schema.BoolAttribute{
				Computed: true,
			},
			"has_wiki": schema.BoolAttribute{
				Computed: true,
			},
			"html_url": schema.StringAttribute{
				Computed: true,
			},
			"id": schema.Int64Attribute{
				Computed: true,
			},
			"ignore_whitespace_conflicts": schema.BoolAttribute{
				Computed: true,
			},
			"internal": schema.BoolAttribute{
				Computed: true,
			},
			"language": schema.StringAttribute{
				Computed: true,
			},
			"languages_url": schema.StringAttribute{
				Computed: true,
			},
			"licenses": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"link": schema.StringAttribute{
				Computed: true,
			},
			"open_issues_count": schema.Int64Attribute{
				Computed: true,
			},
			"open_pr_counter": schema.Int64Attribute{
				Computed: true,
			},
			"projects_mode": schema.StringAttribute{
				Computed: true,
			},
			"release_counter": schema.Int64Attribute{
				Computed: true,
			},
			"size": schema.Int64Attribute{
				Computed: true,
			},
			"ssh_url": schema.StringAttribute{
				Computed: true,
			},
			"stars_count": schema.Int64Attribute{
				Computed: true,
			},
			"topics": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Computed: true,
			},
			"url": schema.StringAttribute{
				Computed: true,
			},
			"watchers_count": schema.Int64Attribute{
				Computed: true,
			},
			"website": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

// Helper function to map Gitea Repository to Terraform model
func mapRepositoryToModel(repo *gitea.Repository, model *repositoryResourceModel) {
	// Basic fields
	model.Id = types.Int64Value(repo.ID)
	model.Name = types.StringValue(repo.Name)
	model.Description = types.StringValue(repo.Description)
	model.Private = types.BoolValue(repo.Private)
	model.DefaultBranch = types.StringValue(repo.DefaultBranch)
	model.Website = types.StringValue(repo.Website)
	model.HtmlUrl = types.StringValue(repo.HTMLURL)
	model.CloneUrl = types.StringValue(repo.CloneURL)
	model.SshUrl = types.StringValue(repo.SSHURL)
	model.Empty = types.BoolValue(repo.Empty)
	model.Fork = types.BoolValue(repo.Fork)
	model.Size = types.Int64Value(int64(repo.Size))
	model.Archived = types.BoolValue(repo.Archived)
	model.StarsCount = types.Int64Value(int64(repo.Stars))
	model.WatchersCount = types.Int64Value(int64(repo.Watchers))
	model.ForksCount = types.Int64Value(int64(repo.Forks))
	model.OpenIssuesCount = types.Int64Value(int64(repo.OpenIssues))
	model.AvatarUrl = types.StringValue(repo.AvatarURL)
	model.Template = types.BoolValue(repo.Template)
	model.Internal = types.BoolValue(repo.Internal)

	// Merge settings
	model.AllowFastForwardOnlyMerge = types.BoolValue(repo.AllowFastForwardOnlyMerge)
	model.AllowMergeCommits = types.BoolValue(repo.AllowMerge)
	model.AllowRebase = types.BoolValue(repo.AllowRebase)
	model.AllowRebaseExplicit = types.BoolValue(repo.AllowRebaseMerge)
	model.AllowSquashMerge = types.BoolValue(repo.AllowSquash)
	model.DefaultDeleteBranchAfterMerge = types.BoolValue(repo.DefaultDeleteBranchAfterMerge)
	model.DefaultMergeStyle = types.StringValue(string(repo.DefaultMergeStyle))
	model.IgnoreWhitespaceConflicts = types.BoolValue(repo.IgnoreWhitespaceConflicts)

	// Feature flags
	model.HasIssues = types.BoolValue(repo.HasIssues)
	model.HasWiki = types.BoolValue(repo.HasWiki)
	model.HasPullRequests = types.BoolValue(repo.HasPullRequests)
	model.HasProjects = types.BoolValue(repo.HasProjects)
	model.HasReleases = types.BoolValue(repo.HasReleases)
	model.HasPackages = types.BoolValue(repo.HasPackages)
	model.HasActions = types.BoolValue(repo.HasActions)

	// Timestamps
	model.CreatedAt = types.StringValue(repo.Created.Format("2006-01-02T15:04:05Z"))
	model.UpdatedAt = types.StringValue(repo.Updated.Format("2006-01-02T15:04:05Z"))

	// Counters
	model.OpenPrCounter = types.Int64Value(int64(repo.OpenPulls))
	model.ReleaseCounter = types.Int64Value(int64(repo.Releases))

	// Projects mode
	if repo.ProjectsMode != nil {
		model.ProjectsMode = types.StringValue(string(*repo.ProjectsMode))
	} else {
		model.ProjectsMode = types.StringNull()
	}

	// Fields not available from API responses (only used during creation)
	// Preserve these values from the existing model (they come from plan during Create, from state during Read)
	// If they're still Unknown, set them to null
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

	// Fields not available in SDK
	model.AllowManualMerge = types.BoolNull()
	model.AllowRebaseUpdate = types.BoolNull()
	model.ArchivedAt = types.StringNull()
	model.AutodetectManualMerge = types.BoolNull()
	model.DefaultAllowMaintainerEdit = types.BoolNull()
	model.HasCode = types.BoolNull()
	model.Language = types.StringNull()
	model.LanguagesUrl = types.StringNull()
	model.Licenses = types.ListNull(types.StringType)
	model.Link = types.StringNull()
	model.Topics = types.ListNull(types.StringType)
	model.Url = types.StringNull()
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
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Private:     plan.Private.ValueBool(),
	}

	if !plan.DefaultBranch.IsNull() {
		createOpts.DefaultBranch = plan.DefaultBranch.ValueString()
	}
	if !plan.AutoInit.IsNull() {
		createOpts.AutoInit = plan.AutoInit.ValueBool()
	}
	if !plan.Gitignores.IsNull() {
		createOpts.Gitignores = plan.Gitignores.ValueString()
	}
	if !plan.IssueLabels.IsNull() {
		createOpts.IssueLabels = plan.IssueLabels.ValueString()
	}
	if !plan.License.IsNull() {
		createOpts.License = plan.License.ValueString()
	}
	if !plan.Readme.IsNull() {
		createOpts.Readme = plan.Readme.ValueString()
	}
	if !plan.Template.IsNull() {
		createOpts.Template = plan.Template.ValueBool()
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
		// Owner is a user (or we'll let it fail with proper error)
		repo, _, err = r.client.CreateRepo(createOpts)
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating repository",
			"Could not create repository, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response to state
	mapRepositoryToModel(repo, &plan)
	plan.Owner = types.StringValue(owner)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *repositoryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state repositoryResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	owner := state.Owner.ValueString()
	repoName := state.Name.ValueString()

	repo, _, err := r.client.GetRepo(owner, repoName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Repository",
			"Could not read repository "+owner+"/"+repoName+": "+err.Error(),
		)
		return
	}

	// Map response to state
	mapRepositoryToModel(repo, &state)
	// Preserve owner from state since it's not in the API response
	state.Owner = types.StringValue(owner)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *repositoryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan repositoryResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state repositoryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	owner := state.Owner.ValueString()
	repoName := state.Name.ValueString()

	editOpts := gitea.EditRepoOption{
		Description: plan.Description.ValueStringPointer(),
		Private:     plan.Private.ValueBoolPointer(),
	}

	if !plan.DefaultBranch.IsNull() {
		editOpts.DefaultBranch = plan.DefaultBranch.ValueStringPointer()
	}

	repo, _, err := r.client.EditRepo(owner, repoName, editOpts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Repository",
			"Could not update repository, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response to state
	mapRepositoryToModel(repo, &plan)
	plan.Owner = types.StringValue(owner)

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
			"Could not delete repository, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *repositoryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: "owner/repo"
	id := req.ID

	// Parse owner/repo
	var owner, repo string
	for i, c := range id {
		if c == '/' {
			owner = id[:i]
			repo = id[i+1:]
			break
		}
	}

	if owner == "" || repo == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Import ID must be in format 'owner/repo', got: %s", id),
		)
		return
	}

	// Fetch the repository
	repository, _, err := r.client.GetRepo(owner, repo)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Importing Repository",
			fmt.Sprintf("Could not import repository %s/%s: %s", owner, repo, err.Error()),
		)
		return
	}

	var data repositoryResourceModel
	mapRepositoryToModel(repository, &data)
	data.Owner = types.StringValue(owner)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
