package provider

import (
	"context"
	"fmt"

	"terraform-provider-gitea/internal/resource_repository"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
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
	resp.Schema = resource_repository.RepositoryResourceSchema(ctx)
}

// Helper function to map Gitea Repository to Terraform model
func mapRepositoryToModel(repo *gitea.Repository, model *resource_repository.RepositoryModel) {
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
	model.ExternalTracker = resource_repository.NewExternalTrackerValueNull()
	model.ExternalWiki = resource_repository.NewExternalWikiValueNull()
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
	model.InternalTracker = resource_repository.NewInternalTrackerValueNull()
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
	model.Owner = resource_repository.NewOwnerValueNull()
	model.Permissions = resource_repository.NewPermissionsValueNull()
	model.ProjectsMode = types.StringNull()
	model.Readme = types.StringNull()
	model.ReleaseCounter = types.Int64Null()
	model.Repo = types.StringNull()
	model.RepoTransfer = resource_repository.NewRepoTransferValueNull()
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
	var plan resource_repository.RepositoryModel

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
	var state resource_repository.RepositoryModel

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
	var plan resource_repository.RepositoryModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse owner/repo from state
	var state resource_repository.RepositoryModel
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
	var state resource_repository.RepositoryModel

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
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
