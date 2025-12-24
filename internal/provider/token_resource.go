package provider

import (
	"context"
	"fmt"
	"strconv"

	"terraform-provider-gitea/internal/resource_token"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
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

func (r *tokenResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_token"
}

func (r *tokenResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	baseSchema := resource_token.TokenResourceSchema(ctx)

	// Make username required and force replacement on change
	baseSchema.Attributes["username"] = schema.StringAttribute{
		Required:            true,
		Description:         "The username of the user for whom the token is created",
		MarkdownDescription: "The username of the user for whom the token is created",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}

	// Add RequiresReplace to name (tokens can't be updated)
	if nameAttr, ok := baseSchema.Attributes["name"].(schema.StringAttribute); ok {
		nameAttr.PlanModifiers = []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		}
		baseSchema.Attributes["name"] = nameAttr
	}

	// Make sha1 sensitive
	if sha1Attr, ok := baseSchema.Attributes["sha1"].(schema.StringAttribute); ok {
		sha1Attr.Sensitive = true
		baseSchema.Attributes["sha1"] = sha1Attr
	}

	resp.Schema = baseSchema
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
	var plan resource_token.TokenModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := gitea.CreateAccessTokenOption{
		Name: plan.Name.ValueString(),
	}

	// Set scopes if provided
	if !plan.Scopes.IsNull() && !plan.Scopes.IsUnknown() {
		var scopes []string
		plan.Scopes.ElementsAs(ctx, &scopes, false)
		// Convert strings to AccessTokenScope
		giteaScopes := make([]gitea.AccessTokenScope, len(scopes))
		for i, s := range scopes {
			giteaScopes[i] = gitea.AccessTokenScope(s)
		}
		opts.Scopes = giteaScopes
	}

	token, _, err := r.client.CreateAccessToken(opts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Token",
			"Could not create token, unexpected error: "+err.Error(),
		)
		return
	}

	// Save the original scopes from plan to preserve order
	originalScopes := plan.Scopes

	mapTokenToModel(token, &plan)

	// Restore the scopes from plan to maintain order consistency
	// (API might return them in different order but same content)
	if !originalScopes.IsNull() && !originalScopes.IsUnknown() {
		plan.Scopes = originalScopes
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *tokenResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resource_token.TokenModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tokenID := state.Id.ValueInt64()

	// List all tokens for the user and find the matching one
	tokens, _, err := r.client.ListAccessTokens(gitea.ListAccessTokensOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Token",
			"Could not list tokens: "+err.Error(),
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

	mapTokenToModel(found, &state)

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
	var state resource_token.TokenModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tokenID := state.Id.ValueInt64()

	_, err := r.client.DeleteAccessToken(tokenID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Token",
			"Could not delete token: "+err.Error(),
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
			"Could not parse token ID: "+err.Error(),
		)
		return
	}

	// List all tokens and find the matching one
	tokens, _, err := r.client.ListAccessTokens(gitea.ListAccessTokensOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Importing Token",
			"Could not list tokens: "+err.Error(),
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
			fmt.Sprintf("Could not find token with ID %d", tokenID),
		)
		return
	}

	var data resource_token.TokenModel
	mapTokenToModel(found, &data)

	// Note: username field won't be populated from import since API doesn't return it
	data.Username = types.StringNull()

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func mapTokenToModel(token *gitea.AccessToken, model *resource_token.TokenModel) {
	model.Id = types.Int64Value(token.ID)
	model.Name = types.StringValue(token.Name)
	model.Sha1 = types.StringValue(token.Token)
	model.TokenLastEight = types.StringValue(token.TokenLastEight)

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
		scopes := make([]types.String, len(token.Scopes))
		for i, s := range token.Scopes {
			scopes[i] = types.StringValue(string(s))
		}
		attrValues := make([]attr.Value, len(scopes))
		for i, v := range scopes {
			attrValues[i] = v
		}
		model.Scopes = types.ListValueMust(types.StringType, attrValues)
	} else {
		model.Scopes = types.ListNull(types.StringType)
	}
	// Set pagination fields to null (not used in resource)
	model.Page = types.Int64Null()
	model.Limit = types.Int64Null()
}
