package provider

import (
	"context"
	"fmt"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = (*forkResource)(nil)
var _ resource.ResourceWithConfigure = (*forkResource)(nil)
var _ resource.ResourceWithImportState = (*forkResource)(nil)

func NewForkResource() resource.Resource {
	return &forkResource{}
}

type forkResource struct {
	client *gitea.Client
}

type forkResourceModel struct {
	// Required
	Owner types.String `tfsdk:"owner"`
	Repo  types.String `tfsdk:"repository"`

	// Optional
	Organization types.String `tfsdk:"organization"`
	Name         types.String `tfsdk:"name"`

	// Computed
	Id       types.Int64  `tfsdk:"id"`
	FullName types.String `tfsdk:"full_name"`
	HtmlUrl  types.String `tfsdk:"html_url"`
	SshUrl   types.String `tfsdk:"ssh_url"`
	CloneUrl types.String `tfsdk:"clone_url"`
}

func (r *forkResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_fork"
}

func (r *forkResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Forks a repository",
		Attributes: map[string]schema.Attribute{
			"owner": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Owner of the source repository",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"repository": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the source repository",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"organization": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Organization name to fork into (if forking into an organization)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Name of the forked repository (defaults to source repository name)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The ID of the forked repository",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"full_name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Full name of the forked repository (owner/name)",
			},
			"html_url": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "HTML URL of the forked repository",
			},
			"ssh_url": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "SSH URL of the forked repository",
			},
			"clone_url": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Clone URL of the forked repository",
			},
		},
	}
}

func (r *forkResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	var data forkResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opt := gitea.CreateForkOption{}
	if !data.Organization.IsNull() && !data.Organization.IsUnknown() {
		org := data.Organization.ValueString()
		opt.Organization = &org
	}
	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		name := data.Name.ValueString()
		opt.Name = &name
	}

	fork, _, err := r.client.CreateFork(data.Owner.ValueString(), data.Repo.ValueString(), opt)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create fork, got error: %s", err))
		return
	}

	data.Id = types.Int64Value(fork.ID)
	data.FullName = types.StringValue(fork.FullName)
	data.HtmlUrl = types.StringValue(fork.HTMLURL)
	data.SshUrl = types.StringValue(fork.SSHURL)
	data.CloneUrl = types.StringValue(fork.CloneURL)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *forkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data forkResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse the full_name to get owner and repo
	parts := strings.Split(data.FullName.ValueString(), "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid State", "Full name is not in the format owner/repo")
		return
	}

	repo, httpResp, err := r.client.GetRepo(parts[0], parts[1])
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read fork, got error: %s", err))
		return
	}

	data.Id = types.Int64Value(repo.ID)
	data.FullName = types.StringValue(repo.FullName)
	data.HtmlUrl = types.StringValue(repo.HTMLURL)
	data.SshUrl = types.StringValue(repo.SSHURL)
	data.CloneUrl = types.StringValue(repo.CloneURL)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *forkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Forks cannot be updated, all attributes require replacement
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Forks cannot be updated. All changes require replacement.",
	)
}

func (r *forkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data forkResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse the full_name to get owner and repo
	parts := strings.Split(data.FullName.ValueString(), "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid State", "Full name is not in the format owner/repo")
		return
	}

	_, err := r.client.DeleteRepo(parts[0], parts[1])
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete fork, got error: %s", err))
		return
	}
}

func (r *forkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: owner/repo (of the fork, not the source)
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid ID format",
			fmt.Sprintf("Expected format: owner/repository, got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("full_name"), req.ID)...)
}
