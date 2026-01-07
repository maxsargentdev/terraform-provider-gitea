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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &userResource{}
	_ resource.ResourceWithConfigure   = &userResource{}
	_ resource.ResourceWithImportState = &userResource{}
)

func NewUserResource() resource.Resource {
	return &userResource{}
}

type userResource struct {
	client *gitea.Client
}

type userResourceModel struct {
	// Required
	Email     types.String `tfsdk:"email"`
	LoginName types.String `tfsdk:"login_name"`
	Password  types.String `tfsdk:"password"`
	Username  types.String `tfsdk:"username"`

	// Optional
	Active                  types.Bool   `tfsdk:"active"`
	Admin                   types.Bool   `tfsdk:"admin"`
	AllowCreateOrganization types.Bool   `tfsdk:"allow_create_organization"`
	AllowGitHook            types.Bool   `tfsdk:"allow_git_hook"`
	AllowImportLocal        types.Bool   `tfsdk:"allow_import_local"`
	Description             types.String `tfsdk:"description"`
	ForcePasswordChange     types.Bool   `tfsdk:"force_password_change"`
	FullName                types.String `tfsdk:"full_name"`
	Location                types.String `tfsdk:"location"`
	MaxRepoCreation         types.Int64  `tfsdk:"max_repo_creation"`
	MustChangePassword      types.Bool   `tfsdk:"must_change_password"`
	ProhibitLogin           types.Bool   `tfsdk:"prohibit_login"`
	Restricted              types.Bool   `tfsdk:"restricted"`
	SendNotification        types.Bool   `tfsdk:"send_notification"`
	Visibility              types.String `tfsdk:"visibility"`

	// Computed
	Id types.String `tfsdk:"id"`
}

func (r *userResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *userResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manages a Gitea user account.",
		MarkdownDescription: "Manages a Gitea user account. This resource allows you to create, update, and delete user accounts in your Gitea instance. Requires admin privileges.",
		Attributes: map[string]schema.Attribute{
			// Required
			"email": schema.StringAttribute{
				Required:            true,
				Description:         "E-Mail Address of the user.",
				MarkdownDescription: "E-Mail Address of the user.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`),
						"must be a valid email address",
					),
				},
			},
			"login_name": schema.StringAttribute{
				Required:            true,
				Description:         "The login name can differ from the username.",
				MarkdownDescription: "The login name can differ from the username.",
			},
			"password": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				Description:         "Password to be set for the user.",
				MarkdownDescription: "Password to be set for the user.",
			},
			"username": schema.StringAttribute{
				Required:            true,
				Description:         "Username of the user to be created.",
				MarkdownDescription: "Username of the user to be created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			// Optional
			"active": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Flag indicating if the user account should be enabled.",
				MarkdownDescription: "Flag indicating if the user account should be enabled.",
			},
			"admin": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Flag indicating if the user should have administrator privileges.",
				MarkdownDescription: "Flag indicating if the user should have administrator privileges.",
			},
			"allow_create_organization": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether the user is allowed to create organizations.",
				MarkdownDescription: "Whether the user is allowed to create organizations.",
			},
			"allow_git_hook": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether the user is allowed to create Git hooks.",
				MarkdownDescription: "Whether the user is allowed to create Git hooks.",
			},
			"allow_import_local": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether the user is allowed to import local repositories.",
				MarkdownDescription: "Whether the user is allowed to import local repositories.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "A description of the user.",
				MarkdownDescription: "A description of the user.",
			},
			"force_password_change": schema.BoolAttribute{
				Optional:            true,
				Description:         "Flag if the user defined password should be overwritten or not.",
				MarkdownDescription: "Flag if the user defined password should be overwritten or not.",
			},
			"full_name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Full name of the user.",
				MarkdownDescription: "Full name of the user.",
			},
			"location": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The location of the user.",
				MarkdownDescription: "The location of the user.",
			},
			"max_repo_creation": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Description:         "The maximum number of repositories this user can create. -1 for unlimited.",
				MarkdownDescription: "The maximum number of repositories this user can create. `-1` for unlimited.",
			},
			"must_change_password": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Flag if the user should change the password after first login.",
				MarkdownDescription: "Flag if the user should change the password after first login.",
			},
			"prohibit_login": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Flag if the user should not be allowed to log in (bot user).",
				MarkdownDescription: "Flag if the user should not be allowed to log in (bot user).",
			},
			"restricted": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether the user has restricted access.",
				MarkdownDescription: "Whether the user has restricted access.",
			},
			"send_notification": schema.BoolAttribute{
				Optional:            true,
				Description:         "Flag to send a notification about the user creation to the defined email.",
				MarkdownDescription: "Flag to send a notification about the user creation to the defined email.",
			},
			"visibility": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Visibility of the user. Can be public, limited or private.",
				MarkdownDescription: "Visibility of the user. Can be `public`, `limited` or `private`.",
				Validators: []validator.String{
					stringvalidator.OneOf("public", "limited", "private"),
				},
			},

			// Computed
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "The ID of this resource.",
				MarkdownDescription: "The ID of this resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Helper function to map Gitea User to Terraform model
func (r *userResource) mapUserToModel(user *gitea.User, model *userResourceModel) {
	model.Id = types.StringValue(fmt.Sprintf("%d", user.ID))
	model.Username = types.StringValue(user.UserName)
	model.Email = types.StringValue(user.Email)
	model.FullName = types.StringValue(user.FullName)
	model.Description = types.StringValue(user.Description)
	model.Location = types.StringValue(user.Location)
	model.Active = types.BoolValue(user.IsActive)
	model.Admin = types.BoolValue(user.IsAdmin)
	model.ProhibitLogin = types.BoolValue(user.ProhibitLogin)
	model.Restricted = types.BoolValue(user.Restricted)
	model.Visibility = types.StringValue(string(user.Visibility))

	// Note: login_name and max_repo_creation are not returned by GET user API
	// We preserve them from existing state
	if model.LoginName.IsUnknown() {
		model.LoginName = types.StringNull()
	}
	if model.MaxRepoCreation.IsUnknown() {
		model.MaxRepoCreation = types.Int64Null()
	}
	if model.MustChangePassword.IsUnknown() {
		model.MustChangePassword = types.BoolNull()
	}
	if model.AllowGitHook.IsUnknown() {
		model.AllowGitHook = types.BoolNull()
	}
	if model.AllowImportLocal.IsUnknown() {
		model.AllowImportLocal = types.BoolNull()
	}
	if model.AllowCreateOrganization.IsUnknown() {
		model.AllowCreateOrganization = types.BoolNull()
	}
	if model.Password.IsUnknown() {
		model.Password = types.StringNull()
	}
	if model.SendNotification.IsUnknown() {
		model.SendNotification = types.BoolNull()
	}
	if model.ForcePasswordChange.IsUnknown() {
		model.ForcePasswordChange = types.BoolNull()
	}
}

func (r *userResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
		LoginName:          plan.LoginName.ValueString(),
		FullName:           plan.FullName.ValueString(),
		MustChangePassword: plan.MustChangePassword.ValueBoolPointer(),
		SendNotify:         plan.SendNotification.ValueBool(),
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
	hasEditFields := !plan.Description.IsNull() || !plan.Location.IsNull() ||
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
	r.mapUserToModel(user, &plan)

	// Preserve fields that are not returned by the API
	// (password, login_name, send_notification, force_password_change are preserved from plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state userResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve values from state that API doesn't return
	preservePassword := state.Password
	preserveLoginName := state.LoginName
	preserveSendNotification := state.SendNotification
	preserveForcePasswordChange := state.ForcePasswordChange
	preserveMaxRepoCreation := state.MaxRepoCreation

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
	r.mapUserToModel(user, &state)

	// Restore preserved values
	state.Password = preservePassword
	state.LoginName = preserveLoginName
	state.SendNotification = preserveSendNotification
	state.ForcePasswordChange = preserveForcePasswordChange
	state.MaxRepoCreation = preserveMaxRepoCreation

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan userResourceModel
	var state userResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update user via Gitea API
	editOpts := gitea.EditUserOption{
		LoginName:               plan.LoginName.ValueString(),
		Email:                   plan.Email.ValueStringPointer(),
		FullName:                plan.FullName.ValueStringPointer(),
		Description:             plan.Description.ValueStringPointer(),
		MustChangePassword:      plan.MustChangePassword.ValueBoolPointer(),
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

	// Only set password if force_password_change is true or password changed
	if plan.ForcePasswordChange.ValueBool() || (plan.Password.ValueString() != state.Password.ValueString()) {
		editOpts.Password = plan.Password.ValueString()
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
	r.mapUserToModel(user, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete user via Gitea API
	_, err := r.client.AdminDeleteUser(state.Username.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting User",
			"Could not delete user "+state.Username.ValueString()+": "+err.Error(),
		)
		return
	}
}

func (r *userResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import using the username
	username := req.ID

	// Fetch the user from Gitea
	user, httpResp, err := r.client.GetUserInfo(username)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			resp.Diagnostics.AddError(
				"User Not Found",
				fmt.Sprintf("User '%s' does not exist or is not accessible", username),
			)
			return
		}
		resp.Diagnostics.AddError(
			"Error Importing User",
			"Could not import user "+username+": "+err.Error(),
		)
		return
	}

	// Map to model
	var data userResourceModel
	r.mapUserToModel(user, &data)

	// Set login_name to username as default for imports
	data.LoginName = types.StringValue(user.UserName)

	// These fields are required but not available from API, set to null
	data.Password = types.StringNull()
	data.SendNotification = types.BoolNull()
	data.ForcePasswordChange = types.BoolNull()

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
