package provider

import (
	"context"
	"fmt"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = (*orgResource)(nil)
var _ resource.ResourceWithConfigure = (*orgResource)(nil)
var _ resource.ResourceWithImportState = (*orgResource)(nil)

func NewOrgResource() resource.Resource {
	return &orgResource{}
}

// Helper function to map Gitea Organization to Terraform model
func mapOrgToModel(org *gitea.Organization, model *orgResourceModel) {
	model.Id = types.Int64Value(org.ID)
	model.Name = types.StringValue(org.UserName)
	model.DisplayName = types.StringValue(org.FullName)
	model.Description = types.StringValue(org.Description)
	model.Website = types.StringValue(org.Website)
	model.Location = types.StringValue(org.Location)
	model.AvatarUrl = types.StringValue(org.AvatarURL)
	model.Visibility = types.StringValue(org.Visibility)
	model.RepoAdminChangeTeamAccess = types.BoolNull()
	model.Email = types.StringNull()
}

type orgResource struct {
	client *gitea.Client
}

type orgResourceModel struct {
	Name                      types.String `tfsdk:"name"`
	AvatarUrl                 types.String `tfsdk:"avatar_url"`
	Description               types.String `tfsdk:"description"`
	Email                     types.String `tfsdk:"email"`
	DisplayName               types.String `tfsdk:"display_name"`
	Id                        types.Int64  `tfsdk:"id"`
	Location                  types.String `tfsdk:"location"`
	RepoAdminChangeTeamAccess types.Bool   `tfsdk:"repo_admin_change_team_access"`
	Visibility                types.String `tfsdk:"visibility"`
	Website                   types.String `tfsdk:"website"`
}

func (r *orgResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_org"
}

func (r *orgResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{

			// required - these are fundamental configuration options
			"name": schema.StringAttribute{
				Required:            true,
				Description:         "The name of the organization",
				MarkdownDescription: "The name of the organization",
			},

			// optional - these tweak the created resource away from its defaults
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The description of the organization",
				MarkdownDescription: "The description of the organization",
			},
			"email": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The email address of the organization",
				MarkdownDescription: "The email address of the organization",
			},
			"display_name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The full display name of the organization",
				MarkdownDescription: "The full display name of the organization",
			},
			"location": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The location of the organization",
				MarkdownDescription: "The location of the organization",
			},
			"repo_admin_change_team_access": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Whether repository administrators can change team access",
				MarkdownDescription: "Whether repository administrators can change team access",
			},
			"visibility": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "possible values are `public` (default), `limited` or `private`",
				MarkdownDescription: "possible values are `public` (default), `limited` or `private`",
				Validators: []validator.String{
					stringvalidator.OneOf(
						"public",
						"limited",
						"private",
					),
				},
			},
			"website": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The website URL of the organization",
				MarkdownDescription: "The website URL of the organization",
			},

			// computed - these are available to read back after creation but are really just metadata
			"avatar_url": schema.StringAttribute{
				Computed:            true,
				Description:         "The URL of the organization's avatar",
				MarkdownDescription: "The URL of the organization's avatar",
			},
			"id": schema.Int64Attribute{
				Computed:            true,
				Description:         "The unique identifier of the organization",
				MarkdownDescription: "The unique identifier of the organization",
			},
		},
	}
}

func (r *orgResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *orgResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data orgResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create org via Gitea API
	createOpts := gitea.CreateOrgOption{
		Name:                      data.Name.ValueString(),
		FullName:                  data.DisplayName.ValueString(),
		Description:               data.Description.ValueString(),
		Website:                   data.Website.ValueString(),
		Location:                  data.Location.ValueString(),
		Visibility:                gitea.VisibleType(data.Visibility.ValueString()),
		RepoAdminChangeTeamAccess: data.RepoAdminChangeTeamAccess.ValueBool(),
	}

	org, _, err := r.client.CreateOrg(createOpts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Organization",
			"Could not create organization, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response to model
	mapOrgToModel(org, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *orgResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data orgResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get org from Gitea API
	org, _, err := r.client.GetOrg(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Organization",
			"Could not read organization "+data.Name.ValueString()+": "+err.Error(),
		)
		return
	}

	// Map response to model
	mapOrgToModel(org, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *orgResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data orgResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update org via Gitea API
	editOpts := gitea.EditOrgOption{
		FullName:    data.DisplayName.ValueString(),
		Description: data.Description.ValueString(),
		Website:     data.Website.ValueString(),
		Location:    data.Location.ValueString(),
		Visibility:  gitea.VisibleType(data.Visibility.ValueString()),
	}

	_, err := r.client.EditOrg(data.Name.ValueString(), editOpts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Organization",
			"Could not update organization "+data.Name.ValueString()+": "+err.Error(),
		)
		return
	}

	// Read back the org to get updated values
	org, _, err := r.client.GetOrg(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Organization After Update",
			"Could not read organization "+data.Name.ValueString()+": "+err.Error(),
		)
		return
	}

	// Map response to model
	mapOrgToModel(org, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *orgResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data orgResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete org via Gitea API
	_, err := r.client.DeleteOrg(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Organization",
			"Could not delete organization "+data.Name.ValueString()+": "+err.Error(),
		)
		return
	}
}

func (r *orgResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import using the organization username
	orgName := req.ID

	// Fetch the organization from Gitea
	org, _, err := r.client.GetOrg(orgName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Importing Organization",
			"Could not import organization "+orgName+": "+err.Error(),
		)
		return
	}

	// Map to model
	var data orgResourceModel
	mapOrgToModel(org, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
