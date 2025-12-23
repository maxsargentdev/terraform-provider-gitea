package provider

import (
	"context"
	"fmt"
	"strconv"

	"terraform-provider-gitea/internal/resource_team"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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

	// Override permission validator to accept "none" for unit-based permissions
	if permAttr, ok := baseSchema.Attributes["permission"].(schema.StringAttribute); ok {
		permAttr.Validators = []validator.String{
			stringvalidator.OneOf("read", "write", "admin", "none"),
		}
		baseSchema.Attributes["permission"] = permAttr
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
	}

	// Set permission
	permValue := plan.Permission.ValueString()
	if permValue == "none" || permValue == "" {
		opts.Permission = gitea.AccessModeRead
	} else {
		opts.Permission = gitea.AccessMode(permValue)
	}

	// Check if units_map is specified for per-unit permissions
	if !plan.UnitsMap.IsNull() && !plan.UnitsMap.IsUnknown() {
		unitsMap := make(map[string]string)
		plan.UnitsMap.ElementsAs(ctx, &unitsMap, false)
		opts.UnitsMap = unitsMap
	} else if !plan.Units.IsNull() {
		var unitStrs []string
		plan.Units.ElementsAs(ctx, &unitStrs, false)
		units := make([]gitea.RepoUnitType, len(unitStrs))
		for i, u := range unitStrs {
			units[i] = gitea.RepoUnitType(u)
		}
		opts.Units = units
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

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
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
	}

	if !plan.Permission.IsNull() {
		perm := gitea.AccessMode(plan.Permission.ValueString())
		opts.Permission = perm
	}

	// Check if units_map is specified for per-unit permissions
	if !plan.UnitsMap.IsNull() && !plan.UnitsMap.IsUnknown() {
		unitsMap := make(map[string]string)
		plan.UnitsMap.ElementsAs(ctx, &unitsMap, false)
		opts.UnitsMap = unitsMap
	} else if !plan.Units.IsNull() {
		var unitStrs []string
		plan.Units.ElementsAs(ctx, &unitStrs, false)
		units := make([]gitea.RepoUnitType, len(unitStrs))
		for i, u := range unitStrs {
			units[i] = gitea.RepoUnitType(u)
		}
		opts.Units = units
	}

	_, err := r.client.EditTeam(plan.Id.ValueInt64(), opts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Team",
			"Could not update team, unexpected error: "+err.Error(),
		)
		return
	}

	// Read back the updated team
	team, _, err := r.client.GetTeam(plan.Id.ValueInt64())
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

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

func mapTeamToModel(ctx context.Context, team *gitea.Team, model *resource_team.TeamModel) {
	model.Id = types.Int64Value(team.ID)
	model.Name = types.StringValue(team.Name)
	model.Description = types.StringValue(team.Description)

	// Gitea has two permission models:
	// 1. Admin access (permission = "admin")
	// 2. General access with units (permission = "none" from API, but we map based on units)
	// When units are specified, Gitea returns "none" but we should keep the original permission value
	if team.Permission == "none" && len(team.Units) > 0 {
		// Unit-based permissions - keep the existing permission value from model
		// Don't overwrite it
	} else {
		// Admin or explicit permission - use what API returns
		model.Permission = types.StringValue(string(team.Permission))
	}

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
