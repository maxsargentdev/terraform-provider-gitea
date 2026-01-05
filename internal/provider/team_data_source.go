package provider

import (
	"context"
	"fmt"

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

func (d *teamDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team"
}

func (d *teamDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"can_create_org_repo": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether the team can create repositories in the organization",
				MarkdownDescription: "Whether the team can create repositories in the organization",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				Description:         "The description of the team",
				MarkdownDescription: "The description of the team",
			},
			"id": schema.Int64Attribute{
				Required:            true,
				Description:         "id of the team to get",
				MarkdownDescription: "id of the team to get",
			},
			"includes_all_repositories": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether the team has access to all repositories in the organization",
				MarkdownDescription: "Whether the team has access to all repositories in the organization",
			},
			"name": schema.StringAttribute{
				Computed:            true,
				Description:         "The name of the team",
				MarkdownDescription: "The name of the team",
			},
			"units": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"units_map": schema.MapAttribute{
				ElementType: types.StringType,
				Computed:    true,
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
	var data TeamDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get team by ID
	team, _, err := d.client.GetTeam(data.Id.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Team",
			"Could not read team: "+err.Error(),
		)
		return
	}

	mapTeamToDataSourceModel(ctx, team, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func mapTeamToDataSourceModel(ctx context.Context, team *gitea.Team, model *TeamDataSourceModel) {
	model.Id = types.Int64Value(team.ID)
	model.Name = types.StringValue(team.Name)
	model.Description = types.StringValue(team.Description)
	model.CanCreateOrgRepo = types.BoolValue(team.CanCreateOrgRepo)
	model.IncludesAllRepositories = types.BoolValue(team.IncludesAllRepositories)

	// Map units list
	if len(team.Units) > 0 {
		elements := make([]attr.Value, len(team.Units))
		for i, v := range team.Units {
			elements[i] = types.StringValue(string(v))
		}
		model.Units, _ = types.ListValue(types.StringType, elements)
	} else {
		model.Units = types.ListNull(types.StringType)
	}

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

type TeamDataSourceModel struct {
	CanCreateOrgRepo        types.Bool   `tfsdk:"can_create_org_repo"`
	Description             types.String `tfsdk:"description"`
	Id                      types.Int64  `tfsdk:"id"`
	IncludesAllRepositories types.Bool   `tfsdk:"includes_all_repositories"`
	Name                    types.String `tfsdk:"name"`
	Units                   types.List   `tfsdk:"units"`
	UnitsMap                types.Map    `tfsdk:"units_map"`
}
