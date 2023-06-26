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
	"strings"

	"github.com/hanneshayashi/gsm/gsmcibeta"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &gdriveOrgUnitMembershipResource{}
var _ resource.ResourceWithImportState = &gdriveOrgUnitMembershipResource{}

const fieldsOrgUnitMembership = "parents,mimeType,driveId,name,id"

func newOrgUnitMembership() resource.Resource {
	return &gdriveOrgUnitMembershipResource{}
}

// gdriveOrgUnitMembershipResource defines the resource implementation.
type gdriveOrgUnitMembershipResource struct {
	client *http.Client
}

// gdriveOrgUnitMembershipResourceModel describes the resource data model.
type gdriveOrgUnitMembershipResourceModel struct {
	Parent    types.String `tfsdk:"parent"`
	Id        types.String `tfsdk:"id"`
	OrgUnitId types.String `tfsdk:"org_unit_id"`
	DriveId   types.String `tfsdk:"drive_id"`
}

func (r *gdriveOrgUnitMembershipResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_drive_ou_membership"
}

func (r *gdriveOrgUnitMembershipResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Sets the membership of a Shared Drive in an organizational unit.

The resource will move the Shared Drive to the specified OU in your Admin Console.

Some things to note:
* You need to specify the **ID** of the OU (**not the path**).
  * You can find the ID via the Admin SDK (or https://gsm.hayashi-ke.online/gsm/orgunits/list/).
* If you move the Shared Drive outside of Terraform, the resource will be re-created.
* A destroy of this resource will not do anything.`,
		Attributes: map[string]schema.Attribute{
			"id": rsId(),
			"org_unit_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the OrgUnit (OrgUnitId)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"drive_id": schema.StringAttribute{
				MarkdownDescription: "ID of the Shared Drive",
				Required:            true,
			},
			"parent": schema.StringAttribute{
				MarkdownDescription: "ID of the organizational unit (NOT the path!)",
				Required:            true,
			},
		},
	}
}

func (r *gdriveOrgUnitMembershipResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *gdriveOrgUnitMembershipResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := &gdriveOrgUnitMembershipResourceModel{}
	resp.Diagnostics.Append(req.Plan.Get(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(plan.move()...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *gdriveOrgUnitMembershipResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	state := &gdriveOrgUnitMembershipResourceModel{}
	resp.Diagnostics.Append(req.State.Get(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := state.Id.ValueString()
	found := false
	memberships, err := gsmcibeta.ListOrgUnitMemberships(id[0:strings.Index(id, "/memberships")], "customers/my_customer", "", "", 1)
	for i := range memberships {
		if i.Name == id {
			found = true
			break
		}
	}
	e := <-err
	if e != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list Org Unit memberships, got error: %s", e))
		return
	}
	if !found {
		state.Parent = types.StringUnknown()
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *gdriveOrgUnitMembershipResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := &gdriveOrgUnitMembershipResourceModel{}
	resp.Diagnostics.Append(req.Plan.Get(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(plan.move()...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *gdriveOrgUnitMembershipResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	return
}

func (r *gdriveOrgUnitMembershipResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
