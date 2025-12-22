package provider

import (
	"context"
	"terraform-provider-gitea/internal/datasource_user"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = (*userDataSource)(nil)

func NewUserDataSource() datasource.DataSource {
	return &userDataSource{}
}

type userDataSource struct{}

func (d *userDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *userDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_user.UserDataSourceSchema(ctx)
}

func (d *userDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data datasource_user.UserModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(callGiteaUserDataSourceAPI(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Typically this method would contain logic that makes an HTTP call to a remote API, and then stores
// computed results back to the data model. For example purposes, this function just sets computed User
// values to mock values to avoid data consistency errors.
func callGiteaUserDataSourceAPI(ctx context.Context, user *datasource_user.UserModel) diag.Diagnostics {
	user.Id = types.StringValue("1")

	return nil
}
