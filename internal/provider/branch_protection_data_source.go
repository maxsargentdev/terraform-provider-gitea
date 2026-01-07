package provider

import (
	"context"
	"fmt"
	"net/http"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &branchProtectionDataSource{}
	_ datasource.DataSourceWithConfigure = &branchProtectionDataSource{}
)

func NewBranchProtectionDataSource() datasource.DataSource {
	return &branchProtectionDataSource{}
}

type branchProtectionDataSource struct {
	client *gitea.Client
}

type branchProtectionDataSourceModel struct {
	ApprovalsWhitelistTeams       types.List   `tfsdk:"approvals_whitelist_teams"`
	ApprovalsWhitelistUsername    types.List   `tfsdk:"approvals_whitelist_username"`
	BlockAdminMergeOverride       types.Bool   `tfsdk:"block_admin_merge_override"`
	BlockOnOfficialReviewRequests types.Bool   `tfsdk:"block_on_official_review_requests"`
	BlockOnOutdatedBranch         types.Bool   `tfsdk:"block_on_outdated_branch"`
	BlockOnRejectedReviews        types.Bool   `tfsdk:"block_on_rejected_reviews"`
	BranchName                    types.String `tfsdk:"branch_name"`
	CreatedAt                     types.String `tfsdk:"created_at"`
	DismissStaleApprovals         types.Bool   `tfsdk:"dismiss_stale_approvals"`
	EnableApprovalsWhitelist      types.Bool   `tfsdk:"enable_approvals_whitelist"`
	EnableForcePush               types.Bool   `tfsdk:"enable_force_push"`
	EnableForcePushAllowlist      types.Bool   `tfsdk:"enable_force_push_allowlist"`
	EnableMergeWhitelist          types.Bool   `tfsdk:"enable_merge_whitelist"`
	EnablePush                    types.Bool   `tfsdk:"enable_push"`
	EnablePushWhitelist           types.Bool   `tfsdk:"enable_push_whitelist"`
	EnableStatusCheck             types.Bool   `tfsdk:"enable_status_check"`
	ForcePushAllowlistDeployKeys  types.Bool   `tfsdk:"force_push_allowlist_deploy_keys"`
	ForcePushAllowlistTeams       types.List   `tfsdk:"force_push_allowlist_teams"`
	ForcePushAllowlistUsernames   types.List   `tfsdk:"force_push_allowlist_usernames"`
	IgnoreStaleApprovals          types.Bool   `tfsdk:"ignore_stale_approvals"`
	MergeWhitelistTeams           types.List   `tfsdk:"merge_whitelist_teams"`
	MergeWhitelistUsernames       types.List   `tfsdk:"merge_whitelist_usernames"`
	Name                          types.String `tfsdk:"name"`
	Owner                         types.String `tfsdk:"owner"`
	Priority                      types.Int64  `tfsdk:"priority"`
	ProtectedFilePatterns         types.String `tfsdk:"protected_file_patterns"`
	PushWhitelistDeployKeys       types.Bool   `tfsdk:"push_whitelist_deploy_keys"`
	PushWhitelistTeams            types.List   `tfsdk:"push_whitelist_teams"`
	PushWhitelistUsernames        types.List   `tfsdk:"push_whitelist_usernames"`
	Repo                          types.String `tfsdk:"repo"`
	RequireSignedCommits          types.Bool   `tfsdk:"require_signed_commits"`
	RequiredApprovals             types.Int64  `tfsdk:"required_approvals"`
	RuleName                      types.String `tfsdk:"rule_name"`
	StatusCheckContexts           types.List   `tfsdk:"status_check_contexts"`
	UnprotectedFilePatterns       types.String `tfsdk:"unprotected_file_patterns"`
	UpdatedAt                     types.String `tfsdk:"updated_at"`
}

func (d *branchProtectionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_branch_protection"
}

func (d *branchProtectionDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Use this data source to retrieve information about branch protection rules for a Gitea repository.",
		MarkdownDescription: "Use this data source to retrieve information about branch protection rules for a Gitea repository.",
		Attributes: map[string]schema.Attribute{
			// Required inputs
			"owner": schema.StringAttribute{
				Required:            true,
				Description:         "The owner of the repository (username or organization name).",
				MarkdownDescription: "The owner of the repository (username or organization name).",
			},
			"repo": schema.StringAttribute{
				Required:            true,
				Description:         "The name of the repository.",
				MarkdownDescription: "The name of the repository.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				Description:         "The name or pattern of the protected branch rule to look up.",
				MarkdownDescription: "The name or pattern of the protected branch rule to look up.",
			},

			// Computed outputs
			"branch_name": schema.StringAttribute{
				Computed:            true,
				DeprecationMessage:  "Use 'name' instead. This field is deprecated and will be removed in a future version.",
				Description:         "Deprecated: Use 'name' instead. The branch name pattern for this protection rule.",
				MarkdownDescription: "**Deprecated:** Use `name` instead. The branch name pattern for this protection rule.",
			},
			"rule_name": schema.StringAttribute{
				Computed:            true,
				Description:         "The name of the branch protection rule.",
				MarkdownDescription: "The name of the branch protection rule.",
			},
			"priority": schema.Int64Attribute{
				Computed:            true,
				Description:         "The priority of this branch protection rule (higher values take precedence).",
				MarkdownDescription: "The priority of this branch protection rule (higher values take precedence).",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				Description:         "The timestamp when the branch protection rule was created.",
				MarkdownDescription: "The timestamp when the branch protection rule was created.",
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				Description:         "The timestamp when the branch protection rule was last updated.",
				MarkdownDescription: "The timestamp when the branch protection rule was last updated.",
			},

			// Push settings
			"enable_push": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether pushing to the branch is enabled.",
				MarkdownDescription: "Whether pushing to the branch is enabled.",
			},
			"enable_push_whitelist": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether push access is restricted to a whitelist of users and teams.",
				MarkdownDescription: "Whether push access is restricted to a whitelist of users and teams.",
			},
			"push_whitelist_usernames": schema.ListAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				Description:         "List of usernames allowed to push to the protected branch.",
				MarkdownDescription: "List of usernames allowed to push to the protected branch.",
			},
			"push_whitelist_teams": schema.ListAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				Description:         "List of team names allowed to push to the protected branch.",
				MarkdownDescription: "List of team names allowed to push to the protected branch.",
			},
			"push_whitelist_deploy_keys": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether deploy keys can push to the protected branch.",
				MarkdownDescription: "Whether deploy keys can push to the protected branch.",
			},

			// Force push settings
			"enable_force_push": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether force pushing is allowed.",
				MarkdownDescription: "Whether force pushing is allowed.",
			},
			"enable_force_push_allowlist": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether force push access is restricted to a whitelist.",
				MarkdownDescription: "Whether force push access is restricted to a whitelist.",
			},
			"force_push_allowlist_usernames": schema.ListAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				Description:         "List of usernames allowed to force push to the protected branch.",
				MarkdownDescription: "List of usernames allowed to force push to the protected branch.",
			},
			"force_push_allowlist_teams": schema.ListAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				Description:         "List of team names allowed to force push to the protected branch.",
				MarkdownDescription: "List of team names allowed to force push to the protected branch.",
			},
			"force_push_allowlist_deploy_keys": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether deploy keys can force push to the protected branch.",
				MarkdownDescription: "Whether deploy keys can force push to the protected branch.",
			},

			// Merge settings
			"enable_merge_whitelist": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether merge access is restricted to a whitelist of users and teams.",
				MarkdownDescription: "Whether merge access is restricted to a whitelist of users and teams.",
			},
			"merge_whitelist_usernames": schema.ListAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				Description:         "List of usernames allowed to merge pull requests.",
				MarkdownDescription: "List of usernames allowed to merge pull requests.",
			},
			"merge_whitelist_teams": schema.ListAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				Description:         "List of team names allowed to merge pull requests.",
				MarkdownDescription: "List of team names allowed to merge pull requests.",
			},

			// Review/approval settings
			"enable_approvals_whitelist": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether only whitelisted users/teams count toward required approvals.",
				MarkdownDescription: "Whether only whitelisted users/teams count toward required approvals.",
			},
			"approvals_whitelist_username": schema.ListAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				Description:         "List of usernames whose approvals count toward the required approval count.",
				MarkdownDescription: "List of usernames whose approvals count toward the required approval count.",
			},
			"approvals_whitelist_teams": schema.ListAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				Description:         "List of team names whose members' approvals count toward the required approval count.",
				MarkdownDescription: "List of team names whose members' approvals count toward the required approval count.",
			},
			"required_approvals": schema.Int64Attribute{
				Computed:            true,
				Description:         "The number of approvals required before a pull request can be merged.",
				MarkdownDescription: "The number of approvals required before a pull request can be merged.",
			},
			"dismiss_stale_approvals": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether to dismiss existing approvals when new commits are pushed.",
				MarkdownDescription: "Whether to dismiss existing approvals when new commits are pushed.",
			},
			"ignore_stale_approvals": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether to ignore stale approvals instead of dismissing them.",
				MarkdownDescription: "Whether to ignore stale approvals instead of dismissing them.",
			},
			"block_on_rejected_reviews": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether to block merging if there are rejected reviews.",
				MarkdownDescription: "Whether to block merging if there are rejected reviews.",
			},
			"block_on_official_review_requests": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether to block merging if there are pending official review requests.",
				MarkdownDescription: "Whether to block merging if there are pending official review requests.",
			},
			"block_on_outdated_branch": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether to block merging if the branch is not up to date with the base branch.",
				MarkdownDescription: "Whether to block merging if the branch is not up to date with the base branch.",
			},
			"block_admin_merge_override": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether to prevent administrators from bypassing branch protection.",
				MarkdownDescription: "Whether to prevent administrators from bypassing branch protection.",
			},

			// Status check settings
			"enable_status_check": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether status checks must pass before merging.",
				MarkdownDescription: "Whether status checks must pass before merging.",
			},
			"status_check_contexts": schema.ListAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				Description:         "List of status check contexts that must pass before merging.",
				MarkdownDescription: "List of status check contexts that must pass before merging.",
			},

			// Other settings
			"require_signed_commits": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether commits must be signed.",
				MarkdownDescription: "Whether commits must be signed.",
			},
			"protected_file_patterns": schema.StringAttribute{
				Computed:            true,
				Description:         "Glob patterns of files that are protected from changes.",
				MarkdownDescription: "Glob patterns of files that are protected from changes.",
			},
			"unprotected_file_patterns": schema.StringAttribute{
				Computed:            true,
				Description:         "Glob patterns of files that are exempt from protection.",
				MarkdownDescription: "Glob patterns of files that are exempt from protection.",
			},
		},
	}
}

func (d *branchProtectionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *branchProtectionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data branchProtectionDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	owner := data.Owner.ValueString()
	repo := data.Repo.ValueString()
	branchName := data.Name.ValueString()

	protection, httpResp, err := d.client.GetBranchProtection(owner, repo, branchName)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			resp.Diagnostics.AddError(
				"Branch Protection Not Found",
				fmt.Sprintf("Branch protection rule '%s' does not exist for repository %s/%s or you do not have permission to access it.", branchName, owner, repo),
			)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Branch Protection",
			fmt.Sprintf("Could not read branch protection '%s' for repository %s/%s: %s", branchName, owner, repo, err.Error()),
		)
		return
	}

	mapBranchProtectionToDataSourceModel(ctx, protection, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Helper function to map Gitea BranchProtection to data source model
func mapBranchProtectionToDataSourceModel(ctx context.Context, protection *gitea.BranchProtection, model *branchProtectionDataSourceModel) {
	// Note: owner, repo, and name need to be preserved from config (not overwritten from API)
	model.RuleName = types.StringValue(protection.RuleName)
	model.BranchName = types.StringValue(protection.BranchName)
	model.EnablePush = types.BoolValue(protection.EnablePush)
	model.EnablePushWhitelist = types.BoolValue(protection.EnablePushWhitelist)
	model.EnableMergeWhitelist = types.BoolValue(protection.EnableMergeWhitelist)
	model.EnableStatusCheck = types.BoolValue(protection.EnableStatusCheck)
	model.RequireSignedCommits = types.BoolValue(protection.RequireSignedCommits)
	model.CreatedAt = types.StringValue(protection.Created.String())
	model.UpdatedAt = types.StringValue(protection.Updated.String())

	// Map boolean fields from API
	model.PushWhitelistDeployKeys = types.BoolValue(protection.PushWhitelistDeployKeys)
	model.BlockOnRejectedReviews = types.BoolValue(protection.BlockOnRejectedReviews)
	model.BlockOnOfficialReviewRequests = types.BoolValue(protection.BlockOnOfficialReviewRequests)
	model.BlockOnOutdatedBranch = types.BoolValue(protection.BlockOnOutdatedBranch)
	model.DismissStaleApprovals = types.BoolValue(protection.DismissStaleApprovals)
	model.EnableApprovalsWhitelist = types.BoolValue(protection.EnableApprovalsWhitelist)

	// Map integer fields
	model.RequiredApprovals = types.Int64Value(protection.RequiredApprovals)

	// Map string fields
	model.ProtectedFilePatterns = types.StringValue(protection.ProtectedFilePatterns)
	model.UnprotectedFilePatterns = types.StringValue(protection.UnprotectedFilePatterns)

	// Map list fields
	if len(protection.PushWhitelistUsernames) > 0 {
		elements := make([]attr.Value, len(protection.PushWhitelistUsernames))
		for i, v := range protection.PushWhitelistUsernames {
			elements[i] = types.StringValue(v)
		}
		model.PushWhitelistUsernames, _ = types.ListValue(types.StringType, elements)
	} else {
		model.PushWhitelistUsernames = types.ListNull(types.StringType)
	}

	if len(protection.PushWhitelistTeams) > 0 {
		elements := make([]attr.Value, len(protection.PushWhitelistTeams))
		for i, v := range protection.PushWhitelistTeams {
			elements[i] = types.StringValue(v)
		}
		model.PushWhitelistTeams, _ = types.ListValue(types.StringType, elements)
	} else {
		model.PushWhitelistTeams = types.ListNull(types.StringType)
	}

	if len(protection.MergeWhitelistUsernames) > 0 {
		elements := make([]attr.Value, len(protection.MergeWhitelistUsernames))
		for i, v := range protection.MergeWhitelistUsernames {
			elements[i] = types.StringValue(v)
		}
		model.MergeWhitelistUsernames, _ = types.ListValue(types.StringType, elements)
	} else {
		model.MergeWhitelistUsernames = types.ListNull(types.StringType)
	}

	if len(protection.MergeWhitelistTeams) > 0 {
		elements := make([]attr.Value, len(protection.MergeWhitelistTeams))
		for i, v := range protection.MergeWhitelistTeams {
			elements[i] = types.StringValue(v)
		}
		model.MergeWhitelistTeams, _ = types.ListValue(types.StringType, elements)
	} else {
		model.MergeWhitelistTeams = types.ListNull(types.StringType)
	}

	if len(protection.ApprovalsWhitelistUsernames) > 0 {
		elements := make([]attr.Value, len(protection.ApprovalsWhitelistUsernames))
		for i, v := range protection.ApprovalsWhitelistUsernames {
			elements[i] = types.StringValue(v)
		}
		model.ApprovalsWhitelistUsername, _ = types.ListValue(types.StringType, elements)
	} else {
		model.ApprovalsWhitelistUsername = types.ListNull(types.StringType)
	}

	if len(protection.ApprovalsWhitelistTeams) > 0 {
		elements := make([]attr.Value, len(protection.ApprovalsWhitelistTeams))
		for i, v := range protection.ApprovalsWhitelistTeams {
			elements[i] = types.StringValue(v)
		}
		model.ApprovalsWhitelistTeams, _ = types.ListValue(types.StringType, elements)
	} else {
		model.ApprovalsWhitelistTeams = types.ListNull(types.StringType)
	}

	if len(protection.StatusCheckContexts) > 0 {
		elements := make([]attr.Value, len(protection.StatusCheckContexts))
		for i, v := range protection.StatusCheckContexts {
			elements[i] = types.StringValue(v)
		}
		model.StatusCheckContexts, _ = types.ListValue(types.StringType, elements)
	} else {
		model.StatusCheckContexts = types.ListNull(types.StringType)
	}

	// Set remaining fields as null (not returned by API)
	model.ForcePushAllowlistUsernames = types.ListNull(types.StringType)
	model.ForcePushAllowlistTeams = types.ListNull(types.StringType)
	model.BlockAdminMergeOverride = types.BoolNull()
	model.EnableForcePush = types.BoolNull()
	model.EnableForcePushAllowlist = types.BoolNull()
	model.ForcePushAllowlistDeployKeys = types.BoolNull()
	model.IgnoreStaleApprovals = types.BoolNull()
	model.Priority = types.Int64Null()
}
