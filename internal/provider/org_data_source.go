package provider

import (
	"context"
	"fmt"
	"net/http"

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
func mapOrgToDataSourceModel(org *gitea.Organization, model *orgDataSourceModel) {
	model.Id = types.Int64Value(org.ID)
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

type orgDataSourceModel struct {
	AvatarUrl                 types.String `tfsdk:"avatar_url"`
	Description               types.String `tfsdk:"description"`
	Email                     types.String `tfsdk:"email"`
	FullName                  types.String `tfsdk:"full_name"`
	Id                        types.Int64  `tfsdk:"id"`
	Location                  types.String `tfsdk:"location"`
	Name                      types.String `tfsdk:"name"`
	RepoAdminChangeTeamAccess types.Bool   `tfsdk:"repo_admin_change_team_access"`
	Visibility                types.String `tfsdk:"visibility"`
	Website                   types.String `tfsdk:"website"`
}

func (d *orgDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_org"
}

func (d *orgDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Use this data source to retrieve information about an existing Gitea organization.",
		MarkdownDescription: "Use this data source to retrieve information about an existing Gitea organization.",
		Attributes: map[string]schema.Attribute{
			// Required input
			"name": schema.StringAttribute{
				Required:            true,
				Description:         "The name of the organization to look up.",
				MarkdownDescription: "The name of the organization to look up.",
			},

			// Computed outputs
			"id": schema.Int64Attribute{
				Computed:            true,
				Description:         "The numeric ID of the organization.",
				MarkdownDescription: "The numeric ID of the organization.",
			},
			"avatar_url": schema.StringAttribute{
				Computed:            true,
				Description:         "URL to the organization's avatar image.",
				MarkdownDescription: "URL to the organization's avatar image.",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				Description:         "The description of the organization.",
				MarkdownDescription: "The description of the organization.",
			},
			"email": schema.StringAttribute{
				Computed:            true,
				Description:         "The email address of the organization.",
				MarkdownDescription: "The email address of the organization.",
			},
			"full_name": schema.StringAttribute{
				Computed:            true,
				Description:         "The full display name of the organization.",
				MarkdownDescription: "The full display name of the organization.",
			},
			"location": schema.StringAttribute{
				Computed:            true,
				Description:         "The location of the organization.",
				MarkdownDescription: "The location of the organization.",
			},
			"repo_admin_change_team_access": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether repository administrators can change team access permissions.",
				MarkdownDescription: "Whether repository administrators can change team access permissions.",
			},
			"visibility": schema.StringAttribute{
				Computed:            true,
				Description:         "The visibility level of the organization (public, limited, or private).",
				MarkdownDescription: "The visibility level of the organization (`public`, `limited`, or `private`).",
			},
			"website": schema.StringAttribute{
				Computed:            true,
				Description:         "The website URL of the organization.",
				MarkdownDescription: "The website URL of the organization.",
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
	var data orgDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgName := data.Name.ValueString()

	// Get org from Gitea API
	org, httpResp, err := d.client.GetOrg(orgName)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			resp.Diagnostics.AddError(
				"Organization Not Found",
				fmt.Sprintf("Organization '%s' does not exist or you do not have permission to access it.", orgName),
			)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Organization",
			fmt.Sprintf("Could not read organization '%s': %s", orgName, err.Error()),
		)
		return
	}

	// Map response to model
	mapOrgToDataSourceModel(org, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
