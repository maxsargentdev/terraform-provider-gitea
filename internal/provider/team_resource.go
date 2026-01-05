package provider

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	"terraform-provider-gitea/internal/resource_team"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &teamResource{}
	_ resource.ResourceWithConfigure   = &teamResource{}
	_ resource.ResourceWithImportState = &teamResource{}
)

func NewTeamResource() resource.Resource {
	return &teamResource{}
}

type teamResource struct {
	client *gitea.Client
}

// TeamModel wraps the generated model and adds the org field
type TeamModel struct {
	resource_team.TeamModel
	Org types.String `tfsdk:"org"`
}

func (r *teamResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team"
}

func (r *teamResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	baseSchema := resource_team.TeamResourceSchema(ctx)

	// Add org as a required string field - this is the input for which org to create the team in
	baseSchema.Attributes["org"] = schema.StringAttribute{
		Required:            true,
		Description:         "The name of the organization to create the team in",
		MarkdownDescription: "The name of the organization to create the team in",
	}

	resp.Schema = baseSchema
}

func (r *teamResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*gitea.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *gitea.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *teamResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan TeamModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgName := plan.Org.ValueString()

	opts := gitea.CreateTeamOption{
		Name:                    plan.Name.ValueString(),
		Description:             plan.Description.ValueString(),
		CanCreateOrgRepo:        plan.CanCreateOrgRepo.ValueBool(),
		IncludesAllRepositories: plan.IncludesAllRepositories.ValueBool(),
		Permission:              gitea.AccessModeNone,
	}

	// units_map is required for specifying permissions
	if !plan.UnitsMap.IsNull() && !plan.UnitsMap.IsUnknown() {
		unitsMap := make(map[string]string)
		plan.UnitsMap.ElementsAs(ctx, &unitsMap, false)
		if len(unitsMap) > 0 {
			opts.UnitsMap = unitsMap
		}
	}

	team, _, err := r.client.CreateTeam(orgName, opts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Team",
			"Could not create team, unexpected error: "+err.Error(),
		)
		return
	}

	mapTeamToModel(ctx, team, &plan.TeamModel)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *teamResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state TeamModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	team, _, err := r.client.GetTeam(state.Id.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Team",
			"Could not read team: "+err.Error(),
		)
		return
	}

	mapTeamToModel(ctx, team, &state.TeamModel)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *teamResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan TeamModel
	var state TeamModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use ID from state since it's computed and won't be in plan
	teamID := state.Id.ValueInt64()
	if teamID == 0 {
		resp.Diagnostics.AddError(
			"Missing Team ID",
			"Team ID is missing from state - cannot update",
		)
		return
	}

	desc := plan.Description.ValueString()
	canCreate := plan.CanCreateOrgRepo.ValueBool()
	inclAll := plan.IncludesAllRepositories.ValueBool()

	opts := gitea.EditTeamOption{
		Name:                    plan.Name.ValueString(),
		Description:             &desc,
		CanCreateOrgRepo:        &canCreate,
		IncludesAllRepositories: &inclAll,
		Permission:              gitea.AccessModeNone,
	}

	// units_map is required for specifying permissions
	if !plan.UnitsMap.IsNull() && !plan.UnitsMap.IsUnknown() {
		unitsMap := make(map[string]string)
		plan.UnitsMap.ElementsAs(ctx, &unitsMap, false)
		if len(unitsMap) > 0 {
			opts.UnitsMap = unitsMap
		}
	}

	_, err := r.client.EditTeam(teamID, opts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Team",
			fmt.Sprintf("Could not update team (ID: %d, Name: %s), unexpected error: %s", teamID, plan.Name.ValueString(), err.Error()),
		)
		return
	}

	// Read back the updated team
	team, _, err := r.client.GetTeam(teamID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Updated Team",
			"Could not read team after update: "+err.Error(),
		)
		return
	}

	mapTeamToModel(ctx, team, &plan.TeamModel)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *teamResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state TeamModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DeleteTeam(state.Id.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Team",
			"Could not delete team, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *teamResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Importing Team",
			"Could not parse team ID: "+err.Error(),
		)
		return
	}

	// Fetch the full team details
	team, _, err := r.client.GetTeam(id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Importing Team",
			fmt.Sprintf("Could not fetch team with ID %d: %s", id, err.Error()),
		)
		return
	}

	// Map to model
	var state TeamModel
	mapTeamToModel(ctx, team, &state.TeamModel)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func mapTeamToModel(ctx context.Context, team *gitea.Team, model *resource_team.TeamModel) {
	model.Id = types.Int64Value(team.ID)
	model.Name = types.StringValue(team.Name)
	model.Description = types.StringValue(team.Description)
	model.CanCreateOrgRepo = types.BoolValue(team.CanCreateOrgRepo)
	model.IncludesAllRepositories = types.BoolValue(team.IncludesAllRepositories)

	// Map units list
	if len(team.Units) > 0 {
		// Sort units for consistent ordering
		unitStrs := make([]string, len(team.Units))
		for i, v := range team.Units {
			unitStrs[i] = string(v)
		}
		sort.Strings(unitStrs)

		elements := make([]attr.Value, len(unitStrs))
		for i, v := range unitStrs {
			elements[i] = types.StringValue(v)
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
		// When units_map is not present, keep existing value if set, otherwise null
		if model.UnitsMap.IsNull() || model.UnitsMap.IsUnknown() {
			model.UnitsMap = types.MapNull(types.StringType)
		}
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

		orgValue, diags := resource_team.NewOrganizationValue(
			resource_team.OrganizationValue{}.AttributeTypes(ctx),
			orgAttrs,
		)
		if !diags.HasError() {
			model.Organization = orgValue
		}
	} else {
		model.Organization = resource_team.NewOrganizationValueNull()
	}
}
