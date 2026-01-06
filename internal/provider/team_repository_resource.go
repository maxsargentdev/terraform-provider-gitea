package provider

import (
	"context"
	"fmt"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = &TeamRepositoryResource{}
	_ resource.ResourceWithConfigure = &TeamRepositoryResource{}
)

func NewTeamRepositoryResource() resource.Resource {
	return &TeamRepositoryResource{}
}

type TeamRepositoryResource struct {
	client *gitea.Client
}

type teamRepositoryResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Org            types.String `tfsdk:"org"`
	TeamName       types.String `tfsdk:"team_name"`
	RepositoryName types.String `tfsdk:"repository_name"`
}

func (r *TeamRepositoryResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team_repository"
}

func (r *TeamRepositoryResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a team repository association in Gitea.",
		Attributes: map[string]schema.Attribute{
			"org": schema.StringAttribute{
				Required:            true,
				Description:         "The name of the organization.",
				MarkdownDescription: "The name of the organization.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"team_name": schema.StringAttribute{
				Required:            true,
				Description:         "The name of the team.",
				MarkdownDescription: "The name of the team.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"repository_name": schema.StringAttribute{
				Required:            true,
				Description:         "The name of the repository.",
				MarkdownDescription: "The name of the repository.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "The ID of the team repository association.",
				MarkdownDescription: "The ID of the team repository association.",
			},
		},
	}
}

func (r *TeamRepositoryResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *TeamRepositoryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan teamRepositoryResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	org := plan.Org.ValueString()
	teamName := plan.TeamName.ValueString()
	repoName := plan.RepositoryName.ValueString()

	// Get team ID by listing org teams and finding by name
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

	// Add repository to team
	_, err = r.client.AddTeamRepository(teamID, org, repoName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Adding Repository to Team",
			fmt.Sprintf("Could not add repository '%s' to team '%s': %s", repoName, teamName, err.Error()),
		)
		return
	}

	// Set plan ID to a composite of org/team/repo
	plan.ID = types.StringValue(fmt.Sprintf("%s/%s/%s", org, teamName, repoName))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *TeamRepositoryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state teamRepositoryResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	org := state.Org.ValueString()
	teamName := state.TeamName.ValueString()
	repoName := state.RepositoryName.ValueString()

	// Get team ID
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

	// Get team repositories to verify association exists
	repos, _, err := r.client.ListTeamRepositories(teamID, gitea.ListTeamRepositoriesOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Listing Team Repositories",
			fmt.Sprintf("Could not list repositories for team '%s': %s", teamName, err.Error()),
		)
		return
	}

	// Check if repository is still assigned to team
	found := false
	for _, repo := range repos {
		if repo.Name == repoName {
			found = true
			break
		}
	}

	if !found {
		// Repository no longer assigned to team, remove from state
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *TeamRepositoryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Update is a no-op because all attributes have RequiresReplace
	// Any change will trigger a delete + create cycle
	var plan teamRepositoryResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *TeamRepositoryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state teamRepositoryResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	org := state.Org.ValueString()
	teamName := state.TeamName.ValueString()
	repoName := state.RepositoryName.ValueString()

	// Get team ID
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

	// Remove repository from team
	_, err = r.client.RemoveTeamRepository(teamID, org, repoName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Removing Repository from Team",
			fmt.Sprintf("Could not remove repository '%s' from team '%s': %s", repoName, teamName, err.Error()),
		)
		return
	}
}
