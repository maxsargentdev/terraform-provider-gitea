package provider

import (
	"context"
	"fmt"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = (*repositoriesDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*repositoriesDataSource)(nil)

func NewRepositoriesDataSource() datasource.DataSource {
	return &repositoriesDataSource{}
}

type repositoriesDataSource struct {
	client *gitea.Client
}

type repositoriesDataSourceModel struct {
	Username types.String       `tfsdk:"username"`
	Org      types.String       `tfsdk:"org"`
	Search   types.String       `tfsdk:"search"`
	Repos    []repositoryIDName `tfsdk:"repositories"`
}

type repositoryIDName struct {
	Id          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	FullName    types.String `tfsdk:"full_name"`
	Owner       types.String `tfsdk:"owner"`
	Description types.String `tfsdk:"description"`
	Private     types.Bool   `tfsdk:"private"`
	Fork        types.Bool   `tfsdk:"fork"`
	Mirror      types.Bool   `tfsdk:"mirror"`
	HtmlUrl     types.String `tfsdk:"html_url"`
	SshUrl      types.String `tfsdk:"ssh_url"`
	CloneUrl    types.String `tfsdk:"clone_url"`
}

func (d *repositoriesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repositories"
}

func (d *repositoriesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a list of repositories. Can list repos for a user, organization, or search all repos.",
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Username to list repositories for (mutually exclusive with org and search)",
			},
			"org": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Organization to list repositories for (mutually exclusive with username and search)",
			},
			"search": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Search query for repositories (mutually exclusive with username and org)",
			},
			"repositories": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of repositories",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Repository ID",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Repository name",
						},
						"full_name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Full name (owner/name)",
						},
						"owner": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Repository owner",
						},
						"description": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Repository description",
						},
						"private": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether the repository is private",
						},
						"fork": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether the repository is a fork",
						},
						"mirror": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether the repository is a mirror",
						},
						"html_url": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "HTML URL of the repository",
						},
						"ssh_url": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "SSH URL of the repository",
						},
						"clone_url": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Clone URL of the repository",
						},
					},
				},
			},
		},
	}
}

func (d *repositoriesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *repositoriesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data repositoriesDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that only one filter is specified
	filtersSet := 0
	if !data.Username.IsNull() {
		filtersSet++
	}
	if !data.Org.IsNull() {
		filtersSet++
	}
	if !data.Search.IsNull() {
		filtersSet++
	}

	if filtersSet > 1 {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"Only one of 'username', 'org', or 'search' can be specified",
		)
		return
	}

	var repos []*gitea.Repository
	var err error

	if !data.Username.IsNull() {
		// List repos for a specific user
		repos, _, err = d.client.ListUserRepos(data.Username.ValueString(), gitea.ListReposOptions{
			ListOptions: gitea.ListOptions{Page: -1},
		})
	} else if !data.Org.IsNull() {
		// List repos for an organization
		repos, _, err = d.client.ListOrgRepos(data.Org.ValueString(), gitea.ListOrgReposOptions{
			ListOptions: gitea.ListOptions{Page: -1},
		})
	} else if !data.Search.IsNull() {
		// Search for repos
		repos, _, err = d.client.SearchRepos(gitea.SearchRepoOptions{
			Keyword:     data.Search.ValueString(),
			ListOptions: gitea.ListOptions{Page: -1},
		})
	} else {
		// List all repos for the authenticated user
		repos, _, err = d.client.ListMyRepos(gitea.ListReposOptions{
			ListOptions: gitea.ListOptions{Page: -1},
		})
	}

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list repositories, got error: %s", err))
		return
	}

	// Map repos to model
	data.Repos = make([]repositoryIDName, len(repos))
	for i, repo := range repos {
		data.Repos[i] = repositoryIDName{
			Id:          types.Int64Value(repo.ID),
			Name:        types.StringValue(repo.Name),
			FullName:    types.StringValue(repo.FullName),
			Owner:       types.StringValue(repo.Owner.UserName),
			Description: types.StringValue(repo.Description),
			Private:     types.BoolValue(repo.Private),
			Fork:        types.BoolValue(repo.Fork),
			Mirror:      types.BoolValue(repo.Mirror),
			HtmlUrl:     types.StringValue(repo.HTMLURL),
			SshUrl:      types.StringValue(repo.SSHURL),
			CloneUrl:    types.StringValue(repo.CloneURL),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
