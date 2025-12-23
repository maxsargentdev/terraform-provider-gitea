package provider

import (
	"context"
	"fmt"
	"terraform-provider-gitea/internal/datasource_user"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = (*userDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*userDataSource)(nil)

func NewUserDataSource() datasource.DataSource {
	return &userDataSource{}
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
	data.Id = types.StringValue(user.UserName)
	data.Login = types.StringValue(user.UserName)
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

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
