//TODO: We need to verify the user running the provider has admin

package provider

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
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
	GiteaUsername      types.String `tfsdk:"gitea_username"`
	GiteaPassword      types.String `tfsdk:"gitea_password"`
	GiteaHostname      types.String `tfsdk:"gitea_hostname"`
	InsecureSkipVerify types.Bool   `tfsdk:"insecure_skip_verify"`
	CACertFile         types.String `tfsdk:"ca_cert_file"`
}

func (p *giteaProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"gitea_username": schema.StringAttribute{
				Required:            true,
				Description:         "The username for authentication with the Gitea server.",
				MarkdownDescription: "The username for authentication with the Gitea server.",
			},
			"gitea_password": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				Description:         "The password for authentication with the Gitea server.",
				MarkdownDescription: "The password for authentication with the Gitea server.",
			},
			"gitea_hostname": schema.StringAttribute{
				Required:            true,
				Description:         "The hostname/URL of the Gitea server (e.g., https://gitea.example.com).",
				MarkdownDescription: "The hostname/URL of the Gitea server (e.g., `https://gitea.example.com`).",
			},
			"insecure_skip_verify": schema.BoolAttribute{
				Optional:            true,
				Description:         "Skip TLS certificate verification. Not recommended for production use.",
				MarkdownDescription: "Skip TLS certificate verification. **Not recommended for production use.**",
			},
			"ca_cert_file": schema.StringAttribute{
				Optional:            true,
				Description:         "Path to a custom CA certificate file to use for TLS verification.",
				MarkdownDescription: "Path to a custom CA certificate file to use for TLS verification.",
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

	// Configure TLS settings
	tlsConfig := &tls.Config{}

	// Handle insecure skip verify
	if !data.InsecureSkipVerify.IsNull() && data.InsecureSkipVerify.ValueBool() {
		tlsConfig.InsecureSkipVerify = true
	}

	// Handle custom CA certificate
	if !data.CACertFile.IsNull() && data.CACertFile.ValueString() != "" {
		caCertFile := data.CACertFile.ValueString()
		caCert, err := os.ReadFile(caCertFile)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Read CA Certificate File",
				fmt.Sprintf("Could not read CA certificate file '%s': %s", caCertFile, err.Error()),
			)
			return
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			resp.Diagnostics.AddError(
				"Invalid CA Certificate",
				fmt.Sprintf("Could not parse CA certificate from file '%s'", caCertFile),
			)
			return
		}
		tlsConfig.RootCAs = caCertPool
	}

	// Create HTTP client with TLS configuration
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	client, err := gitea.NewClient(
		giteaHostname,
		gitea.SetHTTPClient(httpClient),
		gitea.SetBasicAuth(giteaUsername, giteaPassword),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Gitea API Client",
			"An unexpected error occurred when creating the Gitea API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Gitea Client Error: "+err.Error(),
		)
		return
	}

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
		NewRepositoriesDataSource,
		NewTeamsDataSource,
	}
}

func (p *giteaProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewUserResource,
		NewOrgResource,
		NewRepositoryResource,
		NewRepositoryBranchProtectionResource,
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
