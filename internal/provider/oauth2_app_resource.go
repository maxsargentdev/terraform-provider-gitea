package provider

import (
	"context"
	"fmt"
	"strconv"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = (*oauth2AppResource)(nil)
var _ resource.ResourceWithConfigure = (*oauth2AppResource)(nil)
var _ resource.ResourceWithImportState = (*oauth2AppResource)(nil)

func NewOAuth2AppResource() resource.Resource {
	return &oauth2AppResource{}
}

type oauth2AppResource struct {
	client *gitea.Client
}

type oauth2AppResourceModel struct {
	// Required
	Name         types.String `tfsdk:"name"`
	RedirectUris types.List   `tfsdk:"redirect_uris"`

	// Optional
	ConfidentialClient types.Bool `tfsdk:"confidential_client"`

	// Computed
	Id           types.Int64  `tfsdk:"id"`
	ClientId     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
	Created      types.String `tfsdk:"created"`
}

func (r *oauth2AppResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_oauth2_app"
}

func (r *oauth2AppResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an OAuth2 application",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the OAuth2 application",
			},
			"redirect_uris": schema.ListAttribute{
				Required:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "List of redirect URIs",
			},
			"confidential_client": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether this is a confidential client",
			},
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The ID of the OAuth2 application",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"client_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The client ID",
			},
			"client_secret": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "The client secret",
			},
			"created": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Timestamp when the application was created",
			},
		},
	}
}

func (r *oauth2AppResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *oauth2AppResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data oauth2AppResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var redirectUris []string
	resp.Diagnostics.Append(data.RedirectUris.ElementsAs(ctx, &redirectUris, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opt := gitea.CreateOauth2Option{
		Name:               data.Name.ValueString(),
		ConfidentialClient: data.ConfidentialClient.ValueBool(),
		RedirectURIs:       redirectUris,
	}

	oauth2, _, err := r.client.CreateOauth2(opt)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create OAuth2 application, got error: %s", err))
		return
	}

	data.Id = types.Int64Value(oauth2.ID)
	data.ClientId = types.StringValue(oauth2.ClientID)
	data.ClientSecret = types.StringValue(oauth2.ClientSecret)
	if !oauth2.Created.IsZero() {
		data.Created = types.StringValue(oauth2.Created.Format("2006-01-02T15:04:05Z07:00"))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *oauth2AppResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data oauth2AppResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	oauth2, httpResp, err := r.client.GetOauth2(data.Id.ValueInt64())
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read OAuth2 application, got error: %s", err))
		return
	}

	data.Name = types.StringValue(oauth2.Name)
	data.ConfidentialClient = types.BoolValue(oauth2.ConfidentialClient)
	data.ClientId = types.StringValue(oauth2.ClientID)
	// Note: ClientSecret is only returned on creation, not on read
	// We preserve the existing value in state
	if !oauth2.Created.IsZero() {
		data.Created = types.StringValue(oauth2.Created.Format("2006-01-02T15:04:05Z07:00"))
	}

	// Convert redirect URIs to list
	redirectUris, diags := types.ListValueFrom(ctx, types.StringType, oauth2.RedirectURIs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.RedirectUris = redirectUris

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *oauth2AppResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data oauth2AppResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var redirectUris []string
	resp.Diagnostics.Append(data.RedirectUris.ElementsAs(ctx, &redirectUris, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opt := gitea.CreateOauth2Option{
		Name:               data.Name.ValueString(),
		ConfidentialClient: data.ConfidentialClient.ValueBool(),
		RedirectURIs:       redirectUris,
	}

	oauth2, _, err := r.client.UpdateOauth2(data.Id.ValueInt64(), opt)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update OAuth2 application, got error: %s", err))
		return
	}

	data.Name = types.StringValue(oauth2.Name)
	data.ConfidentialClient = types.BoolValue(oauth2.ConfidentialClient)
	data.ClientId = types.StringValue(oauth2.ClientID)
	// Note: ClientSecret is only returned on creation
	// We preserve the existing value in state from the previous state
	if !oauth2.Created.IsZero() {
		data.Created = types.StringValue(oauth2.Created.Format("2006-01-02T15:04:05Z07:00"))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *oauth2AppResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data oauth2AppResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DeleteOauth2(data.Id.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete OAuth2 application, got error: %s", err))
		return
	}
}

func (r *oauth2AppResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	appID, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid ID",
			fmt.Sprintf("Expected a numeric ID, got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), appID)...)
}
