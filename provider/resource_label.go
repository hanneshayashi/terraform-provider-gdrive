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
	"reflect"

	"github.com/hanneshayashi/gsm/gsmdrivelabels"
	"github.com/hanneshayashi/gsm/gsmhelpers"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"google.golang.org/api/drivelabels/v2"
)

const adminAttributeLabels = "use_admin_access"

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &gdriveLabelResource{}
var _ resource.ResourceWithImportState = &gdriveLabelResource{}

func newLabel() resource.Resource {
	return &gdriveLabelResource{}
}

// gdriveLabelResource defines the resource implementation.
type gdriveLabelResource struct {
	client *http.Client
}

func (propertiesModel *gdriveLabelResourcePropertiesModel) populate(properties *drivelabels.GoogleAppsDriveLabelsV2LabelProperties) {
	if properties != nil {
		if !(propertiesModel.Title.IsNull() && properties.Title == "") {
			propertiesModel.Title = types.StringValue(properties.Title)
		}
		if !(propertiesModel.Description.IsNull() && properties.Description == "") {
			propertiesModel.Description = types.StringValue(properties.Description)
		}
	}
}

type gdriveLabelResourcePropertiesModel struct {
	Title       types.String `tfsdk:"title"`
	Description types.String `tfsdk:"description"`
}

// gdriveLabelResourceModel describes the resource data model.
type gdriveLabelResourceModel struct {
	Properties     *gdriveLabelResourcePropertiesModel `tfsdk:"properties"`
	LifeCycle      *gdriveLabelLifeCycleModel          `tfsdk:"life_cycle"`
	Id             types.String                        `tfsdk:"id"`
	LabelId        types.String                        `tfsdk:"label_id"`
	Name           types.String                        `tfsdk:"name"`
	LanguageCode   types.String                        `tfsdk:"language_code"`
	LabelType      types.String                        `tfsdk:"label_type"`
	UseAdminAccess types.Bool                          `tfsdk:"use_admin_access"`
}

func (r *gdriveLabelResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_label"
}

func (r *gdriveLabelResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Creates a Drive Label.

The label must be published before it can be assigned to files. This is controlled via the 'life_cycle'
property.

Publishing can only be done via the label resource, NOT the field resources.

If a label is changed via an "external" resource (e.g., a field), those changes must be published
via the label resource, before they are available for files.

This means that, if you have labels and fields in the same Terraform configuration and you make changes
to the fields you may have to apply twice in order to
1. Apply the changes to the fields.
2. Publish the changes via the label.

A label must be deactivated before it can be deleted.`,
		Attributes: map[string]schema.Attribute{
			"id": rsId(),
			"label_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the label.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Resource name of the label.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"use_admin_access": schema.BoolAttribute{
				Optional: true,
				MarkdownDescription: `Set to true in order to use the user's admin credentials.

The server verifies that the user is an admin for the label before allowing access.`,
			},
			"language_code": schema.StringAttribute{
				MarkdownDescription: `The BCP-47 language code to use for evaluating localized field labels.

When not specified, values in the default configured language are used.`,
				Optional: true,
			},
			"label_type": schema.StringAttribute{
				MarkdownDescription: `The type of this label.

The following values are accepted:
* "SHARED"  - Shared labels may be shared with users to apply to Drive items.
* "ADMIN"   - Admin-owned label. Only creatable and editable by admins. Supports some additional admin-only features.`,
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"life_cycle": lifeCycleRS(),
			"properties": schema.SingleNestedBlock{
				MarkdownDescription: "Basic properties of the label.",
				Attributes: map[string]schema.Attribute{
					"title": schema.StringAttribute{
						MarkdownDescription: "Title of the label.",
						Required:            true,
					},
					"description": schema.StringAttribute{
						MarkdownDescription: "The description of the label.",
						Optional:            true,
					},
				},
			},
		},
	}
}

func (r *gdriveLabelResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *gdriveLabelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := &gdriveLabelResourceModel{}
	resp.Diagnostics.Append(req.Plan.Get(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	labelReq := &drivelabels.GoogleAppsDriveLabelsV2Label{
		LabelType: plan.LabelType.ValueString(),
		Properties: &drivelabels.GoogleAppsDriveLabelsV2LabelProperties{
			Title:       plan.Properties.Title.ValueString(),
			Description: plan.Properties.Description.ValueString(),
		},
	}

	languageCode := plan.LanguageCode.ValueString()
	useAdminAccess := plan.UseAdminAccess.ValueBool()
	l, err := gsmdrivelabels.CreateLabel(labelReq, languageCode, "*", useAdminAccess)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create label, got error: %s", err))
		return
	}
	plan.Name = types.StringValue(l.Name)
	plan.LabelId = types.StringValue(l.Id)
	plan.Id = plan.LabelId
	var lifecycle *drivelabels.GoogleAppsDriveLabelsV2Lifecycle
	if plan.LifeCycle != nil {
		lifecycle = plan.LifeCycle.toLifecycle()
		if plan.LifeCycle.DisabledPolicy != nil {
			labelReq.Lifecycle = &drivelabels.GoogleAppsDriveLabelsV2Lifecycle{
				DisabledPolicy: lifecycle.DisabledPolicy,
			}
		}
	}
	if plan.LifeCycle != nil {
		if lifecycle.State == "PUBLISHED" || lifecycle.State == "DISABLED" {
			publishReq := &drivelabels.GoogleAppsDriveLabelsV2PublishLabelRequest{
				LanguageCode:   languageCode,
				UseAdminAccess: useAdminAccess,
			}
			l, err = gsmdrivelabels.Publish(l.Name, "*", publishReq)
			if err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to publish label, got error: %s", err))
				return
			}
		}
		if lifecycle.State == "DISABLED" {
			disableReq := &drivelabels.GoogleAppsDriveLabelsV2DisableLabelRequest{
				LanguageCode:   languageCode,
				UseAdminAccess: useAdminAccess,
				DisabledPolicy: lifecycle.DisabledPolicy,
			}
			l, err = gsmdrivelabels.Disable(l.Name, "*", disableReq)
			if err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to disable label, got error: %s", err))
				return
			}
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *gdriveLabelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	state := &gdriveLabelResourceModel{}
	resp.Diagnostics.Append(req.State.Get(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(state.populate(ctx)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *gdriveLabelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := &gdriveLabelResourceModel{}
	resp.Diagnostics.Append(req.Plan.Get(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state := &gdriveLabelResourceModel{}
	resp.Diagnostics.Append(req.State.Get(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	labelId := gsmhelpers.EnsurePrefix(plan.LabelId.ValueString(), "labels/")
	updateLabelRequest := newUpdateLabelRequest(plan)
	propertiesChanged := false
	if !plan.Properties.Description.Equal(state.Properties.Description) || !plan.Properties.Title.Equal(state.Properties.Title) {
		req := &drivelabels.GoogleAppsDriveLabelsV2DeltaUpdateLabelRequestRequest{
			UpdateLabel: &drivelabels.GoogleAppsDriveLabelsV2DeltaUpdateLabelRequestUpdateLabelPropertiesRequest{
				Properties: &drivelabels.GoogleAppsDriveLabelsV2LabelProperties{
					Description: plan.Properties.Description.ValueString(),
					Title:       plan.Properties.Title.ValueString(),
				},
			},
		}
		if req.UpdateLabel.Properties.Description == "" {
			req.UpdateLabel.Properties.ForceSendFields = append(req.UpdateLabel.Properties.ForceSendFields, "Description")
		}
		if req.UpdateLabel.Properties.Title == "" {
			req.UpdateLabel.Properties.ForceSendFields = append(req.UpdateLabel.Properties.ForceSendFields, "Title")
		}
		updateLabelRequest.Requests = append(updateLabelRequest.Requests, req)
		_, err := gsmdrivelabels.Delta(labelId, "*", updateLabelRequest)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update label, got error: %s", err))
			return
		}
		propertiesChanged = true
	}
	var lifecyclePlan *drivelabels.GoogleAppsDriveLabelsV2Lifecycle
	var lifecycleState *drivelabels.GoogleAppsDriveLabelsV2Lifecycle
	if plan.LifeCycle != nil {
		lifecyclePlan = plan.LifeCycle.toLifecycle()
	}
	if state.LifeCycle != nil {
		lifecycleState = state.LifeCycle.toLifecycle()
	}
	lifeCycleChanged := lifecyclePlan != nil && !reflect.DeepEqual(lifecyclePlan, lifecycleState)
	if lifeCycleChanged && lifecycleState != nil && lifecyclePlan.State == "PUBLISHED" && lifecycleState.State == "DISABLED" {
		enableReq := &drivelabels.GoogleAppsDriveLabelsV2EnableLabelRequest{
			LanguageCode:   plan.getLanguageCode(),
			UseAdminAccess: plan.getUseAdminAccess(),
		}
		_, err := gsmdrivelabels.Enable(labelId, "*", enableReq)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to enable label, got error: %s", err))
			return
		}
	}
	if (lifeCycleChanged || propertiesChanged) && lifecyclePlan != nil && lifecyclePlan.State == "PUBLISHED" {
		publishReq := &drivelabels.GoogleAppsDriveLabelsV2PublishLabelRequest{
			LanguageCode:   plan.getLanguageCode(),
			UseAdminAccess: plan.getUseAdminAccess(),
		}
		_, err := gsmdrivelabels.Publish(labelId, "*", publishReq)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to publish label, got error: %s", err))
			return
		}
	}
	if lifeCycleChanged && lifecyclePlan.State == "DISABLED" {
		disable := &drivelabels.GoogleAppsDriveLabelsV2DisableLabelRequest{
			LanguageCode:   plan.getLanguageCode(),
			UseAdminAccess: plan.getUseAdminAccess(),
			DisabledPolicy: lifecyclePlan.DisabledPolicy,
		}
		_, err := gsmdrivelabels.Disable(labelId, "*", disable)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to disable label, got error: %s", err))
			return
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *gdriveLabelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	state := &gdriveLabelResourceModel{}
	resp.Diagnostics.Append(req.State.Get(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, err := gsmdrivelabels.DeleteLabel(gsmhelpers.EnsurePrefix(state.LabelId.ValueString(), "labels/"), "", state.UseAdminAccess.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete label, got error: %s", err))
		return
	}
}

func (r *gdriveLabelResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(importSplitId(ctx, req, resp, adminAttributeLabels, "label_id")...)
}
