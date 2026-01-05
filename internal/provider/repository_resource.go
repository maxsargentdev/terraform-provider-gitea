package provider

import (
	"context"
	"fmt"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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

func (r *repositoryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repository"
}

func (r *repositoryResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
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
			"auto_init": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether the repository should be auto-initialized?",
				MarkdownDescription: "Whether the repository should be auto-initialized?",
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
			"default_branch": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "DefaultBranch of the repository (used when initializes and in template)",
				MarkdownDescription: "DefaultBranch of the repository (used when initializes and in template)",
			},
			"default_delete_branch_after_merge": schema.BoolAttribute{
				Computed: true,
			},
			"default_merge_style": schema.StringAttribute{
				Computed: true,
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Description of the repository to create",
				MarkdownDescription: "Description of the repository to create",
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
			"full_name": schema.StringAttribute{
				Computed: true,
			},
			"gitignores": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Gitignores to use",
				MarkdownDescription: "Gitignores to use",
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
			"issue_labels": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Label-Set to use",
				MarkdownDescription: "Label-Set to use",
			},
			"language": schema.StringAttribute{
				Computed: true,
			},
			"languages_url": schema.StringAttribute{
				Computed: true,
			},
			"license": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "License to use",
				MarkdownDescription: "License to use",
			},
			"licenses": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"link": schema.StringAttribute{
				Computed: true,
			},
			"mirror": schema.BoolAttribute{
				Computed: true,
			},
			"mirror_interval": schema.StringAttribute{
				Computed: true,
			},
			"mirror_updated": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required:            true,
				Description:         "Name of the repository to create",
				MarkdownDescription: "Name of the repository to create",
			},
			"object_format_name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "ObjectFormatName of the underlying git repository",
				MarkdownDescription: "ObjectFormatName of the underlying git repository",
				Validators: []validator.String{
					stringvalidator.OneOf(
						"sha1",
						"sha256",
					),
				},
			},
			"open_issues_count": schema.Int64Attribute{
				Computed: true,
			},
			"open_pr_counter": schema.Int64Attribute{
				Computed: true,
			},
			"original_url": schema.StringAttribute{
				Computed: true,
			},
			"private": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether the repository is private",
				MarkdownDescription: "Whether the repository is private",
			},
			"projects_mode": schema.StringAttribute{
				Computed: true,
			},
			"readme": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Readme of the repository to create",
				MarkdownDescription: "Readme of the repository to create",
			},
			"release_counter": schema.Int64Attribute{
				Computed: true,
			},
			"repo": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "name of the repo",
				MarkdownDescription: "name of the repo",
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
			"template": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether the repository is template",
				MarkdownDescription: "Whether the repository is template",
			},
			"topics": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"trust_model": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "TrustModel of the repository",
				MarkdownDescription: "TrustModel of the repository",
				Validators: []validator.String{
					stringvalidator.OneOf(
						"default",
						"collaborator",
						"committer",
						"collaboratorcommitter",
					),
				},
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
func mapRepositoryToModel(repo *gitea.Repository, model *RepositoryModel) {
	model.Id = types.Int64Value(repo.ID)
	model.Name = types.StringValue(repo.Name)
	model.FullName = types.StringValue(repo.FullName)
	model.Description = types.StringValue(repo.Description)
	model.Private = types.BoolValue(repo.Private)
	model.DefaultBranch = types.StringValue(repo.DefaultBranch)
	model.Website = types.StringValue(repo.Website)
	model.HtmlUrl = types.StringValue(repo.HTMLURL)
	model.CloneUrl = types.StringValue(repo.CloneURL)
	model.SshUrl = types.StringValue(repo.SSHURL)
	model.Url = types.StringNull() // Not available in SDK
	model.Empty = types.BoolValue(repo.Empty)
	model.Fork = types.BoolValue(repo.Fork)
	model.Mirror = types.BoolValue(repo.Mirror)
	model.Size = types.Int64Value(int64(repo.Size))
	model.Archived = types.BoolValue(repo.Archived)
	model.StarsCount = types.Int64Value(int64(repo.Stars))
	model.WatchersCount = types.Int64Value(int64(repo.Watchers))
	model.ForksCount = types.Int64Value(int64(repo.Forks))
	model.OpenIssuesCount = types.Int64Value(int64(repo.OpenIssues))
	model.Language = types.StringNull() // Not available in SDK
	model.AvatarUrl = types.StringValue(repo.AvatarURL)
	model.Template = types.BoolValue(repo.Template)
	model.Internal = types.BoolValue(repo.Internal)

	// Set remaining computed fields as null/empty
	model.AutoInit = types.BoolNull()
	model.AllowFastForwardOnlyMerge = types.BoolNull()
	model.AllowManualMerge = types.BoolNull()
	model.AllowMergeCommits = types.BoolNull()
	model.AllowRebase = types.BoolNull()
	model.AllowRebaseExplicit = types.BoolNull()
	model.AllowRebaseUpdate = types.BoolNull()
	model.AllowSquashMerge = types.BoolNull()
	model.ArchivedAt = types.StringNull()
	model.AutodetectManualMerge = types.BoolNull()
	model.CreatedAt = types.StringNull()
	model.DefaultAllowMaintainerEdit = types.BoolNull()
	model.DefaultDeleteBranchAfterMerge = types.BoolNull()
	model.DefaultMergeStyle = types.StringNull()
	model.Gitignores = types.StringNull()
	model.HasActions = types.BoolNull()
	model.HasCode = types.BoolNull()
	model.HasIssues = types.BoolNull()
	model.HasPackages = types.BoolNull()
	model.HasProjects = types.BoolNull()
	model.HasPullRequests = types.BoolNull()
	model.HasReleases = types.BoolNull()
	model.HasWiki = types.BoolNull()
	model.IgnoreWhitespaceConflicts = types.BoolNull()
	model.IssueLabels = types.StringNull()
	model.LanguagesUrl = types.StringNull()
	model.License = types.StringNull()
	model.Licenses = types.ListNull(types.StringType)
	model.Link = types.StringNull()
	model.MirrorInterval = types.StringNull()
	model.MirrorUpdated = types.StringNull()
	model.ObjectFormatName = types.StringNull()
	model.OpenPrCounter = types.Int64Null()
	model.OriginalUrl = types.StringNull()
	model.ProjectsMode = types.StringNull()
	model.Readme = types.StringNull()
	model.ReleaseCounter = types.Int64Null()
	model.Repo = types.StringNull()
	model.Topics = types.ListNull(types.StringType)
	model.TrustModel = types.StringNull()
	model.UpdatedAt = types.StringNull()
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
	var plan RepositoryModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create repository using /user/repos endpoint
	createOpts := gitea.CreateRepoOption{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Private:     plan.Private.ValueBool(),
	}

	if !plan.DefaultBranch.IsNull() {
		createOpts.DefaultBranch = plan.DefaultBranch.ValueString()
	}

	repo, _, err := r.client.CreateRepo(createOpts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating repository",
			"Could not create repository, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response to state
	mapRepositoryToModel(repo, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *repositoryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RepositoryModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse owner/repo from full_name if available, otherwise use authenticated user
	owner := ""
	repoName := state.Name.ValueString()

	if !state.FullName.IsNull() && state.FullName.ValueString() != "" {
		// Parse owner/repo from full_name
		fullName := state.FullName.ValueString()
		// Simple split - in production you'd want more robust parsing
		for i, c := range fullName {
			if c == '/' {
				owner = fullName[:i]
				repoName = fullName[i+1:]
				break
			}
		}
	}

	var repo *gitea.Repository
	var err error

	if owner == "" {
		// Get current user to construct owner/repo path
		user, _, err := r.client.GetMyUserInfo()
		if err != nil {
			resp.Diagnostics.AddError("Error getting user info", err.Error())
			return
		}
		owner = user.UserName
	}

	repo, _, err = r.client.GetRepo(owner, repoName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Repository",
			"Could not read repository "+owner+"/"+repoName+": "+err.Error(),
		)
		return
	}

	// Map response to state
	mapRepositoryToModel(repo, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *repositoryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan RepositoryModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse owner/repo from state
	var state RepositoryModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	owner := ""
	repoName := state.Name.ValueString()

	if !state.FullName.IsNull() && state.FullName.ValueString() != "" {
		fullName := state.FullName.ValueString()
		for i, c := range fullName {
			if c == '/' {
				owner = fullName[:i]
				repoName = fullName[i+1:]
				break
			}
		}
	}

	if owner == "" {
		user, _, err := r.client.GetMyUserInfo()
		if err != nil {
			resp.Diagnostics.AddError("Error getting user info", err.Error())
			return
		}
		owner = user.UserName
	}

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

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *repositoryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RepositoryModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	owner := ""
	repoName := state.Name.ValueString()

	if !state.FullName.IsNull() && state.FullName.ValueString() != "" {
		fullName := state.FullName.ValueString()
		for i, c := range fullName {
			if c == '/' {
				owner = fullName[:i]
				repoName = fullName[i+1:]
				break
			}
		}
	}

	if owner == "" {
		user, _, err := r.client.GetMyUserInfo()
		if err != nil {
			resp.Diagnostics.AddError("Error getting user info", err.Error())
			return
		}
		owner = user.UserName
	}

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

	var data RepositoryModel
	mapRepositoryToModel(repository, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type RepositoryModel struct {
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
	FullName                      types.String `tfsdk:"full_name"`
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
	Mirror                        types.Bool   `tfsdk:"mirror"`
	MirrorInterval                types.String `tfsdk:"mirror_interval"`
	MirrorUpdated                 types.String `tfsdk:"mirror_updated"`
	Name                          types.String `tfsdk:"name"`
	ObjectFormatName              types.String `tfsdk:"object_format_name"`
	OpenIssuesCount               types.Int64  `tfsdk:"open_issues_count"`
	OpenPrCounter                 types.Int64  `tfsdk:"open_pr_counter"`
	OriginalUrl                   types.String `tfsdk:"original_url"`
	Private                       types.Bool   `tfsdk:"private"`
	ProjectsMode                  types.String `tfsdk:"projects_mode"`
	Readme                        types.String `tfsdk:"readme"`
	ReleaseCounter                types.Int64  `tfsdk:"release_counter"`
	Repo                          types.String `tfsdk:"repo"`
	Size                          types.Int64  `tfsdk:"size"`
	SshUrl                        types.String `tfsdk:"ssh_url"`
	StarsCount                    types.Int64  `tfsdk:"stars_count"`
	Template                      types.Bool   `tfsdk:"template"`
	Topics                        types.List   `tfsdk:"topics"`
	TrustModel                    types.String `tfsdk:"trust_model"`
	UpdatedAt                     types.String `tfsdk:"updated_at"`
	Url                           types.String `tfsdk:"url"`
	WatchersCount                 types.Int64  `tfsdk:"watchers_count"`
	Website                       types.String `tfsdk:"website"`
}
