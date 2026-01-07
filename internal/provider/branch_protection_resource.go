package provider

import (
	"context"
	"fmt"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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
	_ resource.Resource                   = &repositoryBranchProtectionResource{}
	_ resource.ResourceWithConfigure      = &repositoryBranchProtectionResource{}
	_ resource.ResourceWithImportState    = &repositoryBranchProtectionResource{}
	_ resource.ResourceWithValidateConfig = &repositoryBranchProtectionResource{}
)

func NewRepositoryBranchProtectionResource() resource.Resource {
	return &repositoryBranchProtectionResource{}
}

type repositoryBranchProtectionResource struct {
	client *gitea.Client
}

// repositoryBranchProtectionResourceModel describes the resource data model.
type repositoryBranchProtectionResourceModel struct {
	// Required - identification fields
	Username types.String `tfsdk:"username"`
	Name     types.String `tfsdk:"name"`
	RuleName types.String `tfsdk:"rule_name"`

	// Optional - Push settings
	EnablePush              types.Bool `tfsdk:"enable_push"`
	PushWhitelistUsers      types.List `tfsdk:"push_whitelist_users"`
	PushWhitelistTeams      types.List `tfsdk:"push_whitelist_teams"`
	PushWhitelistDeployKeys types.Bool `tfsdk:"push_whitelist_deploy_keys"`

	// Optional - Merge settings
	MergeWhitelistUsers types.List `tfsdk:"merge_whitelist_users"`
	MergeWhitelistTeams types.List `tfsdk:"merge_whitelist_teams"`

	// Optional - Status check settings
	StatusCheckPatterns types.List `tfsdk:"status_check_patterns"`

	// Optional - Approval settings
	RequiredApprovals      types.Int64 `tfsdk:"required_approvals"`
	ApprovalWhitelistUsers types.List  `tfsdk:"approval_whitelist_users"`
	ApprovalWhitelistTeams types.List  `tfsdk:"approval_whitelist_teams"`

	// Optional - Review settings
	BlockMergeOnRejectedReviews        types.Bool `tfsdk:"block_merge_on_rejected_reviews"`
	BlockMergeOnOfficialReviewRequests types.Bool `tfsdk:"block_merge_on_official_review_requests"`
	BlockMergeOnOutdatedBranch         types.Bool `tfsdk:"block_merge_on_outdated_branch"`
	DismissStaleApprovals              types.Bool `tfsdk:"dismiss_stale_approvals"`

	// Optional - Other settings
	RequireSignedCommits    types.Bool   `tfsdk:"require_signed_commits"`
	ProtectedFilePatterns   types.String `tfsdk:"protected_file_patterns"`
	UnprotectedFilePatterns types.String `tfsdk:"unprotected_file_patterns"`

	// Computed - read-only metadata
	Id                      types.String `tfsdk:"id"`
	CreatedAt               types.String `tfsdk:"created_at"`
	UpdatedAt               types.String `tfsdk:"updated_at"`
	EnableApprovalWhitelist types.Bool   `tfsdk:"enable_approval_whitelist"`
	EnableMergeWhitelist    types.Bool   `tfsdk:"enable_merge_whitelist"`
	EnablePushWhitelist     types.Bool   `tfsdk:"enable_push_whitelist"`
	EnableStatusCheck       types.Bool   `tfsdk:"enable_status_check"`
}

func (r *repositoryBranchProtectionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repository_branch_protection"
}

func (r *repositoryBranchProtectionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manages branch protection rules for a Gitea repository.",
		MarkdownDescription: "Manages branch protection rules for a Gitea repository.",
		Attributes: map[string]schema.Attribute{
			// Required identification fields
			"username": schema.StringAttribute{
				Required:            true,
				Description:         "User name or organization name.",
				MarkdownDescription: "User name or organization name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				Description:         "Repository name.",
				MarkdownDescription: "Repository name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"rule_name": schema.StringAttribute{
				Required:            true,
				Description:         "Protected Branch Name Pattern.",
				MarkdownDescription: "Protected Branch Name Pattern.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			// Push settings
			"enable_push": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				Description:         "Anyone with write access will be allowed to push to this branch (but not force push).",
				MarkdownDescription: "Anyone with write access will be allowed to push to this branch (but not force push).",
			},
			"push_whitelist_users": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "Allowlisted users for pushing. Requires enable_push to be set to true.",
				MarkdownDescription: "Allowlisted users for pushing. Requires enable_push to be set to true.",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"push_whitelist_teams": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "Allowlisted teams for pushing. Requires enable_push to be set to true.",
				MarkdownDescription: "Allowlisted teams for pushing. Requires enable_push to be set to true.",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"push_whitelist_deploy_keys": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				Description:         "Allow deploy keys with write access to push. Requires enable_push.",
				MarkdownDescription: "Allow deploy keys with write access to push. Requires enable_push.",
			},

			// Merge settings
			"merge_whitelist_users": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "Allow only allowlisted users to merge pull requests into this branch.",
				MarkdownDescription: "Allow only allowlisted users to merge pull requests into this branch.",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"merge_whitelist_teams": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "Allow only allowlisted teams to merge pull requests into this branch.",
				MarkdownDescription: "Allow only allowlisted teams to merge pull requests into this branch.",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},

			// Status check settings
			"status_check_patterns": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "Patterns to specify which status checks must pass before branches can be merged into a branch that matches this rule.",
				MarkdownDescription: "Patterns to specify which status checks must pass before branches can be merged into a branch that matches this rule.",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},

			// Approval settings
			"required_approvals": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
				Description:         "Allow only to merge pull request with enough positive reviews.",
				MarkdownDescription: "Allow only to merge pull request with enough positive reviews.",
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"approval_whitelist_users": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "Only reviews from allowlisted users will count to the required approvals.",
				MarkdownDescription: "Only reviews from allowlisted users will count to the required approvals.",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"approval_whitelist_teams": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "Only reviews from allowlisted teams will count to the required approvals.",
				MarkdownDescription: "Only reviews from allowlisted teams will count to the required approvals.",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},

			// Review settings
			"block_merge_on_rejected_reviews": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				Description:         "Merging will not be possible when changes are requested by official reviewers.",
				MarkdownDescription: "Merging will not be possible when changes are requested by official reviewers.",
			},
			"block_merge_on_official_review_requests": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				Description:         "Merging will not be possible when it has official review requests, even if there are enough approvals.",
				MarkdownDescription: "Merging will not be possible when it has official review requests, even if there are enough approvals.",
			},
			"block_merge_on_outdated_branch": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				Description:         "Merging will not be possible when head branch is behind base branch.",
				MarkdownDescription: "Merging will not be possible when head branch is behind base branch.",
			},
			"dismiss_stale_approvals": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				Description:         "When new commits that change the content of the pull request are pushed to the branch, old approvals will be dismissed.",
				MarkdownDescription: "When new commits that change the content of the pull request are pushed to the branch, old approvals will be dismissed.",
			},

			// Other settings
			"require_signed_commits": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				Description:         "Reject pushes to this branch if they are unsigned or unverifiable.",
				MarkdownDescription: "Reject pushes to this branch if they are unsigned or unverifiable.",
			},
			"protected_file_patterns": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				Description:         "Protected file patterns (separated using semicolon ';').",
				MarkdownDescription: "Protected file patterns (separated using semicolon `;`).",
			},
			"unprotected_file_patterns": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				Description:         "Unprotected file patterns (separated using semicolon ';').",
				MarkdownDescription: "Unprotected file patterns (separated using semicolon `;`).",
			},

			// Computed metadata
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "The ID of this resource.",
				MarkdownDescription: "The ID of this resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				Description:         "Timestamp when the rule was created.",
				MarkdownDescription: "Timestamp when the rule was created.",
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				Description:         "Timestamp when the rule was last updated.",
				MarkdownDescription: "Timestamp when the rule was last updated.",
			},
			"enable_approval_whitelist": schema.BoolAttribute{
				Computed:            true,
				Description:         "True if an approval whitelist is used.",
				MarkdownDescription: "True if an approval whitelist is used.",
			},
			"enable_merge_whitelist": schema.BoolAttribute{
				Computed:            true,
				Description:         "True if a merge whitelist is used.",
				MarkdownDescription: "True if a merge whitelist is used.",
			},
			"enable_push_whitelist": schema.BoolAttribute{
				Computed:            true,
				Description:         "True if a push whitelist is used.",
				MarkdownDescription: "True if a push whitelist is used.",
			},
			"enable_status_check": schema.BoolAttribute{
				Computed:            true,
				Description:         "Require status checks to pass before merging.",
				MarkdownDescription: "Require status checks to pass before merging.",
			},
		},
	}
}

func (r *repositoryBranchProtectionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *repositoryBranchProtectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan repositoryBranchProtectionResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build create options with all fields
	createOpts := gitea.CreateBranchProtectionOption{
		RuleName:                      plan.RuleName.ValueString(),
		EnablePush:                    plan.EnablePush.ValueBool(),
		PushWhitelistDeployKeys:       plan.PushWhitelistDeployKeys.ValueBool(),
		RequiredApprovals:             plan.RequiredApprovals.ValueInt64(),
		BlockOnRejectedReviews:        plan.BlockMergeOnRejectedReviews.ValueBool(),
		BlockOnOfficialReviewRequests: plan.BlockMergeOnOfficialReviewRequests.ValueBool(),
		BlockOnOutdatedBranch:         plan.BlockMergeOnOutdatedBranch.ValueBool(),
		DismissStaleApprovals:         plan.DismissStaleApprovals.ValueBool(),
		RequireSignedCommits:          plan.RequireSignedCommits.ValueBool(),
		ProtectedFilePatterns:         plan.ProtectedFilePatterns.ValueString(),
		UnprotectedFilePatterns:       plan.UnprotectedFilePatterns.ValueString(),
	}

	// Set list fields
	if !plan.PushWhitelistUsers.IsNull() && !plan.PushWhitelistUsers.IsUnknown() {
		var users []string
		resp.Diagnostics.Append(plan.PushWhitelistUsers.ElementsAs(ctx, &users, false)...)
		createOpts.PushWhitelistUsernames = users
	}
	if !plan.PushWhitelistTeams.IsNull() && !plan.PushWhitelistTeams.IsUnknown() {
		var teams []string
		resp.Diagnostics.Append(plan.PushWhitelistTeams.ElementsAs(ctx, &teams, false)...)
		createOpts.PushWhitelistTeams = teams
	}
	if !plan.MergeWhitelistUsers.IsNull() && !plan.MergeWhitelistUsers.IsUnknown() {
		var users []string
		resp.Diagnostics.Append(plan.MergeWhitelistUsers.ElementsAs(ctx, &users, false)...)
		createOpts.MergeWhitelistUsernames = users
	}
	if !plan.MergeWhitelistTeams.IsNull() && !plan.MergeWhitelistTeams.IsUnknown() {
		var teams []string
		resp.Diagnostics.Append(plan.MergeWhitelistTeams.ElementsAs(ctx, &teams, false)...)
		createOpts.MergeWhitelistTeams = teams
	}
	if !plan.StatusCheckPatterns.IsNull() && !plan.StatusCheckPatterns.IsUnknown() {
		var patterns []string
		resp.Diagnostics.Append(plan.StatusCheckPatterns.ElementsAs(ctx, &patterns, false)...)
		createOpts.StatusCheckContexts = patterns
	}
	if !plan.ApprovalWhitelistUsers.IsNull() && !plan.ApprovalWhitelistUsers.IsUnknown() {
		var users []string
		resp.Diagnostics.Append(plan.ApprovalWhitelistUsers.ElementsAs(ctx, &users, false)...)
		createOpts.ApprovalsWhitelistUsernames = users
	}
	if !plan.ApprovalWhitelistTeams.IsNull() && !plan.ApprovalWhitelistTeams.IsUnknown() {
		var teams []string
		resp.Diagnostics.Append(plan.ApprovalWhitelistTeams.ElementsAs(ctx, &teams, false)...)
		createOpts.ApprovalsWhitelistTeams = teams
	}

	if resp.Diagnostics.HasError() {
		return
	}

	protection, _, err := r.client.CreateBranchProtection(
		plan.Username.ValueString(),
		plan.Name.ValueString(),
		createOpts,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Branch Protection",
			fmt.Sprintf("Could not create branch protection rule '%s' for %s/%s: %s",
				plan.RuleName.ValueString(), plan.Username.ValueString(), plan.Name.ValueString(), err.Error()),
		)
		return
	}

	// Map response to state
	r.mapBranchProtectionToModel(ctx, protection, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *repositoryBranchProtectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state repositoryBranchProtectionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	username := state.Username.ValueString()
	name := state.Name.ValueString()
	ruleName := state.RuleName.ValueString()

	protection, httpResp, err := r.client.GetBranchProtection(username, name, ruleName)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Branch Protection",
			fmt.Sprintf("Could not read branch protection rule '%s' for %s/%s: %s",
				ruleName, username, name, err.Error()),
		)
		return
	}

	// Preserve username and name from state
	state.Username = types.StringValue(username)
	state.Name = types.StringValue(name)

	// Map response to state
	r.mapBranchProtectionToModel(ctx, protection, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *repositoryBranchProtectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan repositoryBranchProtectionResourceModel
	var state repositoryBranchProtectionResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build edit options with all fields
	editOpts := gitea.EditBranchProtectionOption{
		EnablePush:                    plan.EnablePush.ValueBoolPointer(),
		PushWhitelistDeployKeys:       plan.PushWhitelistDeployKeys.ValueBoolPointer(),
		BlockOnRejectedReviews:        plan.BlockMergeOnRejectedReviews.ValueBoolPointer(),
		BlockOnOfficialReviewRequests: plan.BlockMergeOnOfficialReviewRequests.ValueBoolPointer(),
		BlockOnOutdatedBranch:         plan.BlockMergeOnOutdatedBranch.ValueBoolPointer(),
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
	if !plan.PushWhitelistUsers.IsNull() && !plan.PushWhitelistUsers.IsUnknown() {
		var users []string
		resp.Diagnostics.Append(plan.PushWhitelistUsers.ElementsAs(ctx, &users, false)...)
		editOpts.PushWhitelistUsernames = users
	}
	if !plan.PushWhitelistTeams.IsNull() && !plan.PushWhitelistTeams.IsUnknown() {
		var teams []string
		resp.Diagnostics.Append(plan.PushWhitelistTeams.ElementsAs(ctx, &teams, false)...)
		editOpts.PushWhitelistTeams = teams
	}
	if !plan.MergeWhitelistUsers.IsNull() && !plan.MergeWhitelistUsers.IsUnknown() {
		var users []string
		resp.Diagnostics.Append(plan.MergeWhitelistUsers.ElementsAs(ctx, &users, false)...)
		editOpts.MergeWhitelistUsernames = users
	}
	if !plan.MergeWhitelistTeams.IsNull() && !plan.MergeWhitelistTeams.IsUnknown() {
		var teams []string
		resp.Diagnostics.Append(plan.MergeWhitelistTeams.ElementsAs(ctx, &teams, false)...)
		editOpts.MergeWhitelistTeams = teams
	}
	if !plan.StatusCheckPatterns.IsNull() && !plan.StatusCheckPatterns.IsUnknown() {
		var patterns []string
		resp.Diagnostics.Append(plan.StatusCheckPatterns.ElementsAs(ctx, &patterns, false)...)
		editOpts.StatusCheckContexts = patterns
	}
	if !plan.ApprovalWhitelistUsers.IsNull() && !plan.ApprovalWhitelistUsers.IsUnknown() {
		var users []string
		resp.Diagnostics.Append(plan.ApprovalWhitelistUsers.ElementsAs(ctx, &users, false)...)
		editOpts.ApprovalsWhitelistUsernames = users
	}
	if !plan.ApprovalWhitelistTeams.IsNull() && !plan.ApprovalWhitelistTeams.IsUnknown() {
		var teams []string
		resp.Diagnostics.Append(plan.ApprovalWhitelistTeams.ElementsAs(ctx, &teams, false)...)
		editOpts.ApprovalsWhitelistTeams = teams
	}

	if resp.Diagnostics.HasError() {
		return
	}

	protection, _, err := r.client.EditBranchProtection(
		state.Username.ValueString(),
		state.Name.ValueString(),
		state.RuleName.ValueString(),
		editOpts,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Branch Protection",
			fmt.Sprintf("Could not update branch protection rule '%s' for %s/%s: %s",
				state.RuleName.ValueString(), state.Username.ValueString(), state.Name.ValueString(), err.Error()),
		)
		return
	}

	// Preserve username and name from plan
	plan.Username = state.Username
	plan.Name = state.Name

	// Map response to state
	r.mapBranchProtectionToModel(ctx, protection, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *repositoryBranchProtectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state repositoryBranchProtectionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	username := state.Username.ValueString()
	name := state.Name.ValueString()
	ruleName := state.RuleName.ValueString()

	httpResp, err := r.client.DeleteBranchProtection(username, name, ruleName)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting Branch Protection",
			fmt.Sprintf("Could not delete branch protection rule '%s' for %s/%s: %s",
				ruleName, username, name, err.Error()),
		)
		return
	}
}

func (r *repositoryBranchProtectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: "username/name/rule_name"
	id := req.ID

	parts := strings.SplitN(id, "/", 3)
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Import ID must be in format 'username/name/rule_name', got: %s", id),
		)
		return
	}

	username := parts[0]
	name := parts[1]
	ruleName := parts[2]

	protection, httpResp, err := r.client.GetBranchProtection(username, name, ruleName)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			resp.Diagnostics.AddError(
				"Branch Protection Not Found",
				fmt.Sprintf("Branch protection rule '%s' not found for %s/%s", ruleName, username, name),
			)
			return
		}
		resp.Diagnostics.AddError(
			"Error Importing Branch Protection",
			fmt.Sprintf("Could not import branch protection rule '%s' for %s/%s: %s",
				ruleName, username, name, err.Error()),
		)
		return
	}

	var data repositoryBranchProtectionResourceModel
	data.Username = types.StringValue(username)
	data.Name = types.StringValue(name)
	r.mapBranchProtectionToModel(ctx, protection, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *repositoryBranchProtectionResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	// No custom validation needed - enable_* flags are now computed
}

// Helper function to map Gitea BranchProtection to Terraform model
func (r *repositoryBranchProtectionResource) mapBranchProtectionToModel(ctx context.Context, protection *gitea.BranchProtection, model *repositoryBranchProtectionResourceModel) {
	// Map the rule name
	model.RuleName = types.StringValue(protection.RuleName)

	// Map the id field (format: username/name/rule_name)
	if !model.Username.IsNull() && !model.Name.IsNull() {
		model.Id = types.StringValue(fmt.Sprintf("%s/%s/%s",
			model.Username.ValueString(),
			model.Name.ValueString(),
			protection.RuleName))
	}

	// Map boolean fields
	model.EnablePush = types.BoolValue(protection.EnablePush)
	model.EnablePushWhitelist = types.BoolValue(protection.EnablePushWhitelist)
	model.PushWhitelistDeployKeys = types.BoolValue(protection.PushWhitelistDeployKeys)
	model.EnableMergeWhitelist = types.BoolValue(protection.EnableMergeWhitelist)
	model.EnableStatusCheck = types.BoolValue(protection.EnableStatusCheck)
	model.EnableApprovalWhitelist = types.BoolValue(protection.EnableApprovalsWhitelist)
	model.BlockMergeOnRejectedReviews = types.BoolValue(protection.BlockOnRejectedReviews)
	model.BlockMergeOnOfficialReviewRequests = types.BoolValue(protection.BlockOnOfficialReviewRequests)
	model.BlockMergeOnOutdatedBranch = types.BoolValue(protection.BlockOnOutdatedBranch)
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
	model.PushWhitelistUsers = stringSliceToListBP(protection.PushWhitelistUsernames)
	model.PushWhitelistTeams = stringSliceToListBP(protection.PushWhitelistTeams)
	model.MergeWhitelistUsers = stringSliceToListBP(protection.MergeWhitelistUsernames)
	model.MergeWhitelistTeams = stringSliceToListBP(protection.MergeWhitelistTeams)
	model.StatusCheckPatterns = stringSliceToListBP(protection.StatusCheckContexts)
	model.ApprovalWhitelistUsers = stringSliceToListBP(protection.ApprovalsWhitelistUsernames)
	model.ApprovalWhitelistTeams = stringSliceToListBP(protection.ApprovalsWhitelistTeams)
}

// stringSliceToListBP converts a Go string slice to a Terraform List.
// Returns an empty list if the slice is nil or empty, never null.
func stringSliceToListBP(slice []string) types.List {
	if len(slice) == 0 {
		return types.ListValueMust(types.StringType, []attr.Value{})
	}

	elements := make([]attr.Value, len(slice))
	for i, v := range slice {
		elements[i] = types.StringValue(v)
	}
	return types.ListValueMust(types.StringType, elements)
}
