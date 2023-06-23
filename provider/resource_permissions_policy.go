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
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &gdrivePermissionPolicyResource{}
var _ resource.ResourceWithImportState = &gdrivePermissionPolicyResource{}

func newPermissionPolicy() resource.Resource {
	return &gdrivePermissionPolicyResource{}
}

// gdrivePermissionPolicyResource defines the resource implementation.
type gdrivePermissionPolicyResource struct {
	client *http.Client
}

// gdrivePermissionPolicyResourceModel describes the resource data model.
type gdrivePermissionPolicyPermissionResourceModel struct {
	PermissionId          types.String `tfsdk:"permission_id"`
	EmailMessage          types.String `tfsdk:"email_message"`
	Type                  types.String `tfsdk:"type"`
	Domain                types.String `tfsdk:"domain"`
	EmailAddress          types.String `tfsdk:"email_address"`
	Role                  types.String `tfsdk:"role"`
	SendNotificationEmail types.Bool   `tfsdk:"send_notification_email"`
	TransferOwnership     types.Bool   `tfsdk:"transfer_ownership"`
	MoveToNewOwnersRoot   types.Bool   `tfsdk:"move_to_new_owners_root"`
}

// gdrivePermissionPolicyResourceModel describes the resource data model.
type gdrivePermissionPolicyResourceModel struct {
	FileId               types.String                                     `tfsdk:"file_id"`
	Id                   types.String                                     `tfsdk:"id"`
	Permissions          []*gdrivePermissionPolicyPermissionResourceModel `tfsdk:"permissions"`
	UseDomainAdminAccess types.Bool                                       `tfsdk:"use_domain_admin_access"`
}

func (r *gdrivePermissionPolicyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_permissions_policy"
}

func (r *gdrivePermissionPolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Creates an authoratative permissions policy on a file or Shared Drive.

**Warning: This resource will set exactly the defined permissions and remove everything else!**

It is HIGHLY recommended that you import the resource and make sure that the owner is properly set before applying it!

You can import the resource using the file's or Shared Drive's id like so:

terraform import [resource address] [fileId]

**Important**: On a *destroy*, this resource will preserve the owner and organizer permissions!`,
		Attributes: map[string]schema.Attribute{
			"id": rsId(),
			"file_id": schema.StringAttribute{
				MarkdownDescription: "ID of the file or Shared Drive.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"use_domain_admin_access": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Use domain admin access.",
			},
			"permissions": schema.SetNestedAttribute{
				Required:            true,
				MarkdownDescription: `Defines the set of permissions to set on the file or Shared Drive.`,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"permission_id": schema.StringAttribute{
							MarkdownDescription: "PermissionID of the trustee.",
							Computed:            true,
						},
						"email_message": schema.StringAttribute{
							MarkdownDescription: "An optional email message that will be sent when the permission is created.",
							Optional:            true,
						},
						"send_notification_email": schema.BoolAttribute{
							MarkdownDescription: "Wether to send a notfication email.",
							Optional:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "The type of the trustee. Can be 'user', 'domain', 'group' or 'anyone'.",
							Optional:            true,
						},
						"domain": schema.StringAttribute{
							MarkdownDescription: "The domain that should be granted access.",
							Optional:            true,
							Validators: []validator.String{
								stringvalidator.ConflictsWith(path.Expressions{
									path.MatchRelative().AtParent().AtName("email_address"),
								}...),
							},
						},
						"email_address": schema.StringAttribute{
							MarkdownDescription: "The email address of the trustee.",
							Optional:            true,
							Validators: []validator.String{
								stringvalidator.ConflictsWith(path.Expressions{
									path.MatchRelative().AtParent().AtName("domain"),
								}...),
							},
						},
						"role": schema.StringAttribute{
							MarkdownDescription: "The role.",
							Required:            true,
						},
						"transfer_ownership": schema.BoolAttribute{
							MarkdownDescription: "Whether to transfer ownership to the specified user.",
							Optional:            true,
						},
						"move_to_new_owners_root": schema.BoolAttribute{
							MarkdownDescription: "This parameter only takes effect if the item is not in a shared drive and the request is attempting to transfer the ownership of the item.",
							Optional:            true,
						},
					},
				},
			},
		},
	}
}

func (r *gdrivePermissionPolicyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *gdrivePermissionPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := &gdrivePermissionPolicyResourceModel{}
	resp.Diagnostics.Append(req.Plan.Get(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	mockState := &gdrivePermissionPolicyResourceModel{
		FileId: plan.FileId,
		Id:     plan.FileId,
	}
	resp.Diagnostics.Append(mockState.populate(ctx)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(setPermissionDiffs(plan, mockState)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Id = plan.FileId
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *gdrivePermissionPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	state := &gdrivePermissionPolicyResourceModel{}
	resp.Diagnostics.Append(req.State.Get(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(state.populate(ctx)...)
	if resp.Diagnostics.HasError() {
		return
	}
	currentPermissionsMap := state.toMap()
	for i := range currentPermissionsMap {
		sP, ok := currentPermissionsMap[combineId(currentPermissionsMap[i].Domain.ValueString(), currentPermissionsMap[i].EmailAddress.ValueString())]
		if ok {
			currentPermissionsMap[i].EmailMessage = sP.EmailMessage
			currentPermissionsMap[i].MoveToNewOwnersRoot = sP.MoveToNewOwnersRoot
			currentPermissionsMap[i].SendNotificationEmail = sP.SendNotificationEmail
			currentPermissionsMap[i].TransferOwnership = sP.TransferOwnership
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *gdrivePermissionPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := &gdrivePermissionPolicyResourceModel{}
	resp.Diagnostics.Append(req.Plan.Get(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state := &gdrivePermissionPolicyResourceModel{}
	resp.Diagnostics.Append(req.State.Get(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(setPermissionDiffs(plan, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *gdrivePermissionPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	plan := &gdrivePermissionPolicyResourceModel{}
	resp.Diagnostics.Append(req.State.Get(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	for i := range plan.Permissions {
		role := plan.Permissions[i].Role.ValueString()
		if role != "owner" && role != "organizer" {
			_, err := gsmdrive.DeletePermission(plan.FileId.ValueString(), plan.Permissions[i].PermissionId.ValueString(), plan.UseDomainAdminAccess.ValueBool())
			if err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete permission, got error: %s", err))
				return
			}
		}
	}
}

func (r *gdrivePermissionPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(importSplitId(ctx, req, resp, adminAttributeDrive, "file_id")...)
}
