/*
Copyright © 2021-2023 Hannes Hayashi

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/

package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hanneshayashi/gsm/gsmdrive"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"google.golang.org/api/drive/v3"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &gdrivePermissionResource{}
var _ resource.ResourceWithImportState = &gdrivePermissionResource{}

const fieldsPermission = "emailAddress,domain,role,type,id,permissionDetails(inherited)"

func newPermission() resource.Resource {
	return &gdrivePermissionResource{}
}

// gdrivePermissionResource defines the resource implementation.
type gdrivePermissionResource struct {
	client *http.Client
}

// gdrivePermissionResourceModel describes the resource data model.
type gdrivePermissionResourceModel struct {
	FileId                types.String `tfsdk:"file_id"`
	PermissionId          types.String `tfsdk:"permission_id"`
	EmailMessage          types.String `tfsdk:"email_message"`
	SendNotificationEmail types.Bool   `tfsdk:"send_notification_email"`
	UseDomainAdminAccess  types.Bool   `tfsdk:"use_domain_admin_access"`
	TransferOwnership     types.Bool   `tfsdk:"transfer_ownership"`
	MoveToNewOwnersRoot   types.Bool   `tfsdk:"move_to_new_owners_root"`
	Id                    types.String `tfsdk:"id"`
	Type                  types.String `tfsdk:"type"`
	Domain                types.String `tfsdk:"domain"`
	EmailAddress          types.String `tfsdk:"email_address"`
	Role                  types.String `tfsdk:"role"`
}

func (r *gdrivePermissionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_permission"
}

func (r *gdrivePermissionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Creates a file or folder with the given MIME type and optionally uploads a local file",
		Attributes: map[string]schema.Attribute{
			"id": rsId(),
			"file_id": schema.StringAttribute{
				MarkdownDescription: "ID of the file or Shared Drive",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"permission_id": schema.StringAttribute{
				MarkdownDescription: "PermissionID of the trustee",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"email_message": schema.StringAttribute{
				MarkdownDescription: "An optional email message that will be sent when the permission is created",
				Optional:            true,
			},
			"send_notification_email": schema.BoolAttribute{
				MarkdownDescription: "Wether to send a notfication email",
				Optional:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of the trustee. Can be 'user', 'domain', 'group' or 'anyone'",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				// Validators: []validator.String{
				// TODO
				// 				ValidateFunc: validatePermissionType,
				// },
			},
			"domain": schema.StringAttribute{
				MarkdownDescription: "The domain that should be granted access",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				// Validators: []validator.String{
				// TODO
				// 				ConflictsWith: []string{"email_address"},
				// },
			},
			"email_address": schema.StringAttribute{
				MarkdownDescription: "The email address of the trustee",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				// Validators: []validator.String{
				// TODO
				// 				ConflictsWith: []string{"domain"},
				// },
			},
			"role": schema.StringAttribute{
				MarkdownDescription: "The role",
				Required:            true,
				// Validators: []validator.String{
				// TODO
				// 				ConflictsWith: []string{"email_address"},
				// },
			},
			"use_domain_admin_access": schema.BoolAttribute{
				MarkdownDescription: "Use domain admin access",
				Optional:            true,
			},
			"transfer_ownership": schema.BoolAttribute{
				MarkdownDescription: "Whether to transfer ownership to the specified user",
				Optional:            true,
			},
			"move_to_new_owners_root": schema.BoolAttribute{
				MarkdownDescription: "This parameter only takes effect if the item is not in a shared drive and the request is attempting to transfer the ownership of the item.",
				Optional:            true,
			},
		},
	}
}

func (r *gdrivePermissionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*http.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *gdrivePermissionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := &gdrivePermissionResourceModel{}
	resp.Diagnostics.Append(req.Plan.Get(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	fileID := plan.FileId.ValueString()
	permissionReq := &drive.Permission{
		Domain:       plan.Domain.ValueString(),
		EmailAddress: plan.EmailAddress.ValueString(),
		Role:         plan.Role.ValueString(),
		Type:         plan.Type.ValueString(),
	}
	p, err := gsmdrive.CreatePermission(fileID, plan.EmailMessage.ValueString(), fieldsPermission, plan.UseDomainAdminAccess.ValueBool(), plan.SendNotificationEmail.ValueBool(), plan.TransferOwnership.ValueBool(), plan.MoveToNewOwnersRoot.ValueBool(), permissionReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set permission on file, got error: %s", err))
		return
	}
	plan.Id = types.StringValue((combineId(fileID, p.Id)))
	plan.PermissionId = types.StringValue(p.Id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *gdrivePermissionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	state := &gdrivePermissionResourceModel{}
	resp.Diagnostics.Append(req.State.Get(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	p, err := gsmdrive.GetPermission(state.FileId.ValueString(), state.PermissionId.ValueString(), fieldsPermission, state.UseDomainAdminAccess.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read permission on file, got error: %s", err))
		return
	}
	if p.EmailAddress != "" {
		state.EmailAddress = types.StringValue(p.EmailAddress)
	}
	if p.Domain != "" {
		state.Domain = types.StringValue(p.Domain)
	}
	state.Role = types.StringValue(p.Role)
	state.Type = types.StringValue(p.Type)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *gdrivePermissionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := &gdrivePermissionResourceModel{}
	resp.Diagnostics.Append(req.Plan.Get(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state := &gdrivePermissionResourceModel{}
	resp.Diagnostics.Append(req.State.Get(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	permissionReq := &drive.Permission{
		Role: plan.Role.ValueString(),
	}
	_, err := gsmdrive.UpdatePermission(plan.FileId.ValueString(), plan.PermissionId.ValueString(), fieldsPermission, plan.UseDomainAdminAccess.ValueBool(), false, permissionReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update permission on file, got error: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *gdrivePermissionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	plan := &gdrivePermissionResourceModel{}
	resp.Diagnostics.Append(req.State.Get(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, err := gsmdrive.DeletePermission(plan.FileId.ValueString(), plan.PermissionId.ValueString(), plan.UseDomainAdminAccess.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete permission, got error: %s", err))
		return
	}
}

func (r *gdrivePermissionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// var validPermissionTypes = []string{
// 	"user",
// 	"group",
// 	"domain",
// 	"anyone",
// }
