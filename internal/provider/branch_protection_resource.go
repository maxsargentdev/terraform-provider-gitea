package provider

import (
	"context"
	"fmt"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &branchProtectionResource{}
	_ resource.ResourceWithConfigure   = &branchProtectionResource{}
	_ resource.ResourceWithImportState = &branchProtectionResource{}
)

func NewBranchProtectionResource() resource.Resource {
	return &branchProtectionResource{}
}

type branchProtectionResource struct {
	client *gitea.Client
}

// branchProtectionResourceModel describes the resource data model.
type branchProtectionResourceModel struct {
	// Required - identification fields
	Owner types.String `tfsdk:"owner"`
	Repo  types.String `tfsdk:"repo"`

	// Required - the rule name (primary identifier for API)
	RuleName types.String `tfsdk:"rule_name"`

	// Deprecated - use rule_name instead
	BranchName types.String `tfsdk:"branch_name"`

	// Push settings
	EnablePush             types.Bool `tfsdk:"enable_push"`
	EnablePushWhitelist    types.Bool `tfsdk:"enable_push_whitelist"`
	PushWhitelistUsernames types.List `tfsdk:"push_whitelist_usernames"`
	PushWhitelistTeams     types.List `tfsdk:"push_whitelist_teams"`
	PushWhitelistDeployKeys types.Bool `tfsdk:"push_whitelist_deploy_keys"`

	// Merge settings
	EnableMergeWhitelist    types.Bool `tfsdk:"enable_merge_whitelist"`
	MergeWhitelistUsernames types.List `tfsdk:"merge_whitelist_usernames"`
	MergeWhitelistTeams     types.List `tfsdk:"merge_whitelist_teams"`

	// Status check settings
	EnableStatusCheck   types.Bool `tfsdk:"enable_status_check"`
	StatusCheckContexts types.List `tfsdk:"status_check_contexts"`

	// Approval settings
	RequiredApprovals          types.Int64 `tfsdk:"required_approvals"`
	EnableApprovalsWhitelist   types.Bool  `tfsdk:"enable_approvals_whitelist"`
	ApprovalsWhitelistUsernames types.List  `tfsdk:"approvals_whitelist_username"`
	ApprovalsWhitelistTeams    types.List  `tfsdk:"approvals_whitelist_teams"`

	// Review settings
	BlockOnRejectedReviews        types.Bool `tfsdk:"block_on_rejected_reviews"`
	BlockOnOfficialReviewRequests types.Bool `tfsdk:"block_on_official_review_requests"`
	BlockOnOutdatedBranch         types.Bool `tfsdk:"block_on_outdated_branch"`
	DismissStaleApprovals         types.Bool `tfsdk:"dismiss_stale_approvals"`

	// Other settings
	RequireSignedCommits    types.Bool   `tfsdk:"require_signed_commits"`
	ProtectedFilePatterns   types.String `tfsdk:"protected_file_patterns"`
	UnprotectedFilePatterns types.String `tfsdk:"unprotected_file_patterns"`

	// Computed - read-only metadata
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func (r *branchProtectionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_branch_protection"
}

func (r *branchProtectionResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manages branch protection rules for a Gitea repository.",
		MarkdownDescription: "Manages branch protection rules for a Gitea repository. Branch protection rules help enforce certain workflows for branches, such as requiring code reviews or passing status checks before merging.",
		Attributes: map[string]schema.Attribute{
			// Required identification fields
			"owner": schema.StringAttribute{
				Required:            true,
				Description:         "The owner of the repository (username or organization name).",
				MarkdownDescription: "The owner of the repository (username or organization name).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"repo": schema.StringAttribute{
				Required:            true,
				Description:         "The name of the repository.",
				MarkdownDescription: "The name of the repository.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"rule_name": schema.StringAttribute{
				Required:            true,
				Description:         "The name of the branch protection rule. This is used as the unique identifier and can be a glob pattern (e.g., 'main', 'release/*').",
				MarkdownDescription: "The name of the branch protection rule. This is used as the unique identifier and can be a glob pattern (e.g., `main`, `release/*`).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			// Deprecated field
			"branch_name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Deprecated: Use rule_name instead. The branch name pattern for this protection rule.",
				MarkdownDescription: "**Deprecated:** Use `rule_name` instead. The branch name pattern for this protection rule.",
				DeprecationMessage:  "The branch_name attribute is deprecated. Use rule_name instead.",
				Default:             stringdefault.StaticString(""),
			},

			// Push settings
			"enable_push": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				Description:         "Whether pushing to the branch is enabled.",
				MarkdownDescription: "Whether pushing to the branch is enabled.",
			},
			"enable_push_whitelist": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				Description:         "Whether to restrict push access to specific users, teams, or deploy keys.",
				MarkdownDescription: "Whether to restrict push access to specific users, teams, or deploy keys.",
			},
			"push_whitelist_usernames": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "List of usernames allowed to push when push whitelist is enabled.",
				MarkdownDescription: "List of usernames allowed to push when push whitelist is enabled.",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"push_whitelist_teams": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "List of team names allowed to push when push whitelist is enabled.",
				MarkdownDescription: "List of team names allowed to push when push whitelist is enabled.",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"push_whitelist_deploy_keys": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				Description:         "Whether deploy keys are allowed to push when push whitelist is enabled.",
				MarkdownDescription: "Whether deploy keys are allowed to push when push whitelist is enabled.",
			},

			// Merge settings
			"enable_merge_whitelist": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				Description:         "Whether to restrict merge access to specific users or teams.",
				MarkdownDescription: "Whether to restrict merge access to specific users or teams.",
			},
			"merge_whitelist_usernames": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "List of usernames allowed to merge when merge whitelist is enabled.",
				MarkdownDescription: "List of usernames allowed to merge when merge whitelist is enabled.",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"merge_whitelist_teams": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "List of team names allowed to merge when merge whitelist is enabled.",
				MarkdownDescription: "List of team names allowed to merge when merge whitelist is enabled.",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},

			// Status check settings
			"enable_status_check": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				Description:         "Whether to require status checks to pass before merging.",
				MarkdownDescription: "Whether to require status checks to pass before merging.",
			},
			"status_check_contexts": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "List of status check context names that must pass before merging.",
				MarkdownDescription: "List of status check context names that must pass before merging.",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},

			// Approval settings
			"required_approvals": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
				Description:         "The number of approvals required before a pull request can be merged. Set to 0 to disable.",
				MarkdownDescription: "The number of approvals required before a pull request can be merged. Set to `0` to disable.",
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"enable_approvals_whitelist": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				Description:         "Whether to restrict approval access to specific users or teams.",
				MarkdownDescription: "Whether to restrict approval access to specific users or teams.",
			},
			"approvals_whitelist_username": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "List of usernames whose approvals count when approval whitelist is enabled.",
				MarkdownDescription: "List of usernames whose approvals count when approval whitelist is enabled.",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"approvals_whitelist_teams": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "List of team names whose members' approvals count when approval whitelist is enabled.",
				MarkdownDescription: "List of team names whose members' approvals count when approval whitelist is enabled.",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},

			// Review settings
			"block_on_rejected_reviews": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				Description:         "Whether to block merging when there are rejected reviews.",
				MarkdownDescription: "Whether to block merging when there are rejected reviews.",
			},
			"block_on_official_review_requests": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				Description:         "Whether to block merging when there are pending official review requests.",
				MarkdownDescription: "Whether to block merging when there are pending official review requests.",
			},
			"block_on_outdated_branch": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				Description:         "Whether to block merging when the branch is behind the base branch.",
				MarkdownDescription: "Whether to block merging when the branch is behind the base branch.",
			},
			"dismiss_stale_approvals": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				Description:         "Whether to dismiss existing approvals when new commits are pushed.",
				MarkdownDescription: "Whether to dismiss existing approvals when new commits are pushed.",
			},

			// Other settings
			"require_signed_commits": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				Description:         "Whether commits must be signed with a GPG key.",
				MarkdownDescription: "Whether commits must be signed with a GPG key.",
			},
			"protected_file_patterns": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				Description:         "Glob patterns of files that cannot be changed. Multiple patterns can be separated by semicolons.",
				MarkdownDescription: "Glob patterns of files that cannot be changed. Multiple patterns can be separated by semicolons.",
			},
			"unprotected_file_patterns": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				Description:         "Glob patterns of files that can be changed even if other rules would prevent it. Multiple patterns can be separated by semicolons.",
				MarkdownDescription: "Glob patterns of files that can be changed even if other rules would prevent it. Multiple patterns can be separated by semicolons.",
			},

			// Computed metadata
			"created_at": schema.StringAttribute{
				Computed:            true,
				Description:         "The timestamp when this branch protection rule was created.",
				MarkdownDescription: "The timestamp when this branch protection rule was created.",
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				Description:         "The timestamp when this branch protection rule was last updated.",
				MarkdownDescription: "The timestamp when this branch protection rule was last updated.",
			},
		},
	}
}

func (r *branchProtectionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*gitea.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *gitea.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *branchProtectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan branchProtectionResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build create options with all fields
	createOpts := gitea.CreateBranchProtectionOption{
		RuleName:                      plan.RuleName.ValueString(),
		EnablePush:                    plan.EnablePush.ValueBool(),
		EnablePushWhitelist:           plan.EnablePushWhitelist.ValueBool(),
		PushWhitelistDeployKeys:       plan.PushWhitelistDeployKeys.ValueBool(),
		EnableMergeWhitelist:          plan.EnableMergeWhitelist.ValueBool(),
		EnableStatusCheck:             plan.EnableStatusCheck.ValueBool(),
		RequiredApprovals:             plan.RequiredApprovals.ValueInt64(),
		EnableApprovalsWhitelist:      plan.EnableApprovalsWhitelist.ValueBool(),
		BlockOnRejectedReviews:        plan.BlockOnRejectedReviews.ValueBool(),
		BlockOnOfficialReviewRequests: plan.BlockOnOfficialReviewRequests.ValueBool(),
		BlockOnOutdatedBranch:         plan.BlockOnOutdatedBranch.ValueBool(),
		DismissStaleApprovals:         plan.DismissStaleApprovals.ValueBool(),
		RequireSignedCommits:          plan.RequireSignedCommits.ValueBool(),
		ProtectedFilePatterns:         plan.ProtectedFilePatterns.ValueString(),
		UnprotectedFilePatterns:       plan.UnprotectedFilePatterns.ValueString(),
	}

	// Handle deprecated branch_name field - use it if provided and rule_name is not set
	if !plan.BranchName.IsNull() && plan.BranchName.ValueString() != "" {
		createOpts.BranchName = plan.BranchName.ValueString()
	}

	// Set list fields
	if !plan.PushWhitelistUsernames.IsNull() && !plan.PushWhitelistUsernames.IsUnknown() {
		var usernames []string
		resp.Diagnostics.Append(plan.PushWhitelistUsernames.ElementsAs(ctx, &usernames, false)...)
		createOpts.PushWhitelistUsernames = usernames
	}
	if !plan.PushWhitelistTeams.IsNull() && !plan.PushWhitelistTeams.IsUnknown() {
		var teams []string
		resp.Diagnostics.Append(plan.PushWhitelistTeams.ElementsAs(ctx, &teams, false)...)
		createOpts.PushWhitelistTeams = teams
	}
	if !plan.MergeWhitelistUsernames.IsNull() && !plan.MergeWhitelistUsernames.IsUnknown() {
		var usernames []string
		resp.Diagnostics.Append(plan.MergeWhitelistUsernames.ElementsAs(ctx, &usernames, false)...)
		createOpts.MergeWhitelistUsernames = usernames
	}
	if !plan.MergeWhitelistTeams.IsNull() && !plan.MergeWhitelistTeams.IsUnknown() {
		var teams []string
		resp.Diagnostics.Append(plan.MergeWhitelistTeams.ElementsAs(ctx, &teams, false)...)
		createOpts.MergeWhitelistTeams = teams
	}
	if !plan.StatusCheckContexts.IsNull() && !plan.StatusCheckContexts.IsUnknown() {
		var contexts []string
		resp.Diagnostics.Append(plan.StatusCheckContexts.ElementsAs(ctx, &contexts, false)...)
		createOpts.StatusCheckContexts = contexts
	}
	if !plan.ApprovalsWhitelistUsernames.IsNull() && !plan.ApprovalsWhitelistUsernames.IsUnknown() {
		var usernames []string
		resp.Diagnostics.Append(plan.ApprovalsWhitelistUsernames.ElementsAs(ctx, &usernames, false)...)
		createOpts.ApprovalsWhitelistUsernames = usernames
	}
	if !plan.ApprovalsWhitelistTeams.IsNull() && !plan.ApprovalsWhitelistTeams.IsUnknown() {
		var teams []string
		resp.Diagnostics.Append(plan.ApprovalsWhitelistTeams.ElementsAs(ctx, &teams, false)...)
		createOpts.ApprovalsWhitelistTeams = teams
	}

	if resp.Diagnostics.HasError() {
		return
	}

	protection, _, err := r.client.CreateBranchProtection(
		plan.Owner.ValueString(),
		plan.Repo.ValueString(),
		createOpts,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Branch Protection",
			fmt.Sprintf("Could not create branch protection rule '%s' for %s/%s: %s",
				plan.RuleName.ValueString(), plan.Owner.ValueString(), plan.Repo.ValueString(), err.Error()),
		)
		return
	}

	// Map response to state
	mapBranchProtectionToModel(ctx, protection, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *branchProtectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state branchProtectionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	owner := state.Owner.ValueString()
	repo := state.Repo.ValueString()
	ruleName := state.RuleName.ValueString()

	// API uses rule_name as the identifier
	protection, httpResp, err := r.client.GetBranchProtection(owner, repo, ruleName)
	if err != nil {
		// Handle 404 - resource was deleted outside of Terraform
		if httpResp != nil && httpResp.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Branch Protection",
			fmt.Sprintf("Could not read branch protection rule '%s' for %s/%s: %s",
				ruleName, owner, repo, err.Error()),
		)
		return
	}

	// Preserve owner and repo from state since they're not in the API response
	state.Owner = types.StringValue(owner)
	state.Repo = types.StringValue(repo)

	// Map response to state
	mapBranchProtectionToModel(ctx, protection, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *branchProtectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan branchProtectionResourceModel
	var state branchProtectionResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build edit options with all fields
	editOpts := gitea.EditBranchProtectionOption{
		EnablePush:                    plan.EnablePush.ValueBoolPointer(),
		EnablePushWhitelist:           plan.EnablePushWhitelist.ValueBoolPointer(),
		PushWhitelistDeployKeys:       plan.PushWhitelistDeployKeys.ValueBoolPointer(),
		EnableMergeWhitelist:          plan.EnableMergeWhitelist.ValueBoolPointer(),
		EnableStatusCheck:             plan.EnableStatusCheck.ValueBoolPointer(),
		EnableApprovalsWhitelist:      plan.EnableApprovalsWhitelist.ValueBoolPointer(),
		BlockOnRejectedReviews:        plan.BlockOnRejectedReviews.ValueBoolPointer(),
		BlockOnOfficialReviewRequests: plan.BlockOnOfficialReviewRequests.ValueBoolPointer(),
		BlockOnOutdatedBranch:         plan.BlockOnOutdatedBranch.ValueBoolPointer(),
		DismissStaleApprovals:         plan.DismissStaleApprovals.ValueBoolPointer(),
		RequireSignedCommits:          plan.RequireSignedCommits.ValueBoolPointer(),
	}

	// Set integer fields
	if !plan.RequiredApprovals.IsNull() {
		val := plan.RequiredApprovals.ValueInt64()
		editOpts.RequiredApprovals = &val
	}

	// Set string fields
	if !plan.ProtectedFilePatterns.IsNull() {
		val := plan.ProtectedFilePatterns.ValueString()
		editOpts.ProtectedFilePatterns = &val
	}
	if !plan.UnprotectedFilePatterns.IsNull() {
		val := plan.UnprotectedFilePatterns.ValueString()
		editOpts.UnprotectedFilePatterns = &val
	}

	// Set list fields
	if !plan.PushWhitelistUsernames.IsNull() && !plan.PushWhitelistUsernames.IsUnknown() {
		var usernames []string
		resp.Diagnostics.Append(plan.PushWhitelistUsernames.ElementsAs(ctx, &usernames, false)...)
		editOpts.PushWhitelistUsernames = usernames
	}
	if !plan.PushWhitelistTeams.IsNull() && !plan.PushWhitelistTeams.IsUnknown() {
		var teams []string
		resp.Diagnostics.Append(plan.PushWhitelistTeams.ElementsAs(ctx, &teams, false)...)
		editOpts.PushWhitelistTeams = teams
	}
	if !plan.MergeWhitelistUsernames.IsNull() && !plan.MergeWhitelistUsernames.IsUnknown() {
		var usernames []string
		resp.Diagnostics.Append(plan.MergeWhitelistUsernames.ElementsAs(ctx, &usernames, false)...)
		editOpts.MergeWhitelistUsernames = usernames
	}
	if !plan.MergeWhitelistTeams.IsNull() && !plan.MergeWhitelistTeams.IsUnknown() {
		var teams []string
		resp.Diagnostics.Append(plan.MergeWhitelistTeams.ElementsAs(ctx, &teams, false)...)
		editOpts.MergeWhitelistTeams = teams
	}
	if !plan.StatusCheckContexts.IsNull() && !plan.StatusCheckContexts.IsUnknown() {
		var contexts []string
		resp.Diagnostics.Append(plan.StatusCheckContexts.ElementsAs(ctx, &contexts, false)...)
		editOpts.StatusCheckContexts = contexts
	}
	if !plan.ApprovalsWhitelistUsernames.IsNull() && !plan.ApprovalsWhitelistUsernames.IsUnknown() {
		var usernames []string
		resp.Diagnostics.Append(plan.ApprovalsWhitelistUsernames.ElementsAs(ctx, &usernames, false)...)
		editOpts.ApprovalsWhitelistUsernames = usernames
	}
	if !plan.ApprovalsWhitelistTeams.IsNull() && !plan.ApprovalsWhitelistTeams.IsUnknown() {
		var teams []string
		resp.Diagnostics.Append(plan.ApprovalsWhitelistTeams.ElementsAs(ctx, &teams, false)...)
		editOpts.ApprovalsWhitelistTeams = teams
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// API uses rule_name as the identifier - use current name from state
	protection, _, err := r.client.EditBranchProtection(
		state.Owner.ValueString(),
		state.Repo.ValueString(),
		state.RuleName.ValueString(),
		editOpts,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Branch Protection",
			fmt.Sprintf("Could not update branch protection rule '%s' for %s/%s: %s",
				state.RuleName.ValueString(), state.Owner.ValueString(), state.Repo.ValueString(), err.Error()),
		)
		return
	}

	// Preserve owner and repo from plan
	plan.Owner = state.Owner
	plan.Repo = state.Repo

	// Map response to state
	mapBranchProtectionToModel(ctx, protection, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *branchProtectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state branchProtectionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	owner := state.Owner.ValueString()
	repo := state.Repo.ValueString()
	ruleName := state.RuleName.ValueString()

	// API uses rule_name as the identifier
	httpResp, err := r.client.DeleteBranchProtection(owner, repo, ruleName)
	if err != nil {
		// If already deleted (404), consider it a success
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting Branch Protection",
			fmt.Sprintf("Could not delete branch protection rule '%s' for %s/%s: %s",
				ruleName, owner, repo, err.Error()),
		)
		return
	}
}

func (r *branchProtectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: "owner/repo/rule_name"
	id := req.ID

	parts := strings.SplitN(id, "/", 3)
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Import ID must be in format 'owner/repo/rule_name', got: %s", id),
		)
		return
	}

	owner := parts[0]
	repo := parts[1]
	ruleName := parts[2]

	// Fetch the branch protection
	protection, httpResp, err := r.client.GetBranchProtection(owner, repo, ruleName)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			resp.Diagnostics.AddError(
				"Branch Protection Not Found",
				fmt.Sprintf("Branch protection rule '%s' not found for %s/%s", ruleName, owner, repo),
			)
			return
		}
		resp.Diagnostics.AddError(
			"Error Importing Branch Protection",
			fmt.Sprintf("Could not import branch protection rule '%s' for %s/%s: %s",
				ruleName, owner, repo, err.Error()),
		)
		return
	}

	var data branchProtectionResourceModel
	data.Owner = types.StringValue(owner)
	data.Repo = types.StringValue(repo)
	mapBranchProtectionToModel(ctx, protection, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *branchProtectionResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config branchProtectionResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that whitelist settings have corresponding lists when enabled
	if !config.EnablePushWhitelist.IsNull() && config.EnablePushWhitelist.ValueBool() {
		pushUsersEmpty := config.PushWhitelistUsernames.IsNull() || len(config.PushWhitelistUsernames.Elements()) == 0
		pushTeamsEmpty := config.PushWhitelistTeams.IsNull() || len(config.PushWhitelistTeams.Elements()) == 0
		pushDeployKeysDisabled := config.PushWhitelistDeployKeys.IsNull() || !config.PushWhitelistDeployKeys.ValueBool()

		if pushUsersEmpty && pushTeamsEmpty && pushDeployKeysDisabled {
			resp.Diagnostics.AddAttributeWarning(
				path.Root("enable_push_whitelist"),
				"Push Whitelist Enabled Without Entries",
				"enable_push_whitelist is enabled but no push_whitelist_usernames, push_whitelist_teams, or push_whitelist_deploy_keys are configured. No one will be able to push.",
			)
		}
	}

	if !config.EnableMergeWhitelist.IsNull() && config.EnableMergeWhitelist.ValueBool() {
		mergeUsersEmpty := config.MergeWhitelistUsernames.IsNull() || len(config.MergeWhitelistUsernames.Elements()) == 0
		mergeTeamsEmpty := config.MergeWhitelistTeams.IsNull() || len(config.MergeWhitelistTeams.Elements()) == 0

		if mergeUsersEmpty && mergeTeamsEmpty {
			resp.Diagnostics.AddAttributeWarning(
				path.Root("enable_merge_whitelist"),
				"Merge Whitelist Enabled Without Entries",
				"enable_merge_whitelist is enabled but no merge_whitelist_usernames or merge_whitelist_teams are configured. No one will be able to merge.",
			)
		}
	}

	if !config.EnableApprovalsWhitelist.IsNull() && config.EnableApprovalsWhitelist.ValueBool() {
		approvalUsersEmpty := config.ApprovalsWhitelistUsernames.IsNull() || len(config.ApprovalsWhitelistUsernames.Elements()) == 0
		approvalTeamsEmpty := config.ApprovalsWhitelistTeams.IsNull() || len(config.ApprovalsWhitelistTeams.Elements()) == 0

		if approvalUsersEmpty && approvalTeamsEmpty {
			resp.Diagnostics.AddAttributeWarning(
				path.Root("enable_approvals_whitelist"),
				"Approvals Whitelist Enabled Without Entries",
				"enable_approvals_whitelist is enabled but no approvals_whitelist_username or approvals_whitelist_teams are configured. No approvals will count.",
			)
		}
	}

	if !config.EnableStatusCheck.IsNull() && config.EnableStatusCheck.ValueBool() {
		contextsEmpty := config.StatusCheckContexts.IsNull() || len(config.StatusCheckContexts.Elements()) == 0

		if contextsEmpty {
			resp.Diagnostics.AddAttributeWarning(
				path.Root("enable_status_check"),
				"Status Check Enabled Without Contexts",
				"enable_status_check is enabled but no status_check_contexts are configured. No status checks will be required.",
			)
		}
	}
}

// Ensure resource implements ValidateConfig interface
var _ resource.ResourceWithValidateConfig = &branchProtectionResource{}

// Helper function to map Gitea BranchProtection to Terraform model
func mapBranchProtectionToModel(ctx context.Context, protection *gitea.BranchProtection, model *branchProtectionResourceModel) {
	// Map the rule name - this is the primary identifier
	model.RuleName = types.StringValue(protection.RuleName)

	// Map deprecated branch_name field
	model.BranchName = types.StringValue(protection.BranchName)

	// Map boolean fields
	model.EnablePush = types.BoolValue(protection.EnablePush)
	model.EnablePushWhitelist = types.BoolValue(protection.EnablePushWhitelist)
	model.PushWhitelistDeployKeys = types.BoolValue(protection.PushWhitelistDeployKeys)
	model.EnableMergeWhitelist = types.BoolValue(protection.EnableMergeWhitelist)
	model.EnableStatusCheck = types.BoolValue(protection.EnableStatusCheck)
	model.EnableApprovalsWhitelist = types.BoolValue(protection.EnableApprovalsWhitelist)
	model.BlockOnRejectedReviews = types.BoolValue(protection.BlockOnRejectedReviews)
	model.BlockOnOfficialReviewRequests = types.BoolValue(protection.BlockOnOfficialReviewRequests)
	model.BlockOnOutdatedBranch = types.BoolValue(protection.BlockOnOutdatedBranch)
	model.DismissStaleApprovals = types.BoolValue(protection.DismissStaleApprovals)
	model.RequireSignedCommits = types.BoolValue(protection.RequireSignedCommits)

	// Map integer fields
	model.RequiredApprovals = types.Int64Value(protection.RequiredApprovals)

	// Map string fields
	model.ProtectedFilePatterns = types.StringValue(protection.ProtectedFilePatterns)
	model.UnprotectedFilePatterns = types.StringValue(protection.UnprotectedFilePatterns)

	// Map timestamp fields
	if !protection.Created.IsZero() {
		model.CreatedAt = types.StringValue(protection.Created.Format("2006-01-02T15:04:05Z07:00"))
	} else {
		model.CreatedAt = types.StringNull()
	}
	if !protection.Updated.IsZero() {
		model.UpdatedAt = types.StringValue(protection.Updated.Format("2006-01-02T15:04:05Z07:00"))
	} else {
		model.UpdatedAt = types.StringNull()
	}

	// Map list fields - always set them even if empty to ensure consistent state
	model.PushWhitelistUsernames = stringSliceToList(protection.PushWhitelistUsernames)
	model.PushWhitelistTeams = stringSliceToList(protection.PushWhitelistTeams)
	model.MergeWhitelistUsernames = stringSliceToList(protection.MergeWhitelistUsernames)
	model.MergeWhitelistTeams = stringSliceToList(protection.MergeWhitelistTeams)
	model.StatusCheckContexts = stringSliceToList(protection.StatusCheckContexts)
	model.ApprovalsWhitelistUsernames = stringSliceToList(protection.ApprovalsWhitelistUsernames)
	model.ApprovalsWhitelistTeams = stringSliceToList(protection.ApprovalsWhitelistTeams)
}

// stringSliceToList converts a Go string slice to a Terraform List.
// Returns an empty list if the slice is nil or empty, never null.
func stringSliceToList(slice []string) types.List {
	if len(slice) == 0 {
		return types.ListValueMust(types.StringType, []attr.Value{})
	}

	elements := make([]attr.Value, len(slice))
	for i, v := range slice {
		elements[i] = types.StringValue(v)
	}
	return types.ListValueMust(types.StringType, elements)
}
