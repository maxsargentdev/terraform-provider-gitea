package provider

import (
	"context"
	"fmt"
	"net/http"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = (*userDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*userDataSource)(nil)

func NewUserDataSource() datasource.DataSource {
	return &userDataSource{}
}

// Helper function to map Gitea User to Terraform data source model
func mapUserToDataSourceModel(user *gitea.User, model *userDataSourceModel) {
	model.Id = types.Int64Value(user.ID)
	model.Username = types.StringValue(user.UserName)
	model.Login = types.StringValue(user.UserName)
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
	if user.LoginName != "" {
		model.LoginName = types.StringValue(user.LoginName)
	} else {
		model.LoginName = types.StringNull()
	}
	model.SourceId = types.Int64Value(user.SourceID)
	model.FollowersCount = types.Int64Null()
	model.FollowingCount = types.Int64Null()
	model.StarredReposCount = types.Int64Null()
}

type userDataSource struct {
	client *gitea.Client
}

type userDataSourceModel struct {
	Id                types.Int64  `tfsdk:"id"`
	Active            types.Bool   `tfsdk:"active"`
	AvatarUrl         types.String `tfsdk:"avatar_url"`
	Created           types.String `tfsdk:"created"`
	Description       types.String `tfsdk:"description"`
	Email             types.String `tfsdk:"email"`
	FollowersCount    types.Int64  `tfsdk:"followers_count"`
	FollowingCount    types.Int64  `tfsdk:"following_count"`
	FullName          types.String `tfsdk:"full_name"`
	HtmlUrl           types.String `tfsdk:"html_url"`
	Username          types.String `tfsdk:"username"`
	IsAdmin           types.Bool   `tfsdk:"is_admin"`
	Language          types.String `tfsdk:"language"`
	LastLogin         types.String `tfsdk:"last_login"`
	Location          types.String `tfsdk:"location"`
	Login             types.String `tfsdk:"login"`
	LoginName         types.String `tfsdk:"login_name"`
	ProhibitLogin     types.Bool   `tfsdk:"prohibit_login"`
	Restricted        types.Bool   `tfsdk:"restricted"`
	SourceId          types.Int64  `tfsdk:"source_id"`
	StarredReposCount types.Int64  `tfsdk:"starred_repos_count"`
	Visibility        types.String `tfsdk:"visibility"`
	Website           types.String `tfsdk:"website"`
}

func (d *userDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *userDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Use this data source to retrieve information about an existing Gitea user.",
		MarkdownDescription: "Use this data source to retrieve information about an existing Gitea user.",
		Attributes: map[string]schema.Attribute{

			// query parameters - at least one must be provided
			"username": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The username of the user to look up. Either username or id must be specified.",
				MarkdownDescription: "The username of the user to look up. Either `username` or `id` must be specified.",
			},
			"id": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Description:         "The numeric ID of the user to look up. Either username or id must be specified.",
				MarkdownDescription: "The numeric ID of the user to look up. Either `username` or `id` must be specified.",
			},

			// computed - these are available to read back after creation but are really just metadata
			"active": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether the user account is active.",
				MarkdownDescription: "Whether the user account is active.",
			},
			"avatar_url": schema.StringAttribute{
				Computed:            true,
				Description:         "URL to the user's avatar image.",
				MarkdownDescription: "URL to the user's avatar image.",
			},
			"created": schema.StringAttribute{
				Computed:            true,
				Description:         "The timestamp when the user account was created.",
				MarkdownDescription: "The timestamp when the user account was created.",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				Description:         "The user's bio/description.",
				MarkdownDescription: "The user's bio/description.",
			},
			"email": schema.StringAttribute{
				Computed:            true,
				Description:         "The user's email address.",
				MarkdownDescription: "The user's email address.",
			},
			"followers_count": schema.Int64Attribute{
				Computed:            true,
				Description:         "The number of followers the user has.",
				MarkdownDescription: "The number of followers the user has.",
			},
			"following_count": schema.Int64Attribute{
				Computed:            true,
				Description:         "The number of users this user is following.",
				MarkdownDescription: "The number of users this user is following.",
			},
			"full_name": schema.StringAttribute{
				Computed:            true,
				Description:         "The user's full display name.",
				MarkdownDescription: "The user's full display name.",
			},
			"html_url": schema.StringAttribute{
				Computed:            true,
				Description:         "URL to the user's Gitea profile page.",
				MarkdownDescription: "URL to the user's Gitea profile page.",
			},
			"is_admin": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether the user is a Gitea administrator.",
				MarkdownDescription: "Whether the user is a Gitea administrator.",
			},
			"language": schema.StringAttribute{
				Computed:            true,
				Description:         "The user's preferred language/locale.",
				MarkdownDescription: "The user's preferred language/locale.",
			},
			"last_login": schema.StringAttribute{
				Computed:            true,
				Description:         "The timestamp of the user's last login.",
				MarkdownDescription: "The timestamp of the user's last login.",
			},
			"location": schema.StringAttribute{
				Computed:            true,
				Description:         "The user's location.",
				MarkdownDescription: "The user's location.",
			},
			"login": schema.StringAttribute{
				Computed:            true,
				Description:         "The user's login name (same as username).",
				MarkdownDescription: "The user's login name (same as `username`).",
			},
			"login_name": schema.StringAttribute{
				Computed:            true,
				Description:         "The identifier provided by an external authenticator (if configured).",
				MarkdownDescription: "The identifier provided by an external authenticator (if configured).",
			},
			"prohibit_login": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether the user is prohibited from logging in.",
				MarkdownDescription: "Whether the user is prohibited from logging in.",
			},
			"restricted": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether the user has restricted permissions.",
				MarkdownDescription: "Whether the user has restricted permissions.",
			},
			"source_id": schema.Int64Attribute{
				Computed:            true,
				Description:         "The ID of the user's authentication source.",
				MarkdownDescription: "The ID of the user's authentication source.",
			},
			"starred_repos_count": schema.Int64Attribute{
				Computed:            true,
				Description:         "The number of repositories the user has starred.",
				MarkdownDescription: "The number of repositories the user has starred.",
			},
			"visibility": schema.StringAttribute{
				Computed:            true,
				Description:         "The user's visibility level (public, limited, or private).",
				MarkdownDescription: "The user's visibility level (`public`, `limited`, or `private`).",
			},
			"website": schema.StringAttribute{
				Computed:            true,
				Description:         "The user's website URL.",
				MarkdownDescription: "The user's website URL.",
			},
		},
	}
}

func (d *userDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*gitea.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *gitea.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *userDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data userDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that at least one of username or id is provided
	if data.Username.IsNull() && data.Id.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Missing Required Attribute",
			"Either 'username' or 'id' must be specified to look up a user.",
		)
		return
	}

	var user *gitea.User
	var httpResp *gitea.Response
	var err error

	if !data.Username.IsNull() {
		// Query by username
		username := data.Username.ValueString()
		user, httpResp, err = d.client.GetUserInfo(username)
		if err != nil {
			if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
				resp.Diagnostics.AddError(
					"User Not Found",
					fmt.Sprintf("User '%s' does not exist or you do not have permission to access it.", username),
				)
				return
			}
			resp.Diagnostics.AddError(
				"Error Reading User",
				fmt.Sprintf("Could not read user '%s': %s", username, err.Error()),
			)
			return
		}
	} else {
		// Query by ID - need to use admin API
		userID := data.Id.ValueInt64()
		user, httpResp, err = d.client.GetUserByID(userID)
		if err != nil {
			if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
				resp.Diagnostics.AddError(
					"User Not Found",
					fmt.Sprintf("User with ID %d does not exist or you do not have permission to access it.", userID),
				)
				return
			}
			resp.Diagnostics.AddError(
				"Error Reading User",
				fmt.Sprintf("Could not read user with ID %d: %s", userID, err.Error()),
			)
			return
		}
	}

	// Map response to model
	mapUserToDataSourceModel(user, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
