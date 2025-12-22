package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ provider.Provider = (*giteaProvider)(nil)

func New() func() provider.Provider {
	return func() provider.Provider {
		return &giteaProvider{}
	}
}

type giteaProvider struct{}

func (p *giteaProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {

}

func (p *giteaProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {

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
