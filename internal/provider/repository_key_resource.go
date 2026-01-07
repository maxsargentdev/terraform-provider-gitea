package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = (*repositoryKeyResource)(nil)
var _ resource.ResourceWithConfigure = (*repositoryKeyResource)(nil)
var _ resource.ResourceWithImportState = (*repositoryKeyResource)(nil)

func NewRepositoryKeyResource() resource.Resource {
	return &repositoryKeyResource{}
}

type repositoryKeyResource struct {
	client *gitea.Client
}

type repositoryKeyResourceModel struct {
	// Required
	Owner types.String `tfsdk:"owner"`
	Repo  types.String `tfsdk:"repository"`
	Title types.String `tfsdk:"title"`
	Key   types.String `tfsdk:"key"`

	// Optional
	ReadOnly types.Bool `tfsdk:"read_only"`

	// Computed
	Id          types.Int64  `tfsdk:"id"`
	Fingerprint types.String `tfsdk:"fingerprint"`
	Created     types.String `tfsdk:"created_at"`
}

func (r *repositoryKeyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repository_key"
}

func (r *repositoryKeyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a repository deploy key",
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
			"title": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Title of the key",
			},
			"key": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "An armored SSH key to add",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"read_only": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Whether the key has only read access or read/write",
			},
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The ID of the deploy key",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"fingerprint": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Fingerprint of the key",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Timestamp when the key was created",
			},
		},
	}
}

func (r *repositoryKeyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *repositoryKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data repositoryKeyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opt := gitea.CreateKeyOption{
		Title:    data.Title.ValueString(),
		Key:      data.Key.ValueString(),
		ReadOnly: data.ReadOnly.ValueBool(),
	}

	key, _, err := r.client.CreateDeployKey(data.Owner.ValueString(), data.Repo.ValueString(), opt)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create repository deploy key, got error: %s", err))
		return
	}

	data.Id = types.Int64Value(key.ID)
	data.Fingerprint = types.StringValue(key.Fingerprint)
	if !key.Created.IsZero() {
		data.Created = types.StringValue(key.Created.Format("2006-01-02T15:04:05Z07:00"))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *repositoryKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data repositoryKeyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	key, httpResp, err := r.client.GetDeployKey(data.Owner.ValueString(), data.Repo.ValueString(), data.Id.ValueInt64())
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read repository deploy key, got error: %s", err))
		return
	}

	data.Title = types.StringValue(key.Title)
	data.Key = types.StringValue(key.Key)
	data.ReadOnly = types.BoolValue(key.ReadOnly)
	data.Fingerprint = types.StringValue(key.Fingerprint)
	if !key.Created.IsZero() {
		data.Created = types.StringValue(key.Created.Format("2006-01-02T15:04:05Z07:00"))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *repositoryKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data repositoryKeyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The Gitea API doesn't support updating deploy keys
	// Only title/read_only could be updated, but the API doesn't expose this
	// The key field has RequiresReplace, so changes to it will trigger recreation

	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Repository deploy keys cannot be updated. This resource must be recreated.",
	)
}

func (r *repositoryKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data repositoryKeyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DeleteDeployKey(data.Owner.ValueString(), data.Repo.ValueString(), data.Id.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete repository deploy key, got error: %s", err))
		return
	}
}

func (r *repositoryKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: owner/repo/keyID
	parts := strings.Split(req.ID, "/")
	if len(parts) != 3 {
		resp.Diagnostics.AddError(
			"Invalid ID format",
			fmt.Sprintf("Expected format: owner/repository/keyID, got: %s", req.ID),
		)
		return
	}

	owner := parts[0]
	repo := parts[1]
	keyID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid key ID",
			fmt.Sprintf("Expected numeric key ID, got: %s", parts[2]),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("owner"), owner)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("repository"), repo)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), keyID)...)
}
