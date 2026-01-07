package provider

import (
	"context"
	"fmt"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = (*userDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*userDataSource)(nil)

func NewUserDataSource() datasource.DataSource {
	return &userDataSource{}
}

// Helper function to map Gitea User to Terraform data source model
func mapUserToDataSourceModel(user *gitea.User, model *userDataSourceModel) {
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
		Attributes: map[string]schema.Attribute{

			// required - these are fundamental configuration options
			"username": schema.StringAttribute{
				Required:            true,
				Description:         "Username of the user whose data is to be listed",
				MarkdownDescription: "Username of the user whose data is to be listed",
			},

			// computed - these are available to read back after creation but are really just metadata
			"active": schema.BoolAttribute{
				Computed:            true,
				Description:         "Is the user active?",
				MarkdownDescription: "Is the user active?",
			},
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
			"email": schema.StringAttribute{
				Computed: true,
			},
			"followers_count": schema.Int64Attribute{
				Computed:            true,
				Description:         "user counts",
				MarkdownDescription: "user counts",
			},
			"following_count": schema.Int64Attribute{
				Computed: true,
			},
			"full_name": schema.StringAttribute{
				Computed:            true,
				Description:         "the user's full name",
				MarkdownDescription: "the user's full name",
			},
			"html_url": schema.StringAttribute{
				Computed:            true,
				Description:         "URL to the user's gitea page",
				MarkdownDescription: "URL to the user's gitea page",
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
			"login_name": schema.StringAttribute{
				Computed:            true,
				Description:         "identifier of the user, provided by the external authenticator (if configured)",
				MarkdownDescription: "identifier of the user, provided by the external authenticator (if configured)",
			},
			"prohibit_login": schema.BoolAttribute{
				Computed:            true,
				Description:         "Is user login prohibited",
				MarkdownDescription: "Is user login prohibited",
			},
			"restricted": schema.BoolAttribute{
				Computed:            true,
				Description:         "Is user restricted",
				MarkdownDescription: "Is user restricted",
			},
			"source_id": schema.Int64Attribute{
				Computed:            true,
				Description:         "The ID of the user's Authentication Source",
				MarkdownDescription: "The ID of the user's Authentication Source",
			},
			"starred_repos_count": schema.Int64Attribute{
				Computed: true,
			},
			"visibility": schema.StringAttribute{
				Computed:            true,
				Description:         "User visibility level option: public, limited, private",
				MarkdownDescription: "User visibility level option: public, limited, private",
			},
			"website": schema.StringAttribute{
				Computed:            true,
				Description:         "the user's website",
				MarkdownDescription: "the user's website",
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

	var username = data.Username.ValueString()

	// Get user from Gitea API
	user, _, err := d.client.GetUserInfo(username)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading User",
			"Could not read user "+username+": "+err.Error(),
		)
		return
	}

	// Map response to model
	mapUserToDataSourceModel(user, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
