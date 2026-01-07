package provider

import (
	"context"
	"fmt"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = (*repositoryActionsVariableResource)(nil)
var _ resource.ResourceWithConfigure = (*repositoryActionsVariableResource)(nil)
var _ resource.ResourceWithImportState = (*repositoryActionsVariableResource)(nil)

func NewRepositoryActionsVariableResource() resource.Resource {
	return &repositoryActionsVariableResource{}
}

type repositoryActionsVariableResource struct {
	client *gitea.Client
}

type repositoryActionsVariableResourceModel struct {
	// Required
	Owner types.String `tfsdk:"owner"`
	Repo  types.String `tfsdk:"repository"`
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

func (r *repositoryActionsVariableResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repository_actions_variable"
}

func (r *repositoryActionsVariableResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a repository actions variable",
		Attributes: map[string]schema.Attribute{
			"owner": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Owner of the repository",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"repository": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the repository",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the variable",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"value": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Value of the variable",
			},
		},
	}
}

func (r *repositoryActionsVariableResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *repositoryActionsVariableResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data repositoryActionsVariableResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.CreateRepoActionVariable(
		data.Owner.ValueString(),
		data.Repo.ValueString(),
		data.Name.ValueString(),
		data.Value.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create repository actions variable, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *repositoryActionsVariableResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data repositoryActionsVariableResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	variable, httpResp, err := r.client.GetRepoActionVariable(
		data.Owner.ValueString(),
		data.Repo.ValueString(),
		data.Name.ValueString(),
	)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read repository actions variable, got error: %s", err))
		return
	}

	data.Name = types.StringValue(variable.Name)
	data.Value = types.StringValue(variable.Value)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *repositoryActionsVariableResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data repositoryActionsVariableResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.UpdateRepoActionVariable(
		data.Owner.ValueString(),
		data.Repo.ValueString(),
		data.Name.ValueString(),
		data.Value.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update repository actions variable, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *repositoryActionsVariableResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data repositoryActionsVariableResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DeleteRepoActionVariable(
		data.Owner.ValueString(),
		data.Repo.ValueString(),
		data.Name.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete repository actions variable, got error: %s", err))
		return
	}
}

func (r *repositoryActionsVariableResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: owner/repo/variableName
	parts := strings.Split(req.ID, "/")
	if len(parts) != 3 {
		resp.Diagnostics.AddError(
			"Invalid ID format",
			fmt.Sprintf("Expected format: owner/repository/variableName, got: %s", req.ID),
		)
		return
	}

	owner := parts[0]
	repo := parts[1]
	variableName := parts[2]

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("owner"), owner)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("repository"), repo)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), variableName)...)
}
