// Code generated manually for team_membership resource

package resource_team_membership

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func TeamMembershipResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "The unique identifier of the team membership (format: team_id:username)",
				MarkdownDescription: "The unique identifier of the team membership (format: team_id:username)",
			},
			"team_id": schema.Int64Attribute{
				Required:            true,
				Description:         "The ID of the team",
				MarkdownDescription: "The ID of the team",
			},
			"username": schema.StringAttribute{
				Required:            true,
				Description:         "The username of the team member",
				MarkdownDescription: "The username of the team member",
			},
		},
	}
}

type TeamMembershipModel struct {
	Id       types.String `tfsdk:"id"`
	TeamId   types.Int64  `tfsdk:"team_id"`
	Username types.String `tfsdk:"username"`
}
