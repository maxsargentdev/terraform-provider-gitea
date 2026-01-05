// Code generated manually for team_membership datasource

package datasource_team_membership

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func TeamMembershipDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description:         "Get information about a team membership (checks if a user is a member of a team)",
		MarkdownDescription: "Get information about a team membership (checks if a user is a member of a team)",
		Attributes: map[string]schema.Attribute{
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
	TeamId   types.Int64  `tfsdk:"team_id"`
	Username types.String `tfsdk:"username"`
}
