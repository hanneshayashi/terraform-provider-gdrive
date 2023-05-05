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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"google.golang.org/api/drive/v3"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &gdriveDriveResource{}
var _ resource.ResourceWithImportState = &gdriveDriveResource{}

const fieldsDrive = "id,name,restrictions"

func newDrive() resource.Resource {
	return &gdriveDriveResource{}
}

// gdriveDriveResource defines the resource implementation.
type gdriveDriveResource struct {
	client *http.Client
}

type driveRestrictionsModel struct {
	AdminManagedRestrictions     types.Bool `tfsdk:"admin_managed_restrictions"`
	CopyRequiresWriterPermission types.Bool `tfsdk:"copy_requires_writer_permission"`
	DomainUsersOnly              types.Bool `tfsdk:"domain_users_only"`
	DriveMembersOnly             types.Bool `tfsdk:"drive_members_only"`
}

// gdriveDriveResourceModelV1 describes the resource data model V1.
type gdriveDriveResourceModelV1 struct {
	Name                 types.String            `tfsdk:"name"`
	UseDomainAdminAccess types.Bool              `tfsdk:"use_domain_admin_access"`
	DriveId              types.String            `tfsdk:"drive_id"`
	Id                   types.String            `tfsdk:"id"`
	Restrictions         *driveRestrictionsModel `tfsdk:"restrictions"`
}

// gdriveDriveResourceModelV0 describes the resource data model V0.
type gdriveDriveResourceModelV0 struct {
	Name                 types.String              `tfsdk:"name"`
	UseDomainAdminAccess types.Bool                `tfsdk:"use_domain_admin_access"`
	Id                   types.String              `tfsdk:"id"`
	WaitAfterCreate      types.Int64               `tfsdk:"wait_after_create"`
	Restrictions         []*driveRestrictionsModel `tfsdk:"restrictions"`
}

func (r *gdriveDriveResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_drive"
}

func (r *gdriveDriveResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             1,
		MarkdownDescription: "Creates a Shared Drive",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the Shared Drive",
				Required:            true,
			},
			"use_domain_admin_access": schema.BoolAttribute{
				MarkdownDescription: "Use domain admin access",
				Optional:            true,
			},
			"drive_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the Shared Drive (driveId)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the Shared Drive (driveId)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"restrictions": schema.SingleNestedBlock{
				MarkdownDescription: "The restrictions that should be set on the Shared Drive",
				Attributes: map[string]schema.Attribute{
					"admin_managed_restrictions": schema.BoolAttribute{
						MarkdownDescription: "Whether administrative privileges on this Shared Drive are required to modify restrictions",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
					},
					"copy_requires_writer_permission": schema.BoolAttribute{
						MarkdownDescription: `Whether the options to copy, print, or download files inside this Shared Drive, should be disabled for readers and commenters.
When this restriction is set to true, it will override the similarly named field to true for any file inside this Shared Drive`,
						Optional: true,
						Computed: true,
						Default:  booldefault.StaticBool(false),
					},
					"domain_users_only": schema.BoolAttribute{
						MarkdownDescription: `Whether access to this Shared Drive and items inside this Shared Drive is restricted to users of the domain to which this Shared Drive belongs.
This restriction may be overridden by other sharing policies controlled outside of this Shared Drive`,
						Optional: true,
						Computed: true,
						Default:  booldefault.StaticBool(false),
					},
					"drive_members_only": schema.BoolAttribute{
						MarkdownDescription: "Whether access to items inside this Shared Drive is restricted to its members",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
					},
				},
			},
		},
	}
}

func (r *gdriveDriveResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: &schema.Schema{
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						Required: true,
					},
					"use_domain_admin_access": schema.BoolAttribute{
						Optional: true,
					},
					"wait_after_create": schema.Int64Attribute{
						Optional: true,
						Computed: true,
						Default:  int64default.StaticInt64(60),
					},
					"id": schema.StringAttribute{
						Computed: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"restrictions": schema.ListNestedAttribute{
						Optional: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"admin_managed_restrictions": schema.BoolAttribute{
									MarkdownDescription: "Whether administrative privileges on this Shared Drive are required to modify restrictions",
									Optional:            true,
								},
								"copy_requires_writer_permission": schema.BoolAttribute{
									MarkdownDescription: `Whether the options to copy, print, or download files inside this Shared Drive, should be disabled for readers and commenters.
When this restriction is set to true, it will override the similarly named field to true for any file inside this Shared Drive`,
									Optional: true,
								},
								"domain_users_only": schema.BoolAttribute{
									MarkdownDescription: `Whether access to this Shared Drive and items inside this Shared Drive is restricted to users of the domain to which this Shared Drive belongs.
This restriction may be overridden by other sharing policies controlled outside of this Shared Drive`,
									Optional: true,
								},
								"drive_members_only": schema.BoolAttribute{
									MarkdownDescription: "Whether access to items inside this Shared Drive is restricted to its members",
									Optional:            true,
								},
							},
						},
					},
				},
			},
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				stateV0 := &gdriveDriveResourceModelV0{}
				resp.Diagnostics.Append(req.State.Get(ctx, stateV0)...)
				if resp.Diagnostics.HasError() {
					return
				}
				stateV1 := &gdriveDriveResourceModelV1{
					Name:                 stateV0.Name,
					UseDomainAdminAccess: stateV0.UseDomainAdminAccess,
					Id:                   stateV0.Id,
					DriveId:              stateV0.Id,
					Restrictions:         stateV0.Restrictions[0],
				}
				resp.Diagnostics.Append(resp.State.Set(ctx, stateV1)...)
			},
		},
	}
}

func (r *gdriveDriveResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *gdriveDriveResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := &gdriveDriveResourceModelV1{}
	resp.Diagnostics.Append(req.Plan.Get(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	driveReq := &drive.Drive{
		Name: plan.Name.ValueString(),
	}
	d, err := gsmdrive.CreateDrive(driveReq, fieldsDrive, false)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create drive, got error: %s", err))
		return
	}
	plan.Id = types.StringValue(d.Id)
	plan.DriveId = types.StringValue(d.Id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if plan.Restrictions != nil {
		driveReq = &drive.Drive{
			Restrictions: plan.Restrictions.toDriveRestrictions(),
		}
		_, err = gsmdrive.UpdateDrive(d.Id, fieldsDrive, plan.UseDomainAdminAccess.ValueBool(), driveReq)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set drive restrictions, got error: %s", err))
			return
		}
	}
}

func (r *gdriveDriveResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	state := &gdriveDriveResourceModelV1{}
	resp.Diagnostics.Append(req.State.Get(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.DriveId = state.Id
	resp.Diagnostics.Append(state.populate()...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *gdriveDriveResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := &gdriveDriveResourceModelV1{}
	resp.Diagnostics.Append(req.Plan.Get(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state := &gdriveDriveResourceModelV1{}
	resp.Diagnostics.Append(req.State.Get(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	driveReq := &drive.Drive{}
	if !plan.Name.Equal(state.Name) {
		driveReq.Name = plan.Name.ValueString()
	}
	if plan.Restrictions != nil {
		driveReq.Restrictions = plan.Restrictions.toDriveRestrictions()
	}
	_, err := gsmdrive.UpdateDrive(plan.Id.ValueString(), fieldsDrive, plan.UseDomainAdminAccess.ValueBool(), driveReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update drive, got error: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *gdriveDriveResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	state := &gdriveDriveResourceModelV1{}
	resp.Diagnostics.Append(req.State.Get(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, err := gsmdrive.DeleteDrive(state.Id.ValueString(), state.UseDomainAdminAccess.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete drive, got error: %s", err))
		return
	}
}

func (r *gdriveDriveResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
