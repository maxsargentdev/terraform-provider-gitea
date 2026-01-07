package provider

import (
	"context"
	"fmt"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

type teamMembershipDataSourceModel struct {
	Org      types.String `tfsdk:"org"`
	TeamName types.String `tfsdk:"team_name"`
	Username types.String `tfsdk:"username"`
}

func (d *teamMembershipDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team_membership"
}

func (d *teamMembershipDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Get information about a team membership (checks if a user is a member of a team)",
		MarkdownDescription: "Get information about a team membership (checks if a user is a member of a team)",
		Attributes: map[string]schema.Attribute{

			// required - these are fundamental configuration options
			"org": schema.StringAttribute{
				Required:            true,
				Description:         "The name of the organization",
				MarkdownDescription: "The name of the organization",
			},
			"team_name": schema.StringAttribute{
				Required:            true,
				Description:         "The name of the team",
				MarkdownDescription: "The name of the team",
			},
			"username": schema.StringAttribute{
				Required:            true,
				Description:         "The username of the team member",
				MarkdownDescription: "The username of the team member",
			},
		},
	}
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
	var data teamMembershipDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org := data.Org.ValueString()
	teamName := data.TeamName.ValueString()
	username := data.Username.ValueString()

	// Get team ID from name
	teams, _, err := d.client.ListOrgTeams(org, gitea.ListTeamsOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Listing Teams",
			fmt.Sprintf("Could not list teams for organization '%s': %s", org, err.Error()),
		)
		return
	}

	var teamID int64
	for _, team := range teams {
		if team.Name == teamName {
			teamID = team.ID
			break
		}
	}

	if teamID == 0 {
		resp.Diagnostics.AddError(
			"Team Not Found",
			fmt.Sprintf("Could not find team '%s' in organization '%s'", teamName, org),
		)
		return
	}

	// Check if the user is a member of the team
	_, _, err = d.client.GetTeamMember(teamID, username)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Team Membership",
			fmt.Sprintf("Could not verify team membership for user %s in team %s: %s", username, teamName, err.Error()),
		)
		return
	}

	// If we get here, the membership exists
	// The data already has the required fields from the config
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
