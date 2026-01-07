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

type repositoryDataSourceModel struct {
	// Required inputs
	Owner types.String `tfsdk:"owner"`
	Name  types.String `tfsdk:"name"`

	// Computed outputs
	Id                            types.Int64  `tfsdk:"id"`
	FullName                      types.String `tfsdk:"full_name"`
	Description                   types.String `tfsdk:"description"`
	Private                       types.Bool   `tfsdk:"private"`
	Fork                          types.Bool   `tfsdk:"fork"`
	Mirror                        types.Bool   `tfsdk:"mirror"`
	Template                      types.Bool   `tfsdk:"template"`
	Internal                      types.Bool   `tfsdk:"internal"`
	Empty                         types.Bool   `tfsdk:"empty"`
	Archived                      types.Bool   `tfsdk:"archived"`
	DefaultBranch                 types.String `tfsdk:"default_branch"`
	Website                       types.String `tfsdk:"website"`
	HtmlUrl                       types.String `tfsdk:"html_url"`
	CloneUrl                      types.String `tfsdk:"clone_url"`
	SshUrl                        types.String `tfsdk:"ssh_url"`
	Size                          types.Int64  `tfsdk:"size"`
	StarsCount                    types.Int64  `tfsdk:"stars_count"`
	WatchersCount                 types.Int64  `tfsdk:"watchers_count"`
	ForksCount                    types.Int64  `tfsdk:"forks_count"`
	OpenIssuesCount               types.Int64  `tfsdk:"open_issues_count"`
	OpenPrCounter                 types.Int64  `tfsdk:"open_pr_counter"`
	ReleaseCounter                types.Int64  `tfsdk:"release_counter"`
	AvatarUrl                     types.String `tfsdk:"avatar_url"`
	ObjectFormatName              types.String `tfsdk:"object_format_name"`
	HasIssues                     types.Bool   `tfsdk:"has_issues"`
	HasWiki                       types.Bool   `tfsdk:"has_wiki"`
	HasPullRequests               types.Bool   `tfsdk:"has_pull_requests"`
	HasProjects                   types.Bool   `tfsdk:"has_projects"`
	HasReleases                   types.Bool   `tfsdk:"has_releases"`
	HasPackages                   types.Bool   `tfsdk:"has_packages"`
	HasActions                    types.Bool   `tfsdk:"has_actions"`
	IgnoreWhitespaceConflicts     types.Bool   `tfsdk:"ignore_whitespace_conflicts"`
	AllowMergeCommits             types.Bool   `tfsdk:"allow_merge_commits"`
	AllowRebase                   types.Bool   `tfsdk:"allow_rebase"`
	AllowRebaseExplicit           types.Bool   `tfsdk:"allow_rebase_explicit"`
	AllowRebaseUpdate             types.Bool   `tfsdk:"allow_rebase_update"`
	AllowSquashMerge              types.Bool   `tfsdk:"allow_squash_merge"`
	AllowFastForwardOnlyMerge     types.Bool   `tfsdk:"allow_fast_forward_only_merge"`
	AllowManualMerge              types.Bool   `tfsdk:"allow_manual_merge"`
	AutodetectManualMerge         types.Bool   `tfsdk:"autodetect_manual_merge"`
	DefaultDeleteBranchAfterMerge types.Bool   `tfsdk:"default_delete_branch_after_merge"`
	DefaultMergeStyle             types.String `tfsdk:"default_merge_style"`
	DefaultAllowMaintainerEdit    types.Bool   `tfsdk:"default_allow_maintainer_edit"`
}

func (d *repositoryDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repository"
}

func (d *repositoryDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Use this data source to retrieve information about an existing Gitea repository.",
		MarkdownDescription: "Use this data source to retrieve information about an existing Gitea repository.",
		Attributes: map[string]schema.Attribute{
			// Required inputs
			"owner": schema.StringAttribute{
				Required:            true,
				Description:         "The owner of the repository (username or organization name).",
				MarkdownDescription: "The owner of the repository (username or organization name).",
			},
			"name": schema.StringAttribute{
				Required:            true,
				Description:         "The name of the repository.",
				MarkdownDescription: "The name of the repository.",
			},

			// Computed outputs
			"id": schema.Int64Attribute{
				Computed:            true,
				Description:         "The numeric ID of the repository.",
				MarkdownDescription: "The numeric ID of the repository.",
			},
			"full_name": schema.StringAttribute{
				Computed:            true,
				Description:         "The full name of the repository in owner/repo format.",
				MarkdownDescription: "The full name of the repository in `owner/repo` format.",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				Description:         "The description of the repository.",
				MarkdownDescription: "The description of the repository.",
			},
			"private": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether the repository is private.",
				MarkdownDescription: "Whether the repository is private.",
			},
			"fork": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether the repository is a fork.",
				MarkdownDescription: "Whether the repository is a fork.",
			},
			"mirror": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether the repository is a mirror.",
				MarkdownDescription: "Whether the repository is a mirror.",
			},
			"template": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether the repository is a template.",
				MarkdownDescription: "Whether the repository is a template.",
			},
			"internal": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether the repository is internal.",
				MarkdownDescription: "Whether the repository is internal.",
			},
			"empty": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether the repository is empty.",
				MarkdownDescription: "Whether the repository is empty.",
			},
			"archived": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether the repository is archived.",
				MarkdownDescription: "Whether the repository is archived.",
			},
			"default_branch": schema.StringAttribute{
				Computed:            true,
				Description:         "The default branch of the repository.",
				MarkdownDescription: "The default branch of the repository.",
			},
			"website": schema.StringAttribute{
				Computed:            true,
				Description:         "The website URL associated with the repository.",
				MarkdownDescription: "The website URL associated with the repository.",
			},
			"html_url": schema.StringAttribute{
				Computed:            true,
				Description:         "The URL to view the repository in the web UI.",
				MarkdownDescription: "The URL to view the repository in the web UI.",
			},
			"clone_url": schema.StringAttribute{
				Computed:            true,
				Description:         "The HTTPS URL to clone the repository.",
				MarkdownDescription: "The HTTPS URL to clone the repository.",
			},
			"ssh_url": schema.StringAttribute{
				Computed:            true,
				Description:         "The SSH URL to clone the repository.",
				MarkdownDescription: "The SSH URL to clone the repository.",
			},
			"size": schema.Int64Attribute{
				Computed:            true,
				Description:         "The size of the repository in KB.",
				MarkdownDescription: "The size of the repository in KB.",
			},
			"stars_count": schema.Int64Attribute{
				Computed:            true,
				Description:         "The number of stars the repository has.",
				MarkdownDescription: "The number of stars the repository has.",
			},
			"watchers_count": schema.Int64Attribute{
				Computed:            true,
				Description:         "The number of watchers the repository has.",
				MarkdownDescription: "The number of watchers the repository has.",
			},
			"forks_count": schema.Int64Attribute{
				Computed:            true,
				Description:         "The number of forks of the repository.",
				MarkdownDescription: "The number of forks of the repository.",
			},
			"open_issues_count": schema.Int64Attribute{
				Computed:            true,
				Description:         "The number of open issues in the repository.",
				MarkdownDescription: "The number of open issues in the repository.",
			},
			"open_pr_counter": schema.Int64Attribute{
				Computed:            true,
				Description:         "The number of open pull requests in the repository.",
				MarkdownDescription: "The number of open pull requests in the repository.",
			},
			"release_counter": schema.Int64Attribute{
				Computed:            true,
				Description:         "The number of releases in the repository.",
				MarkdownDescription: "The number of releases in the repository.",
			},
			"avatar_url": schema.StringAttribute{
				Computed:            true,
				Description:         "The URL of the repository's avatar.",
				MarkdownDescription: "The URL of the repository's avatar.",
			},
			"object_format_name": schema.StringAttribute{
				Computed:            true,
				Description:         "The object format name of the underlying git repository (sha1 or sha256).",
				MarkdownDescription: "The object format name of the underlying git repository (sha1 or sha256).",
			},
			"has_issues": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether the repository has issues enabled.",
				MarkdownDescription: "Whether the repository has issues enabled.",
			},
			"has_wiki": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether the repository has wiki enabled.",
				MarkdownDescription: "Whether the repository has wiki enabled.",
			},
			"has_pull_requests": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether the repository has pull requests enabled.",
				MarkdownDescription: "Whether the repository has pull requests enabled.",
			},
			"has_projects": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether the repository has projects enabled.",
				MarkdownDescription: "Whether the repository has projects enabled.",
			},
			"has_releases": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether the repository has releases enabled.",
				MarkdownDescription: "Whether the repository has releases enabled.",
			},
			"has_packages": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether the repository has packages enabled.",
				MarkdownDescription: "Whether the repository has packages enabled.",
			},
			"has_actions": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether the repository has actions enabled.",
				MarkdownDescription: "Whether the repository has actions enabled.",
			},
			"ignore_whitespace_conflicts": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether the repository ignores whitespace conflicts.",
				MarkdownDescription: "Whether the repository ignores whitespace conflicts.",
			},
			"allow_merge_commits": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether merge commits are allowed.",
				MarkdownDescription: "Whether merge commits are allowed.",
			},
			"allow_rebase": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether rebase merging is allowed.",
				MarkdownDescription: "Whether rebase merging is allowed.",
			},
			"allow_rebase_explicit": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether explicit rebase merging is allowed.",
				MarkdownDescription: "Whether explicit rebase merging is allowed.",
			},
			"allow_rebase_update": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether rebase update is allowed.",
				MarkdownDescription: "Whether rebase update is allowed.",
			},
			"allow_squash_merge": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether squash merging is allowed.",
				MarkdownDescription: "Whether squash merging is allowed.",
			},
			"allow_fast_forward_only_merge": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether fast-forward only merging is allowed.",
				MarkdownDescription: "Whether fast-forward only merging is allowed.",
			},
			"allow_manual_merge": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether manual merging is allowed.",
				MarkdownDescription: "Whether manual merging is allowed.",
			},
			"autodetect_manual_merge": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether manual merge is autodetected.",
				MarkdownDescription: "Whether manual merge is autodetected.",
			},
			"default_delete_branch_after_merge": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether branches are deleted after merge by default.",
				MarkdownDescription: "Whether branches are deleted after merge by default.",
			},
			"default_merge_style": schema.StringAttribute{
				Computed:            true,
				Description:         "The default merge style for pull requests.",
				MarkdownDescription: "The default merge style for pull requests.",
			},
			"default_allow_maintainer_edit": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether maintainers can edit pull requests by default.",
				MarkdownDescription: "Whether maintainers can edit pull requests by default.",
			},
		},
	}
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
	var data repositoryDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	owner := data.Owner.ValueString()
	repoName := data.Name.ValueString()

	repo, httpResp, err := d.client.GetRepo(owner, repoName)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			resp.Diagnostics.AddError(
				"Repository Not Found",
				fmt.Sprintf("Repository %s/%s does not exist or you do not have permission to access it.", owner, repoName),
			)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Repository",
			fmt.Sprintf("Could not read repository %s/%s: %s", owner, repoName, err.Error()),
		)
		return
	}

	// Map response to model
	data.Id = types.Int64Value(repo.ID)
	data.Name = types.StringValue(repo.Name)
	data.FullName = types.StringValue(repo.FullName)
	data.Description = types.StringValue(repo.Description)
	data.Private = types.BoolValue(repo.Private)
	data.Fork = types.BoolValue(repo.Fork)
	data.Mirror = types.BoolValue(repo.Mirror)
	data.Template = types.BoolValue(repo.Template)
	data.Internal = types.BoolValue(repo.Internal)
	data.Empty = types.BoolValue(repo.Empty)
	data.Archived = types.BoolValue(repo.Archived)
	data.DefaultBranch = types.StringValue(repo.DefaultBranch)
	data.Website = types.StringValue(repo.Website)
	data.HtmlUrl = types.StringValue(repo.HTMLURL)
	data.CloneUrl = types.StringValue(repo.CloneURL)
	data.SshUrl = types.StringValue(repo.SSHURL)
	data.Size = types.Int64Value(int64(repo.Size))
	data.StarsCount = types.Int64Value(int64(repo.Stars))
	data.WatchersCount = types.Int64Value(int64(repo.Watchers))
	data.ForksCount = types.Int64Value(int64(repo.Forks))
	data.OpenIssuesCount = types.Int64Value(int64(repo.OpenIssues))
	data.OpenPrCounter = types.Int64Value(int64(repo.OpenPulls))
	data.ReleaseCounter = types.Int64Value(int64(repo.Releases))
	data.AvatarUrl = types.StringValue(repo.AvatarURL)
	data.ObjectFormatName = types.StringValue(repo.ObjectFormatName)
	data.HasIssues = types.BoolValue(repo.HasIssues)
	data.HasWiki = types.BoolValue(repo.HasWiki)
	data.HasPullRequests = types.BoolValue(repo.HasPullRequests)
	data.HasProjects = types.BoolValue(repo.HasProjects)
	data.HasReleases = types.BoolValue(repo.HasReleases)
	data.HasPackages = types.BoolValue(repo.HasPackages)
	data.HasActions = types.BoolValue(repo.HasActions)
	data.IgnoreWhitespaceConflicts = types.BoolValue(repo.IgnoreWhitespaceConflicts)
	data.AllowMergeCommits = types.BoolValue(repo.AllowMerge)
	data.AllowRebase = types.BoolValue(repo.AllowRebase)
	data.AllowRebaseExplicit = types.BoolValue(repo.AllowRebaseMerge)
	data.AllowRebaseUpdate = types.BoolValue(false)             // Field not available in SDK
	data.AllowSquashMerge = types.BoolValue(repo.AllowSquash)
	data.AllowFastForwardOnlyMerge = types.BoolValue(repo.AllowFastForwardOnlyMerge)
	data.AllowManualMerge = types.BoolValue(false)             // Field not available in SDK
	data.AutodetectManualMerge = types.BoolValue(false)        // Field not available in SDK
	data.DefaultDeleteBranchAfterMerge = types.BoolValue(repo.DefaultDeleteBranchAfterMerge)
	data.DefaultMergeStyle = types.StringValue(string(repo.DefaultMergeStyle))
	data.DefaultAllowMaintainerEdit = types.BoolValue(false)   // Field not available in SDK

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
