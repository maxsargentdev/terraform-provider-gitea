package provider

import (
	"context"
	"fmt"
	"terraform-provider-gitea/internal/datasource_org"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = (*orgDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*orgDataSource)(nil)

func NewOrgDataSource() datasource.DataSource {
	return &orgDataSource{}
}

// Helper function to map Gitea Organization to Terraform data source model
func mapOrgToDataSourceModel(org *gitea.Organization, model *datasource_org.OrgModel) {
	model.Id = types.Int64Value(org.ID)
	model.Org = types.StringValue(org.UserName)
	model.Username = types.StringValue(org.UserName)
	model.Name = types.StringValue(org.UserName)
	model.FullName = types.StringValue(org.FullName)
	model.Description = types.StringValue(org.Description)
	model.Website = types.StringValue(org.Website)
	model.Location = types.StringValue(org.Location)
	model.AvatarUrl = types.StringValue(org.AvatarURL)
	model.Visibility = types.StringValue(org.Visibility)
	model.RepoAdminChangeTeamAccess = types.BoolNull()
	model.Email = types.StringNull()
}

type orgDataSource struct {
	client *gitea.Client
}

func (d *orgDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_org"
}

func (d *orgDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_org.OrgDataSourceSchema(ctx)
}

func (d *orgDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *orgDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data datasource_org.OrgModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The schema uses "org" field as the org name to query
	orgName := data.Org.ValueString()

	// Get org from Gitea API
	org, _, err := d.client.GetOrg(orgName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Organization",
			"Could not read organization "+orgName+": "+err.Error(),
		)
		return
	}

	// Map response to model
	mapOrgToDataSourceModel(org, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
