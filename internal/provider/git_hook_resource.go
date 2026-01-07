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

var _ resource.Resource = (*gitHookResource)(nil)
var _ resource.ResourceWithConfigure = (*gitHookResource)(nil)
var _ resource.ResourceWithImportState = (*gitHookResource)(nil)

func NewGitHookResource() resource.Resource {
	return &gitHookResource{}
}

type gitHookResource struct {
	client *gitea.Client
}

type gitHookResourceModel struct {
	// Required
	Owner   types.String `tfsdk:"owner"`
	Repo    types.String `tfsdk:"repository"`
	Name    types.String `tfsdk:"name"`
	Content types.String `tfsdk:"content"`

	// Computed
	IsActive types.Bool `tfsdk:"is_active"`
}

func (r *gitHookResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_git_hook"
}

func (r *gitHookResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a repository git hook",
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
				MarkdownDescription: "Name of the git hook (e.g., pre-receive, update, post-receive)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"content": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Content/script of the git hook",
			},
			"is_active": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the git hook is active",
			},
		},
	}
}

func (r *gitHookResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *gitHookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data gitHookResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opt := gitea.EditGitHookOption{
		Content: data.Content.ValueString(),
	}

	_, err := r.client.EditRepoGitHook(
		data.Owner.ValueString(),
		data.Repo.ValueString(),
		data.Name.ValueString(),
		opt,
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create git hook, got error: %s", err))
		return
	}

	// Read back to get the is_active status
	hook, _, err := r.client.GetRepoGitHook(
		data.Owner.ValueString(),
		data.Repo.ValueString(),
		data.Name.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read created git hook, got error: %s", err))
		return
	}

	data.IsActive = types.BoolValue(hook.IsActive)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *gitHookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data gitHookResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hook, httpResp, err := r.client.GetRepoGitHook(
		data.Owner.ValueString(),
		data.Repo.ValueString(),
		data.Name.ValueString(),
	)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read git hook, got error: %s", err))
		return
	}

	data.Name = types.StringValue(hook.Name)
	data.Content = types.StringValue(hook.Content)
	data.IsActive = types.BoolValue(hook.IsActive)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *gitHookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data gitHookResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opt := gitea.EditGitHookOption{
		Content: data.Content.ValueString(),
	}

	_, err := r.client.EditRepoGitHook(
		data.Owner.ValueString(),
		data.Repo.ValueString(),
		data.Name.ValueString(),
		opt,
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update git hook, got error: %s", err))
		return
	}

	// Read back to get the updated is_active status
	hook, _, err := r.client.GetRepoGitHook(
		data.Owner.ValueString(),
		data.Repo.ValueString(),
		data.Name.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read updated git hook, got error: %s", err))
		return
	}

	data.IsActive = types.BoolValue(hook.IsActive)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *gitHookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data gitHookResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DeleteRepoGitHook(
		data.Owner.ValueString(),
		data.Repo.ValueString(),
		data.Name.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete git hook, got error: %s", err))
		return
	}
}

func (r *gitHookResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: owner/repo/hookName
	parts := strings.Split(req.ID, "/")
	if len(parts) != 3 {
		resp.Diagnostics.AddError(
			"Invalid ID format",
			fmt.Sprintf("Expected format: owner/repository/hookName, got: %s", req.ID),
		)
		return
	}

	owner := parts[0]
	repo := parts[1]
	hookName := parts[2]

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("owner"), owner)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("repository"), repo)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), hookName)...)
}
