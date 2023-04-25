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
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(60),
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
	data := &gdriveDriveResourceModelV1{}
	diags := req.Plan.Get(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	driveReq := &drive.Drive{
		Name: data.Name.ValueString(),
	}
	d, err := gsmdrive.CreateDrive(driveReq, "id,name,restrictions", false)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create drive, got error: %s", err))
		return
	}
	data.Id = types.StringValue(d.Id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if data.Restrictions != nil {
		driveReq = &drive.Drive{
			Restrictions: &drive.DriveRestrictions{},
		}
		if !data.Restrictions.AdminManagedRestrictions.IsNull() {
			driveReq.Restrictions.AdminManagedRestrictions = data.Restrictions.AdminManagedRestrictions.ValueBool()
			if !driveReq.Restrictions.AdminManagedRestrictions {
				driveReq.Restrictions.ForceSendFields = append(driveReq.Restrictions.ForceSendFields, "AdminManagedRestrictions")
			}
		}
		if !data.Restrictions.CopyRequiresWriterPermission.IsNull() {
			driveReq.Restrictions.CopyRequiresWriterPermission = data.Restrictions.CopyRequiresWriterPermission.ValueBool()
			if !driveReq.Restrictions.CopyRequiresWriterPermission {
				driveReq.Restrictions.ForceSendFields = append(driveReq.Restrictions.ForceSendFields, "CopyRequiresWriterPermission")
			}
		}
		if !data.Restrictions.DomainUsersOnly.IsNull() {
			driveReq.Restrictions.DomainUsersOnly = data.Restrictions.DomainUsersOnly.ValueBool()
			if !driveReq.Restrictions.DomainUsersOnly {
				driveReq.Restrictions.ForceSendFields = append(driveReq.Restrictions.ForceSendFields, "DomainUsersOnly")
			}
		}
		if !data.Restrictions.DriveMembersOnly.IsNull() {
			driveReq.Restrictions.DriveMembersOnly = data.Restrictions.DriveMembersOnly.ValueBool()
			if !driveReq.Restrictions.DriveMembersOnly {
				driveReq.Restrictions.ForceSendFields = append(driveReq.Restrictions.ForceSendFields, "DriveMembersOnly")
			}
		}
		d, err = gsmdrive.UpdateDrive(d.Id, "id,name,restrictions", false, driveReq)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set drive restrictions, got error: %s", err))
			return
		}
	}
}

func (r *gdriveDriveResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	data := &gdriveDriveResourceModelV1{}
	diags := req.State.Get(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	drive, err := gsmdrive.GetDrive(data.Id.ValueString(), "id,name,restrictions", data.UseDomainAdminAccess.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create drive, got error: %s", err))
		return
	}
	data.Name = types.StringValue(drive.Name)
	if drive.Restrictions != nil && (drive.Restrictions.AdminManagedRestrictions || drive.Restrictions.CopyRequiresWriterPermission || drive.Restrictions.DomainUsersOnly || drive.Restrictions.DriveMembersOnly) {
		adminManagedRestrictions := types.BoolValue(drive.Restrictions.AdminManagedRestrictions)
		copyRequiresWriterPermission := types.BoolValue(drive.Restrictions.CopyRequiresWriterPermission)
		domainUsersOnly := types.BoolValue(drive.Restrictions.DomainUsersOnly)
		driveMembersOnly := types.BoolValue(drive.Restrictions.DriveMembersOnly)
		data.Restrictions = &restrictionsModel{
			AdminManagedRestrictions:     adminManagedRestrictions,
			CopyRequiresWriterPermission: copyRequiresWriterPermission,
			DomainUsersOnly:              domainUsersOnly,
			DriveMembersOnly:             driveMembersOnly,
		}
	}
	diags = resp.State.Set(ctx, &data)
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
	_, err := gsmdrive.UpdateDrive(plan.Id.ValueString(), "id,name,restrictions", plan.UseDomainAdminAccess.ValueBool(), driveReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set drive restrictions, got error: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *gdriveDriveResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *gdriveDriveResourceModelV1

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete example, got error: %s", err))
	//     return
	// }
}

func (r *gdriveDriveResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// func resourceDrive() *schema.Resource {
// 	return &schema.Resource{
// 		Description: "Creates a Shared Drive",
// 		Schema: map[string]*schema.Schema{
// 			"restrictions": {
// 				Type:     schema.TypeList,
// 				Optional: true,
// 				DiffSuppressFunc: func(_, oldValue, newValue string, _ *schema.ResourceData) bool {
// 					if (oldValue == "true" && newValue == "false") || (oldValue == "false" && newValue == "true") || (oldValue == "" && newValue != "") || (oldValue == "0" && newValue == "1") {
// 						return false
// 					}
// 					return true
// 				},
// 				Description: "The restrictions that should be set on the Shared Drive",
// 				MaxItems:    1,
// 				Elem: &schema.Resource{
// 					Schema: map[string]*schema.Schema{
// 						"admin_managed_restrictions": {
// 							Type:        schema.TypeBool,
// 							Optional:    true,
// 							Description: "Whether administrative privileges on this Shared Drive are required to modify restrictions",
// 						},
// 						"copy_requires_writer_permission": {
// 							Type:     schema.TypeBool,
// 							Optional: true,
// 							Description: `Whether the options to copy, print, or download files inside this Shared Drive, should be disabled for readers and commenters.
// When this restriction is set to true, it will override the similarly named field to true for any file inside this Shared Drive`,
// 						},
// 						"domain_users_only": {
// 							Type:     schema.TypeBool,
// 							Optional: true,
// 							Description: `Whether access to this Shared Drive and items inside this Shared Drive is restricted to users of the domain to which this Shared Drive belongs.
// This restriction may be overridden by other sharing policies controlled outside of this Shared Drive`,
// 						},
// 						"drive_members_only": {
// 							Type:        schema.TypeBool,
// 							Optional:    true,
// 							Description: "Whether access to items inside this Shared Drive is restricted to its members",
// 						},
// 					},
// 				},
// 			},
// 		},
// 		Create: resourceCreateDrive,
// 		Read:   resourceReadDrive,
// 		Update: resourceUpdateDrive,
// 		Delete: resourceDeleteDrive,
// 		Exists: resourceExistsDrive,
// 		Importer: &schema.ResourceImporter{
// 			StateContext: schema.ImportStatePassthroughContext,
// 		},
// 	}
// }

// func dataToDrive(d *schema.ResourceData, update bool) (*drive.Drive, error) {
// 	newDrive := &drive.Drive{}
// 	if d.HasChange("name") {
// 		newDrive.Name = d.Get("name").(string)
// 		if newDrive.Name == "" {
// 			newDrive.ForceSendFields = append(newDrive.ForceSendFields, "Name")
// 		}
// 	}
// 	if update {
// 		if d.HasChange("restrictions") {
// 			newDrive.Restrictions = &drive.DriveRestrictions{}
// 			restrictions := d.Get("restrictions").([]any)
// 			if len(restrictions) > 0 {
// 				if d.HasChange("restrictions.0.admin_managed_restrictions") {
// 					newDrive.Restrictions.AdminManagedRestrictions = d.Get("restrictions.0.admin_managed_restrictions").(bool)
// 					if !newDrive.Restrictions.AdminManagedRestrictions {
// 						newDrive.Restrictions.ForceSendFields = append(newDrive.Restrictions.ForceSendFields, "AdminManagedRestrictions")
// 					}
// 				}
// 				if d.HasChange("restrictions.0.copy_requires_writer_permission") {
// 					newDrive.Restrictions.CopyRequiresWriterPermission = d.Get("restrictions.0.copy_requires_writer_permission").(bool)
// 					if !newDrive.Restrictions.CopyRequiresWriterPermission {
// 						newDrive.Restrictions.ForceSendFields = append(newDrive.Restrictions.ForceSendFields, "CopyRequiresWriterPermission")
// 					}
// 				}
// 				if d.HasChange("restrictions.0.domain_users_only") {
// 					newDrive.Restrictions.DomainUsersOnly = d.Get("restrictions.0.domain_users_only").(bool)
// 					if !newDrive.Restrictions.DomainUsersOnly {
// 						newDrive.Restrictions.ForceSendFields = append(newDrive.Restrictions.ForceSendFields, "DomainUsersOnly")
// 					}
// 				}
// 				if d.HasChange("restrictions.0.drive_members_only") {
// 					newDrive.Restrictions.DriveMembersOnly = d.Get("restrictions.0.drive_members_only").(bool)
// 					if !newDrive.Restrictions.DriveMembersOnly {
// 						newDrive.Restrictions.ForceSendFields = append(newDrive.Restrictions.ForceSendFields, "DriveMembersOnly")
// 					}
// 				}
// 			} else {
// 				newDrive.Restrictions.ForceSendFields = append(newDrive.Restrictions.ForceSendFields, "AdminManagedRestrictions", "CopyRequiresWriterPermission", "DomainUsersOnly", "DriveMembersOnly")
// 			}
// 		}
// 	}
// 	return newDrive, nil
// }

// func resourceCreateDrive(d *schema.ResourceData, _ any) error {
// 	var driveResult *drive.Drive
// 	var err error
// 	driveRequest, err := dataToDrive(d, false)
// 	if err != nil {
// 		return err
// 	}
// 	driveResult, err = gsmdrive.CreateDrive(driveRequest, "*", true)
// 	if err != nil {
// 		return err
// 	}
// 	d.SetId(driveResult.Id)
// 	time.Sleep(time.Duration(d.Get("wait_after_create").(int)) * time.Second)
// 	if d.HasChange("restrictions") {
// 		return resourceUpdateDrive(d, nil)
// 	}
// 	return resourceReadDrive(d, nil)
// }

// func resourceReadDrive(d *schema.ResourceData, _ any) error {
// 	r, err := gsmdrive.GetDrive(d.Id(), "*", d.Get("use_domain_admin_access").(bool))
// 	if err != nil {
// 		return err
// 	}
// 	d.Set("name", r.Name)
// 	restrictions := []map[string]bool{
// 		{
// 			"admin_managed_restrictions":      r.Restrictions.AdminManagedRestrictions,
// 			"copy_requires_writer_permission": r.Restrictions.CopyRequiresWriterPermission,
// 			"domain_users_only":               r.Restrictions.DomainUsersOnly,
// 			"drive_members_only":              r.Restrictions.DriveMembersOnly,
// 		},
// 	}
// 	d.Set("restrictions", restrictions)
// 	return nil
// }

// func resourceUpdateDrive(d *schema.ResourceData, _ any) error {
// 	driveRequest, err := dataToDrive(d, true)
// 	if err != nil {
// 		return err
// 	}
// 	_, err = gsmdrive.UpdateDrive(d.Id(), "id", d.Get("use_domain_admin_access").(bool), driveRequest)
// 	if err != nil {
// 		return err
// 	}
// 	return resourceReadDrive(d, nil)
// }

// func resourceDeleteDrive(d *schema.ResourceData, _ any) error {
// 	_, err := gsmdrive.DeleteDrive(d.Id())
// 	return err
// }

// func resourceExistsDrive(d *schema.ResourceData, _ any) (bool, error) {
// 	_, err := gsmdrive.GetDrive(d.Id(), "id", d.Get("use_domain_admin_access").(bool))
// 	if err != nil {
// 		return false, err
// 	}
// 	return true, nil
// }
