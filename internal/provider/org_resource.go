package provider

import (
	"context"
	"fmt"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &orgResource{}
	_ resource.ResourceWithConfigure   = &orgResource{}
	_ resource.ResourceWithImportState = &orgResource{}
)

func NewOrgResource() resource.Resource {
	return &orgResource{}
}

type orgResource struct {
	client *gitea.Client
}

type orgResourceModel struct {
	// Required
	Name types.String `tfsdk:"name"`

	// Optional
	Description types.String `tfsdk:"description"`
	FullName    types.String `tfsdk:"full_name"`
	Location    types.String `tfsdk:"location"`
	Visibility  types.String `tfsdk:"visibility"`
	Website     types.String `tfsdk:"website"`

	// Computed
	AvatarUrl types.String `tfsdk:"avatar_url"`
	Id        types.String `tfsdk:"id"`
	Repos     types.List   `tfsdk:"repos"`
}

func (r *orgResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_org"
}

func (r *orgResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manages a Gitea organization.",
		MarkdownDescription: "Manages a Gitea organization. This resource allows you to create, update, and delete organizations in Gitea.",
		Attributes: map[string]schema.Attribute{
			// Required
			"name": schema.StringAttribute{
				Required:            true,
				Description:         "The name of the organization.",
				MarkdownDescription: "The name of the organization.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			// Optional
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				Description:         "Description of the organization.",
				MarkdownDescription: "Description of the organization.",
			},
			"full_name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The full (display) name of the organization.",
				MarkdownDescription: "The full (display) name of the organization.",
			},
			"location": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				Description:         "Location of the organization.",
				MarkdownDescription: "Location of the organization.",
			},
			"visibility": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("public"),
				Description:         "Visibility of the organization (public, limited, private).",
				MarkdownDescription: "Visibility of the organization (`public`, `limited`, `private`).",
				Validators: []validator.String{
					stringvalidator.OneOf("public", "limited", "private"),
				},
			},
			"website": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				Description:         "Website of the organization.",
				MarkdownDescription: "Website of the organization.",
			},

			// Computed
			"avatar_url": schema.StringAttribute{
				Computed:            true,
				Description:         "The URL of the organization's avatar.",
				MarkdownDescription: "The URL of the organization's avatar.",
			},
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "The ID of the organization.",
				MarkdownDescription: "The ID of the organization.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"repos": schema.ListAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				Description:         "List of repository names belonging to the organization.",
				MarkdownDescription: "List of repository names belonging to the organization.",
			},
		},
	}
}

func (r *orgResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Helper function to get organization repos
func (r *orgResource) getOrgRepos(ctx context.Context, orgName string) ([]string, error) {
	repos, _, err := r.client.ListOrgRepos(orgName, gitea.ListOrgReposOptions{
		ListOptions: gitea.ListOptions{Page: -1},
	})
	if err != nil {
		return nil, err
	}

	repoNames := make([]string, 0, len(repos))
	for _, repo := range repos {
		repoNames = append(repoNames, repo.Name)
	}
	return repoNames, nil
}

// Helper function to map Gitea Organization to Terraform model
func (r *orgResource) mapOrgToModel(ctx context.Context, org *gitea.Organization, model *orgResourceModel) error {
	model.Id = types.StringValue(fmt.Sprintf("%d", org.ID))
	model.Name = types.StringValue(org.UserName)
	model.Description = types.StringValue(org.Description)
	model.FullName = types.StringValue(org.FullName)
	model.Location = types.StringValue(org.Location)
	model.Visibility = types.StringValue(string(org.Visibility))
	model.Website = types.StringValue(org.Website)
	model.AvatarUrl = types.StringValue(org.AvatarURL)

	// Get repos
	repoNames, err := r.getOrgRepos(ctx, org.UserName)
	if err != nil {
		return err
	}

	reposList, diags := types.ListValueFrom(ctx, types.StringType, repoNames)
	if diags.HasError() {
		return fmt.Errorf("error converting repos to list")
	}
	model.Repos = reposList

	return nil
}

func (r *orgResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan orgResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createOpts := gitea.CreateOrgOption{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		FullName:    plan.FullName.ValueString(),
		Location:    plan.Location.ValueString(),
		Visibility:  gitea.VisibleType(plan.Visibility.ValueString()),
		Website:     plan.Website.ValueString(),
	}

	org, _, err := r.client.CreateOrg(createOpts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Organization",
			"Could not create organization, unexpected error: "+err.Error(),
		)
		return
	}

	if err := r.mapOrgToModel(ctx, org, &plan); err != nil {
		resp.Diagnostics.AddError(
			"Error Mapping Organization",
			"Could not map organization response: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *orgResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state orgResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org, httpResp, err := r.client.GetOrg(state.Name.ValueString())
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Organization",
			"Could not read organization: "+err.Error(),
		)
		return
	}

	if err := r.mapOrgToModel(ctx, org, &state); err != nil {
		resp.Diagnostics.AddError(
			"Error Mapping Organization",
			"Could not map organization response: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *orgResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan orgResourceModel
	var state orgResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	editOpts := gitea.EditOrgOption{
		Description: plan.Description.ValueString(),
		FullName:    plan.FullName.ValueString(),
		Location:    plan.Location.ValueString(),
		Visibility:  gitea.VisibleType(plan.Visibility.ValueString()),
		Website:     plan.Website.ValueString(),
	}

	_, err := r.client.EditOrg(state.Name.ValueString(), editOpts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Organization",
			"Could not update organization: "+err.Error(),
		)
		return
	}

	// Read back the organization
	org, _, err := r.client.GetOrg(state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Organization After Update",
			"Could not read organization: "+err.Error(),
		)
		return
	}

	if err := r.mapOrgToModel(ctx, org, &plan); err != nil {
		resp.Diagnostics.AddError(
			"Error Mapping Organization",
			"Could not map organization response: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *orgResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state orgResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DeleteOrg(state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Organization",
			"Could not delete organization: "+err.Error(),
		)
		return
	}
}

func (r *orgResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by organization name
	orgName := req.ID

	if orgName == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Organization name cannot be empty",
		)
		return
	}

	// Fetch the organization
	org, httpResp, err := r.client.GetOrg(orgName)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			resp.Diagnostics.AddError(
				"Organization Not Found",
				fmt.Sprintf("Organization '%s' does not exist or is not accessible", orgName),
			)
			return
		}
		resp.Diagnostics.AddError(
			"Error Importing Organization",
			fmt.Sprintf("Could not import organization '%s': %s", orgName, err.Error()),
		)
		return
	}

	var data orgResourceModel
	if err := r.mapOrgToModel(ctx, org, &data); err != nil {
		resp.Diagnostics.AddError(
			"Error Mapping Organization",
			"Could not map organization response: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
