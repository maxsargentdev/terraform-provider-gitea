package provider

import (
	"context"
	"fmt"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &repositoryDataSource{}
	_ datasource.DataSourceWithConfigure = &repositoryDataSource{}
)

func NewRepositoryDataSource() datasource.DataSource {
	return &repositoryDataSource{}
}

// Helper function to map Gitea Repository to Terraform data source model
func mapRepositoryToDataSourceModel(repo *gitea.Repository, model *RepositoryModel) {
	model.Id = types.Int64Value(repo.ID)
	model.Name = types.StringValue(repo.Name)
	model.FullName = types.StringValue(repo.FullName)
	model.Description = types.StringValue(repo.Description)
	model.Private = types.BoolValue(repo.Private)
	model.DefaultBranch = types.StringValue(repo.DefaultBranch)
	model.Website = types.StringValue(repo.Website)
	model.HtmlUrl = types.StringValue(repo.HTMLURL)
	model.CloneUrl = types.StringValue(repo.CloneURL)
	model.SshUrl = types.StringValue(repo.SSHURL)
	model.Empty = types.BoolValue(repo.Empty)
	model.Fork = types.BoolValue(repo.Fork)
	model.Mirror = types.BoolValue(repo.Mirror)
	model.Size = types.Int64Value(int64(repo.Size))
	model.Archived = types.BoolValue(repo.Archived)
	model.StarsCount = types.Int64Value(int64(repo.Stars))
	model.WatchersCount = types.Int64Value(int64(repo.Watchers))
	model.ForksCount = types.Int64Value(int64(repo.Forks))
	model.OpenIssuesCount = types.Int64Value(int64(repo.OpenIssues))
	model.AvatarUrl = types.StringValue(repo.AvatarURL)
	model.Template = types.BoolValue(repo.Template)
	model.Internal = types.BoolValue(repo.Internal)
}

type repositoryDataSource struct {
	client *gitea.Client
}

func (d *repositoryDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repository"
}

func (d *repositoryDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = RepositoryDataSourceSchema(ctx)
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
	var data RepositoryModel

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
	mapRepositoryToDataSourceModel(repo, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func RepositoryDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"allow_fast_forward_only_merge": schema.BoolAttribute{
				Computed: true,
			},
			"allow_manual_merge": schema.BoolAttribute{
				Computed: true,
			},
			"allow_merge_commits": schema.BoolAttribute{
				Computed: true,
			},
			"allow_rebase": schema.BoolAttribute{
				Computed: true,
			},
			"allow_rebase_explicit": schema.BoolAttribute{
				Computed: true,
			},
			"allow_rebase_update": schema.BoolAttribute{
				Computed: true,
			},
			"allow_squash_merge": schema.BoolAttribute{
				Computed: true,
			},
			"archived": schema.BoolAttribute{
				Computed: true,
			},
			"archived_at": schema.StringAttribute{
				Computed: true,
			},
			"autodetect_manual_merge": schema.BoolAttribute{
				Computed: true,
			},
			"avatar_url": schema.StringAttribute{
				Computed: true,
			},
			"clone_url": schema.StringAttribute{
				Computed: true,
			},
			"created_at": schema.StringAttribute{
				Computed: true,
			},
			"default_allow_maintainer_edit": schema.BoolAttribute{
				Computed: true,
			},
			"default_branch": schema.StringAttribute{
				Computed: true,
			},
			"default_delete_branch_after_merge": schema.BoolAttribute{
				Computed: true,
			},
			"default_merge_style": schema.StringAttribute{
				Computed: true,
			},
			"description": schema.StringAttribute{
				Computed: true,
			},
			"empty": schema.BoolAttribute{
				Computed: true,
			},
			"fork": schema.BoolAttribute{
				Computed: true,
			},
			"forks_count": schema.Int64Attribute{
				Computed: true,
			},
			"full_name": schema.StringAttribute{
				Computed: true,
			},
			"has_actions": schema.BoolAttribute{
				Computed: true,
			},
			"has_code": schema.BoolAttribute{
				Computed: true,
			},
			"has_issues": schema.BoolAttribute{
				Computed: true,
			},
			"has_packages": schema.BoolAttribute{
				Computed: true,
			},
			"has_projects": schema.BoolAttribute{
				Computed: true,
			},
			"has_pull_requests": schema.BoolAttribute{
				Computed: true,
			},
			"has_releases": schema.BoolAttribute{
				Computed: true,
			},
			"has_wiki": schema.BoolAttribute{
				Computed: true,
			},
			"html_url": schema.StringAttribute{
				Computed: true,
			},
			"id": schema.Int64Attribute{
				Computed: true,
			},
			"ignore_whitespace_conflicts": schema.BoolAttribute{
				Computed: true,
			},
			"internal": schema.BoolAttribute{
				Computed: true,
			},
			"language": schema.StringAttribute{
				Computed: true,
			},
			"languages_url": schema.StringAttribute{
				Computed: true,
			},
			"licenses": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"link": schema.StringAttribute{
				Computed: true,
			},
			"mirror": schema.BoolAttribute{
				Computed: true,
			},
			"mirror_interval": schema.StringAttribute{
				Computed: true,
			},
			"mirror_updated": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Computed: true,
			},
			"object_format_name": schema.StringAttribute{
				Computed:            true,
				Description:         "ObjectFormatName of the underlying git repository",
				MarkdownDescription: "ObjectFormatName of the underlying git repository",
			},
			"open_issues_count": schema.Int64Attribute{
				Computed: true,
			},
			"open_pr_counter": schema.Int64Attribute{
				Computed: true,
			},
			"original_url": schema.StringAttribute{
				Computed: true,
			},
			"private": schema.BoolAttribute{
				Computed: true,
			},
			"projects_mode": schema.StringAttribute{
				Computed: true,
			},
			"release_counter": schema.Int64Attribute{
				Computed: true,
			},
			"repo": schema.StringAttribute{
				Required:            true,
				Description:         "name of the repo",
				MarkdownDescription: "name of the repo",
			},
			"size": schema.Int64Attribute{
				Computed: true,
			},
			"ssh_url": schema.StringAttribute{
				Computed: true,
			},
			"stars_count": schema.Int64Attribute{
				Computed: true,
			},
			"template": schema.BoolAttribute{
				Computed: true,
			},
			"topics": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Computed: true,
			},
			"url": schema.StringAttribute{
				Computed: true,
			},
			"watchers_count": schema.Int64Attribute{
				Computed: true,
			},
			"website": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}
