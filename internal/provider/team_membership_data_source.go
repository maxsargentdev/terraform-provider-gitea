package provider

import (
	"context"
	"fmt"

	"github.com/maxsargendev/terraform-provider-gitea/internal/datasource_team_membership"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var (
	_ datasource.DataSource              = &teamMembershipDataSource{}
	_ datasource.DataSourceWithConfigure = &teamMembershipDataSource{}
)

func NewTeamMembershipDataSource() datasource.DataSource {
	return &teamMembershipDataSource{}
}

type teamMembershipDataSource struct {
	client *gitea.Client
}

func (d *teamMembershipDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team_membership"
}

func (d *teamMembershipDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_team_membership.TeamMembershipDataSourceSchema(ctx)
}

func (d *teamMembershipDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *teamMembershipDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data datasource_team_membership.TeamMembershipModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	teamID := data.TeamId.ValueInt64()
	username := data.Username.ValueString()

	// Check if the user is a member of the team
	_, _, err := d.client.GetTeamMember(teamID, username)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Team Membership",
			fmt.Sprintf("Could not verify team membership for user %s in team %d: %s", username, teamID, err.Error()),
		)
		return
	}

	// If we get here, the membership exists
	// The data already has the required fields from the config
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
