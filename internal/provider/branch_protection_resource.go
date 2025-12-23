package provider

import (
	"context"
	"fmt"

	"terraform-provider-gitea/internal/resource_branch_protection"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
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

func (r *branchProtectionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_branch_protection"
}

func (r *branchProtectionResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_branch_protection.BranchProtectionResourceSchema(ctx)
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
	var plan resource_branch_protection.BranchProtectionModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create branch protection
	createOpts := gitea.CreateBranchProtectionOption{
		BranchName: plan.BranchName.ValueString(),
		RuleName:   plan.RuleName.ValueString(),
	}

	// Set optional boolean fields
	if !plan.EnablePush.IsNull() {
		createOpts.EnablePush = plan.EnablePush.ValueBool()
	}
	if !plan.EnablePushWhitelist.IsNull() {
		createOpts.EnablePushWhitelist = plan.EnablePushWhitelist.ValueBool()
	}
	if !plan.EnableMergeWhitelist.IsNull() {
		createOpts.EnableMergeWhitelist = plan.EnableMergeWhitelist.ValueBool()
	}
	if !plan.EnableStatusCheck.IsNull() {
		createOpts.EnableStatusCheck = plan.EnableStatusCheck.ValueBool()
	}
	if !plan.RequireSignedCommits.IsNull() {
		createOpts.RequireSignedCommits = plan.RequireSignedCommits.ValueBool()
	}

	// Set list fields
	if !plan.PushWhitelistUsernames.IsNull() {
		var usernames []string
		plan.PushWhitelistUsernames.ElementsAs(ctx, &usernames, false)
		createOpts.PushWhitelistUsernames = usernames
	}
	if !plan.MergeWhitelistUsernames.IsNull() {
		var usernames []string
		plan.MergeWhitelistUsernames.ElementsAs(ctx, &usernames, false)
		createOpts.MergeWhitelistUsernames = usernames
	}

	protection, _, err := r.client.CreateBranchProtection(
		plan.Owner.ValueString(),
		plan.Repo.ValueString(),
		createOpts,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating branch protection",
			"Could not create branch protection, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response to state
	mapBranchProtectionToModel(ctx, protection, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *branchProtectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resource_branch_protection.BranchProtectionModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// API uses rule_name as the identifier after creation, not branch_name
	protection, _, err := r.client.GetBranchProtection(
		state.Owner.ValueString(),
		state.Repo.ValueString(),
		state.RuleName.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Branch Protection",
			"Could not read branch protection: "+err.Error(),
		)
		return
	}

	// Map response to state
	mapBranchProtectionToModel(ctx, protection, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *branchProtectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan resource_branch_protection.BranchProtectionModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	editOpts := gitea.EditBranchProtectionOption{}

	// Set all fields from plan
	if !plan.EnablePush.IsNull() {
		editOpts.EnablePush = plan.EnablePush.ValueBoolPointer()
	}
	if !plan.EnablePushWhitelist.IsNull() {
		editOpts.EnablePushWhitelist = plan.EnablePushWhitelist.ValueBoolPointer()
	}
	if !plan.PushWhitelistUsernames.IsNull() {
		plan.PushWhitelistUsernames.ElementsAs(ctx, &editOpts.PushWhitelistUsernames, false)
	}
	if !plan.PushWhitelistTeams.IsNull() {
		plan.PushWhitelistTeams.ElementsAs(ctx, &editOpts.PushWhitelistTeams, false)
	}
	if !plan.PushWhitelistDeployKeys.IsNull() {
		editOpts.PushWhitelistDeployKeys = plan.PushWhitelistDeployKeys.ValueBoolPointer()
	}
	if !plan.EnableMergeWhitelist.IsNull() {
		editOpts.EnableMergeWhitelist = plan.EnableMergeWhitelist.ValueBoolPointer()
	}
	if !plan.MergeWhitelistUsernames.IsNull() {
		plan.MergeWhitelistUsernames.ElementsAs(ctx, &editOpts.MergeWhitelistUsernames, false)
	}
	if !plan.MergeWhitelistTeams.IsNull() {
		plan.MergeWhitelistTeams.ElementsAs(ctx, &editOpts.MergeWhitelistTeams, false)
	}
	if !plan.EnableStatusCheck.IsNull() {
		editOpts.EnableStatusCheck = plan.EnableStatusCheck.ValueBoolPointer()
	}
	if !plan.StatusCheckContexts.IsNull() {
		plan.StatusCheckContexts.ElementsAs(ctx, &editOpts.StatusCheckContexts, false)
	}
	if !plan.RequiredApprovals.IsNull() {
		val := plan.RequiredApprovals.ValueInt64()
		editOpts.RequiredApprovals = &val
	}
	if !plan.EnableApprovalsWhitelist.IsNull() {
		editOpts.EnableApprovalsWhitelist = plan.EnableApprovalsWhitelist.ValueBoolPointer()
	}
	if !plan.ApprovalsWhitelistUsername.IsNull() {
		plan.ApprovalsWhitelistUsername.ElementsAs(ctx, &editOpts.ApprovalsWhitelistUsernames, false)
	}
	if !plan.ApprovalsWhitelistTeams.IsNull() {
		plan.ApprovalsWhitelistTeams.ElementsAs(ctx, &editOpts.ApprovalsWhitelistTeams, false)
	}
	if !plan.BlockOnRejectedReviews.IsNull() {
		editOpts.BlockOnRejectedReviews = plan.BlockOnRejectedReviews.ValueBoolPointer()
	}
	if !plan.BlockOnOfficialReviewRequests.IsNull() {
		editOpts.BlockOnOfficialReviewRequests = plan.BlockOnOfficialReviewRequests.ValueBoolPointer()
	}
	if !plan.BlockOnOutdatedBranch.IsNull() {
		editOpts.BlockOnOutdatedBranch = plan.BlockOnOutdatedBranch.ValueBoolPointer()
	}
	if !plan.DismissStaleApprovals.IsNull() {
		editOpts.DismissStaleApprovals = plan.DismissStaleApprovals.ValueBoolPointer()
	}
	if !plan.RequireSignedCommits.IsNull() {
		editOpts.RequireSignedCommits = plan.RequireSignedCommits.ValueBoolPointer()
	}
	if !plan.ProtectedFilePatterns.IsNull() {
		val := plan.ProtectedFilePatterns.ValueString()
		editOpts.ProtectedFilePatterns = &val
	}
	if !plan.UnprotectedFilePatterns.IsNull() {
		val := plan.UnprotectedFilePatterns.ValueString()
		editOpts.UnprotectedFilePatterns = &val
	}

	// API uses rule_name as the identifier
	protection, _, err := r.client.EditBranchProtection(
		plan.Owner.ValueString(),
		plan.Repo.ValueString(),
		plan.RuleName.ValueString(),
		editOpts,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Branch Protection",
			"Could not update branch protection, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response to state
	mapBranchProtectionToModel(ctx, protection, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *branchProtectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resource_branch_protection.BranchProtectionModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// API uses rule_name as the identifier
	_, err := r.client.DeleteBranchProtection(
		state.Owner.ValueString(),
		state.Repo.ValueString(),
		state.RuleName.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Branch Protection",
			"Could not delete branch protection, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *branchProtectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Helper function to map Gitea BranchProtection to Terraform model
func mapBranchProtectionToModel(ctx context.Context, protection *gitea.BranchProtection, model *resource_branch_protection.BranchProtectionModel) {
	// Note: owner, repo, and branch_name need to be preserved from the plan/state (not overwritten from API)
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
