package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &publicKeyResource{}
	_ resource.ResourceWithConfigure   = &publicKeyResource{}
	_ resource.ResourceWithImportState = &publicKeyResource{}
)

func NewPublicKeyResource() resource.Resource {
	return &publicKeyResource{}
}

type publicKeyResource struct {
	client *gitea.Client
}

type publicKeyResourceModel struct {
	// Required
	Key      types.String `tfsdk:"key"`
	Title    types.String `tfsdk:"title"`
	Username types.String `tfsdk:"username"`

	// Optional
	ReadOnly types.Bool `tfsdk:"read_only"`

	// Computed
	Created     types.String `tfsdk:"created"`
	Fingerprint types.String `tfsdk:"fingerprint"`
	Id          types.String `tfsdk:"id"`
	Type        types.String `tfsdk:"type"`
}

func (r *publicKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_public_key"
}

func (r *publicKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manages a user's SSH public key.",
		MarkdownDescription: "Manages a user's SSH public key in Gitea.",
		Attributes: map[string]schema.Attribute{
			// Required
			"key": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				Description:         "An armored SSH key to add.",
				MarkdownDescription: "An armored SSH key to add.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"title": schema.StringAttribute{
				Required:            true,
				Description:         "Title of the key to add.",
				MarkdownDescription: "Title of the key to add.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"username": schema.StringAttribute{
				Required:            true,
				Description:         "User to associate with the added key.",
				MarkdownDescription: "User to associate with the added key.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			// Optional
			"read_only": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Describe if the key has only read access or read/write.",
				MarkdownDescription: "Describe if the key has only read access or read/write.",
			},

			// Computed
			"created": schema.StringAttribute{
				Computed:            true,
				Description:         "Timestamp when the key was created.",
				MarkdownDescription: "Timestamp when the key was created.",
			},
			"fingerprint": schema.StringAttribute{
				Computed:            true,
				Description:         "SSH key fingerprint.",
				MarkdownDescription: "SSH key fingerprint.",
			},
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "The ID of this resource.",
				MarkdownDescription: "The ID of this resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				Computed:            true,
				Description:         "Type of SSH key.",
				MarkdownDescription: "Type of SSH key.",
			},
		},
	}
}

func (r *publicKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	var plan publicKeyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	username := plan.Username.ValueString()

	opt := gitea.CreateKeyOption{
		Title:    plan.Title.ValueString(),
		Key:      plan.Key.ValueString(),
		ReadOnly: plan.ReadOnly.ValueBool(),
	}

	key, _, err := r.client.AdminCreateUserPublicKey(username, opt)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Public Key",
			fmt.Sprintf("Could not create public key for user '%s': %s", username, err.Error()),
		)
		return
	}

	plan.Id = types.StringValue(fmt.Sprintf("%d", key.ID))
	plan.Fingerprint = types.StringValue(key.Fingerprint)
	plan.Type = types.StringValue(key.KeyType)
	plan.ReadOnly = types.BoolValue(key.ReadOnly)
	if !key.Created.IsZero() {
		plan.Created = types.StringValue(key.Created.Format("2006-01-02T15:04:05Z07:00"))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *publicKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state publicKeyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse the ID
	keyID, err := strconv.ParseInt(state.Id.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Public Key",
			fmt.Sprintf("Invalid key ID: %s", state.Id.ValueString()),
		)
		return
	}

	// List keys for the user and find the matching one
	username := state.Username.ValueString()
	keys, _, err := r.client.ListPublicKeys(username, gitea.ListPublicKeysOptions{
		ListOptions: gitea.ListOptions{Page: -1},
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Public Key",
			fmt.Sprintf("Could not list public keys for user '%s': %s", username, err.Error()),
		)
		return
	}

	var foundKey *gitea.PublicKey
	for _, k := range keys {
		if k.ID == keyID {
			foundKey = k
			break
		}
	}

	if foundKey == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Title = types.StringValue(foundKey.Title)
	state.Key = types.StringValue(foundKey.Key)
	state.ReadOnly = types.BoolValue(foundKey.ReadOnly)
	state.Fingerprint = types.StringValue(foundKey.Fingerprint)
	state.Type = types.StringValue(foundKey.KeyType)
	if !foundKey.Created.IsZero() {
		state.Created = types.StringValue(foundKey.Created.Format("2006-01-02T15:04:05Z07:00"))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *publicKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Public keys cannot be updated
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Public keys cannot be updated. This resource must be recreated.",
	)
}

func (r *publicKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state publicKeyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse the ID
	keyID, err := strconv.ParseInt(state.Id.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Public Key",
			fmt.Sprintf("Invalid key ID: %s", state.Id.ValueString()),
		)
		return
	}

	username := state.Username.ValueString()
	_, err = r.client.AdminDeleteUserPublicKey(username, int(keyID))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Public Key",
			fmt.Sprintf("Could not delete public key %d for user '%s': %s", keyID, username, err.Error()),
		)
		return
	}
}

func (r *publicKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: "username/key_id"
	id := req.ID
	parts := strings.Split(id, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Import ID must be in format 'username/key_id', got: %s", id),
		)
		return
	}

	username := parts[0]
	keyIDStr := parts[1]

	keyID, err := strconv.ParseInt(keyIDStr, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("key_id must be a valid number, got: %s", keyIDStr),
		)
		return
	}

	// List keys for the user and find the matching one
	keys, _, err := r.client.ListPublicKeys(username, gitea.ListPublicKeysOptions{
		ListOptions: gitea.ListOptions{Page: -1},
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Importing Public Key",
			fmt.Sprintf("Could not list public keys for user '%s': %s", username, err.Error()),
		)
		return
	}

	var foundKey *gitea.PublicKey
	for _, k := range keys {
		if k.ID == keyID {
			foundKey = k
			break
		}
	}

	if foundKey == nil {
		resp.Diagnostics.AddError(
			"Public Key Not Found",
			fmt.Sprintf("Public key %d not found for user '%s'", keyID, username),
		)
		return
	}

	state := publicKeyResourceModel{
		Id:          types.StringValue(fmt.Sprintf("%d", foundKey.ID)),
		Username:    types.StringValue(username),
		Title:       types.StringValue(foundKey.Title),
		Key:         types.StringValue(foundKey.Key),
		ReadOnly:    types.BoolValue(foundKey.ReadOnly),
		Fingerprint: types.StringValue(foundKey.Fingerprint),
		Type:        types.StringValue(foundKey.KeyType),
	}
	if !foundKey.Created.IsZero() {
		state.Created = types.StringValue(foundKey.Created.Format("2006-01-02T15:04:05Z07:00"))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
