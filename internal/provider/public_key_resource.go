package provider

import (
	"context"
	"fmt"
	"strconv"

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

var _ resource.Resource = (*publicKeyResource)(nil)
var _ resource.ResourceWithConfigure = (*publicKeyResource)(nil)
var _ resource.ResourceWithImportState = (*publicKeyResource)(nil)

func NewPublicKeyResource() resource.Resource {
	return &publicKeyResource{}
}

type publicKeyResource struct {
	client *gitea.Client
}

type publicKeyResourceModel struct {
	// Required
	Title types.String `tfsdk:"title"`
	Key   types.String `tfsdk:"key"`

	// Optional
	ReadOnly types.Bool `tfsdk:"read_only"`

	// Computed
	Id          types.Int64  `tfsdk:"id"`
	Fingerprint types.String `tfsdk:"fingerprint"`
	KeyType     types.String `tfsdk:"key_type"`
	Created     types.String `tfsdk:"created_at"`
}

func (r *publicKeyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_public_key"
}

func (r *publicKeyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a user's SSH public key",
		Attributes: map[string]schema.Attribute{
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
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether the key has only read access or read/write",
			},
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The ID of the public key",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"fingerprint": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Fingerprint of the key",
			},
			"key_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Type of the key",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Timestamp when the key was created",
			},
		},
	}
}

func (r *publicKeyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *publicKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data publicKeyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opt := gitea.CreateKeyOption{
		Title:    data.Title.ValueString(),
		Key:      data.Key.ValueString(),
		ReadOnly: data.ReadOnly.ValueBool(),
	}

	key, _, err := r.client.CreatePublicKey(opt)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create public key, got error: %s", err))
		return
	}

	data.Id = types.Int64Value(key.ID)
	data.Fingerprint = types.StringValue(key.Fingerprint)
	data.KeyType = types.StringValue(key.KeyType)
	if !key.Created.IsZero() {
		data.Created = types.StringValue(key.Created.Format("2006-01-02T15:04:05Z07:00"))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *publicKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data publicKeyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	key, httpResp, err := r.client.GetPublicKey(data.Id.ValueInt64())
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read public key, got error: %s", err))
		return
	}

	data.Title = types.StringValue(key.Title)
	data.Key = types.StringValue(key.Key)
	data.ReadOnly = types.BoolValue(key.ReadOnly)
	data.Fingerprint = types.StringValue(key.Fingerprint)
	data.KeyType = types.StringValue(key.KeyType)
	if !key.Created.IsZero() {
		data.Created = types.StringValue(key.Created.Format("2006-01-02T15:04:05Z07:00"))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *publicKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data publicKeyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The Gitea API doesn't support updating public keys
	// Only the title could theoretically be updated, but the API doesn't expose this
	// So we need to recreate the key (which is handled by RequiresReplace on the key field)

	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Public keys cannot be updated. This resource must be recreated.",
	)
}

func (r *publicKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data publicKeyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DeletePublicKey(data.Id.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete public key, got error: %s", err))
		return
	}
}

func (r *publicKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	keyID, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid ID",
			fmt.Sprintf("Expected a numeric ID, got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), keyID)...)
}
