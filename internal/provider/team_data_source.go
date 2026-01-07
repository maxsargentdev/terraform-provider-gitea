package provider

import (
	"context"
	"fmt"
	"net/http"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &teamDataSource{}
	_ datasource.DataSourceWithConfigure = &teamDataSource{}
)

func NewTeamDataSource() datasource.DataSource {
	return &teamDataSource{}
}

type teamDataSource struct {
	client *gitea.Client
}

type teamDataSourceModel struct {
	Org                     types.String `tfsdk:"org"`
	Name                    types.String `tfsdk:"name"`
	CanCreateOrgRepo        types.Bool   `tfsdk:"can_create_org_repo"`
	Description             types.String `tfsdk:"description"`
	Id                      types.Int64  `tfsdk:"id"`
	IncludesAllRepositories types.Bool   `tfsdk:"includes_all_repositories"`
	UnitsMap                types.Map    `tfsdk:"units_map"`
}

func (d *teamDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team"
}

func (d *teamDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Use this data source to retrieve information about an existing team within a Gitea organization.",
		MarkdownDescription: "Use this data source to retrieve information about an existing team within a Gitea organization.",
		Attributes: map[string]schema.Attribute{

			// required - these are fundamental configuration options
			"org": schema.StringAttribute{
				Required:            true,
				Description:         "The name of the organization that owns the team.",
				MarkdownDescription: "The name of the organization that owns the team.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				Description:         "The name of the team to look up.",
				MarkdownDescription: "The name of the team to look up.",
			},

			// computed - these are available to read back after creation but are really just metadata
			"id": schema.Int64Attribute{
				Computed:            true,
				Description:         "The numeric ID of the team.",
				MarkdownDescription: "The numeric ID of the team.",
			},
			"can_create_org_repo": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether team members can create repositories in the organization.",
				MarkdownDescription: "Whether team members can create repositories in the organization.",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				Description:         "The description of the team.",
				MarkdownDescription: "The description of the team.",
			},
			"includes_all_repositories": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether the team has access to all repositories in the organization.",
				MarkdownDescription: "Whether the team has access to all repositories in the organization.",
			},
			"units_map": schema.MapAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				Description:         "A map of unit names to permission levels (e.g., 'repo.code': 'read', 'repo.issues': 'write').",
				MarkdownDescription: "A map of unit names to permission levels (e.g., `repo.code`: `read`, `repo.issues`: `write`).",
			},
		},
	}
}

func (d *teamDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *teamDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data teamDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org := data.Org.ValueString()
	teamName := data.Name.ValueString()

	// Get team by org and name
	teams, httpResp, err := d.client.ListOrgTeams(org, gitea.ListTeamsOptions{})
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			resp.Diagnostics.AddError(
				"Organization Not Found",
				fmt.Sprintf("Organization '%s' does not exist or you do not have permission to access it.", org),
			)
			return
		}
		resp.Diagnostics.AddError(
			"Error Listing Teams",
			fmt.Sprintf("Could not list teams for organization '%s': %s", org, err.Error()),
		)
		return
	}

	var team *gitea.Team
	for _, t := range teams {
		if t.Name == teamName {
			team = t
			break
		}
	}

	if team == nil {
		resp.Diagnostics.AddError(
			"Team Not Found",
			fmt.Sprintf("Team '%s' does not exist in organization '%s' or you do not have permission to access it.", teamName, org),
		)
		return
	}

	mapTeamToDataSourceModel(ctx, team, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func mapTeamToDataSourceModel(ctx context.Context, team *gitea.Team, model *teamDataSourceModel) {
	model.Id = types.Int64Value(team.ID)
	model.Name = types.StringValue(team.Name)
	model.Description = types.StringValue(team.Description)
	model.CanCreateOrgRepo = types.BoolValue(team.CanCreateOrgRepo)
	model.IncludesAllRepositories = types.BoolValue(team.IncludesAllRepositories)

	// Map units_map if present
	if len(team.UnitsMap) > 0 {
		mapElements := make(map[string]attr.Value)
		for k, v := range team.UnitsMap {
			mapElements[k] = types.StringValue(v)
		}
		model.UnitsMap, _ = types.MapValue(types.StringType, mapElements)
	} else {
		model.UnitsMap = types.MapNull(types.StringType)
	}

}
