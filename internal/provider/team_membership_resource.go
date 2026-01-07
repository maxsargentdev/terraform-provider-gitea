package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &teamMembershipResource{}
	_ resource.ResourceWithConfigure   = &teamMembershipResource{}
	_ resource.ResourceWithImportState = &teamMembershipResource{}
)

func NewTeamMembershipResource() resource.Resource {
	return &teamMembershipResource{}
}

type teamMembershipResource struct {
	client *gitea.Client
}

type teamMembershipResourceModel struct {
	// Required
	TeamId   types.Int64  `tfsdk:"team_id"`
	Username types.String `tfsdk:"username"`

	// Computed
	Id types.String `tfsdk:"id"`
}

func (r *teamMembershipResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team_membership"
}

func (r *teamMembershipResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manages a team membership in Gitea.",
		MarkdownDescription: "Manages a team membership in Gitea. This resource adds or removes a user from a Gitea team.",
		Attributes: map[string]schema.Attribute{
			// Required
			"team_id": schema.Int64Attribute{
				Required:            true,
				Description:         "The ID of the team.",
				MarkdownDescription: "The ID of the team.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"username": schema.StringAttribute{
				Required:            true,
				Description:         "The username of the team member.",
				MarkdownDescription: "The username of the team member.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			// Computed
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "The ID of this resource.",
				MarkdownDescription: "The ID of this resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
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

	teamID := plan.TeamId.ValueInt64()
	username := plan.Username.ValueString()

	_, err := r.client.AddTeamMember(teamID, username)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Adding Team Member",
			fmt.Sprintf("Could not add user '%s' to team %d: %s", username, teamID, err.Error()),
		)
		return
	}

	// Set computed ID
	plan.Id = types.StringValue(fmt.Sprintf("%d/%s", teamID, username))

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *teamMembershipResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state teamMembershipResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	teamID := state.TeamId.ValueInt64()
	username := state.Username.ValueString()

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
			fmt.Sprintf("Could not read team membership for user '%s' in team %d: %s", username, teamID, err.Error()),
		)
		return
	}

	// Ensure ID is set (handles imports and upgrades from old state)
	state.Id = types.StringValue(fmt.Sprintf("%d/%s", teamID, username))

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

	teamID := state.TeamId.ValueInt64()
	username := state.Username.ValueString()

	httpResp, err := r.client.RemoveTeamMember(teamID, username)
	if err != nil {
		// If already deleted (404), consider it a success
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		resp.Diagnostics.AddError(
			"Error Removing Team Member",
			fmt.Sprintf("Could not remove user '%s' from team %d: %s", username, teamID, err.Error()),
		)
		return
	}
}

func (r *teamMembershipResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: "team_id/username"
	id := req.ID

	// Parse the import ID
	parts := strings.Split(id, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Import ID must be in format 'team_id/username', got: %s", id),
		)
		return
	}

	teamIDStr := parts[0]
	username := parts[1]

	teamID, err := strconv.ParseInt(teamIDStr, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("team_id must be a valid number, got: %s", teamIDStr),
		)
		return
	}

	// Validate inputs
	if teamID == 0 || username == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Import ID components cannot be empty. Format: 'team_id/username'",
		)
		return
	}

	// Verify the membership exists
	_, httpResp, err := r.client.GetTeamMember(teamID, username)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			resp.Diagnostics.AddError(
				"Team Membership Not Found",
				fmt.Sprintf("User '%s' is not a member of team %d.", username, teamID),
			)
			return
		}
		resp.Diagnostics.AddError(
			"Error Verifying Team Membership",
			fmt.Sprintf("Could not verify team membership for user '%s' in team %d: %s", username, teamID, err.Error()),
		)
		return
	}

	state := teamMembershipResourceModel{
		Id:       types.StringValue(fmt.Sprintf("%d/%s", teamID, username)),
		TeamId:   types.Int64Value(teamID),
		Username: types.StringValue(username),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
