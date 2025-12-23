package provider

import (
	"context"
	"fmt"

	"terraform-provider-gitea/internal/datasource_repository"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &repositoryDataSource{}
	_ datasource.DataSourceWithConfigure = &repositoryDataSource{}
)

func NewRepositoryDataSource() datasource.DataSource {
	return &repositoryDataSource{}
}

type repositoryDataSource struct {
	client *gitea.Client
}

func (d *repositoryDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repository"
}

func (d *repositoryDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_repository.RepositoryDataSourceSchema(ctx)
}

func (d *repositoryDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *repositoryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data datasource_repository.RepositoryModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse owner and repo from config
	// For data sources, we expect owner/repo to be provided
	owner := ""
	repoName := data.Name.ValueString()

	if !data.FullName.IsNull() && data.FullName.ValueString() != "" {
		fullName := data.FullName.ValueString()
		for i, c := range fullName {
			if c == '/' {
				owner = fullName[:i]
				repoName = fullName[i+1:]
				break
			}
		}
	}

	if owner == "" {
		resp.Diagnostics.AddError(
			"Missing Owner",
			"Repository owner must be specified via full_name (owner/repo format)",
		)
		return
	}

	repo, _, err := d.client.GetRepo(owner, repoName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Repository",
			"Could not read repository "+owner+"/"+repoName+": "+err.Error(),
		)
		return
	}

	// Map response to data
	data.Id = types.Int64Value(repo.ID)
	data.Name = types.StringValue(repo.Name)
	data.FullName = types.StringValue(repo.FullName)
	data.Description = types.StringValue(repo.Description)
	data.Private = types.BoolValue(repo.Private)
	data.DefaultBranch = types.StringValue(repo.DefaultBranch)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
