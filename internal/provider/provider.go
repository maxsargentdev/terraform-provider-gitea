package provider

import (
	"context"
	"os"

	//"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	//"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	//"github.com/hashicorp/terraform-plugin-framework/function"
	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ provider.Provider = (*giteaProvider)(nil)

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &giteaProvider{
			Version: version,
		}
	}
}

type giteaProvider struct {
	Version string
}

type giteaProviderModel struct {
	giteaUsername types.String `tfsdk:"gitea_username"`
	giteaPassword types.String `tfsdk:"gitea_password"`
	giteaHostname types.String `tfsdk:"gitea_hostname"`
}

func (p *giteaProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"gitea_username": schema.StringAttribute{
				Optional: false,
			},
			"gitea_password": schema.StringAttribute{
				Optional: false,
			},
			"gitea_hostname": schema.StringAttribute{
				Optional: false,
			},
		},
	}
}

func (p *giteaProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	giteaUsername := os.Getenv("GITEA_USERNAME")
	giteaPassword := os.Getenv("GITEA_PASSWORD")
	giteaHostname := os.Getenv("GITEA_HOSTNAME")

	var data giteaProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if data.giteaUsername.ValueString() != "" {
		giteaUsername = data.giteaUsername.ValueString()
	}

	if data.giteaPassword.ValueString() != "" {
		giteaPassword = data.giteaPassword.ValueString()
	}

	if data.giteaHostname.ValueString() != "" {
		giteaHostname = data.giteaHostname.ValueString()
	}

	if giteaUsername == "" {
		resp.Diagnostics.AddError(
			"Missing Username Configuration",
			"While configuring the provider, the username was not found in "+
				"the GITEA_USERNAME environment variable or provider "+
				"configuration block gitea_username attribute.",
		)
	}

	if giteaPassword == "" {
		resp.Diagnostics.AddError(
			"Missing Password Configuration",
			"While configuring the provider, the password was not found in "+
				"the GITEA_PASSWORD environment variable or provider "+
				"configuration block gitea_password attribute.",
		)
	}

	if giteaHostname == "" {
		resp.Diagnostics.AddError(
			"Missing Hostname Configuration",
			"While configuring the provider, the username was not found in "+
				"the GITEA_HOSTNAME environment variable or provider "+
				"configuration block gitea_hostname attribute.",
		)
	}

	client, err := gitea.NewClient(giteaHostname)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Gitea API Client",
			"An unexpected error occurred when creating the Gitea API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Gitea Client Error: "+err.Error(),
		)
		return
	}

	client.SetBasicAuth(giteaUsername, giteaPassword)

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *giteaProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "gitea"
}

func (p *giteaProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewUserDataSource,
	}
}

func (p *giteaProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewUserResource,
	}
}
