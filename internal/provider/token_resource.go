package provider

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &tokenResource{}
var _ resource.ResourceWithImportState = &tokenResource{}

func NewTokenResource() resource.Resource {
	return &tokenResource{}
}

type tokenResource struct {
	client *gitea.Client
}

type tokenResourceModel struct {
	CreatedAt      types.String `tfsdk:"created_at"`
	Id             types.Int64  `tfsdk:"id"`
	LastUsedAt     types.String `tfsdk:"last_used_at"`
	Name           types.String `tfsdk:"name"`
	Scopes         types.Set    `tfsdk:"scopes"` // Override: Set instead of List
	Sha1           types.String `tfsdk:"sha1"`
	TokenLastEight types.String `tfsdk:"token_last_eight"`
}

func (r *tokenResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_token"
}

func (r *tokenResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manages an API access token for the authenticated user.",
		MarkdownDescription: "Manages an API access token for the authenticated user.",
		Attributes: map[string]schema.Attribute{

			// required - these are fundamental configuration options
			"name": schema.StringAttribute{
				Required:            true,
				Description:         "The name of the access token. Must be unique for the user.",
				MarkdownDescription: "The name of the access token. Must be unique for the user.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"scopes": schema.SetAttribute{
				ElementType:         types.StringType,
				Required:            true,
				Description:         "The scopes granted to this access token (e.g., 'repo', 'admin:org', 'user').",
				MarkdownDescription: "The scopes granted to this access token (e.g., `repo`, `admin:org`, `user`).",
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
			},

			// computed - these are available to read back after creation but are really just metadata
			"created_at": schema.StringAttribute{
				Computed:            true,
				Description:         "The timestamp when the token was created.",
				MarkdownDescription: "The timestamp when the token was created.",
			},
			"id": schema.Int64Attribute{
				Computed:            true,
				Description:         "The unique numeric identifier of the access token.",
				MarkdownDescription: "The unique numeric identifier of the access token.",
			},
			"last_used_at": schema.StringAttribute{
				Computed:            true,
				Description:         "The timestamp when the token was last used.",
				MarkdownDescription: "The timestamp when the token was last used.",
			},
			"sha1": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				Description:         "The actual token value. Only available immediately after creation; cannot be retrieved later from the API.",
				MarkdownDescription: "The actual token value. **Only available immediately after creation**; cannot be retrieved later from the API.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"token_last_eight": schema.StringAttribute{
				Computed:            true,
				Description:         "The last eight characters of the token for identification.",
				MarkdownDescription: "The last eight characters of the token for identification.",
			},
		},
	}

}

func (r *tokenResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *tokenResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan tokenResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract scopes from the plan
	var scopes []string
	resp.Diagnostics.Append(plan.Scopes.ElementsAs(ctx, &scopes, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert to gitea.AccessTokenScope slice
	apiScopes := make([]gitea.AccessTokenScope, len(scopes))
	for i, s := range scopes {
		apiScopes[i] = gitea.AccessTokenScope(s)
	}

	opts := gitea.CreateAccessTokenOption{
		Name:   plan.Name.ValueString(),
		Scopes: apiScopes,
	}

	token, _, err := r.client.CreateAccessToken(opts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Token",
			fmt.Sprintf("Could not create token '%s': %s", plan.Name.ValueString(), err.Error()),
		)
		return
	}

	mapTokenToModel(token, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *tokenResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state tokenResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tokenID := state.Id.ValueInt64()

	// List all tokens for the user and find the matching one
	// Note: The Gitea API does not provide a direct endpoint to get a single token by ID
	tokens, _, err := r.client.ListAccessTokens(gitea.ListAccessTokensOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Token",
			fmt.Sprintf("Could not list tokens: %s", err.Error()),
		)
		return
	}

	// Find the token with matching ID
	var found *gitea.AccessToken
	for i := range tokens {
		if tokens[i].ID == tokenID {
			found = tokens[i]
			break
		}
	}

	if found == nil {
		// Token not found, remove from state
		resp.State.RemoveResource(ctx)
		return
	}

	// Preserve sha1 from state since API doesn't return it after creation
	sha1Value := state.Sha1

	mapTokenToModel(found, &state)

	// Restore sha1 from previous state
	state.Sha1 = sha1Value

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *tokenResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// This should never be called because all attributes have RequiresReplace plan modifiers
	// Terraform will automatically use delete+create pattern instead of update
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Tokens cannot be updated and should trigger replacement. This is a provider bug.",
	)
}

func (r *tokenResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state tokenResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tokenID := state.Id.ValueInt64()

	httpResp, err := r.client.DeleteAccessToken(tokenID)
	if err != nil {
		// If already deleted (404), treat as success
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting Token",
			fmt.Sprintf("Could not delete token with ID %d: %s", tokenID, err.Error()),
		)
		return
	}
}

func (r *tokenResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import using the token ID
	tokenID, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Token ID",
			fmt.Sprintf("Could not parse token ID '%s': %s. Token ID must be a numeric value.", req.ID, err.Error()),
		)
		return
	}

	// List all tokens and find the matching one
	tokens, _, err := r.client.ListAccessTokens(gitea.ListAccessTokensOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Importing Token",
			fmt.Sprintf("Could not list tokens: %s", err.Error()),
		)
		return
	}

	var found *gitea.AccessToken
	for i := range tokens {
		if tokens[i].ID == tokenID {
			found = tokens[i]
			break
		}
	}

	if found == nil {
		resp.Diagnostics.AddError(
			"Token Not Found",
			fmt.Sprintf("Could not find token with ID %d. The token may have been deleted or you may not have permission to access it.", tokenID),
		)
		return
	}

	var data tokenResourceModel
	mapTokenToModel(found, &data)

	// Note: sha1 will be empty for imported tokens since API doesn't return it
	resp.Diagnostics.AddWarning(
		"Token Value Not Available",
		"The token value (sha1) is not available for imported tokens. It is only provided at creation time.",
	)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func mapTokenToModel(token *gitea.AccessToken, model *tokenResourceModel) {
	model.Id = types.Int64Value(token.ID)
	model.Name = types.StringValue(token.Name)
	model.TokenLastEight = types.StringValue(token.TokenLastEight)

	// Only set sha1 if it's present (only available on create)
	if token.Token != "" {
		model.Sha1 = types.StringValue(token.Token)
	}

	if !token.Created.IsZero() {
		model.CreatedAt = types.StringValue(token.Created.String())
	} else {
		model.CreatedAt = types.StringNull()
	}

	if !token.Updated.IsZero() {
		model.LastUsedAt = types.StringValue(token.Updated.String())
	} else {
		model.LastUsedAt = types.StringNull()
	}

	if len(token.Scopes) > 0 {
		scopeValues := make([]attr.Value, len(token.Scopes))
		for i, s := range token.Scopes {
			scopeValues[i] = types.StringValue(string(s))
		}
		model.Scopes = types.SetValueMust(types.StringType, scopeValues)
	} else {
		model.Scopes = types.SetNull(types.StringType)
	}

}
