package provider

import (
	"context"
	"fmt"

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

func (d *branchProtectionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_branch_protection"
}

func (d *branchProtectionDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = BranchProtectionDataSourceSchema(ctx)
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
	var data BranchProtectionDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	protection, _, err := d.client.GetBranchProtection(
		data.Owner.ValueString(),
		data.Repo.ValueString(),
		data.BranchName.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Branch Protection",
			"Could not read branch protection: "+err.Error(),
		)
		return
	}

	mapBranchProtectionToDataSourceModel(ctx, protection, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Helper function to map Gitea BranchProtection to data source model
func mapBranchProtectionToDataSourceModel(ctx context.Context, protection *gitea.BranchProtection, model *BranchProtectionDataSourceModel) {
	// Note: owner, repo, and branch_name need to be preserved from config (not overwritten from API)
	model.RuleName = types.StringValue(protection.RuleName)
	model.Name = types.StringValue(protection.BranchName)
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

func BranchProtectionDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"approvals_whitelist_teams": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"approvals_whitelist_username": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"block_admin_merge_override": schema.BoolAttribute{
				Computed: true,
			},
			"block_on_official_review_requests": schema.BoolAttribute{
				Computed: true,
			},
			"block_on_outdated_branch": schema.BoolAttribute{
				Computed: true,
			},
			"block_on_rejected_reviews": schema.BoolAttribute{
				Computed: true,
			},
			"branch_name": schema.StringAttribute{
				Computed:            true,
				Description:         "Deprecated: true",
				MarkdownDescription: "Deprecated: true",
			},
			"created_at": schema.StringAttribute{
				Computed: true,
			},
			"dismiss_stale_approvals": schema.BoolAttribute{
				Computed: true,
			},
			"enable_approvals_whitelist": schema.BoolAttribute{
				Computed: true,
			},
			"enable_force_push": schema.BoolAttribute{
				Computed: true,
			},
			"enable_force_push_allowlist": schema.BoolAttribute{
				Computed: true,
			},
			"enable_merge_whitelist": schema.BoolAttribute{
				Computed: true,
			},
			"enable_push": schema.BoolAttribute{
				Computed: true,
			},
			"enable_push_whitelist": schema.BoolAttribute{
				Computed: true,
			},
			"enable_status_check": schema.BoolAttribute{
				Computed: true,
			},
			"force_push_allowlist_deploy_keys": schema.BoolAttribute{
				Computed: true,
			},
			"force_push_allowlist_teams": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"force_push_allowlist_usernames": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"ignore_stale_approvals": schema.BoolAttribute{
				Computed: true,
			},
			"merge_whitelist_teams": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"merge_whitelist_usernames": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Required:            true,
				Description:         "name of protected branch",
				MarkdownDescription: "name of protected branch",
			},
			"owner": schema.StringAttribute{
				Required:            true,
				Description:         "owner of the repo",
				MarkdownDescription: "owner of the repo",
			},
			"priority": schema.Int64Attribute{
				Computed:            true,
				Description:         "Priority is the priority of this branch protection rule",
				MarkdownDescription: "Priority is the priority of this branch protection rule",
			},
			"protected_file_patterns": schema.StringAttribute{
				Computed: true,
			},
			"push_whitelist_deploy_keys": schema.BoolAttribute{
				Computed: true,
			},
			"push_whitelist_teams": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"push_whitelist_usernames": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"repo": schema.StringAttribute{
				Required:            true,
				Description:         "name of the repo",
				MarkdownDescription: "name of the repo",
			},
			"require_signed_commits": schema.BoolAttribute{
				Computed: true,
			},
			"required_approvals": schema.Int64Attribute{
				Computed: true,
			},
			"rule_name": schema.StringAttribute{
				Computed:            true,
				Description:         "RuleName is the name of the branch protection rule",
				MarkdownDescription: "RuleName is the name of the branch protection rule",
			},
			"status_check_contexts": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"unprotected_file_patterns": schema.StringAttribute{
				Computed: true,
			},
			"updated_at": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

type BranchProtectionDataSourceModel struct {
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
