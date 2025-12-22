package provider

import (
	"context"
	"terraform-provider-gitea/internal/resource_user"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = (*userResource)(nil)

func NewUserResource() resource.Resource {
	return &userResource{}
}

type userResource struct{}

func (r *userResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *userResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_user.UserResourceSchema(ctx)
}

func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resource_user.UserModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(callGiteaUserResourceAPI(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resource_user.UserModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data resource_user.UserModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(callGiteaUserResourceAPI(ctx, &data)...)
	if resp.Diagnostics.HasError() {
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
}

// Typically this method would contain logic that makes an HTTP call to a remote API, and then stores
// computed results back to the data model. For example purposes, this function just sets all unknown
// User values to null to avoid data consistency errors.
func callGiteaUserResourceAPI(ctx context.Context, user *resource_user.UserModel) diag.Diagnostics {
	if user.Id.IsUnknown() {
		user.Id = types.Int64Null()
	}

	if user.Email.IsUnknown() {
		user.Email = types.StringNull()
	}

	if user.Active.IsUnknown() {
		user.Active = types.BoolNull()
	}

	if user.AvatarUrl.IsUnknown() {
		user.AvatarUrl = types.StringNull()
	}

	if user.Created.IsUnknown() {
		user.Created = types.StringNull()
	}

	if user.CreatedAt.IsUnknown() {
		user.CreatedAt = types.StringNull()
	}

	if user.Description.IsUnknown() {
		user.Description = types.StringNull()
	}

	if user.FollowersCount.IsUnknown() {
		user.FollowersCount = types.Int64Null()
	}

	if user.FollowingCount.IsUnknown() {
		user.FollowingCount = types.Int64Null()
	}

	if user.FullName.IsUnknown() {
		user.FullName = types.StringNull()
	}

	if user.HtmlUrl.IsUnknown() {
		user.HtmlUrl = types.StringNull()
	}

	if user.IsAdmin.IsUnknown() {
		user.IsAdmin = types.BoolNull()
	}

	if user.Language.IsUnknown() {
		user.Language = types.StringNull()
	}

	if user.LastLogin.IsUnknown() {
		user.LastLogin = types.StringNull()
	}

	if user.Location.IsUnknown() {
		user.Location = types.StringNull()
	}

	if user.Login.IsUnknown() {
		user.Login = types.StringNull()
	}

	if user.LoginName.IsUnknown() {
		user.LoginName = types.StringNull()
	}

	if user.MustChangePassword.IsUnknown() {
		user.MustChangePassword = types.BoolNull()
	}

	if user.Password.IsUnknown() {
		user.Password = types.StringNull()
	}

	if user.ProhibitLogin.IsUnknown() {
		user.ProhibitLogin = types.BoolNull()
	}

	if user.Restricted.IsUnknown() {
		user.Restricted = types.BoolNull()
	}

	if user.SendNotify.IsUnknown() {
		user.SendNotify = types.BoolNull()
	}

	if user.SourceId.IsUnknown() {
		user.SourceId = types.Int64Null()
	}

	if user.StarredReposCount.IsUnknown() {
		user.StarredReposCount = types.Int64Null()
	}

	if user.Username.IsUnknown() {
		user.Username = types.StringNull()
	}

	if user.Visibility.IsUnknown() {
		user.Visibility = types.StringNull()
	}

	if user.Website.IsUnknown() {
		user.Website = types.StringNull()
	}

	return nil
}
