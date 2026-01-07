package provider

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
	// Computed fields
	model.Id = types.Int64Value(user.ID)
	model.AvatarUrl = types.StringValue(user.AvatarURL)
	model.Created = types.StringValue(user.Created.Format("2006-01-02T15:04:05Z07:00"))
	model.LastLogin = types.StringValue(user.LastLogin.Format("2006-01-02T15:04:05Z07:00"))
	model.Language = types.StringValue(user.Language)
	model.FollowersCount = types.Int64Value(int64(user.FollowerCount))
	model.FollowingCount = types.Int64Value(int64(user.FollowingCount))
	model.StarredReposCount = types.Int64Value(int64(user.StarredRepoCount))

	// Optional fields returned by API
	model.Username = types.StringValue(user.UserName)
	model.Email = types.StringValue(user.Email)
	model.FullName = types.StringValue(user.FullName)
	model.Description = types.StringValue(user.Description)
	model.Website = types.StringValue(user.Website)
	model.Location = types.StringValue(user.Location)
	model.Active = types.BoolValue(user.IsActive)
	model.Admin = types.BoolValue(user.IsAdmin)
	model.ProhibitLogin = types.BoolValue(user.ProhibitLogin)
	model.Restricted = types.BoolValue(user.Restricted)
	model.Visibility = types.StringValue(string(user.Visibility))
	model.SourceId = types.Int64Value(user.SourceID)

	// Creation-only fields - preserve from existing model, convert Unknown to null
	if model.Password.IsUnknown() {
		model.Password = types.StringNull()
	}
	if model.LoginName.IsUnknown() {
		model.LoginName = types.StringNull()
	}
	if model.MustChangePassword.IsUnknown() {
		model.MustChangePassword = types.BoolNull()
	}
	if model.SendNotify.IsUnknown() {
		model.SendNotify = types.BoolNull()
	}
	if model.AllowGitHook.IsUnknown() {
		model.AllowGitHook = types.BoolNull()
	}
	if model.AllowImportLocal.IsUnknown() {
		model.AllowImportLocal = types.BoolNull()
	}
	if model.MaxRepoCreation.IsUnknown() {
		model.MaxRepoCreation = types.Int64Null()
	}
	if model.AllowCreateOrganization.IsUnknown() {
		model.AllowCreateOrganization = types.BoolNull()
	}
}

type userResource struct {
	client *gitea.Client
}

type userResourceModel struct {
	// Required
	Username types.String `tfsdk:"username"`
	Email    types.String `tfsdk:"email"`
	Password types.String `tfsdk:"password"`

	// Optional - from CreateUserOption and EditUserOption
	SourceId           types.Int64  `tfsdk:"source_id"`
	LoginName          types.String `tfsdk:"login_name"`
	FullName           types.String `tfsdk:"full_name"`
	MustChangePassword types.Bool   `tfsdk:"must_change_password"`
	SendNotify         types.Bool   `tfsdk:"send_notify"`
	Visibility         types.String `tfsdk:"visibility"`

	// Optional - from EditUserOption only
	Description             types.String `tfsdk:"description"`
	Website                 types.String `tfsdk:"website"`
	Location                types.String `tfsdk:"location"`
	Active                  types.Bool   `tfsdk:"active"`
	Admin                   types.Bool   `tfsdk:"admin"`
	AllowGitHook            types.Bool   `tfsdk:"allow_git_hook"`
	AllowImportLocal        types.Bool   `tfsdk:"allow_import_local"`
	MaxRepoCreation         types.Int64  `tfsdk:"max_repo_creation"`
	ProhibitLogin           types.Bool   `tfsdk:"prohibit_login"`
	AllowCreateOrganization types.Bool   `tfsdk:"allow_create_organization"`
	Restricted              types.Bool   `tfsdk:"restricted"`

	// Computed - key outputs
	Id               types.Int64  `tfsdk:"id"`
	AvatarUrl        types.String `tfsdk:"avatar_url"`
	Created          types.String `tfsdk:"created"`
	LastLogin        types.String `tfsdk:"last_login"`
	Language         types.String `tfsdk:"language"`
	FollowersCount   types.Int64  `tfsdk:"followers_count"`
	FollowingCount   types.Int64  `tfsdk:"following_count"`
	StarredReposCount types.Int64  `tfsdk:"starred_repos_count"`
}

func (r *userResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *userResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manages a Gitea user account. This resource allows you to create, update, and delete user accounts in your Gitea instance. Requires admin privileges.",
		MarkdownDescription: "Manages a **Gitea user account**. This resource allows you to create, update, and delete user accounts in your Gitea instance. Requires admin privileges.",
		Attributes: map[string]schema.Attribute{
			// Required
			"username": schema.StringAttribute{
				Required:            true,
				Description:         "The unique username for the user account. This is used for login and is displayed publicly. Cannot be changed after creation.",
				MarkdownDescription: "The unique **username** for the user account. This is used for login and is displayed publicly. Cannot be changed after creation.",
			},
			"email": schema.StringAttribute{
				Required:            true,
				Description:         "The primary email address associated with the user account. Must be a valid email format (e.g., user@example.com). Used for notifications and account recovery.",
				MarkdownDescription: "The primary **email address** associated with the user account. Must be a valid email format (e.g., `user@example.com`). Used for notifications and account recovery.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`),
						"must be a valid email address (e.g., user@example.com)",
					),
				},
			},
			"password": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				Description:         "The password for the user account. Must meet the password complexity requirements configured in Gitea. This value is write-only and cannot be read back from the API.",
				MarkdownDescription: "The **password** for the user account. Must meet the password complexity requirements configured in Gitea. This value is write-only and cannot be read back from the API.",
			},

			// Optional - from CreateUserOption and EditUserOption
			"source_id": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Description:         "The ID of the authentication source (e.g., LDAP, OAuth, SMTP) to associate with this user. Use 0 or leave unset for local authentication.",
				MarkdownDescription: "The ID of the **authentication source** (e.g., LDAP, OAuth, SMTP) to associate with this user. Use `0` or leave unset for local authentication.",
			},
			"login_name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The login name used for external authentication sources (e.g., LDAP DN, OAuth username). Only applicable when source_id is set to a non-local authentication source.",
				MarkdownDescription: "The **login name** used for external authentication sources (e.g., LDAP DN, OAuth username). Only applicable when `source_id` is set to a non-local authentication source.",
			},
			"full_name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The user's full display name (e.g., 'John Doe'). This is shown in the UI alongside the username.",
				MarkdownDescription: "The user's **full display name** (e.g., 'John Doe'). This is shown in the UI alongside the username.",
			},
			"must_change_password": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "If true, the user will be required to change their password upon next login. Useful for initial account setup or password reset scenarios.",
				MarkdownDescription: "If `true`, the user will be required to change their password upon next login. Useful for initial account setup or password reset scenarios.",
			},
			"send_notify": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "If true, sends a welcome notification email to the user upon account creation. Requires email settings to be configured in Gitea.",
				MarkdownDescription: "If `true`, sends a welcome notification email to the user upon account creation. Requires email settings to be configured in Gitea.",
			},
			"visibility": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The visibility level of the user's profile. Valid values are: 'public' (visible to everyone), 'limited' (visible to logged-in users only), or 'private' (visible only to the user and admins).",
				MarkdownDescription: "The **visibility level** of the user's profile. Valid values are: `public` (visible to everyone), `limited` (visible to logged-in users only), or `private` (visible only to the user and admins).",
				Validators: []validator.String{
					stringvalidator.OneOf("public", "limited", "private"),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "A short biography or description for the user's profile. Displayed on the user's public profile page.",
				MarkdownDescription: "A short **biography or description** for the user's profile. Displayed on the user's public profile page.",
			},
			"website": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The user's personal or professional website URL. Displayed on the user's public profile page.",
				MarkdownDescription: "The user's personal or professional **website URL**. Displayed on the user's public profile page.",
			},
			"location": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The user's geographic location (e.g., 'San Francisco, CA'). Displayed on the user's public profile page.",
				MarkdownDescription: "The user's geographic **location** (e.g., 'San Francisco, CA'). Displayed on the user's public profile page.",
			},
			"active": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether the user account is active. If false, the user cannot log in. Use this to temporarily disable accounts without deleting them.",
				MarkdownDescription: "Whether the user account is **active**. If `false`, the user cannot log in. Use this to temporarily disable accounts without deleting them.",
			},
			"admin": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether the user has administrator privileges. Admins have full access to all Gitea settings, users, and repositories.",
				MarkdownDescription: "Whether the user has **administrator privileges**. Admins have full access to all Gitea settings, users, and repositories.",
			},
			"allow_git_hook": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether the user is allowed to create and manage Git hooks. Git hooks can execute arbitrary code, so this should only be granted to trusted users.",
				MarkdownDescription: "Whether the user is allowed to create and manage **Git hooks**. Git hooks can execute arbitrary code, so this should only be granted to trusted users.",
			},
			"allow_import_local": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether the user is allowed to import repositories from the local filesystem of the Gitea server. This is a privileged operation that requires filesystem access.",
				MarkdownDescription: "Whether the user is allowed to **import repositories from the local filesystem** of the Gitea server. This is a privileged operation that requires filesystem access.",
			},
			"max_repo_creation": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Description:         "The maximum number of repositories this user can create. Set to -1 for unlimited. Defaults to the global setting if not specified.",
				MarkdownDescription: "The **maximum number of repositories** this user can create. Set to `-1` for unlimited. Defaults to the global setting if not specified.",
			},
			"prohibit_login": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether the user is prohibited from logging in. Unlike 'active', this is typically used for bot accounts or accounts that should only be used programmatically.",
				MarkdownDescription: "Whether the user is **prohibited from logging in**. Unlike `active`, this is typically used for bot accounts or accounts that should only be used programmatically.",
			},
			"allow_create_organization": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether the user is allowed to create new organizations. Organizations allow grouping repositories and managing team permissions.",
				MarkdownDescription: "Whether the user is allowed to **create new organizations**. Organizations allow grouping repositories and managing team permissions.",
			},
			"restricted": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether the user has restricted access. Restricted users can only see public repositories and their own repositories, and cannot interact with others.",
				MarkdownDescription: "Whether the user has **restricted access**. Restricted users can only see public repositories and their own repositories, and cannot interact with others.",
			},

			// Computed - key outputs
			"id": schema.Int64Attribute{
				Computed:            true,
				Description:         "The unique numeric identifier for this user in Gitea. This ID is stable and can be used for API operations.",
				MarkdownDescription: "The unique **numeric identifier** for this user in Gitea. This ID is stable and can be used for API operations.",

				// ID doesnt change once set, only computed once so refer to state
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"avatar_url": schema.StringAttribute{
				Computed:            true,
				Description:         "The URL to the user's avatar image. This is either a Gravatar URL based on the user's email or a custom uploaded avatar.",
				MarkdownDescription: "The URL to the user's **avatar image**. This is either a Gravatar URL based on the user's email or a custom uploaded avatar.",
			},
			"created": schema.StringAttribute{
				Computed:            true,
				Description:         "The timestamp when this user account was created, in RFC 3339 format (e.g., '2024-01-15T10:30:00Z').",
				MarkdownDescription: "The **timestamp** when this user account was created, in RFC 3339 format (e.g., `2024-01-15T10:30:00Z`).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_login": schema.StringAttribute{
				Computed:            true,
				Description:         "The timestamp of the user's most recent login, in RFC 3339 format. Useful for identifying inactive accounts.",
				MarkdownDescription: "The **timestamp** of the user's most recent login, in RFC 3339 format. Useful for identifying inactive accounts.",
			},
			"language": schema.StringAttribute{
				Computed:            true,
				Description:         "The user's preferred language/locale setting for the Gitea UI (e.g., 'en-US', 'zh-CN'). Set by the user in their profile settings.",
				MarkdownDescription: "The user's preferred **language/locale** setting for the Gitea UI (e.g., `en-US`, `zh-CN`). Set by the user in their profile settings.",
			},
			"followers_count": schema.Int64Attribute{
				Computed:            true,
				Description:         "The number of users following this user. Reflects the user's community engagement and popularity.",
				MarkdownDescription: "The number of **users following** this user. Reflects the user's community engagement and popularity.",
			},
			"following_count": schema.Int64Attribute{
				Computed:            true,
				Description:         "The number of users this user is following. Reflects the user's engagement with other community members.",
				MarkdownDescription: "The number of **users this user is following**. Reflects the user's engagement with other community members.",
			},
			"starred_repos_count": schema.Int64Attribute{
				Computed:            true,
				Description:         "The number of repositories this user has starred. Starred repositories appear in the user's starred list for quick access.",
				MarkdownDescription: "The number of **repositories this user has starred**. Starred repositories appear in the user's starred list for quick access.",
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
	var plan userResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create user via Gitea API
	createOpts := gitea.CreateUserOption{
		Username:           plan.Username.ValueString(),
		Email:              plan.Email.ValueString(),
		Password:           plan.Password.ValueString(),
		SourceID:           plan.SourceId.ValueInt64(),
		LoginName:          plan.LoginName.ValueString(),
		FullName:           plan.FullName.ValueString(),
		MustChangePassword: plan.MustChangePassword.ValueBoolPointer(),
		SendNotify:         plan.SendNotify.ValueBool(),
		Visibility:         (*gitea.VisibleType)(plan.Visibility.ValueStringPointer()),
	}

	user, _, err := r.client.AdminCreateUser(createOpts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating User",
			"Could not create user, unexpected error: "+err.Error(),
		)
		return
	}

	// Apply EditUserOption fields via update (since some fields aren't in CreateUserOption)
	editOpts := gitea.EditUserOption{
		LoginName:               plan.LoginName.ValueString(),
		Description:             plan.Description.ValueStringPointer(),
		Website:                 plan.Website.ValueStringPointer(),
		Location:                plan.Location.ValueStringPointer(),
		Active:                  plan.Active.ValueBoolPointer(),
		Admin:                   plan.Admin.ValueBoolPointer(),
		AllowGitHook:            plan.AllowGitHook.ValueBoolPointer(),
		AllowImportLocal:        plan.AllowImportLocal.ValueBoolPointer(),
		ProhibitLogin:           plan.ProhibitLogin.ValueBoolPointer(),
		AllowCreateOrganization: plan.AllowCreateOrganization.ValueBoolPointer(),
		Restricted:              plan.Restricted.ValueBoolPointer(),
	}

	if !plan.MaxRepoCreation.IsNull() && !plan.MaxRepoCreation.IsUnknown() {
		maxRepoInt := int(plan.MaxRepoCreation.ValueInt64())
		editOpts.MaxRepoCreation = &maxRepoInt
	}

	// Only call edit if we have fields to update
	hasEditFields := !plan.Description.IsNull() || !plan.Website.IsNull() || !plan.Location.IsNull() ||
		!plan.Active.IsNull() || !plan.Admin.IsNull() || !plan.AllowGitHook.IsNull() ||
		!plan.AllowImportLocal.IsNull() || !plan.MaxRepoCreation.IsNull() || !plan.ProhibitLogin.IsNull() ||
		!plan.AllowCreateOrganization.IsNull() || !plan.Restricted.IsNull()

	if hasEditFields {
		_, err = r.client.AdminEditUser(plan.Username.ValueString(), editOpts)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Updating User After Creation",
				"User was created but could not apply additional settings: "+err.Error(),
			)
			return
		}

		// Re-read the user to get updated values
		user, _, err = r.client.GetUserInfo(plan.Username.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading User After Update",
				"Could not read user "+plan.Username.ValueString()+": "+err.Error(),
			)
			return
		}
	}

	// Map response to model
	mapUserToModel(user, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state userResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get user from Gitea API
	user, response, err := r.client.GetUserInfo(state.Username.ValueString())
	if err != nil {
		// If user was deleted externally, remove from state
		if response != nil && response.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading User",
			"Could not read user "+state.Username.ValueString()+": "+err.Error(),
		)
		return
	}

	// Map response to model
	mapUserToModel(user, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan userResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update user via Gitea API
	editOpts := gitea.EditUserOption{
		LoginName:               plan.LoginName.ValueString(),
		SourceID:                plan.SourceId.ValueInt64(),
		Email:                   plan.Email.ValueStringPointer(),
		FullName:                plan.FullName.ValueStringPointer(),
		Password:                plan.Password.ValueString(),
		Description:             plan.Description.ValueStringPointer(),
		MustChangePassword:      plan.MustChangePassword.ValueBoolPointer(),
		Website:                 plan.Website.ValueStringPointer(),
		Location:                plan.Location.ValueStringPointer(),
		Active:                  plan.Active.ValueBoolPointer(),
		Admin:                   plan.Admin.ValueBoolPointer(),
		AllowGitHook:            plan.AllowGitHook.ValueBoolPointer(),
		AllowImportLocal:        plan.AllowImportLocal.ValueBoolPointer(),
		ProhibitLogin:           plan.ProhibitLogin.ValueBoolPointer(),
		AllowCreateOrganization: plan.AllowCreateOrganization.ValueBoolPointer(),
		Restricted:              plan.Restricted.ValueBoolPointer(),
		Visibility:              (*gitea.VisibleType)(plan.Visibility.ValueStringPointer()),
	}

	if !plan.MaxRepoCreation.IsNull() && !plan.MaxRepoCreation.IsUnknown() {
		maxRepoInt := int(plan.MaxRepoCreation.ValueInt64())
		editOpts.MaxRepoCreation = &maxRepoInt
	}

	_, err := r.client.AdminEditUser(plan.Username.ValueString(), editOpts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating User",
			"Could not update user "+plan.Username.ValueString()+": "+err.Error(),
		)
		return
	}

	// Read back the user to get updated values
	user, _, err := r.client.GetUserInfo(plan.Username.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading User After Update",
			"Could not read user "+plan.Username.ValueString()+": "+err.Error(),
		)
		return
	}

	// Map response to model
	mapUserToModel(user, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
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
