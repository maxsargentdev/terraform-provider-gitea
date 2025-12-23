package provider

import (
	"context"
	"fmt"

	"terraform-provider-gitea/internal/datasource_team"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
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
	resp.Schema = datasource_team.TeamDataSourceSchema(ctx)
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
	var data datasource_team.TeamModel

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

func mapTeamToDataSourceModel(ctx context.Context, team *gitea.Team, model *datasource_team.TeamModel) {
	model.Id = types.Int64Value(team.ID)
	model.Name = types.StringValue(team.Name)
	model.Description = types.StringValue(team.Description)
	model.Permission = types.StringValue(string(team.Permission))
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

	// Map organization nested object
	if team.Organization != nil {
		orgAttrs := map[string]attr.Value{
			"id":                            types.Int64Value(team.Organization.ID),
			"name":                          types.StringValue(team.Organization.UserName),
			"username":                      types.StringValue(team.Organization.UserName),
			"full_name":                     types.StringValue(team.Organization.FullName),
			"avatar_url":                    types.StringValue(team.Organization.AvatarURL),
			"description":                   types.StringValue(team.Organization.Description),
			"website":                       types.StringValue(team.Organization.Website),
			"location":                      types.StringValue(team.Organization.Location),
			"visibility":                    types.StringValue(team.Organization.Visibility),
			"email":                         types.StringNull(),
			"repo_admin_change_team_access": types.BoolNull(),
		}

		orgValue, diags := datasource_team.NewOrganizationValue(
			datasource_team.OrganizationValue{}.AttributeTypes(ctx),
			orgAttrs,
		)
		if !diags.HasError() {
			model.Organization = orgValue
		}
	} else {
		model.Organization = datasource_team.NewOrganizationValueNull()
	}
}
