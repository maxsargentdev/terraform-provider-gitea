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

	// Map response to data - use same mapping function as resource
	data.BranchName = types.StringValue(protection.BranchName)
	data.RuleName = types.StringValue(protection.RuleName)
	data.Name = types.StringValue(protection.BranchName)
	data.EnablePush = types.BoolValue(protection.EnablePush)
	data.EnablePushWhitelist = types.BoolValue(protection.EnablePushWhitelist)
	data.EnableMergeWhitelist = types.BoolValue(protection.EnableMergeWhitelist)
	data.EnableStatusCheck = types.BoolValue(protection.EnableStatusCheck)
	data.RequireSignedCommits = types.BoolValue(protection.RequireSignedCommits)
	data.CreatedAt = types.StringValue(protection.Created.String())
	data.UpdatedAt = types.StringValue(protection.Updated.String())

	// Map boolean fields from API
	data.PushWhitelistDeployKeys = types.BoolValue(protection.PushWhitelistDeployKeys)
	data.BlockOnRejectedReviews = types.BoolValue(protection.BlockOnRejectedReviews)
	data.BlockOnOfficialReviewRequests = types.BoolValue(protection.BlockOnOfficialReviewRequests)
	data.BlockOnOutdatedBranch = types.BoolValue(protection.BlockOnOutdatedBranch)
	data.DismissStaleApprovals = types.BoolValue(protection.DismissStaleApprovals)
	data.EnableApprovalsWhitelist = types.BoolValue(protection.EnableApprovalsWhitelist)

	// Map integer fields
	data.RequiredApprovals = types.Int64Value(protection.RequiredApprovals)

	// Map string fields
	data.ProtectedFilePatterns = types.StringValue(protection.ProtectedFilePatterns)
	data.UnprotectedFilePatterns = types.StringValue(protection.UnprotectedFilePatterns)

	// Map list fields
	if len(protection.PushWhitelistUsernames) > 0 {
		elements := make([]attr.Value, len(protection.PushWhitelistUsernames))
		for i, v := range protection.PushWhitelistUsernames {
			elements[i] = types.StringValue(v)
		}
		data.PushWhitelistUsernames, _ = types.ListValue(types.StringType, elements)
	} else {
		data.PushWhitelistUsernames = types.ListNull(types.StringType)
	}

	if len(protection.PushWhitelistTeams) > 0 {
		elements := make([]attr.Value, len(protection.PushWhitelistTeams))
		for i, v := range protection.PushWhitelistTeams {
			elements[i] = types.StringValue(v)
		}
		data.PushWhitelistTeams, _ = types.ListValue(types.StringType, elements)
	} else {
		data.PushWhitelistTeams = types.ListNull(types.StringType)
	}

	if len(protection.MergeWhitelistUsernames) > 0 {
		elements := make([]attr.Value, len(protection.MergeWhitelistUsernames))
		for i, v := range protection.MergeWhitelistUsernames {
			elements[i] = types.StringValue(v)
		}
		data.MergeWhitelistUsernames, _ = types.ListValue(types.StringType, elements)
	} else {
		data.MergeWhitelistUsernames = types.ListNull(types.StringType)
	}

	if len(protection.MergeWhitelistTeams) > 0 {
		elements := make([]attr.Value, len(protection.MergeWhitelistTeams))
		for i, v := range protection.MergeWhitelistTeams {
			elements[i] = types.StringValue(v)
		}
		data.MergeWhitelistTeams, _ = types.ListValue(types.StringType, elements)
	} else {
		data.MergeWhitelistTeams = types.ListNull(types.StringType)
	}

	if len(protection.ApprovalsWhitelistUsernames) > 0 {
		elements := make([]attr.Value, len(protection.ApprovalsWhitelistUsernames))
		for i, v := range protection.ApprovalsWhitelistUsernames {
			elements[i] = types.StringValue(v)
		}
		data.ApprovalsWhitelistUsername, _ = types.ListValue(types.StringType, elements)
	} else {
		data.ApprovalsWhitelistUsername = types.ListNull(types.StringType)
	}

	if len(protection.ApprovalsWhitelistTeams) > 0 {
		elements := make([]attr.Value, len(protection.ApprovalsWhitelistTeams))
		for i, v := range protection.ApprovalsWhitelistTeams {
			elements[i] = types.StringValue(v)
		}
		data.ApprovalsWhitelistTeams, _ = types.ListValue(types.StringType, elements)
	} else {
		data.ApprovalsWhitelistTeams = types.ListNull(types.StringType)
	}

	if len(protection.StatusCheckContexts) > 0 {
		elements := make([]attr.Value, len(protection.StatusCheckContexts))
		for i, v := range protection.StatusCheckContexts {
			elements[i] = types.StringValue(v)
		}
		data.StatusCheckContexts, _ = types.ListValue(types.StringType, elements)
	} else {
		data.StatusCheckContexts = types.ListNull(types.StringType)
	}

	// Set remaining fields as null (not returned by API)
	data.ForcePushAllowlistUsernames = types.ListNull(types.StringType)
	data.ForcePushAllowlistTeams = types.ListNull(types.StringType)
	data.BlockAdminMergeOverride = types.BoolNull()
	data.EnableForcePush = types.BoolNull()
	data.EnableForcePushAllowlist = types.BoolNull()
	data.ForcePushAllowlistDeployKeys = types.BoolNull()
	data.IgnoreStaleApprovals = types.BoolNull()
	data.Priority = types.Int64Null()

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
