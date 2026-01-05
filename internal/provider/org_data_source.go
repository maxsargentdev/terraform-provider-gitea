package provider

import (
	"context"
	"fmt"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = (*orgDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*orgDataSource)(nil)

func NewOrgDataSource() datasource.DataSource {
	return &orgDataSource{}
}

// Helper function to map Gitea Organization to Terraform data source model
func mapOrgToDataSourceModel(org *gitea.Organization, model *OrgDataSourceModel) {
	model.Id = types.Int64Value(org.ID)
	model.Org = types.StringValue(org.UserName)
	model.Username = types.StringValue(org.UserName)
	model.Name = types.StringValue(org.UserName)
	model.FullName = types.StringValue(org.FullName)
	model.Description = types.StringValue(org.Description)
	model.Website = types.StringValue(org.Website)
	model.Location = types.StringValue(org.Location)
	model.AvatarUrl = types.StringValue(org.AvatarURL)
	model.Visibility = types.StringValue(org.Visibility)
	model.RepoAdminChangeTeamAccess = types.BoolNull()
	model.Email = types.StringNull()
}

type orgDataSource struct {
	client *gitea.Client
}

func (d *orgDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_org"
}

func (d *orgDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"avatar_url": schema.StringAttribute{
				Computed:            true,
				Description:         "The URL of the organization's avatar",
				MarkdownDescription: "The URL of the organization's avatar",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				Description:         "The description of the organization",
				MarkdownDescription: "The description of the organization",
			},
			"email": schema.StringAttribute{
				Computed:            true,
				Description:         "The email address of the organization",
				MarkdownDescription: "The email address of the organization",
			},
			"full_name": schema.StringAttribute{
				Computed:            true,
				Description:         "The full display name of the organization",
				MarkdownDescription: "The full display name of the organization",
			},
			"id": schema.Int64Attribute{
				Computed:            true,
				Description:         "The unique identifier of the organization",
				MarkdownDescription: "The unique identifier of the organization",
			},
			"location": schema.StringAttribute{
				Computed:            true,
				Description:         "The location of the organization",
				MarkdownDescription: "The location of the organization",
			},
			"name": schema.StringAttribute{
				Computed:            true,
				Description:         "The name of the organization",
				MarkdownDescription: "The name of the organization",
			},
			"org": schema.StringAttribute{
				Required:            true,
				Description:         "name of the organization to get",
				MarkdownDescription: "name of the organization to get",
			},
			"repo_admin_change_team_access": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether repository administrators can change team access",
				MarkdownDescription: "Whether repository administrators can change team access",
			},
			"username": schema.StringAttribute{
				Computed:            true,
				Description:         "username of the organization\ndeprecated",
				MarkdownDescription: "username of the organization\ndeprecated",
			},
			"visibility": schema.StringAttribute{
				Computed:            true,
				Description:         "The visibility level of the organization (public, limited, private)",
				MarkdownDescription: "The visibility level of the organization (public, limited, private)",
			},
			"website": schema.StringAttribute{
				Computed:            true,
				Description:         "The website URL of the organization",
				MarkdownDescription: "The website URL of the organization",
			},
		},
	}
}

func (d *orgDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*gitea.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *gitea.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *orgDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data OrgDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The schema uses "org" field as the org name to query
	orgName := data.Org.ValueString()

	// Get org from Gitea API
	org, _, err := d.client.GetOrg(orgName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Organization",
			"Could not read organization "+orgName+": "+err.Error(),
		)
		return
	}

	// Map response to model
	mapOrgToDataSourceModel(org, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type OrgDataSourceModel struct {
	AvatarUrl                 types.String `tfsdk:"avatar_url"`
	Description               types.String `tfsdk:"description"`
	Email                     types.String `tfsdk:"email"`
	FullName                  types.String `tfsdk:"full_name"`
	Id                        types.Int64  `tfsdk:"id"`
	Location                  types.String `tfsdk:"location"`
	Name                      types.String `tfsdk:"name"`
	Org                       types.String `tfsdk:"org"`
	RepoAdminChangeTeamAccess types.Bool   `tfsdk:"repo_admin_change_team_access"`
	Username                  types.String `tfsdk:"username"`
	Visibility                types.String `tfsdk:"visibility"`
	Website                   types.String `tfsdk:"website"`
}
