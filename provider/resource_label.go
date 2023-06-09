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

	"github.com/hanneshayashi/gsm/gsmdrivelabels"
	"github.com/hanneshayashi/gsm/gsmhelpers"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"google.golang.org/api/drivelabels/v2"
)

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

type gdriveLabelResourcePropertiesRSModel struct {
	Title       types.String `tfsdk:"title"`
	Description types.String `tfsdk:"description"`
}

// gdriveLabelResourceModel describes the resource data model.
type gdriveLabelResourceModel struct {
	Properties     *gdriveLabelResourcePropertiesRSModel `tfsdk:"properties"`
	Id             types.String                          `tfsdk:"id"`
	LabelId        types.String                          `tfsdk:"label_id"`
	Name           types.String                          `tfsdk:"name"`
	LanguageCode   types.String                          `tfsdk:"language_code"`
	LabelType      types.String                          `tfsdk:"label_type"`
	UseAdminAccess types.Bool                            `tfsdk:"use_admin_access"`
}

func (r *gdriveLabelResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_label"
}

func (r *gdriveLabelResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Creates a Drive Label",
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
				Description: `Set to true in order to use the user's admin credentials.
The server verifies that the user is an admin for the label before allowing access.
Requires setting the 'use_labels_admin_scope' property to 'true' in the provider config.`,
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
	l, err := gsmdrivelabels.CreateLabel(labelReq, plan.LanguageCode.ValueString(), "*", plan.UseAdminAccess.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create label, got error: %s", err))
		return
	}
	plan.Name = types.StringValue(l.Name)
	plan.LabelId = types.StringValue(l.Id)
	plan.Id = plan.LabelId
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
	updateLabelRequest := newUpdateLabelRequest(plan)
	if plan.Properties != nil && state.Properties != nil && (!plan.Properties.Description.Equal(state.Properties.Description) || !plan.Properties.Title.Equal(state.Properties.Title)) {
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
		_, err := gsmdrivelabels.Delta(gsmhelpers.EnsurePrefix(plan.Id.ValueString(), "labels/"), "*", updateLabelRequest)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update label, got error: %s", err))
			return
		}
	}
	// fieldsPlanM := make(map[string]*gdriveLabelResourceFieldsModel)
	// fieldsStateM := make(map[string]*gdriveLabelResourceFieldsModel)
	// for i := range plan.Field {
	// 	tflog.Debug(ctx, fmt.Sprintf("planFieldXXX: %s", plan.Field[i].Id))
	// 	if plan.Field[i].Id.IsUnknown() {
	// 		req := &drivelabels.GoogleAppsDriveLabelsV2DeltaUpdateLabelRequestRequest{
	// 			CreateField: &drivelabels.GoogleAppsDriveLabelsV2DeltaUpdateLabelRequestCreateFieldRequest{
	// 				Field: plan.Field[i].toField(),
	// 			},
	// 		}
	// 		createFieldsReq := newUpdateLabelRequest(plan)
	// 		createFieldsReq.Requests = append(createFieldsReq.Requests, req)
	// 		updatedLabel, err := gsmdrivelabels.Delta(gsmhelpers.EnsurePrefix(labelId, "labels/"), "*", createFieldsReq)
	// 		if err != nil {
	// 			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update label (create new fields), got error: %s", err))
	// 			return
	// 		}
	// 		newField := (updatedLabel.UpdatedLabel.Fields[len(updatedLabel.UpdatedLabel.Fields)-1])
	// 		plan.Field[i].Id = types.StringValue(newField.Id)
	// 		plan.Field[i].FieldId = plan.Field[i].Id
	// 		plan.Field[i].QueryKey = types.StringValue(newField.QueryKey)
	// 	} else {
	// 		fieldsPlanM[plan.Field[i].Id.ValueString()] = plan.Field[i]
	// 	}
	// }
	// deleteOldFieldsRequest := newUpdateLabelRequest(plan)
	// for i := range state.Field {
	// 	fieldsStateM[state.Field[i].FieldId.ValueString()] = state.Field[i]
	// 	tflog.Debug(ctx, fmt.Sprintf("stateFieldXXX: %s", state.Field[i].Id))
	// 	fieldId := state.Field[i].FieldId.ValueString()
	// 	_, ok := fieldsPlanM[fieldId]
	// 	if !ok {
	// 		req := &drivelabels.GoogleAppsDriveLabelsV2DeltaUpdateLabelRequestRequest{
	// 			DeleteField: &drivelabels.GoogleAppsDriveLabelsV2DeltaUpdateLabelRequestDeleteFieldRequest{
	// 				Id: fieldId,
	// 			},
	// 		}
	// 		deleteOldFieldsRequest.Requests = append(deleteOldFieldsRequest.Requests, req)
	// 	}
	// }
	// if len(deleteOldFieldsRequest.Requests) > 0 {
	// 	_, err := gsmdrivelabels.Delta(gsmhelpers.EnsurePrefix(labelId, "labels/"), "*", deleteOldFieldsRequest)
	// 	if err != nil {
	// 		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update label (delete old fields), got error: %s", err))
	// 		return
	// 	}
	// }
	// modFieldsRequest := newUpdateLabelRequest(plan)
	// for i := range fieldsPlanM {
	// 	if fieldsPlanM[i].DateOptions != nil && fieldsStateM[i].DateOptions != nil {
	// 		if !fieldsPlanM[i].DateOptions.DateFormatType.Equal(fieldsStateM[i].DateOptions.DateFormatType) {
	// 			modFieldsRequest.Requests = append(modFieldsRequest.Requests, &drivelabels.GoogleAppsDriveLabelsV2DeltaUpdateLabelRequestRequest{
	// 				UpdateFieldType: &drivelabels.GoogleAppsDriveLabelsV2DeltaUpdateLabelRequestUpdateFieldTypeRequest{
	// 					Id: fieldsPlanM[i].FieldId.ValueString(),
	// 					DateOptions: &drivelabels.GoogleAppsDriveLabelsV2FieldDateOptions{
	// 						DateFormatType: fieldsPlanM[i].DateOptions.DateFormatType.ValueString(),
	// 					},
	// 				},
	// 			})
	// 		}
	// 	} else if fieldsPlanM[i].UserOptions != nil && fieldsPlanM[i].UserOptions.ListOptions != nil && fieldsStateM[i].UserOptions.ListOptions != nil {
	// 		if fieldsPlanM[i].UserOptions.ListOptions.MaxEntries.Equal(fieldsStateM[i].UserOptions.ListOptions.MaxEntries) {

	// 		}
	// 	}
	// }
	// fixOrderReq := newUpdateLabelRequest(plan)
	// for i := range plan.Field {
	// 	req := &drivelabels.GoogleAppsDriveLabelsV2DeltaUpdateLabelRequestRequest{
	// 		UpdateField: &drivelabels.GoogleAppsDriveLabelsV2DeltaUpdateLabelRequestUpdateFieldPropertiesRequest{
	// 			Id: plan.Field[i].Id.ValueString(),
	// 			Properties: &drivelabels.GoogleAppsDriveLabelsV2FieldProperties{
	// 				DisplayName: plan.Field[i].Properties.DisplayName.ValueString(),
	// 				Required:    plan.Field[i].Properties.Required.ValueBool(),
	// 			},
	// 		},
	// 	}
	// 	if i < len(plan.Field)-1 {
	// 		req.UpdateField.Properties.InsertBeforeField = plan.Field[i+1].Id.ValueString()
	// 	}
	// 	fixOrderReq.Requests = append(fixOrderReq.Requests, req)
	// }
	// if len(fixOrderReq.Requests) > 0 {
	// 	_, err := gsmdrivelabels.Delta(gsmhelpers.EnsurePrefix(labelId, "labels/"), "*", fixOrderReq)
	// 	if err != nil {
	// 		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update label (fix field order), got error: %s", err))
	// 		return
	// 	}
	// }
	// resp.Diagnostics.Append(plan.populate(ctx)...)
	// if resp.Diagnostics.HasError() {
	// 	return
	// }
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *gdriveLabelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	state := &gdriveLabelResourceModel{}
	resp.Diagnostics.Append(req.State.Get(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, err := gsmdrivelabels.DeleteLabel(gsmhelpers.EnsurePrefix(state.Id.ValueString(), "labels/"), "", state.UseAdminAccess.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete label, got error: %s", err))
		return
	}
}

func (r *gdriveLabelResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
