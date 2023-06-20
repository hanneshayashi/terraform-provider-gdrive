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
	"strconv"
	"strings"

	"github.com/hanneshayashi/gsm/gsmdrivelabels"
	"github.com/hanneshayashi/gsm/gsmhelpers"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"google.golang.org/api/drivelabels/v2"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &gdriveLabelPermissionResourceModel{}
var _ resource.ResourceWithImportState = &gdriveLabelPermissionResourceModel{}

func newLabelPermission() resource.Resource {
	return &gdriveLabelPermissionResourceModel{}
}

// gdriveLabelPermissionResourceModel defines the resource implementation.
type gdriveLabelPermissionResource struct {
	client *http.Client
}

// gdriveLabelPermissionResourceModel describes the resource data model.
type gdriveLabelPermissionResourceModel struct {
	Id             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Parent         types.String `tfsdk:"parent"`
	Email          types.String `tfsdk:"email"`
	Audience       types.String `tfsdk:"audience"`
	Role           types.String `tfsdk:"role"`
	UseAdminAccess types.Bool   `tfsdk:"use_admin_access"`
}

func (r *gdriveLabelPermissionResourceModel) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_label_permission"
}

func (r *gdriveLabelPermissionResourceModel) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Updates a Label's permissions.
If a permission for the indicated principal doesn't exist, a new Label Permission is created, otherwise the existing permission is updated.
Permissions affect the Label resource as a whole, are not revisioned, and do not require publishing.`,
		Attributes: map[string]schema.Attribute{
			"id": rsId(),
			"name": schema.StringAttribute{
				MarkdownDescription: "Resource name of this permission.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"parent": schema.StringAttribute{
				MarkdownDescription: "The ID of the label.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"use_admin_access": schema.BoolAttribute{
				Optional: true,
				Description: `Set to true in order to use the user's admin credentials.
The server verifies that the user is an admin for the label before allowing access.`,
			},
			"email": schema.StringAttribute{
				Optional: true,
				Description: `Specifies the email address for a user or group pricinpal.
User and Group permissions may only be inserted using email address.
On update requests, if email address is specified, no principal should be specified.`,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.Expressions{
						path.MatchRoot("audience"),
						path.MatchRoot("email"),
					}...),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"audience": schema.StringAttribute{
				MarkdownDescription: `Audience to grant a role to.
The magic value of audiences/default may be used to apply the role to the default audience in the context of the organization that owns the Label.`,
				Optional: true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.Expressions{
						path.MatchRoot("audience"),
						path.MatchRoot("email"),
					}...),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role": schema.StringAttribute{
				Required: true,
				Description: `The role the principal should have. Possible values are:

* READER     - A reader can read the label and associated metadata applied to Drive items.
* APPLIER    - An applier can write associated metadata on Drive items in which they also have write access to. Implies READER.
* ORGANIZER  - An organizer can pin this label in shared drives they manage and add new appliers to the label.
* EDITOR     - Editors can make any update including deleting the label which also deletes the associated Drive item metadata. Implies APPLIER.`,
			},
		},
	}
}

func (r *gdriveLabelPermissionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (permissionModel *gdriveLabelPermissionResourceModel) toPermission() *drivelabels.GoogleAppsDriveLabelsV2LabelPermission {
	permission := &drivelabels.GoogleAppsDriveLabelsV2LabelPermission{
		Role: permissionModel.Role.ValueString(),
	}
	if !permissionModel.Audience.IsNull() {
		permission.Audience = permissionModel.Audience.ValueString()
	}
	if !permissionModel.Email.IsNull() {
		permission.Email = permissionModel.Email.ValueString()
	}
	return permission
}

func (r *gdriveLabelPermissionResourceModel) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := &gdriveLabelPermissionResourceModel{}
	resp.Diagnostics.Append(req.Plan.Get(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	p, err := gsmdrivelabels.CreateLabelPermission(gsmhelpers.EnsurePrefix(plan.Parent.ValueString(), "labels/"), "", plan.UseAdminAccess.ValueBool(), plan.toPermission())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create label permission, got error: %s", err))
		return
	}
	plan.Name = types.StringValue(p.Name)
	plan.Id = plan.Name
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *gdriveLabelPermissionResourceModel) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	state := &gdriveLabelPermissionResourceModel{}
	resp.Diagnostics.Append(req.State.Get(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	name := state.Id.ValueString()
	currentP, err := gsmdrivelabels.ListLabelPermissions(gsmhelpers.EnsurePrefix(state.Parent.ValueString(), "labels/"), "", state.UseAdminAccess.ValueBool(), 1)
	for i := range currentP {
		if i.Name == name {
			if i.Email != "" {
				state.Email = types.StringValue(i.Email)
			}
			if i.Audience != "" && state.Audience.ValueString() != "audiences/default" {
				state.Audience = types.StringValue(i.Audience)
			}
			state.Role = types.StringValue(i.Role)
		}
	}
	e := <-err
	if e != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list permissions on label, got error: %s", e))
	}
	state.Name = state.Id
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *gdriveLabelPermissionResourceModel) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := &gdriveLabelPermissionResourceModel{}
	resp.Diagnostics.Append(req.Plan.Get(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state := &gdriveLabelPermissionResourceModel{}
	resp.Diagnostics.Append(req.State.Get(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	useAdminAccess := plan.UseAdminAccess.ValueBool()
	parent := gsmhelpers.EnsurePrefix(plan.Parent.ValueString(), "labels/")
	updateReq := &drivelabels.GoogleAppsDriveLabelsV2BatchUpdateLabelPermissionsRequest{
		UseAdminAccess: useAdminAccess,
		Requests: []*drivelabels.GoogleAppsDriveLabelsV2UpdateLabelPermissionRequest{
			{
				UseAdminAccess:  useAdminAccess,
				Parent:          parent,
				LabelPermission: plan.toPermission(),
			},
		},
	}
	_, err := gsmdrivelabels.BatchUpdateLabelPermissions(parent, "", updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update permissions on label, got error: %s", err))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *gdriveLabelPermissionResourceModel) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	state := &gdriveLabelPermissionResourceModel{}
	resp.Diagnostics.Append(req.State.Get(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	useAdminAccess := state.UseAdminAccess.ValueBool()
	deleteReq := &drivelabels.GoogleAppsDriveLabelsV2BatchDeleteLabelPermissionsRequest{
		UseAdminAccess: useAdminAccess,
		Requests: []*drivelabels.GoogleAppsDriveLabelsV2DeleteLabelPermissionRequest{
			{
				Name:           state.Name.ValueString(),
				UseAdminAccess: useAdminAccess,
			},
		},
	}
	gsmdrivelabels.BatchDeleteLabelPermissions(gsmhelpers.EnsurePrefix(state.Parent.ValueString(), "labels/"), deleteReq)
}

func (r *gdriveLabelPermissionResourceModel) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")
	if len(idParts) < 2 {
		resp.Diagnostics.AddError("Unexpected Import Identifier", fmt.Sprintf("Expected import identifier with format: 'useAdminAccess,name'. Got: %q", req.ID))
		return
	}
	for i := range idParts {
		if idParts[i] == "" {
			resp.Diagnostics.AddError("Unexpected Import Identifier", fmt.Sprintf("Expected import identifier with format: 'useAdminAccess,name'. Got: %q", req.ID))
			return
		}
	}
	useAdminAccess, err := strconv.ParseBool(idParts[0])
	if err != nil {
		resp.Diagnostics.AddError("Unexpected Import Identifier", fmt.Sprintf("Unable to parse '%s' as bool: %v", idParts[0], err))
		return
	}
	id := strings.Join(idParts[1:], "")
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("parent"), id[:strings.LastIndex(id, "/permissions/")])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("use_admin_access"), useAdminAccess)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
