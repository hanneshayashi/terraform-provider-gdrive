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

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &gdriveLabelDateFieldResource{}
var _ resource.ResourceWithImportState = &gdriveLabelDateFieldResource{}

func newLabelDateField() resource.Resource {
	return &gdriveLabelDateFieldResource{}
}

// gdriveLabelDateFieldResource defines the resource implementation.
type gdriveLabelDateFieldResource struct {
	client *http.Client
}

type gdriveLabelDateOptionsRSModel struct {
	DateFormatType types.String `tfsdk:"date_format_type"`
}

type gdriveLabelDateFieldResourceModel struct {
	LifeCycle      *gdriveLabelLifeCycleModel        `tfsdk:"life_cycle"`
	Properties     *gdriveLabelFieldPropertieseModel `tfsdk:"properties"`
	DateOptions    *gdriveLabelDateOptionsRSModel    `tfsdk:"date_options"`
	Id             types.String                      `tfsdk:"id"`
	FieldId        types.String                      `tfsdk:"field_id"`
	LabelId        types.String                      `tfsdk:"label_id"`
	QueryKey       types.String                      `tfsdk:"query_key"`
	LanguageCode   types.String                      `tfsdk:"language_code"`
	UseAdminAccess types.Bool                        `tfsdk:"use_admin_access"`
}

func (fieldModel *gdriveLabelDateFieldResourceModel) toField() (field *drivelabels.GoogleAppsDriveLabelsV2Field) {
	field = &drivelabels.GoogleAppsDriveLabelsV2Field{
		DateOptions: &drivelabels.GoogleAppsDriveLabelsV2FieldDateOptions{},
	}
	if fieldModel.DateOptions != nil {
		field.DateOptions = fieldModel.DateOptions.toDateOptions()
	}
	if fieldModel.LifeCycle != nil {
		field.Lifecycle = fieldModel.LifeCycle.toLifecycle()
	}
	if fieldModel.Properties != nil {
		field.Properties = fieldModel.Properties.toProperties()
	}
	return
}

func (dateOptionsModel *gdriveLabelDateOptionsRSModel) toDateOptions() (dateOptions *drivelabels.GoogleAppsDriveLabelsV2FieldDateOptions) {
	dateOptions = &drivelabels.GoogleAppsDriveLabelsV2FieldDateOptions{
		DateFormatType: dateOptionsModel.DateFormatType.ValueString(),
	}
	if dateOptions.DateFormatType == "" {
		dateOptions.DateFormatType = "DATE_FORMAT_UNSPECIFIED"
	}
	return
}

func (fieldModel *gdriveLabelDateFieldResourceModel) getLifeCycle() (lifecycle lifeCycleInterface) {
	return fieldModel.LifeCycle
}

func (fieldModel *gdriveLabelDateFieldResourceModel) setProperties(properties *gdriveLabelFieldPropertieseModel) {
	fieldModel.Properties = properties
}

func (fieldModel *gdriveLabelDateFieldResourceModel) setIds(labelId, fieldId, queryKey string) {
	fieldModel.Id = types.StringValue(combineId(labelId, fieldId))
	fieldModel.LabelId = types.StringValue(labelId)
	fieldModel.FieldId = types.StringValue(fieldId)
	fieldModel.QueryKey = types.StringValue(queryKey)
}

func (fieldModel *gdriveLabelDateFieldResourceModel) getProperties() *gdriveLabelFieldPropertieseModel {
	return fieldModel.Properties
}

func (fieldModel *gdriveLabelDateFieldResourceModel) getId() string {
	return fieldModel.Id.ValueString()
}

func (fieldModel *gdriveLabelDateFieldResourceModel) getFieldId() string {
	return fieldModel.FieldId.ValueString()
}

func (fieldModel *gdriveLabelDateFieldResourceModel) getLabelId() string {
	return fieldModel.LabelId.ValueString()
}

func (fieldModel *gdriveLabelDateFieldResourceModel) getLanguageCode() string {
	return fieldModel.LanguageCode.ValueString()
}

func (fieldModel *gdriveLabelDateFieldResourceModel) getUseAdminAccess() bool {
	return fieldModel.UseAdminAccess.ValueBool()
}

func (r *gdriveLabelDateFieldResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_label_date_field"
}

func (r *gdriveLabelDateFieldResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Creates a Date Field for a Drive Label.

Changes made to a Field must be published via the Field's Label before they are available for files.

Publishing can only be done via the Label resource, NOT the Field resources.

This means that, if you have Labels and Fields in the same Terraform configuration and you make changes
to the Fields you may have to apply twice in order to
1. Apply the changes to the Fields.
2. Publish the changes via the Label.

A Field must be deactivated before it can be deleted.`,
		Attributes: map[string]schema.Attribute{
			"id": rsId(),
			"field_id": schema.StringAttribute{
				MarkdownDescription: `The key of the field, unique within a label or library.

This value is autogenerated. Matches the regex: ([a-zA-Z0-9])+`,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"query_key": schema.StringAttribute{
				MarkdownDescription: `The key to use when constructing Drive search queries to find files based on values defined for this field on files. For example, "{queryKey} > 2001-01-01".`,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"label_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the label.",
				Required:            true,
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
		},
		Blocks: map[string]schema.Block{
			"life_cycle": lifeCycleRS(),
			"properties": fieldProperties(),
			"date_options": schema.SingleNestedBlock{
				MarkdownDescription: `Date field options.`,
				Attributes: map[string]schema.Attribute{
					"date_format_type": schema.StringAttribute{
						MarkdownDescription: `Localized date formatting option. Field values are rendered in this format according to their locale.

The following values are accepted:
* "LONG_DATE"   - Includes full month name. For example, January 12, 1999 (MMMM d, y)
* "SHORT_DATE"  - Short, numeric, representation. For example, 12/13/99 (M/d/yy)`,
						Optional: true,
					},
				},
			},
		},
	}
}

func (r *gdriveLabelDateFieldResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *gdriveLabelDateFieldResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := &gdriveLabelDateFieldResourceModel{}
	resp.Diagnostics.Append(req.Plan.Get(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(createLabelField(plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *gdriveLabelDateFieldResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	state := &gdriveLabelDateFieldResourceModel{}
	resp.Diagnostics.Append(req.State.Get(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	field, err := populateField(state)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get label field, got error: %s", err))
		return
	}
	if field.DateOptions != nil && field.DateOptions.DateFormatType != "" {
		state.DateOptions = &gdriveLabelDateOptionsRSModel{
			DateFormatType: types.StringValue(field.DateOptions.DateFormatType),
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *gdriveLabelDateFieldResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := &gdriveLabelDateFieldResourceModel{}
	resp.Diagnostics.Append(req.Plan.Get(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state := &gdriveLabelDateFieldResourceModel{}
	resp.Diagnostics.Append(req.State.Get(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	updateLabelRequest := newUpdateFieldRequest(plan, state)
	planField := plan.toField()
	stateField := state.toField()
	fieldId := plan.getFieldId()
	if !reflect.DeepEqual(planField.DateOptions, stateField.DateOptions) {
		updateDateOptionsRequest := &drivelabels.GoogleAppsDriveLabelsV2DeltaUpdateLabelRequestRequest{
			UpdateFieldType: &drivelabels.GoogleAppsDriveLabelsV2DeltaUpdateLabelRequestUpdateFieldTypeRequest{
				Id:          fieldId,
				DateOptions: &drivelabels.GoogleAppsDriveLabelsV2FieldDateOptions{},
			},
		}
		if plan.DateOptions != nil {
			updateDateOptionsRequest.UpdateFieldType.DateOptions = plan.DateOptions.toDateOptions()
		}
		updateLabelRequest.Requests = append(updateLabelRequest.Requests, updateDateOptionsRequest)
	}
	_, err := gsmdrivelabels.Delta(gsmhelpers.EnsurePrefix(plan.getLabelId(), "labels/"), "*", updateLabelRequest)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update date field, got error: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *gdriveLabelDateFieldResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	state := &gdriveLabelDateFieldResourceModel{}
	resp.Diagnostics.Append(req.State.Get(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(deleteLabelField(state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *gdriveLabelDateFieldResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(importSplitId(ctx, req, resp, adminAttributeLabels, "label_id/field_id")...)
}
