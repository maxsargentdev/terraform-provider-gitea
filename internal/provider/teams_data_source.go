package provider

import (
	"context"
	"fmt"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = (*teamsDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*teamsDataSource)(nil)

func NewTeamsDataSource() datasource.DataSource {
	return &teamsDataSource{}
}

type teamsDataSource struct {
	client *gitea.Client
}

type teamsDataSourceModel struct {
	Org   types.String    `tfsdk:"org"`
	Teams []teamShortInfo `tfsdk:"teams"`
}

type teamShortInfo struct {
	Id                      types.Int64  `tfsdk:"id"`
	Name                    types.String `tfsdk:"name"`
	Description             types.String `tfsdk:"description"`
	Permission              types.String `tfsdk:"permission"`
	CanCreateOrgRepo        types.Bool   `tfsdk:"can_create_org_repo"`
	IncludesAllRepositories types.Bool   `tfsdk:"includes_all_repositories"`
}

func (d *teamsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_teams"
}

func (d *teamsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a list of teams. Can list teams for an organization or all teams for the authenticated user.",
		Attributes: map[string]schema.Attribute{
			"org": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Organization to list teams for. If not specified, lists all teams for the authenticated user.",
			},
			"teams": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of teams",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Team ID",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Team name",
						},
						"description": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Team description",
						},
						"permission": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Team permission level",
						},
						"can_create_org_repo": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether team members can create org repositories",
						},
						"includes_all_repositories": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether the team has access to all repositories",
						},
					},
				},
			},
		},
	}
}

func (d *teamsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *teamsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data teamsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var teams []*gitea.Team
	var err error

	if !data.Org.IsNull() {
		// List teams for a specific organization
		teams, _, err = d.client.ListOrgTeams(data.Org.ValueString(), gitea.ListTeamsOptions{
			ListOptions: gitea.ListOptions{Page: -1},
		})
	} else {
		// List all teams for the authenticated user
		teams, _, err = d.client.ListMyTeams(&gitea.ListTeamsOptions{
			ListOptions: gitea.ListOptions{Page: -1},
		})
	}

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list teams, got error: %s", err))
		return
	}

	// Map teams to model
	data.Teams = make([]teamShortInfo, len(teams))
	for i, team := range teams {
		data.Teams[i] = teamShortInfo{
			Id:                      types.Int64Value(team.ID),
			Name:                    types.StringValue(team.Name),
			Description:             types.StringValue(team.Description),
			Permission:              types.StringValue(string(team.Permission)),
			CanCreateOrgRepo:        types.BoolValue(team.CanCreateOrgRepo),
			IncludesAllRepositories: types.BoolValue(team.IncludesAllRepos),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
