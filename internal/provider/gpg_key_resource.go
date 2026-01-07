package provider

import (
	"context"
	"fmt"
	"strconv"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = (*gpgKeyResource)(nil)
var _ resource.ResourceWithConfigure = (*gpgKeyResource)(nil)
var _ resource.ResourceWithImportState = (*gpgKeyResource)(nil)

func NewGPGKeyResource() resource.Resource {
	return &gpgKeyResource{}
}

type gpgKeyResource struct {
	client *gitea.Client
}

type gpgKeyResourceModel struct {
	// Required
	ArmoredPublicKey types.String `tfsdk:"armored_public_key"`

	// Computed
	Id                types.Int64  `tfsdk:"id"`
	PrimaryKeyId      types.String `tfsdk:"primary_key_id"`
	KeyId             types.String `tfsdk:"key_id"`
	CanSign           types.Bool   `tfsdk:"can_sign"`
	CanEncryptComms   types.Bool   `tfsdk:"can_encrypt_comms"`
	CanEncryptStorage types.Bool   `tfsdk:"can_encrypt_storage"`
	CanCertify        types.Bool   `tfsdk:"can_certify"`
	Created           types.String `tfsdk:"created_at"`
	Expires           types.String `tfsdk:"expires_at"`
}

func (r *gpgKeyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gpg_key"
}

func (r *gpgKeyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a user's GPG key for signing commits",
		Attributes: map[string]schema.Attribute{
			"armored_public_key": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "An armored GPG key to add",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The ID of the GPG key",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"primary_key_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Primary key ID",
			},
			"key_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Key ID",
			},
			"can_sign": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the key can sign",
			},
			"can_encrypt_comms": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the key can encrypt communications",
			},
			"can_encrypt_storage": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the key can encrypt storage",
			},
			"can_certify": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the key can certify",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Timestamp when the key was created",
			},
			"expires_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Timestamp when the key expires",
			},
		},
	}
}

func (r *gpgKeyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *gpgKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data gpgKeyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opt := gitea.CreateGPGKeyOption{
		ArmoredKey: data.ArmoredPublicKey.ValueString(),
	}

	key, _, err := r.client.CreateGPGKey(opt)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create GPG key, got error: %s", err))
		return
	}

	data.Id = types.Int64Value(key.ID)
	data.PrimaryKeyId = types.StringValue(key.PrimaryKeyID)
	data.KeyId = types.StringValue(key.KeyID)
	data.CanSign = types.BoolValue(key.CanSign)
	data.CanEncryptComms = types.BoolValue(key.CanEncryptComms)
	data.CanEncryptStorage = types.BoolValue(key.CanEncryptStorage)
	data.CanCertify = types.BoolValue(key.CanCertify)
	if !key.Created.IsZero() {
		data.Created = types.StringValue(key.Created.Format("2006-01-02T15:04:05Z07:00"))
	}
	if !key.Expires.IsZero() {
		data.Expires = types.StringValue(key.Expires.Format("2006-01-02T15:04:05Z07:00"))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *gpgKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data gpgKeyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	key, httpResp, err := r.client.GetGPGKey(data.Id.ValueInt64())
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read GPG key, got error: %s", err))
		return
	}

	data.PrimaryKeyId = types.StringValue(key.PrimaryKeyID)
	data.KeyId = types.StringValue(key.KeyID)
	data.CanSign = types.BoolValue(key.CanSign)
	data.CanEncryptComms = types.BoolValue(key.CanEncryptComms)
	data.CanEncryptStorage = types.BoolValue(key.CanEncryptStorage)
	data.CanCertify = types.BoolValue(key.CanCertify)
	if !key.Created.IsZero() {
		data.Created = types.StringValue(key.Created.Format("2006-01-02T15:04:05Z07:00"))
	}
	if !key.Expires.IsZero() {
		data.Expires = types.StringValue(key.Expires.Format("2006-01-02T15:04:05Z07:00"))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *gpgKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// GPG keys cannot be updated via the Gitea API
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"GPG keys cannot be updated. This resource must be recreated.",
	)
}

func (r *gpgKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data gpgKeyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DeleteGPGKey(data.Id.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete GPG key, got error: %s", err))
		return
	}
}

func (r *gpgKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
