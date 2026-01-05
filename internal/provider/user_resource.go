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
var _ resource.ResourceWithImportState = (*userResource)(nil)

func NewUserResource() resource.Resource {
	return &userResource{}
}

// Helper function to map Gitea User to Terraform model
func mapUserToModel(user *gitea.User, model *resource_user.UserModel) {
	model.Id = types.Int64Value(user.ID)
	model.Username = types.StringValue(user.UserName)
	model.Email = types.StringValue(user.Email)
	model.FullName = types.StringValue(user.FullName)
	model.AvatarUrl = types.StringValue(user.AvatarURL)
	model.IsAdmin = types.BoolValue(user.IsAdmin)
	model.Active = types.BoolValue(user.IsActive)
	model.Description = types.StringValue(user.Description)
	model.Location = types.StringValue(user.Location)
	model.Website = types.StringValue(user.Website)
	model.Language = types.StringValue(user.Language)
	model.Visibility = types.StringValue(string(user.Visibility))
	model.Created = types.StringValue(user.Created.String())
	model.LastLogin = types.StringValue(user.LastLogin.String())
	model.ProhibitLogin = types.BoolValue(user.ProhibitLogin)
	model.Restricted = types.BoolValue(user.Restricted)
	model.HtmlUrl = types.StringValue("")
	model.Login = types.StringValue(user.UserName)
	model.SourceId = types.Int64Value(user.SourceID)
	model.FollowersCount = types.Int64Null()
	model.FollowingCount = types.Int64Null()
	model.StarredReposCount = types.Int64Null()
	model.SendNotify = types.BoolNull()
	model.MustChangePassword = types.BoolNull()
	model.CreatedAt = types.StringNull()
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

	// Preserve plan values for fields not returned by API
	sendNotify := data.SendNotify
	mustChangePassword := data.MustChangePassword

	// Create user via Gitea API
	createOpts := gitea.CreateUserOption{
		Username:           data.Username.ValueString(),
		Email:              data.Email.ValueString(),
		Password:           data.Password.ValueString(),
		MustChangePassword: data.MustChangePassword.ValueBoolPointer(),
		SendNotify:         data.SendNotify.ValueBool(),
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
	mapUserToModel(user, &data)

	// Restore plan values for fields not returned by API, but only if they were specified
	if !sendNotify.IsUnknown() {
		data.SendNotify = sendNotify
	} else {
		data.SendNotify = types.BoolNull()
	}
	if !mustChangePassword.IsUnknown() {
		data.MustChangePassword = mustChangePassword
	} else {
		data.MustChangePassword = types.BoolNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resource_user.UserModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve state values for fields not returned by API
	sendNotify := data.SendNotify
	mustChangePassword := data.MustChangePassword

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
	mapUserToModel(user, &data)

	// Restore state values for fields not returned by API, but only if they were known
	if !sendNotify.IsUnknown() {
		data.SendNotify = sendNotify
	} else {
		data.SendNotify = types.BoolNull()
	}
	if !mustChangePassword.IsUnknown() {
		data.MustChangePassword = mustChangePassword
	} else {
		data.MustChangePassword = types.BoolNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data resource_user.UserModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve plan values for fields not returned by API
	sendNotify := data.SendNotify
	mustChangePassword := data.MustChangePassword

	// Update user via Gitea API
	editOpts := gitea.EditUserOption{
		Email:     data.Email.ValueStringPointer(),
		FullName:  data.FullName.ValueStringPointer(),
		Active:    data.Active.ValueBoolPointer(),
		LoginName: data.Username.ValueString(),
	}

	_, err := r.client.AdminEditUser(data.Username.ValueString(), editOpts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating User",
			"Could not update user "+data.Username.ValueString()+": "+err.Error(),
		)
		return
	}

	// Read back the user to get updated values
	user, _, err := r.client.GetUserInfo(data.Username.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading User After Update",
			"Could not read user "+data.Username.ValueString()+": "+err.Error(),
		)
		return
	}

	// Map response to model
	mapUserToModel(user, &data)

	// Restore plan values for fields not returned by API, but only if they were specified
	if !sendNotify.IsUnknown() {
		data.SendNotify = sendNotify
	} else {
		data.SendNotify = types.BoolNull()
	}
	if !mustChangePassword.IsUnknown() {
		data.MustChangePassword = mustChangePassword
	} else {
		data.MustChangePassword = types.BoolNull()
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

func (r *userResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import using the username
	username := req.ID

	// Fetch the user from Gitea
	user, _, err := r.client.GetUserInfo(username)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Importing User",
			"Could not import user "+username+": "+err.Error(),
		)
		return
	}

	// Map to model
	var data resource_user.UserModel
	mapUserToModel(user, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Removed callGiteaUserResourceAPI - no longer needed
