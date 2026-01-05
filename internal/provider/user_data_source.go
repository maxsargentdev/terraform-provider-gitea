package provider

import (
	"context"
	"fmt"
	"github.com/maxsargendev/terraform-provider-gitea/internal/datasource_user"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = (*userDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*userDataSource)(nil)

func NewUserDataSource() datasource.DataSource {
	return &userDataSource{}
}

// Helper function to map Gitea User to Terraform data source model
func mapUserToDataSourceModel(user *gitea.User, model *datasource_user.UserModel) {
	model.Id = types.StringValue(user.UserName)
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

func (d *userDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *userDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_user.UserDataSourceSchema(ctx)
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
	var data datasource_user.UserModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The schema uses "id" field as the username to query
	username := data.Id.ValueString()

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
