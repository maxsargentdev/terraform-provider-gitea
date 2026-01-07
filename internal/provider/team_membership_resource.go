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
	Id       types.String `tfsdk:"id"`
	Org      types.String `tfsdk:"org"`
	TeamName types.String `tfsdk:"team_name"`
	Username types.String `tfsdk:"username"`
}

func (r *teamMembershipResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team_membership"
}

func (r *teamMembershipResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manages a team membership, assigning a user to a team within an organization.",
		MarkdownDescription: "Manages a team membership, assigning a user to a team within an organization.\n\nThis resource adds or removes a user from a Gitea team. When destroyed, the user is removed from the team but the user account itself is not affected.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "The unique identifier for the team membership in the format 'org/team_name/username'.",
				MarkdownDescription: "The unique identifier for the team membership in the format `org/team_name/username`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"org": schema.StringAttribute{
				Required:            true,
				Description:         "The name of the organization that owns the team.",
				MarkdownDescription: "The name of the organization that owns the team.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"team_name": schema.StringAttribute{
				Required:            true,
				Description:         "The name of the team to add the user to.",
				MarkdownDescription: "The name of the team to add the user to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"username": schema.StringAttribute{
				Required:            true,
				Description:         "The username of the user to add to the team.",
				MarkdownDescription: "The username of the user to add to the team.",
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

// getTeamByName finds a team by name within an organization.
// Uses pagination to handle large organizations with many teams.
// Returns the team ID, a boolean indicating if the team was found, and any error encountered.
func (r *teamMembershipResource) getTeamByName(org, teamName string) (int64, bool, error) {
	teams, httpResp, err := r.client.ListOrgTeams(org, gitea.ListTeamsOptions{
		ListOptions: gitea.ListOptions{Page: -1}, // Get all teams with pagination
	})
	if err != nil {
		// Check for 404 - organization doesn't exist
		if httpResp != nil && httpResp.StatusCode == 404 {
			return 0, false, fmt.Errorf("organization '%s' not found", org)
		}
		return 0, false, err
	}

	for _, team := range teams {
		if team.Name == teamName {
			return team.ID, true, nil
		}
	}

	return 0, false, nil
}

// computeId generates the resource ID from org, team_name, and username.
func computeId(org, teamName, username string) string {
	return fmt.Sprintf("%s/%s/%s", org, teamName, username)
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

	// Get team ID from name using helper function with pagination
	teamID, found, err := r.getTeamByName(org, teamName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Looking Up Team",
			fmt.Sprintf("Could not look up team '%s' in organization '%s': %s", teamName, org, err.Error()),
		)
		return
	}

	if !found {
		resp.Diagnostics.AddError(
			"Team Not Found",
			fmt.Sprintf("Could not find team '%s' in organization '%s'. Please verify the team exists.", teamName, org),
		)
		return
	}

	_, err = r.client.AddTeamMember(teamID, username)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Adding Team Member",
			fmt.Sprintf("Could not add user '%s' to team '%s': %s", username, teamName, err.Error()),
		)
		return
	}

	// Set computed ID
	plan.Id = types.StringValue(computeId(org, teamName, username))

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

	// Get team ID from name using helper function with pagination
	teamID, found, err := r.getTeamByName(org, teamName)
	if err != nil {
		// If the org doesn't exist, remove from state
		if strings.Contains(err.Error(), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Looking Up Team",
			fmt.Sprintf("Could not look up team '%s' in organization '%s': %s", teamName, org, err.Error()),
		)
		return
	}

	if !found {
		// Team no longer exists, remove from state
		resp.State.RemoveResource(ctx)
		return
	}

	// Check if the user is still a member of the team
	_, httpResp, err := r.client.GetTeamMember(teamID, username)
	if err != nil {
		// Handle 404 - membership was deleted outside of Terraform
		if httpResp != nil && httpResp.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Team Membership",
			fmt.Sprintf("Could not read team membership for user '%s' in team '%s': %s", username, teamName, err.Error()),
		)
		return
	}

	// Ensure ID is set (handles imports and upgrades from old state)
	state.Id = types.StringValue(computeId(org, teamName, username))

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

	// Get team ID from name using helper function with pagination
	teamID, found, err := r.getTeamByName(org, teamName)
	if err != nil {
		// If the org doesn't exist, consider deletion successful
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError(
			"Error Looking Up Team",
			fmt.Sprintf("Could not look up team '%s' in organization '%s': %s", teamName, org, err.Error()),
		)
		return
	}

	if !found {
		// Team no longer exists, nothing to delete
		return
	}

	httpResp, err := r.client.RemoveTeamMember(teamID, username)
	if err != nil {
		// If already deleted (404), consider it a success
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		resp.Diagnostics.AddError(
			"Error Removing Team Member",
			fmt.Sprintf("Could not remove user '%s' from team '%s': %s", username, teamName, err.Error()),
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

	// Validate inputs
	if org == "" || teamName == "" || username == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Import ID components cannot be empty. Format: 'org/team_name/username'",
		)
		return
	}

	// Get team ID from name using helper function with pagination
	teamID, found, err := r.getTeamByName(org, teamName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Looking Up Team",
			fmt.Sprintf("Could not look up team '%s' in organization '%s': %s", teamName, org, err.Error()),
		)
		return
	}

	if !found {
		resp.Diagnostics.AddError(
			"Team Not Found",
			fmt.Sprintf("Could not find team '%s' in organization '%s'. Please verify the team exists.", teamName, org),
		)
		return
	}

	// Verify the membership exists
	_, httpResp, err := r.client.GetTeamMember(teamID, username)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			resp.Diagnostics.AddError(
				"Team Membership Not Found",
				fmt.Sprintf("User '%s' is not a member of team '%s' in organization '%s'.", username, teamName, org),
			)
			return
		}
		resp.Diagnostics.AddError(
			"Error Verifying Team Membership",
			fmt.Sprintf("Could not verify team membership for user '%s' in team '%s': %s", username, teamName, err.Error()),
		)
		return
	}

	state := teamMembershipResourceModel{
		Id:       types.StringValue(computeId(org, teamName, username)),
		Org:      types.StringValue(org),
		TeamName: types.StringValue(teamName),
		Username: types.StringValue(username),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
