package provider

import (
	"context"
	"fmt"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &teamMembershipResource{}

func NewTeamMembershipResource() resource.Resource {
	return &teamMembershipResource{}
}

type teamMembershipResource struct {
	client *gitea.Client
}

type TeamMembershipModel struct {
	TeamId   types.Int64  `tfsdk:"team_id"`
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
			"team_id": schema.Int64Attribute{
				Required:            true,
				Description:         "The ID of the team",
				MarkdownDescription: "The ID of the team",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
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
	var plan TeamMembershipModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	teamID := plan.TeamId.ValueInt64()
	username := plan.Username.ValueString()

	_, err := r.client.AddTeamMember(teamID, username)
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
	var state TeamMembershipModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	teamID := state.TeamId.ValueInt64()
	username := state.Username.ValueString()

	// Check if the user is still a member of the team
	_, _, err := r.client.GetTeamMember(teamID, username)
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
	var state TeamMembershipModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	teamID := state.TeamId.ValueInt64()
	username := state.Username.ValueString()

	_, err := r.client.RemoveTeamMember(teamID, username)
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
	// Import format: "team_id:username"
	id := req.ID

	// Parse the import ID
	var teamID int64
	var username string

	_, err := fmt.Sscanf(id, "%d:%s", &teamID, &username)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Import ID must be in format 'team_id:username', got: %s", id),
		)
		return
	}

	state := TeamMembershipModel{
		TeamId:   types.Int64Value(teamID),
		Username: types.StringValue(username),
	}

	// Verify the membership exists
	_, _, err = r.client.GetTeamMember(teamID, username)
	if err != nil {
		resp.Diagnostics.AddError(
			"Team Membership Not Found",
			fmt.Sprintf("Could not find team membership for team %d and user %s: %s", teamID, username, err.Error()),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
