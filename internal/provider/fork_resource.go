package provider

import (
	"context"
	"fmt"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &forkResource{}
	_ resource.ResourceWithConfigure   = &forkResource{}
	_ resource.ResourceWithImportState = &forkResource{}
)

func NewForkResource() resource.Resource {
	return &forkResource{}
}

type forkResource struct {
	client *gitea.Client
}

type forkResourceModel struct {
	// Required
	Owner types.String `tfsdk:"owner"`
	Repo  types.String `tfsdk:"repo"`

	// Optional
	Organization types.String `tfsdk:"organization"`

	// Computed
	Id types.String `tfsdk:"id"`
}

func (r *forkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_fork"
}

func (r *forkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Forks a repository.",
		MarkdownDescription: "Forks a repository in Gitea.",
		Attributes: map[string]schema.Attribute{
			// Required
			"owner": schema.StringAttribute{
				Required:            true,
				Description:         "The owner or owning organization of the repository to fork.",
				MarkdownDescription: "The owner or owning organization of the repository to fork.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"repo": schema.StringAttribute{
				Required:            true,
				Description:         "The name of the repository to fork.",
				MarkdownDescription: "The name of the repository to fork.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			// Optional
			"organization": schema.StringAttribute{
				Optional:            true,
				Description:         "The organization that owns the forked repo.",
				MarkdownDescription: "The organization that owns the forked repo.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			// Computed
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "The ID of this resource.",
				MarkdownDescription: "The ID of this resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *forkResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *forkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan forkResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opt := gitea.CreateForkOption{}
	if !plan.Organization.IsNull() && !plan.Organization.IsUnknown() {
		org := plan.Organization.ValueString()
		opt.Organization = &org
	}

	fork, _, err := r.client.CreateFork(plan.Owner.ValueString(), plan.Repo.ValueString(), opt)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Fork",
			fmt.Sprintf("Could not fork repository %s/%s: %s", plan.Owner.ValueString(), plan.Repo.ValueString(), err.Error()),
		)
		return
	}

	plan.Id = types.StringValue(fmt.Sprintf("%d", fork.ID))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *forkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state forkResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Determine the fork owner (either organization or current user)
	var forkOwner string
	if !state.Organization.IsNull() && !state.Organization.IsUnknown() {
		forkOwner = state.Organization.ValueString()
	} else {
		// Get current user
		user, _, err := r.client.GetMyUserInfo()
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading Fork",
				fmt.Sprintf("Could not get current user: %s", err.Error()),
			)
			return
		}
		forkOwner = user.UserName
	}

	// The fork name defaults to the source repo name
	repoName := state.Repo.ValueString()

	repo, httpResp, err := r.client.GetRepo(forkOwner, repoName)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Fork",
			fmt.Sprintf("Could not read fork %s/%s: %s", forkOwner, repoName, err.Error()),
		)
		return
	}

	// Update computed fields
	state.Id = types.StringValue(fmt.Sprintf("%d", repo.ID))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *forkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Forks cannot be updated, all attributes require replacement
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Forks cannot be updated. All changes require replacement.",
	)
}

func (r *forkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state forkResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Determine the fork owner (either organization or current user)
	var forkOwner string
	if !state.Organization.IsNull() && !state.Organization.IsUnknown() {
		forkOwner = state.Organization.ValueString()
	} else {
		// Get current user
		user, _, err := r.client.GetMyUserInfo()
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Deleting Fork",
				fmt.Sprintf("Could not get current user: %s", err.Error()),
			)
			return
		}
		forkOwner = user.UserName
	}

	// The fork name defaults to the source repo name
	repoName := state.Repo.ValueString()

	_, err := r.client.DeleteRepo(forkOwner, repoName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Fork",
			fmt.Sprintf("Could not delete fork %s/%s: %s", forkOwner, repoName, err.Error()),
		)
		return
	}
}

func (r *forkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: owner/repo (of the fork, not the source)
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid ID Format",
			fmt.Sprintf("Expected format: owner/repository, got: %s", req.ID),
		)
		return
	}

	forkOwner := parts[0]
	forkRepo := parts[1]

	// Fetch the repository to get full details
	repo, httpResp, err := r.client.GetRepo(forkOwner, forkRepo)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			resp.Diagnostics.AddError(
				"Repository Not Found",
				fmt.Sprintf("Repository %s does not exist or is not accessible", req.ID),
			)
			return
		}
		resp.Diagnostics.AddError(
			"Error Importing Fork",
			fmt.Sprintf("Could not import fork %s: %s", req.ID, err.Error()),
		)
		return
	}

	// Verify this is actually a fork
	if !repo.Fork || repo.Parent == nil {
		resp.Diagnostics.AddError(
			"Not a Fork",
			fmt.Sprintf("Repository %s is not a fork", req.ID),
		)
		return
	}

	var data forkResourceModel
	data.Id = types.StringValue(fmt.Sprintf("%d", repo.ID))
	data.Repo = types.StringValue(repo.Parent.Name)

	// Set source owner from parent
	if repo.Parent.Owner != nil {
		data.Owner = types.StringValue(repo.Parent.Owner.UserName)
	}

	// Check if fork owner is an organization or the current user
	user, _, _ := r.client.GetMyUserInfo()
	if user != nil && user.UserName == forkOwner {
		// Fork is owned by current user, no organization specified
		data.Organization = types.StringNull()
	} else {
		// Fork is owned by an organization
		data.Organization = types.StringValue(forkOwner)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
