package provider

import (
	"context"
	"fmt"
	"strconv"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
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

type teamResourceModel struct {
	Org                     types.String `tfsdk:"org"`
	Name                    types.String `tfsdk:"name"`
	UnitsMap                types.Map    `tfsdk:"units_map"`
	CanCreateOrgRepo        types.Bool   `tfsdk:"can_create_org_repo"`
	Description             types.String `tfsdk:"description"`
	IncludesAllRepositories types.Bool   `tfsdk:"includes_all_repositories"`
	Id                      types.Int64  `tfsdk:"id"`
}

func (r *teamResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team"
}

func (r *teamResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manages a Gitea team within an organization. Teams allow you to group users and assign permissions to repositories.",
		MarkdownDescription: "Manages a Gitea team within an organization. Teams allow you to group users and assign permissions to repositories.",
		Attributes: map[string]schema.Attribute{

			// required - these are fundamental configuration options
			"org": schema.StringAttribute{
				Required:            true,
				Description:         "The name of the organization the team belongs to. This is required for creating and managing the team.",
				MarkdownDescription: "The name of the organization the team belongs to. This is required for creating and managing the team.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				Description:         "The name of the team. Must be unique within the organization.",
				MarkdownDescription: "The name of the team. Must be unique within the organization.",
			},
			"units_map": schema.MapAttribute{
				Required: true,
				Description: "A map of repository units to their permission levels. Keys are unit names (e.g., 'repo.code', 'repo.issues', 'repo.pulls', 'repo.releases', 'repo.wiki', 'repo.ext_wiki', 'repo.ext_issues', 'repo.projects', 'repo.packages', 'repo.actions'). Values must be one of: 'none' (no access), 'read' (read-only access), 'write' (read and write access), 'admin' (full administrative access).",
				MarkdownDescription: `A map of repository units to their permission levels.

**Unit names:**
- ` + "`repo.code`" + ` - Repository code/files
- ` + "`repo.issues`" + ` - Issue tracker
- ` + "`repo.pulls`" + ` - Pull requests
- ` + "`repo.releases`" + ` - Releases
- ` + "`repo.wiki`" + ` - Wiki
- ` + "`repo.ext_wiki`" + ` - External wiki
- ` + "`repo.ext_issues`" + ` - External issue tracker
- ` + "`repo.projects`" + ` - Projects
- ` + "`repo.packages`" + ` - Packages
- ` + "`repo.actions`" + ` - Actions

**Permission values:**
- ` + "`none`" + ` - No access
- ` + "`read`" + ` - Read-only access
- ` + "`write`" + ` - Read and write access
- ` + "`admin`" + ` - Full administrative access`,
				ElementType: types.StringType,
			},

			// optional - these tweak the created resource away from its defaults
			"can_create_org_repo": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether team members can create new repositories in the organization. Defaults to false if not specified.",
				MarkdownDescription: "Whether team members can create new repositories in the organization. Defaults to `false` if not specified.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "A description of the team's purpose. Optional, but recommended for documentation.",
				MarkdownDescription: "A description of the team's purpose. Optional, but recommended for documentation.",
			},
			"includes_all_repositories": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether the team automatically has access to all repositories in the organization, including newly created ones. Defaults to false if not specified.",
				MarkdownDescription: "Whether the team automatically has access to all repositories in the organization, including newly created ones. Defaults to `false` if not specified.",
			},

			// computed - these are available to read back after creation but are really just metadata
			"id": schema.Int64Attribute{
				Computed:            true,
				Description:         "The unique numeric identifier of the team in Gitea.",
				MarkdownDescription: "The unique numeric identifier of the team in Gitea.",

				// ID doesnt change once set, only computed once so refer to state
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
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
	var plan teamResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgName := plan.Org.ValueString()

	// Validate units_map is provided (required attribute)
	if plan.UnitsMap.IsNull() || plan.UnitsMap.IsUnknown() {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"units_map is required but was not provided",
		)
		return
	}

	// Extract units_map for specifying permissions
	unitsMap := make(map[string]string)
	plan.UnitsMap.ElementsAs(ctx, &unitsMap, false)

	// Validate permission values - only "none", "read", "write", "admin" are allowed
	validPermissions := map[string]bool{"none": true, "read": true, "write": true, "admin": true}
	for unit, permission := range unitsMap {
		if !validPermissions[permission] {
			resp.Diagnostics.AddError(
				"Invalid Permission Value",
				fmt.Sprintf("Unit '%s' has invalid permission '%s'. Valid values are: 'none', 'read', 'write', 'admin'", unit, permission),
			)
		}
	}
	if resp.Diagnostics.HasError() {
		return
	}

	opts := gitea.CreateTeamOption{
		Name:                    plan.Name.ValueString(),
		Description:             plan.Description.ValueString(),
		CanCreateOrgRepo:        plan.CanCreateOrgRepo.ValueBool(),
		IncludesAllRepositories: plan.IncludesAllRepositories.ValueBool(),
		Permission:              gitea.AccessModeNone,
		UnitsMap:                unitsMap,
	}

	team, _, err := r.client.CreateTeam(orgName, opts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Team",
			"Could not create team, unexpected error: "+err.Error(),
		)
		return
	}

	mapTeamToModel(ctx, team, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *teamResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state teamResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	team, httpResp, err := r.client.GetTeam(state.Id.ValueInt64())
	if err != nil {
		// Handle 404 - team was deleted outside of Terraform
		if httpResp != nil && httpResp.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Team",
			fmt.Sprintf("Could not read team (ID: %d): %s", state.Id.ValueInt64(), err.Error()),
		)
		return
	}

	mapTeamToModel(ctx, team, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *teamResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan teamResourceModel
	var state teamResourceModel

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

	// Validate units_map is provided (required attribute)
	if plan.UnitsMap.IsNull() || plan.UnitsMap.IsUnknown() {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"units_map is required but was not provided",
		)
		return
	}

	// Extract units_map for specifying permissions
	unitsMap := make(map[string]string)
	plan.UnitsMap.ElementsAs(ctx, &unitsMap, false)

	// Validate permission values - only "none", "read", "write", "admin" are allowed
	validPermissions := map[string]bool{"none": true, "read": true, "write": true, "admin": true}
	for unit, permission := range unitsMap {
		if !validPermissions[permission] {
			resp.Diagnostics.AddError(
				"Invalid Permission Value",
				fmt.Sprintf("Unit '%s' has invalid permission '%s'. Valid values are: 'none', 'read', 'write', 'admin'", unit, permission),
			)
		}
	}
	if resp.Diagnostics.HasError() {
		return
	}

	opts := gitea.EditTeamOption{
		Name:                    plan.Name.ValueString(),
		Description:             &desc,
		CanCreateOrgRepo:        &canCreate,
		IncludesAllRepositories: &inclAll,
		Permission:              gitea.AccessModeNone,
		UnitsMap:                unitsMap,
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

	mapTeamToModel(ctx, team, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *teamResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state teamResourceModel

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
			fmt.Sprintf("Could not parse team ID '%s': %s. The import ID must be a numeric team ID.", req.ID, err.Error()),
		)
		return
	}

	// Fetch the full team details
	team, httpResp, err := r.client.GetTeam(id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			resp.Diagnostics.AddError(
				"Team Not Found",
				fmt.Sprintf("Team with ID %d does not exist or you don't have permission to access it.", id),
			)
			return
		}
		resp.Diagnostics.AddError(
			"Error Importing Team",
			fmt.Sprintf("Could not fetch team with ID %d: %s", id, err.Error()),
		)
		return
	}

	// Map to model
	var state teamResourceModel
	mapTeamToModel(ctx, team, &state)

	// Verify that org was populated (required for the resource to function)
	if state.Org.IsNull() || state.Org.IsUnknown() || state.Org.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Import Error",
			fmt.Sprintf("Could not determine organization for team ID %d. The API response did not include organization information.", id),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func mapTeamToModel(ctx context.Context, team *gitea.Team, model *teamResourceModel) {
	model.Id = types.Int64Value(team.ID)
	model.Name = types.StringValue(team.Name)
	model.Description = types.StringValue(team.Description)
	model.CanCreateOrgRepo = types.BoolValue(team.CanCreateOrgRepo)
	model.IncludesAllRepositories = types.BoolValue(team.IncludesAllRepositories)

	// Set organization name from API response if available (important for import)
	if team.Organization != nil {
		model.Org = types.StringValue(team.Organization.UserName)
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
}
