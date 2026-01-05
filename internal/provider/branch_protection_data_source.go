package provider

import (
	"context"
	"fmt"

	"terraform-provider-gitea/internal/datasource_branch_protection"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
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
	resp.Schema = datasource_branch_protection.BranchProtectionDataSourceSchema(ctx)
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
	var data datasource_branch_protection.BranchProtectionModel

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
func mapBranchProtectionToDataSourceModel(ctx context.Context, protection *gitea.BranchProtection, model *datasource_branch_protection.BranchProtectionModel) {
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
