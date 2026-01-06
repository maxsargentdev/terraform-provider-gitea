package provider

import (
	"context"
	"fmt"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = (*userResource)(nil)
var _ resource.ResourceWithConfigure = (*userResource)(nil)
var _ resource.ResourceWithImportState = (*userResource)(nil)

func NewUserResource() resource.Resource {
	return &userResource{}
}

// Helper function to map Gitea User to Terraform model
func mapUserToModel(user *gitea.User, model *userResourceModel) {
	model.Active = types.BoolValue(user.IsActive)
	model.AvatarUrl = types.StringValue(user.AvatarURL)
	model.Created = types.StringValue(user.Created.String())
	model.CreatedAt = types.StringNull()
	model.Description = types.StringValue(user.Description)
	model.Email = types.StringValue(user.Email)
	model.FollowersCount = types.Int64Null()
	model.FollowingCount = types.Int64Null()
	model.FullName = types.StringValue(user.FullName)
	model.HtmlUrl = types.StringValue("")
	model.Id = types.Int64Value(user.ID)
	model.IsAdmin = types.BoolValue(user.IsAdmin)
	model.Language = types.StringValue(user.Language)
	model.LastLogin = types.StringValue(user.LastLogin.String())
	model.Location = types.StringValue(user.Location)
	model.Login = types.StringValue(user.UserName)
	model.MustChangePassword = types.BoolNull()
	model.ProhibitLogin = types.BoolValue(user.ProhibitLogin)
	model.Restricted = types.BoolValue(user.Restricted)
	model.SendNotify = types.BoolNull()
	model.SourceId = types.Int64Value(user.SourceID)
	model.StarredReposCount = types.Int64Null()
	model.Username = types.StringValue(user.UserName)
	model.Visibility = types.StringValue(string(user.Visibility))
	model.Website = types.StringValue(user.Website)
}

type userResource struct {
	client *gitea.Client
}

type userResourceModel struct {
	Active             types.Bool   `tfsdk:"active"`
	AvatarUrl          types.String `tfsdk:"avatar_url"`
	Created            types.String `tfsdk:"created"`
	CreatedAt          types.String `tfsdk:"created_at"`
	Description        types.String `tfsdk:"description"`
	Email              types.String `tfsdk:"email"`
	FollowersCount     types.Int64  `tfsdk:"followers_count"`
	FollowingCount     types.Int64  `tfsdk:"following_count"`
	FullName           types.String `tfsdk:"full_name"`
	HtmlUrl            types.String `tfsdk:"html_url"`
	Id                 types.Int64  `tfsdk:"id"`
	IsAdmin            types.Bool   `tfsdk:"is_admin"`
	Language           types.String `tfsdk:"language"`
	LastLogin          types.String `tfsdk:"last_login"`
	Location           types.String `tfsdk:"location"`
	Login              types.String `tfsdk:"login"`
	MustChangePassword types.Bool   `tfsdk:"must_change_password"`
	Password           types.String `tfsdk:"password"`
	ProhibitLogin      types.Bool   `tfsdk:"prohibit_login"`
	Restricted         types.Bool   `tfsdk:"restricted"`
	SendNotify         types.Bool   `tfsdk:"send_notify"`
	SourceId           types.Int64  `tfsdk:"source_id"`
	StarredReposCount  types.Int64  `tfsdk:"starred_repos_count"`
	Username           types.String `tfsdk:"username"`
	Visibility         types.String `tfsdk:"visibility"`
	Website            types.String `tfsdk:"website"`
}

func (r *userResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *userResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{

			// required - these are fundamental configuration options
			"email": schema.StringAttribute{
				Required:            true,
				Description:         "The email of the user to create.",
				MarkdownDescription: "The email address of the user to create.",
			},
			"username": schema.StringAttribute{
				Required:            true,
				Description:         "Username of the user",
				MarkdownDescription: "Username of the user",
			},
			"password": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				Description:         "The plain text password for the user. This is write-only and cannot be read back.",
				MarkdownDescription: "The plain text password for the user. This is write-only and cannot be read back.",
			},

			// optional - these tweak the created resource away from its defaults
			"active": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Is user active (can login)",
				MarkdownDescription: "Is user active (can login)",
			},
			"created_at": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "For explicitly setting the user creation timestamp. Useful when users are\nmigrated from other systems. When omitted, the user's creation timestamp\nwill be set to \"now\".",
				MarkdownDescription: "For explicitly setting the user creation timestamp. Useful when users are\nmigrated from other systems. When omitted, the user's creation timestamp\nwill be set to \"now\".",
			},
			"full_name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The full display name of the user",
				MarkdownDescription: "The full display name of the user",
			},
			"must_change_password": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether the user must change password on first login",
				MarkdownDescription: "Whether the user must change password on first login",
			},
			"restricted": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether the user has restricted access privileges",
				MarkdownDescription: "Whether the user has restricted access privileges",
			},
			"send_notify": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether to send welcome notification email to the user",
				MarkdownDescription: "Whether to send welcome notification email to the user",
			},
			"source_id": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Description:         "The authentication source ID to associate with the user",
				MarkdownDescription: "The authentication source ID to associate with the user",
			},
			"visibility": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "User visibility level: public, limited, or private",
				MarkdownDescription: "User visibility level: public, limited, or private",
			},

			// computed - these are available to read back after creation but are really just metadata
			"avatar_url": schema.StringAttribute{
				Computed:            true,
				Description:         "URL to the user's avatar",
				MarkdownDescription: "URL to the user's avatar",
			},
			"created": schema.StringAttribute{
				Computed: true,
			},
			"description": schema.StringAttribute{
				Computed:            true,
				Description:         "the user's description",
				MarkdownDescription: "the user's description",
			},
			"followers_count": schema.Int64Attribute{
				Computed:            true,
				Description:         "user counts",
				MarkdownDescription: "user counts",
			},
			"following_count": schema.Int64Attribute{
				Computed: true,
			},
			"html_url": schema.StringAttribute{
				Computed:            true,
				Description:         "URL to the user's gitea page",
				MarkdownDescription: "URL to the user's gitea page",
			},
			"id": schema.Int64Attribute{
				Computed:            true,
				Description:         "the user's id",
				MarkdownDescription: "the user's id",
			},
			"is_admin": schema.BoolAttribute{
				Computed:            true,
				Description:         "Is the user an administrator",
				MarkdownDescription: "Is the user an administrator",
			},
			"language": schema.StringAttribute{
				Computed:            true,
				Description:         "User locale",
				MarkdownDescription: "User locale",
			},
			"last_login": schema.StringAttribute{
				Computed: true,
			},
			"location": schema.StringAttribute{
				Computed:            true,
				Description:         "the user's location",
				MarkdownDescription: "the user's location",
			},
			"login": schema.StringAttribute{
				Computed:            true,
				Description:         "login of the user, same as `username`",
				MarkdownDescription: "login of the user, same as `username`",
			},
			"prohibit_login": schema.BoolAttribute{
				Computed:            true,
				Description:         "Is user login prohibited",
				MarkdownDescription: "Is user login prohibited",
			},
			"starred_repos_count": schema.Int64Attribute{
				Computed: true,
			},
			"website": schema.StringAttribute{
				Computed:            true,
				Description:         "the user's website",
				MarkdownDescription: "the user's website",
			},
		},
	}
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
	var data userResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve plan values for fields not returned by API or set during creation
	sendNotify := data.SendNotify
	mustChangePassword := data.MustChangePassword
	restricted := data.Restricted
	visibility := data.Visibility
	sourceId := data.SourceId
	createdAt := data.CreatedAt

	// Create user via Gitea API
	createOpts := gitea.CreateUserOption{
		Username:           data.Username.ValueString(),
		Email:              data.Email.ValueString(),
		Password:           data.Password.ValueString(),
		MustChangePassword: data.MustChangePassword.ValueBoolPointer(),
		SendNotify:         data.SendNotify.ValueBool(),
		Visibility:         (*gitea.VisibleType)(data.Visibility.ValueStringPointer()),
		SourceID:           data.SourceId.ValueInt64(),
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

	// If restricted was set in the plan, we need to update the user after creation
	// because CreateUserOption doesn't support the restricted field
	if !restricted.IsNull() && !restricted.IsUnknown() {
		editOpts := gitea.EditUserOption{
			LoginName:  data.Username.ValueString(),
			Restricted: restricted.ValueBoolPointer(),
		}
		_, err := r.client.AdminEditUser(data.Username.ValueString(), editOpts)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Setting User Restricted Status",
				"User was created but could not set restricted status: "+err.Error(),
			)
			return
		}

		// Re-read the user to get updated values
		user, _, err = r.client.GetUserInfo(data.Username.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading User After Setting Restricted",
				"Could not read user "+data.Username.ValueString()+": "+err.Error(),
			)
			return
		}
	}

	// Map response to model
	mapUserToModel(user, &data)

	// Restore plan values for fields not returned by API or that we want to preserve
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
	if !visibility.IsUnknown() && !visibility.IsNull() {
		data.Visibility = visibility
	}
	if !sourceId.IsUnknown() && !sourceId.IsNull() {
		data.SourceId = sourceId
	}
	if !createdAt.IsUnknown() && !createdAt.IsNull() {
		data.CreatedAt = createdAt
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data userResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve state values for fields not returned by API
	sendNotify := data.SendNotify
	mustChangePassword := data.MustChangePassword
	restricted := data.Restricted
	visibility := data.Visibility
	sourceId := data.SourceId
	createdAt := data.CreatedAt

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
	if !restricted.IsUnknown() && !restricted.IsNull() {
		data.Restricted = restricted
	}
	if !visibility.IsUnknown() && !visibility.IsNull() {
		data.Visibility = visibility
	}
	if !sourceId.IsUnknown() && !sourceId.IsNull() {
		data.SourceId = sourceId
	}
	if !createdAt.IsUnknown() && !createdAt.IsNull() {
		data.CreatedAt = createdAt
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data userResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve plan values for fields not returned by API
	sendNotify := data.SendNotify
	mustChangePassword := data.MustChangePassword
	restricted := data.Restricted
	visibility := data.Visibility
	sourceId := data.SourceId
	createdAt := data.CreatedAt

	// Update user via Gitea API
	editOpts := gitea.EditUserOption{
		Email:              data.Email.ValueStringPointer(),
		FullName:           data.FullName.ValueStringPointer(),
		LoginName:          data.Username.ValueString(),
		Restricted:         data.Restricted.ValueBoolPointer(),
		Visibility:         (*gitea.VisibleType)(data.Visibility.ValueStringPointer()),
		SourceID:           data.SourceId.ValueInt64(),
		Password:           data.Password.ValueString(),
		MustChangePassword: data.MustChangePassword.ValueBoolPointer(),
	}

	// Only set Active if explicitly provided (not null/unknown)
	if !data.Active.IsNull() && !data.Active.IsUnknown() {
		editOpts.Active = data.Active.ValueBoolPointer()
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
	if !restricted.IsUnknown() && !restricted.IsNull() {
		data.Restricted = restricted
	}
	if !visibility.IsUnknown() && !visibility.IsNull() {
		data.Visibility = visibility
	}
	if !sourceId.IsUnknown() && !sourceId.IsNull() {
		data.SourceId = sourceId
	}
	if !createdAt.IsUnknown() && !createdAt.IsNull() {
		data.CreatedAt = createdAt
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data userResourceModel

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
	var data userResourceModel
	mapUserToModel(user, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
