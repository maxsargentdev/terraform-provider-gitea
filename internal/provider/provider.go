//TODO: We need to verify the user running the provider has admin

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

// Ensure the implementation satisfies the expected interfaces.
var _ provider.Provider = (*giteaProvider)(nil)

// New is a helper function to simplify provider server and testing implementation.
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
	GiteaUsername types.String `tfsdk:"gitea_username"`
	GiteaPassword types.String `tfsdk:"gitea_password"`
	GiteaHostname types.String `tfsdk:"gitea_hostname"`
}

func (p *giteaProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"gitea_username": schema.StringAttribute{
				Required: true,
			},
			"gitea_password": schema.StringAttribute{
				Required: true,
			},
			"gitea_hostname": schema.StringAttribute{
				Required: true,
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

	if data.GiteaUsername.ValueString() != "" {
		giteaUsername = data.GiteaUsername.ValueString()
	}

	if data.GiteaPassword.ValueString() != "" {
		giteaPassword = data.GiteaPassword.ValueString()
	}

	if data.GiteaHostname.ValueString() != "" {
		giteaHostname = data.GiteaHostname.ValueString()
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
		NewOrgDataSource,
		NewRepositoryDataSource,
		NewBranchProtectionDataSource,
		NewTeamDataSource,
		NewTeamMembershipDataSource,
	}
}

func (p *giteaProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewUserResource,
		NewOrgResource,
		NewRepositoryResource,
		NewBranchProtectionResource,
		NewTeamResource,
		NewTeamRepositoryResource,
		NewTokenResource,
		NewTeamMembershipResource,
		NewPublicKeyResource,
		NewGPGKeyResource,
		NewRepositoryKeyResource,
		NewOAuth2AppResource,
		NewRepositoryWebhookResource,
		NewRepositoryActionsSecretResource,
		NewRepositoryActionsVariableResource,
		NewForkResource,
		NewGitHookResource,
	}
}
