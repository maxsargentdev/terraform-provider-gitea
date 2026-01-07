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

var (
	_ resource.Resource                = &tokenResource{}
	_ resource.ResourceWithConfigure   = &tokenResource{}
	_ resource.ResourceWithImportState = &tokenResource{}
)

func NewTokenResource() resource.Resource {
	return &tokenResource{}
}

type tokenResource struct {
	client *gitea.Client
}

type tokenResourceModel struct {
	// Required
	Name   types.String `tfsdk:"name"`
	Scopes types.Set    `tfsdk:"scopes"`

	// Computed
	Id        types.String `tfsdk:"id"`
	LastEight types.String `tfsdk:"last_eight"`
	Token     types.String `tfsdk:"token"`
}

func (r *tokenResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_token"
}

func (r *tokenResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manages an API access token.",
		MarkdownDescription: "Manages an API access token. Note: This resource requires username/password authentication; token-based provider configuration cannot be used.",
		Attributes: map[string]schema.Attribute{
			// Required
			"name": schema.StringAttribute{
				Required:            true,
				Description:         "The name of the Access Token.",
				MarkdownDescription: "The name of the Access Token.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"scopes": schema.SetAttribute{
				ElementType:         types.StringType,
				Required:            true,
				Description:         "List of string representations of scopes for the token.",
				MarkdownDescription: "List of string representations of scopes for the token.",
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
			},

			// Computed
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "The ID of this resource.",
				MarkdownDescription: "The ID of this resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_eight": schema.StringAttribute{
				Computed:            true,
				Description:         "Final eight characters of the token.",
				MarkdownDescription: "Final eight characters of the token.",
			},
			"token": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				Description:         "The actual Access Token.",
				MarkdownDescription: "The actual Access Token. Only available immediately after creation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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

	r.mapTokenToModel(token, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *tokenResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state tokenResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tokenID, err := strconv.ParseInt(state.Id.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Token",
			fmt.Sprintf("Invalid token ID: %s", state.Id.ValueString()),
		)
		return
	}

	// List all tokens for the user and find the matching one
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

	// Preserve token from state since API doesn't return it after creation
	tokenValue := state.Token

	r.mapTokenToModel(found, &state)

	// Restore token from previous state
	state.Token = tokenValue

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *tokenResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// This should never be called because all attributes have RequiresReplace plan modifiers
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Tokens cannot be updated and should trigger replacement.",
	)
}

func (r *tokenResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state tokenResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tokenID, err := strconv.ParseInt(state.Id.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Token",
			fmt.Sprintf("Invalid token ID: %s", state.Id.ValueString()),
		)
		return
	}

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
			fmt.Sprintf("Could not parse token ID '%s': token ID must be a numeric value.", req.ID),
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
			fmt.Sprintf("Could not find token with ID %d.", tokenID),
		)
		return
	}

	var data tokenResourceModel
	r.mapTokenToModel(found, &data)

	// Note: token will be empty for imported tokens since API doesn't return it
	resp.Diagnostics.AddWarning(
		"Token Value Not Available",
		"The token value is not available for imported tokens. It is only provided at creation time.",
	)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *tokenResource) mapTokenToModel(token *gitea.AccessToken, model *tokenResourceModel) {
	model.Id = types.StringValue(fmt.Sprintf("%d", token.ID))
	model.Name = types.StringValue(token.Name)
	model.LastEight = types.StringValue(token.TokenLastEight)

	// Only set token if it's present (only available on create)
	if token.Token != "" {
		model.Token = types.StringValue(token.Token)
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
