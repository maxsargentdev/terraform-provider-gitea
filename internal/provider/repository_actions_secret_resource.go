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

var _ resource.Resource = (*repositoryActionsSecretResource)(nil)
var _ resource.ResourceWithConfigure = (*repositoryActionsSecretResource)(nil)
var _ resource.ResourceWithImportState = (*repositoryActionsSecretResource)(nil)

func NewRepositoryActionsSecretResource() resource.Resource {
	return &repositoryActionsSecretResource{}
}

type repositoryActionsSecretResource struct {
	client *gitea.Client
}

type repositoryActionsSecretResourceModel struct {
	// Required
	Owner types.String `tfsdk:"owner"`
	Repo  types.String `tfsdk:"repository"`
	Name  types.String `tfsdk:"name"`
	Data  types.String `tfsdk:"data"`

	// Optional
	Description types.String `tfsdk:"description"`

	// Computed
	Created types.String `tfsdk:"created_at"`
}

func (r *repositoryActionsSecretResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repository_actions_secret"
}

func (r *repositoryActionsSecretResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a repository actions secret",
		Attributes: map[string]schema.Attribute{
			"owner": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Owner of the repository",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"repository": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the repository",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the secret (max 30 characters)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"data": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: "Value of the secret",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Description of the secret",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Timestamp when the secret was created",
			},
		},
	}
}

func (r *repositoryActionsSecretResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *repositoryActionsSecretResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data repositoryActionsSecretResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opt := gitea.CreateSecretOption{
		Name:        data.Name.ValueString(),
		Data:        data.Data.ValueString(),
		Description: data.Description.ValueString(),
	}

	_, err := r.client.CreateRepoActionSecret(data.Owner.ValueString(), data.Repo.ValueString(), opt)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create repository actions secret, got error: %s", err))
		return
	}

	// Read back to get created timestamp
	secrets, _, err := r.client.ListRepoActionSecret(data.Owner.ValueString(), data.Repo.ValueString(), gitea.ListRepoActionSecretOption{})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read created secret, got error: %s", err))
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

func (r *repositoryActionsSecretResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data repositoryActionsSecretResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	secrets, _, err := r.client.ListRepoActionSecret(data.Owner.ValueString(), data.Repo.ValueString(), gitea.ListRepoActionSecretOption{})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read repository actions secrets, got error: %s", err))
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

func (r *repositoryActionsSecretResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data repositoryActionsSecretResourceModel

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

	_, err := r.client.CreateRepoActionSecret(data.Owner.ValueString(), data.Repo.ValueString(), opt)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update repository actions secret, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *repositoryActionsSecretResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data repositoryActionsSecretResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DeleteRepoActionSecret(data.Owner.ValueString(), data.Repo.ValueString(), data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete repository actions secret, got error: %s", err))
		return
	}
}

func (r *repositoryActionsSecretResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: owner/repo/secretName
	parts := strings.Split(req.ID, "/")
	if len(parts) != 3 {
		resp.Diagnostics.AddError(
			"Invalid ID format",
			fmt.Sprintf("Expected format: owner/repository/secretName, got: %s", req.ID),
		)
		return
	}

	owner := parts[0]
	repo := parts[1]
	secretName := parts[2]

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("owner"), owner)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("repository"), repo)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), secretName)...)
}
