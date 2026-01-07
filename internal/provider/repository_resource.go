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
	// Required
	Owner types.String `tfsdk:"owner"`
	Name  types.String `tfsdk:"name"`

	// Optional - from CreateRepoOption
	Description      types.String `tfsdk:"description"`
	Private          types.Bool   `tfsdk:"private"`
	IssueLabels      types.String `tfsdk:"issue_labels"`
	AutoInit         types.Bool   `tfsdk:"auto_init"`
	Template         types.Bool   `tfsdk:"template"`
	Gitignores       types.String `tfsdk:"gitignores"`
	License          types.String `tfsdk:"license"`
	Readme           types.String `tfsdk:"readme"`
	DefaultBranch    types.String `tfsdk:"default_branch"`
	TrustModel       types.String `tfsdk:"trust_model"`
	ObjectFormatName types.String `tfsdk:"object_format_name"`

	// Computed - key outputs
	Id       types.Int64  `tfsdk:"id"`
	HtmlUrl  types.String `tfsdk:"html_url"`
	SshUrl   types.String `tfsdk:"ssh_url"`
	CloneUrl types.String `tfsdk:"clone_url"`
}

func (r *repositoryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repository"
}

func (r *repositoryResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Gitea repository.",
		Attributes: map[string]schema.Attribute{
			// Required
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

			// Optional - from CreateRepoOption
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Description of the repository to create",
				MarkdownDescription: "Description of the repository to create",
			},
			"private": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether the repository is private",
				MarkdownDescription: "Whether the repository is private",
			},
			"issue_labels": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Issue Label set to use",
				MarkdownDescription: "Issue Label set to use",
			},
			"auto_init": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether the repository should be auto-initialized?",
				MarkdownDescription: "Whether the repository should be auto-initialized?",
			},
			"template": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether the repository is template",
				MarkdownDescription: "Whether the repository is template",
			},
			"gitignores": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Gitignores to use",
				MarkdownDescription: "Gitignores to use",
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
			"default_branch": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "DefaultBranch of the repository (used when initializes and in template)",
				MarkdownDescription: "DefaultBranch of the repository (used when initializes and in template)",
			},
			"trust_model": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "TrustModel of the repository",
				MarkdownDescription: "TrustModel of the repository",
			},
			"object_format_name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "ObjectFormatName of the underlying git repository (sha1 or sha256)",
				MarkdownDescription: "ObjectFormatName of the underlying git repository (sha1 or sha256)",
			},

			// Computed - key outputs
			"id": schema.Int64Attribute{
				Computed:            true,
				Description:         "The ID of the repository",
				MarkdownDescription: "The ID of the repository",
			},
			"html_url": schema.StringAttribute{
				Computed:            true,
				Description:         "The URL to the repository in the web UI",
				MarkdownDescription: "The URL to the repository in the web UI",
			},
			"ssh_url": schema.StringAttribute{
				Computed:            true,
				Description:         "The SSH URL to clone the repository",
				MarkdownDescription: "The SSH URL to clone the repository",
			},
			"clone_url": schema.StringAttribute{
				Computed:            true,
				Description:         "The HTTPS URL to clone the repository",
				MarkdownDescription: "The HTTPS URL to clone the repository",
			},
		},
	}
}

// Helper function to map Gitea Repository to Terraform model
func mapRepositoryToModel(repo *gitea.Repository, model *repositoryResourceModel) {
	// Basic fields
	model.Id = types.Int64Value(repo.ID)
	model.Name = types.StringValue(repo.Name)
	model.HtmlUrl = types.StringValue(repo.HTMLURL)
	model.CloneUrl = types.StringValue(repo.CloneURL)
	model.SshUrl = types.StringValue(repo.SSHURL)

	// Optional fields that are returned by API
	model.Description = types.StringValue(repo.Description)
	model.Private = types.BoolValue(repo.Private)
	model.DefaultBranch = types.StringValue(repo.DefaultBranch)
	model.Template = types.BoolValue(repo.Template)
	model.ObjectFormatName = types.StringValue(repo.ObjectFormatName)

	// Creation-only fields - preserve from existing model, convert Unknown to null
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
