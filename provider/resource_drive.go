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

type restrictionsModel struct {
	AdminManagedRestrictions     types.Bool `tfsdk:"admin_managed_restrictions"`
	CopyRequiresWriterPermission types.Bool `tfsdk:"copy_requires_writer_permission"`
	DomainUsersOnly              types.Bool `tfsdk:"domain_users_only"`
	DriveMembersOnly             types.Bool `tfsdk:"drive_members_only"`
}

// gdriveDriveResourceModelV1 describes the resource data model V1.
type gdriveDriveResourceModelV1 struct {
	Name                 types.String       `tfsdk:"name"`
	UseDomainAdminAccess types.Bool         `tfsdk:"use_domain_admin_access"`
	Id                   types.String       `tfsdk:"id"`
	WaitAfterCreate      types.Int64        `tfsdk:"wait_after_create"`
	Restrictions         *restrictionsModel `tfsdk:"restrictions"`
}

// gdriveDriveResourceModelV0 describes the resource data model V0.
type gdriveDriveResourceModelV0 struct {
	Name                 types.String         `tfsdk:"name"`
	UseDomainAdminAccess types.Bool           `tfsdk:"use_domain_admin_access"`
	Id                   types.String         `tfsdk:"id"`
	WaitAfterCreate      types.Int64          `tfsdk:"wait_after_create"`
	Restrictions         []*restrictionsModel `tfsdk:"restrictions"`
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
			"wait_after_create": schema.Int64Attribute{
				MarkdownDescription: `The Drive API returns a Shared Drive object immediately after creation, even though it is often not ready or visibile in other APIs.
In order to prevent 404 errors after the creation of a Shared Drive, the provider will wait the specified number of seconds after the creation of a Shared Drive and before returning or attempting further operations.
This value is only used for the initial creation and not used for updates. Changing this value after the initial creation has no effect.`,
				Optional:           true,
				DeprecationMessage: "Remove this attribute's configuration as it no longer is used and the attribute will be removed in the next major version of the provider.",
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
				dataV0 := &gdriveDriveResourceModelV0{}
				diags := req.State.Get(ctx, dataV0)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
				dataV1 := &gdriveDriveResourceModelV1{
					Name:                 dataV0.Name,
					UseDomainAdminAccess: dataV0.UseDomainAdminAccess,
					Id:                   dataV0.Id,
					WaitAfterCreate:      dataV0.WaitAfterCreate,
					Restrictions:         dataV0.Restrictions[0],
				}
				resp.Diagnostics.Append(resp.State.Set(ctx, dataV1)...)
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
	diags := req.Plan.Get(ctx, plan)
	resp.Diagnostics.Append(diags...)
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
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if plan.Restrictions != nil {
		driveReq = &drive.Drive{
			Restrictions: &drive.DriveRestrictions{},
		}
		if !plan.Restrictions.AdminManagedRestrictions.IsNull() {
			driveReq.Restrictions.AdminManagedRestrictions = plan.Restrictions.AdminManagedRestrictions.ValueBool()
			if !driveReq.Restrictions.AdminManagedRestrictions {
				driveReq.Restrictions.ForceSendFields = append(driveReq.Restrictions.ForceSendFields, "AdminManagedRestrictions")
			}
		}
		if !plan.Restrictions.CopyRequiresWriterPermission.IsNull() {
			driveReq.Restrictions.CopyRequiresWriterPermission = plan.Restrictions.CopyRequiresWriterPermission.ValueBool()
			if !driveReq.Restrictions.CopyRequiresWriterPermission {
				driveReq.Restrictions.ForceSendFields = append(driveReq.Restrictions.ForceSendFields, "CopyRequiresWriterPermission")
			}
		}
		if !plan.Restrictions.DomainUsersOnly.IsNull() {
			driveReq.Restrictions.DomainUsersOnly = plan.Restrictions.DomainUsersOnly.ValueBool()
			if !driveReq.Restrictions.DomainUsersOnly {
				driveReq.Restrictions.ForceSendFields = append(driveReq.Restrictions.ForceSendFields, "DomainUsersOnly")
			}
		}
		if !plan.Restrictions.DriveMembersOnly.IsNull() {
			driveReq.Restrictions.DriveMembersOnly = plan.Restrictions.DriveMembersOnly.ValueBool()
			if !driveReq.Restrictions.DriveMembersOnly {
				driveReq.Restrictions.ForceSendFields = append(driveReq.Restrictions.ForceSendFields, "DriveMembersOnly")
			}
		}
		_, err = gsmdrive.UpdateDrive(d.Id, fieldsDrive, plan.UseDomainAdminAccess.ValueBool(), driveReq)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set drive restrictions, got error: %s", err))
			return
		}
	}
}

func (r *gdriveDriveResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	plan := &gdriveDriveResourceModelV1{}
	diags := req.State.Get(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	d, err := gsmdrive.GetDrive(plan.Id.ValueString(), fieldsDrive, plan.UseDomainAdminAccess.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create drive, got error: %s", err))
		return
	}
	plan.Name = types.StringValue(d.Name)
	if d.Restrictions != nil && (d.Restrictions.AdminManagedRestrictions || d.Restrictions.CopyRequiresWriterPermission || d.Restrictions.DomainUsersOnly || d.Restrictions.DriveMembersOnly) {
		adminManagedRestrictions := types.BoolValue(d.Restrictions.AdminManagedRestrictions)
		copyRequiresWriterPermission := types.BoolValue(d.Restrictions.CopyRequiresWriterPermission)
		domainUsersOnly := types.BoolValue(d.Restrictions.DomainUsersOnly)
		driveMembersOnly := types.BoolValue(d.Restrictions.DriveMembersOnly)
		plan.Restrictions = &restrictionsModel{
			AdminManagedRestrictions:     adminManagedRestrictions,
			CopyRequiresWriterPermission: copyRequiresWriterPermission,
			DomainUsersOnly:              domainUsersOnly,
			DriveMembersOnly:             driveMembersOnly,
		}
	}
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *gdriveDriveResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := &gdriveDriveResourceModelV1{}
	resp.Diagnostics.Append(req.Plan.Get(ctx, plan)...)
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
		driveReq = &drive.Drive{
			Restrictions: &drive.DriveRestrictions{},
		}
		if !plan.Restrictions.AdminManagedRestrictions.IsNull() {
			driveReq.Restrictions.AdminManagedRestrictions = plan.Restrictions.AdminManagedRestrictions.ValueBool()
			if !driveReq.Restrictions.AdminManagedRestrictions {
				driveReq.Restrictions.ForceSendFields = append(driveReq.Restrictions.ForceSendFields, "AdminManagedRestrictions")
			}
		}
		if !plan.Restrictions.CopyRequiresWriterPermission.IsNull() {
			driveReq.Restrictions.CopyRequiresWriterPermission = plan.Restrictions.CopyRequiresWriterPermission.ValueBool()
			if !driveReq.Restrictions.CopyRequiresWriterPermission {
				driveReq.Restrictions.ForceSendFields = append(driveReq.Restrictions.ForceSendFields, "CopyRequiresWriterPermission")
			}
		}
		if !plan.Restrictions.DomainUsersOnly.IsNull() {
			driveReq.Restrictions.DomainUsersOnly = plan.Restrictions.DomainUsersOnly.ValueBool()
			if !driveReq.Restrictions.DomainUsersOnly {
				driveReq.Restrictions.ForceSendFields = append(driveReq.Restrictions.ForceSendFields, "DomainUsersOnly")
			}
		}
		if !plan.Restrictions.DriveMembersOnly.IsNull() {
			driveReq.Restrictions.DriveMembersOnly = plan.Restrictions.DriveMembersOnly.ValueBool()
			if !driveReq.Restrictions.DriveMembersOnly {
				driveReq.Restrictions.ForceSendFields = append(driveReq.Restrictions.ForceSendFields, "DriveMembersOnly")
			}
		}
	}
	_, err := gsmdrive.UpdateDrive(plan.Id.ValueString(), fieldsDrive, plan.UseDomainAdminAccess.ValueBool(), driveReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set drive restrictions, got error: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *gdriveDriveResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	plan := &gdriveDriveResourceModelV1{}
	resp.Diagnostics.Append(req.State.Get(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, err := gsmdrive.DeleteDrive(plan.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete drive, got error: %s", err))
		return
	}
}

func (r *gdriveDriveResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
