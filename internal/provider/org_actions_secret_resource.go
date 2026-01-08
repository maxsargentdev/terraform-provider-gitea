package provider

import (
	"context"
	"fmt"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = (*orgActionsSecretResource)(nil)
var _ resource.ResourceWithConfigure = (*orgActionsSecretResource)(nil)
var _ resource.ResourceWithImportState = (*orgActionsSecretResource)(nil)

func NewOrgActionsSecretResource() resource.Resource {
	return &orgActionsSecretResource{}
}

type orgActionsSecretResource struct {
	client *gitea.Client
}

type orgActionsSecretResourceModel struct {
	// Required
	Org  types.String `tfsdk:"org"`
	Name types.String `tfsdk:"name"`
	Data types.String `tfsdk:"data"`

	// Optional
	Description types.String `tfsdk:"description"`

	// Computed
	Created types.String `tfsdk:"created_at"`
}

func (r *orgActionsSecretResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_org_actions_secret"
}

func (r *orgActionsSecretResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manages an organization actions secret in Gitea. Organization secrets are available to all repositories within the organization for GitHub Actions workflows.",
		MarkdownDescription: "Manages an organization actions secret in Gitea. Organization secrets are available to all repositories within the organization for GitHub Actions workflows.\n\n## Import\n\nOrganization action secrets can be imported using the format `org/secretName`:\n\n```shell\nterraform import gitea_org_actions_secret.example myorg/MY_SECRET\n```",
		Attributes: map[string]schema.Attribute{
			"org": schema.StringAttribute{
				Required:            true,
				Description:         "Name of the organization",
				MarkdownDescription: "Name of the organization that owns the secret.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				Description:         "Name of the secret (max 30 characters)",
				MarkdownDescription: "Name of the secret. Must be unique within the organization and cannot exceed 30 characters.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"data": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				Description:         "Value of the secret",
				MarkdownDescription: "The secret value. This is sensitive and will not be displayed in logs or state output.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Description:         "Description of the secret",
				MarkdownDescription: "Optional description of what this secret is used for.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				Description:         "Timestamp when the secret was created",
				MarkdownDescription: "Timestamp when the secret was created in RFC3339 format.",
			},
		},
	}
}

func (r *orgActionsSecretResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *orgActionsSecretResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data orgActionsSecretResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opt := gitea.CreateSecretOption{
		Name:        data.Name.ValueString(),
		Data:        data.Data.ValueString(),
		Description: data.Description.ValueString(),
	}

	_, err := r.client.CreateOrgActionSecret(data.Org.ValueString(), opt)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Organization Actions Secret",
			fmt.Sprintf("Unable to create organization actions secret '%s', got error: %s", data.Name.ValueString(), err),
		)
		return
	}

	// Read back to get created timestamp
	secrets, _, err := r.client.ListOrgActionSecret(data.Org.ValueString(), gitea.ListOrgActionSecretOption{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Created Secret",
			fmt.Sprintf("Unable to read created secret '%s', got error: %s", data.Name.ValueString(), err),
		)
		return
	}

	// Find the created secret
	for _, secret := range secrets {
		if secret.Name == data.Name.ValueString() {
			if !secret.Created.IsZero() {
				data.Created = types.StringValue(secret.Created.Format("2006-01-02T15:04:05Z07:00"))
			}
			break
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *orgActionsSecretResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data orgActionsSecretResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	secrets, _, err := r.client.ListOrgActionSecret(data.Org.ValueString(), gitea.ListOrgActionSecretOption{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Organization Actions Secrets",
			fmt.Sprintf("Unable to read organization actions secrets for '%s', got error: %s", data.Org.ValueString(), err),
		)
		return
	}

	// Find the secret by name
	found := false
	for _, secret := range secrets {
		if secret.Name == data.Name.ValueString() {
			found = true
			data.Description = types.StringValue(secret.Description)
			if !secret.Created.IsZero() {
				data.Created = types.StringValue(secret.Created.Format("2006-01-02T15:04:05Z07:00"))
			}
			// Note: The secret data is not returned by the API for security reasons
			// We preserve the existing value in state
			break
		}
	}

	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *orgActionsSecretResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data orgActionsSecretResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The API uses PUT which creates or updates
	opt := gitea.CreateSecretOption{
		Name:        data.Name.ValueString(),
		Data:        data.Data.ValueString(),
		Description: data.Description.ValueString(),
	}

	_, err := r.client.CreateOrgActionSecret(data.Org.ValueString(), opt)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Organization Actions Secret",
			fmt.Sprintf("Unable to update organization actions secret '%s', got error: %s", data.Name.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *orgActionsSecretResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data orgActionsSecretResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DeleteOrgActionSecret(data.Org.ValueString(), data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Organization Actions Secret",
			fmt.Sprintf("Unable to delete organization actions secret '%s', got error: %s", data.Name.ValueString(), err),
		)
		return
	}
}

func (r *orgActionsSecretResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: org/secretName
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid ID Format",
			fmt.Sprintf("Expected format: org/secretName, got: %s", req.ID),
		)
		return
	}

	org := parts[0]
	secretName := parts[1]

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("org"), org)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), secretName)...)
}
