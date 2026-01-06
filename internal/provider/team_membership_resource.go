package provider

import (
	"context"
	"fmt"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &teamMembershipResource{}
var _ resource.ResourceWithImportState = &teamMembershipResource{}

func NewTeamMembershipResource() resource.Resource {
	return &teamMembershipResource{}
}

type teamMembershipResource struct {
	client *gitea.Client
}

type teamMembershipResourceModel struct {
	Org      types.String `tfsdk:"org"`
	TeamName types.String `tfsdk:"team_name"`
	Username types.String `tfsdk:"username"`
}

func (r *teamMembershipResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team_membership"
}

func (r *teamMembershipResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manages a team membership (assigns a user to a team)",
		MarkdownDescription: "Manages a team membership (assigns a user to a team)",
		Attributes: map[string]schema.Attribute{

			// required - these are fundamental configuration options
			"org": schema.StringAttribute{
				Required:            true,
				Description:         "The name of the organization",
				MarkdownDescription: "The name of the organization",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"team_name": schema.StringAttribute{
				Required:            true,
				Description:         "The name of the team",
				MarkdownDescription: "The name of the team",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"username": schema.StringAttribute{
				Required:            true,
				Description:         "The username of the user to add to the team",
				MarkdownDescription: "The username of the user to add to the team",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *teamMembershipResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *teamMembershipResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan teamMembershipResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org := plan.Org.ValueString()
	teamName := plan.TeamName.ValueString()
	username := plan.Username.ValueString()

	// Get team ID from name
	teams, _, err := r.client.ListOrgTeams(org, gitea.ListTeamsOptions{})
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

	_, err = r.client.AddTeamMember(teamID, username)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Adding Team Member",
			"Could not add user to team, unexpected error: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *teamMembershipResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state teamMembershipResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org := state.Org.ValueString()
	teamName := state.TeamName.ValueString()
	username := state.Username.ValueString()

	// Get team ID from name
	teams, _, err := r.client.ListOrgTeams(org, gitea.ListTeamsOptions{})
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
		// Team no longer exists, remove from state
		resp.State.RemoveResource(ctx)
		return
	}

	// Check if the user is still a member of the team
	_, _, err = r.client.GetTeamMember(teamID, username)
	if err != nil {
		// If 404, remove from state
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *teamMembershipResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Team membership cannot be updated - only created or deleted
	// RequiresReplace plan modifiers ensure this never gets called
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Team membership cannot be updated. Terraform will recreate the resource.",
	)
}

func (r *teamMembershipResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state teamMembershipResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org := state.Org.ValueString()
	teamName := state.TeamName.ValueString()
	username := state.Username.ValueString()

	// Get team ID from name
	teams, _, err := r.client.ListOrgTeams(org, gitea.ListTeamsOptions{})
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
		// Team no longer exists, nothing to delete
		return
	}

	_, err = r.client.RemoveTeamMember(teamID, username)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Removing Team Member",
			"Could not remove user from team: "+err.Error(),
		)
		return
	}
}

// ImportState allows importing existing team memberships
func (r *teamMembershipResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: "org/team_name/username"
	id := req.ID

	// Parse the import ID
	parts := strings.Split(id, "/")
	if len(parts) != 3 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Import ID must be in format 'org/team_name/username', got: %s", id),
		)
		return
	}

	org := parts[0]
	teamName := parts[1]
	username := parts[2]

	// Get team ID from name
	teams, _, err := r.client.ListOrgTeams(org, gitea.ListTeamsOptions{})
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

	// Verify the membership exists
	_, _, err = r.client.GetTeamMember(teamID, username)
	if err != nil {
		resp.Diagnostics.AddError(
			"Team Membership Not Found",
			fmt.Sprintf("Could not find team membership for team '%s' and user '%s': %s", teamName, username, err.Error()),
		)
		return
	}

	state := teamMembershipResourceModel{
		Org:      types.StringValue(org),
		TeamName: types.StringValue(teamName),
		Username: types.StringValue(username),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
