package provider

import (
	"context"
	"fmt"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
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
	Id        types.Int64  `tfsdk:"id"`
	AvatarUrl types.String `tfsdk:"avatar_url"`
}

func (r *userResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *userResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manages a Gitea user.",
		MarkdownDescription: "Manages a Gitea user.",
		Attributes: map[string]schema.Attribute{
			// Required
			"username": schema.StringAttribute{
				Required:            true,
				Description:         "Username of the user.",
				MarkdownDescription: "Username of the user.",
			},
			"email": schema.StringAttribute{
				Required:            true,
				Description:         "The email address of the user",
				MarkdownDescription: "The email address of the user",
			},
			"password": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				Description:         "The plain text password for the user. This is write-only and cannot be read back.",
				MarkdownDescription: "The plain text password for the user. This is write-only and cannot be read back.",
			},

			// Optional - from CreateUserOption and EditUserOption
			"source_id": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Description:         "The authentication source ID to associate with the user.",
				MarkdownDescription: "The authentication source ID to associate with the user.",
			},
			"login_name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The login name for the authentication source.",
				MarkdownDescription: "The login name for the authentication source.",
			},
			"full_name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The full display name of the user.",
				MarkdownDescription: "The full display name of the user.",
			},
			"must_change_password": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether the user must change password on first login.",
				MarkdownDescription: "Whether the user must change password on first login.",
			},
			"send_notify": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether to send welcome notification email to the user.",
				MarkdownDescription: "Whether to send welcome notification email to the user.",
			},
			"visibility": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "User visibility level: public, limited, or private",
				MarkdownDescription: "User visibility level: public, limited, or private",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The user's description.",
				MarkdownDescription: "The user's description.",
			},
			"website": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The user's website.",
				MarkdownDescription: "The user's website.",
			},
			"location": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The user's location.",
				MarkdownDescription: "The user's location.",
			},
			"active": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Is user active (can login).",
				MarkdownDescription: "Is user active (can login).",
			},
			"admin": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Is the user an administrator.",
				MarkdownDescription: "Is the user an administrator.",
			},
			"allow_git_hook": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether the user can use git hooks.",
				MarkdownDescription: "Whether the user can use git hooks.",
			},
			"allow_import_local": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether the user can import local repositories.",
				MarkdownDescription: "Whether the user can import local repositories.",
			},
			"max_repo_creation": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Description:         "Maximum number of repositories the user can create.",
				MarkdownDescription: "Maximum number of repositories the user can create.",
			},
			"prohibit_login": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Is user login prohibited.",
				MarkdownDescription: "Is user login prohibited.",
			},
			"allow_create_organization": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether the user can create organizations.",
				MarkdownDescription: "Whether the user can create organizations.",
			},
			"restricted": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether the user has restricted access privileges.",
				MarkdownDescription: "Whether the user has restricted access privileges.",
			},

			// Computed - key outputs
			"id": schema.Int64Attribute{
				Computed:            true,
				Description:         "The user's ID.",
				MarkdownDescription: "The user's ID.",

				// ID doesnt change once set, only computed once so refer to state
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"avatar_url": schema.StringAttribute{
				Computed:            true,
				Description:         "URL to the user's avatar.",
				MarkdownDescription: "URL to the user's avatar.",
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
	user, _, err := r.client.GetUserInfo(state.Username.ValueString())
	if err != nil {
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
