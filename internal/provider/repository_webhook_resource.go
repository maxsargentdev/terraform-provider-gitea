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

var _ resource.Resource = (*repositoryWebhookResource)(nil)
var _ resource.ResourceWithConfigure = (*repositoryWebhookResource)(nil)
var _ resource.ResourceWithImportState = (*repositoryWebhookResource)(nil)

func NewRepositoryWebhookResource() resource.Resource {
	return &repositoryWebhookResource{}
}

type repositoryWebhookResource struct {
	client *gitea.Client
}

type repositoryWebhookResourceModel struct {
	// Required
	Owner  types.String `tfsdk:"owner"`
	Repo   types.String `tfsdk:"repository"`
	Type   types.String `tfsdk:"type"`
	Config types.Map    `tfsdk:"config"`
	Events types.List   `tfsdk:"events"`

	// Optional
	BranchFilter        types.String `tfsdk:"branch_filter"`
	Active              types.Bool   `tfsdk:"active"`
	AuthorizationHeader types.String `tfsdk:"authorization_header"`

	// Computed
	Id      types.Int64  `tfsdk:"id"`
	Updated types.String `tfsdk:"updated_at"`
	Created types.String `tfsdk:"created_at"`
}

func (r *repositoryWebhookResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repository_webhook"
}

func (r *repositoryWebhookResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a repository webhook",
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
			"type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Type of webhook (gitea, gogs, slack, discord, dingtalk, telegram, msteams, feishu)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"config": schema.MapAttribute{
				Required:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Configuration for the webhook (e.g., url, content_type, secret)",
			},
			"events": schema.ListAttribute{
				Required:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "List of event types that trigger this webhook",
			},
			"branch_filter": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Branch filter for the webhook",
			},
			"active": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Whether the webhook is active",
			},
			"authorization_header": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Authorization header for the webhook",
			},
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The ID of the webhook",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Timestamp when the webhook was last updated",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Timestamp when the webhook was created",
			},
		},
	}
}

func (r *repositoryWebhookResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *repositoryWebhookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data repositoryWebhookResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var config map[string]string
	resp.Diagnostics.Append(data.Config.ElementsAs(ctx, &config, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var events []string
	resp.Diagnostics.Append(data.Events.ElementsAs(ctx, &events, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opt := gitea.CreateHookOption{
		Type:                gitea.HookType(data.Type.ValueString()),
		Config:              config,
		Events:              events,
		BranchFilter:        data.BranchFilter.ValueString(),
		Active:              data.Active.ValueBool(),
		AuthorizationHeader: data.AuthorizationHeader.ValueString(),
	}

	hook, _, err := r.client.CreateRepoHook(data.Owner.ValueString(), data.Repo.ValueString(), opt)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create repository webhook, got error: %s", err))
		return
	}

	data.Id = types.Int64Value(hook.ID)
	if !hook.Updated.IsZero() {
		data.Updated = types.StringValue(hook.Updated.Format("2006-01-02T15:04:05Z07:00"))
	}
	if !hook.Created.IsZero() {
		data.Created = types.StringValue(hook.Created.Format("2006-01-02T15:04:05Z07:00"))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *repositoryWebhookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data repositoryWebhookResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hook, httpResp, err := r.client.GetRepoHook(data.Owner.ValueString(), data.Repo.ValueString(), data.Id.ValueInt64())
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read repository webhook, got error: %s", err))
		return
	}

	data.Type = types.StringValue(hook.Type)
	data.BranchFilter = types.StringValue(hook.BranchFilter)
	data.Active = types.BoolValue(hook.Active)
	data.AuthorizationHeader = types.StringValue(hook.AuthorizationHeader)

	// Convert config to map
	config, diags := types.MapValueFrom(ctx, types.StringType, hook.Config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Config = config

	// Convert events to list
	events, diags := types.ListValueFrom(ctx, types.StringType, hook.Events)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Events = events

	if !hook.Updated.IsZero() {
		data.Updated = types.StringValue(hook.Updated.Format("2006-01-02T15:04:05Z07:00"))
	}
	if !hook.Created.IsZero() {
		data.Created = types.StringValue(hook.Created.Format("2006-01-02T15:04:05Z07:00"))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *repositoryWebhookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data repositoryWebhookResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var config map[string]string
	resp.Diagnostics.Append(data.Config.ElementsAs(ctx, &config, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var events []string
	resp.Diagnostics.Append(data.Events.ElementsAs(ctx, &events, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	active := data.Active.ValueBool()
	opt := gitea.EditHookOption{
		Config:              config,
		Events:              events,
		BranchFilter:        data.BranchFilter.ValueString(),
		Active:              &active,
		AuthorizationHeader: data.AuthorizationHeader.ValueString(),
	}

	_, err := r.client.EditRepoHook(data.Owner.ValueString(), data.Repo.ValueString(), data.Id.ValueInt64(), opt)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update repository webhook, got error: %s", err))
		return
	}

	// Read back the updated webhook
	hook, _, err := r.client.GetRepoHook(data.Owner.ValueString(), data.Repo.ValueString(), data.Id.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read updated repository webhook, got error: %s", err))
		return
	}

	if !hook.Updated.IsZero() {
		data.Updated = types.StringValue(hook.Updated.Format("2006-01-02T15:04:05Z07:00"))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *repositoryWebhookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data repositoryWebhookResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DeleteRepoHook(data.Owner.ValueString(), data.Repo.ValueString(), data.Id.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete repository webhook, got error: %s", err))
		return
	}
}

func (r *repositoryWebhookResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: owner/repo/hookID
	parts := strings.Split(req.ID, "/")
	if len(parts) != 3 {
		resp.Diagnostics.AddError(
			"Invalid ID format",
			fmt.Sprintf("Expected format: owner/repository/hookID, got: %s", req.ID),
		)
		return
	}

	owner := parts[0]
	repo := parts[1]
	hookID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid hook ID",
			fmt.Sprintf("Expected numeric hook ID, got: %s", parts[2]),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("owner"), owner)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("repository"), repo)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), hookID)...)
}
