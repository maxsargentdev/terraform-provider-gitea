package provider

import (
	"context"
	"fmt"
	"terraform-provider-gitea/internal/resource_user"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = (*userResource)(nil)
var _ resource.ResourceWithConfigure = (*userResource)(nil)

func NewUserResource() resource.Resource {
	return &userResource{}
}

type userResource struct {
	client *gitea.Client
}

func (r *userResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *userResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_user.UserResourceSchema(ctx)
}

func (r *userResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resource_user.UserModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create user via Gitea API
	createOpts := gitea.CreateUserOption{
		Username:           data.Username.ValueString(),
		Email:              data.Email.ValueString(),
		Password:           data.Password.ValueString(),
		MustChangePassword: data.MustChangePassword.ValueBoolPointer(),
	}

	if !data.FullName.IsNull() && !data.FullName.IsUnknown() {
		createOpts.FullName = data.FullName.ValueString()
	}

	user, _, err := r.client.AdminCreateUser(createOpts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating User",
			"Could not create user, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response to model
	data.Id = types.Int64Value(user.ID)
	data.Username = types.StringValue(user.UserName)
	data.Email = types.StringValue(user.Email)
	data.FullName = types.StringValue(user.FullName)
	data.AvatarUrl = types.StringValue(user.AvatarURL)
	data.IsAdmin = types.BoolValue(user.IsAdmin)
	data.Active = types.BoolValue(user.IsActive)
	data.Description = types.StringValue(user.Description)
	data.Location = types.StringValue(user.Location)
	data.Website = types.StringValue(user.Website)
	data.Language = types.StringValue(user.Language)
	data.Visibility = types.StringValue(string(user.Visibility))
	data.Created = types.StringValue(user.Created.String())
	data.LastLogin = types.StringValue(user.LastLogin.String())
	data.ProhibitLogin = types.BoolValue(user.ProhibitLogin)
	data.Restricted = types.BoolValue(user.Restricted)
	data.HtmlUrl = types.StringValue("")
	data.Login = types.StringValue(user.UserName)
	if user.LoginName != "" {
		data.LoginName = types.StringValue(user.LoginName)
	} else if data.LoginName.IsNull() || data.LoginName.IsUnknown() {
		data.LoginName = types.StringValue("empty")
	}
	data.SourceId = types.Int64Value(user.SourceID)
	data.FollowersCount = types.Int64Null()
	data.FollowingCount = types.Int64Null()
	data.StarredReposCount = types.Int64Null()
	data.SendNotify = types.BoolNull()
	data.MustChangePassword = types.BoolNull()
	data.CreatedAt = types.StringNull()

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resource_user.UserModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get user from Gitea API
	user, _, err := r.client.GetUserInfo(data.Username.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading User",
			"Could not read user "+data.Username.ValueString()+": "+err.Error(),
		)
		return
	}

	// Map response to model
	data.Id = types.Int64Value(user.ID)
	data.Username = types.StringValue(user.UserName)
	data.Email = types.StringValue(user.Email)
	data.FullName = types.StringValue(user.FullName)
	data.AvatarUrl = types.StringValue(user.AvatarURL)
	data.IsAdmin = types.BoolValue(user.IsAdmin)
	data.Active = types.BoolValue(user.IsActive)
	data.Description = types.StringValue(user.Description)
	data.Location = types.StringValue(user.Location)
	data.Website = types.StringValue(user.Website)
	data.Language = types.StringValue(user.Language)
	data.Visibility = types.StringValue(string(user.Visibility))
	data.Created = types.StringValue(user.Created.String())
	data.LastLogin = types.StringValue(user.LastLogin.String())
	data.ProhibitLogin = types.BoolValue(user.ProhibitLogin)
	data.Restricted = types.BoolValue(user.Restricted)
	data.HtmlUrl = types.StringValue("")
	data.Login = types.StringValue(user.UserName)
	if user.LoginName != "" {
		data.LoginName = types.StringValue(user.LoginName)
	} else if data.LoginName.IsNull() || data.LoginName.IsUnknown() {
		data.LoginName = types.StringValue("empty")
	}
	data.SourceId = types.Int64Value(user.SourceID)
	data.FollowersCount = types.Int64Null()
	data.FollowingCount = types.Int64Null()
	data.StarredReposCount = types.Int64Null()
	data.SendNotify = types.BoolNull()
	data.MustChangePassword = types.BoolNull()
	data.CreatedAt = types.StringNull()

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data resource_user.UserModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update user via Gitea API
	editOpts := gitea.EditUserOption{
		Email:    data.Email.ValueStringPointer(),
		FullName: data.FullName.ValueStringPointer(),
		Active:   data.Active.ValueBoolPointer(),
	}

	_, err := r.client.AdminEditUser(data.Username.ValueString(), editOpts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating User",
			"Could not update user "+data.Username.ValueString()+": "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resource_user.UserModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete user via Gitea API
	_, err := r.client.AdminDeleteUser(data.Username.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting User",
			"Could not delete user "+data.Username.ValueString()+": "+err.Error(),
		)
		return
	}
}

// Removed callGiteaUserResourceAPI - no longer needed
