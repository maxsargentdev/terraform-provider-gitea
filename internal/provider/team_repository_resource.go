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

var (
	_ resource.Resource                = &TeamRepositoryResource{}
	_ resource.ResourceWithConfigure   = &TeamRepositoryResource{}
	_ resource.ResourceWithImportState = &TeamRepositoryResource{}
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
		Description:         "Manages a team repository association in Gitea. This resource assigns a repository to a team, granting team members access to the repository based on the team's permission level.",
		MarkdownDescription: "Manages a team repository association in Gitea. This resource assigns a repository to a team, granting team members access to the repository based on the team's permission level.\n\n## Import\n\nTeam repository associations can be imported using the format `org/team_name/repository_name`:\n\n```shell\nterraform import gitea_team_repository.example myorg/developers/myrepo\n```",
		Attributes: map[string]schema.Attribute{

			// required - these are fundamental configuration options
			"org": schema.StringAttribute{
				Required:            true,
				Description:         "The name of the organization that owns both the team and the repository.",
				MarkdownDescription: "The name of the organization that owns both the team and the repository.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"team_name": schema.StringAttribute{
				Required:            true,
				Description:         "The name of the team to assign the repository to. The team must exist within the specified organization.",
				MarkdownDescription: "The name of the team to assign the repository to. The team must exist within the specified organization.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"repository_name": schema.StringAttribute{
				Required:            true,
				Description:         "The name of the repository to assign to the team. The repository must exist within the specified organization.",
				MarkdownDescription: "The name of the repository to assign to the team. The repository must exist within the specified organization.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			// computed - these are available to read back after creation but are really just metadata
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "The ID of the team repository association, in the format 'org/team_name/repository_name'.",
				MarkdownDescription: "The ID of the team repository association, in the format `org/team_name/repository_name`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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

// findTeamByName searches for a team by name within an organization.
// Returns the team ID if found, 0 if not found, and an error if the API call fails.
func (r *TeamRepositoryResource) findTeamByName(org, teamName string) (int64, error) {
	teams, _, err := r.client.ListOrgTeams(org, gitea.ListTeamsOptions{
		ListOptions: gitea.ListOptions{Page: -1}, // Get all teams with pagination
	})
	if err != nil {
		return 0, err
	}

	for _, team := range teams {
		if team.Name == teamName {
			return team.ID, nil
		}
	}

	return 0, nil
}

// checkRepositoryInTeam checks if a repository is assigned to a team.
// Returns true if found, false otherwise.
func (r *TeamRepositoryResource) checkRepositoryInTeam(teamID int64, repoName string) (bool, error) {
	repos, _, err := r.client.ListTeamRepositories(teamID, gitea.ListTeamRepositoriesOptions{
		ListOptions: gitea.ListOptions{Page: -1}, // Get all repositories with pagination
	})
	if err != nil {
		return false, err
	}

	for _, repo := range repos {
		if repo.Name == repoName {
			return true, nil
		}
	}

	return false, nil
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

	// Get team ID by name
	teamID, err := r.findTeamByName(org, teamName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Listing Teams",
			fmt.Sprintf("Could not list teams for organization '%s': %s", org, err.Error()),
		)
		return
	}

	if teamID == 0 {
		resp.Diagnostics.AddError(
			"Team Not Found",
			fmt.Sprintf("Could not find team '%s' in organization '%s'", teamName, org),
		)
		return
	}

	// Add repository to team
	httpResp, err := r.client.AddTeamRepository(teamID, org, repoName)
	if err != nil {
		// Handle 404 - repository or team not found
		if httpResp != nil && httpResp.StatusCode == 404 {
			resp.Diagnostics.AddError(
				"Repository or Team Not Found",
				fmt.Sprintf("Could not add repository '%s' to team '%s': the repository or team was not found", repoName, teamName),
			)
			return
		}
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

	// Get team ID by name
	teamID, err := r.findTeamByName(org, teamName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Listing Teams",
			fmt.Sprintf("Could not list teams for organization '%s': %s", org, err.Error()),
		)
		return
	}

	if teamID == 0 {
		// Team no longer exists, remove from state
		resp.State.RemoveResource(ctx)
		return
	}

	// Check if repository is still assigned to team
	found, err := r.checkRepositoryInTeam(teamID, repoName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Listing Team Repositories",
			fmt.Sprintf("Could not list repositories for team '%s': %s", teamName, err.Error()),
		)
		return
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

	// Get team ID by name
	teamID, err := r.findTeamByName(org, teamName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Listing Teams",
			fmt.Sprintf("Could not list teams for organization '%s': %s", org, err.Error()),
		)
		return
	}

	if teamID == 0 {
		// Team no longer exists, nothing to delete
		return
	}

	// Remove repository from team
	httpResp, err := r.client.RemoveTeamRepository(teamID, org, repoName)
	if err != nil {
		// Handle 404 gracefully - resource already removed
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		resp.Diagnostics.AddError(
			"Error Removing Repository from Team",
			fmt.Sprintf("Could not remove repository '%s' from team '%s': %s", repoName, teamName, err.Error()),
		)
		return
	}
}

// ImportState allows importing existing team repository associations.
// Import format: "org/team_name/repository_name"
func (r *TeamRepositoryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Parse the import ID
	id := req.ID
	parts := strings.Split(id, "/")
	if len(parts) != 3 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Import ID must be in format 'org/team_name/repository_name', got: %s", id),
		)
		return
	}

	org := parts[0]
	teamName := parts[1]
	repoName := parts[2]

	// Validate org is not empty
	if org == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Organization name cannot be empty",
		)
		return
	}

	// Validate team_name is not empty
	if teamName == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Team name cannot be empty",
		)
		return
	}

	// Validate repository_name is not empty
	if repoName == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Repository name cannot be empty",
		)
		return
	}

	// Get team ID by name to verify team exists
	teamID, err := r.findTeamByName(org, teamName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Listing Teams",
			fmt.Sprintf("Could not list teams for organization '%s': %s", org, err.Error()),
		)
		return
	}

	if teamID == 0 {
		resp.Diagnostics.AddError(
			"Team Not Found",
			fmt.Sprintf("Could not find team '%s' in organization '%s'", teamName, org),
		)
		return
	}

	// Verify the repository is assigned to the team
	found, err := r.checkRepositoryInTeam(teamID, repoName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Listing Team Repositories",
			fmt.Sprintf("Could not list repositories for team '%s': %s", teamName, err.Error()),
		)
		return
	}

	if !found {
		resp.Diagnostics.AddError(
			"Team Repository Association Not Found",
			fmt.Sprintf("Repository '%s' is not assigned to team '%s' in organization '%s'", repoName, teamName, org),
		)
		return
	}

	// Set state
	state := teamRepositoryResourceModel{
		ID:             types.StringValue(fmt.Sprintf("%s/%s/%s", org, teamName, repoName)),
		Org:            types.StringValue(org),
		TeamName:       types.StringValue(teamName),
		RepositoryName: types.StringValue(repoName),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
