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
var _ resource.Resource = &gdriveLabelSelectionFieldResource{}
var _ resource.ResourceWithImportState = &gdriveLabelSelectionFieldResource{}

func newLabelSelectionField() resource.Resource {
	return &gdriveLabelSelectionFieldResource{}
}

// gdriveLabelSelectionFieldResource defines the resource implementation.
type gdriveLabelSelectionFieldResource struct {
	client *http.Client
}

type gdriveLabelSelectionOptionsRSModel struct {
	ListOptions *gdriveLabelListOptionsModel `tfsdk:"list_options"`
}

type gdriveLabelSelectionFieldResourceModel struct {
	LifeCycle        *gdriveLabelLifeCycleModel          `tfsdk:"life_cycle"`
	Properties       *gdriveLabelFieldPropertieseModel   `tfsdk:"properties"`
	SelectionOptions *gdriveLabelSelectionOptionsRSModel `tfsdk:"selection_options"`
	Id               types.String                        `tfsdk:"id"`
	FieldId          types.String                        `tfsdk:"field_id"`
	LabelId          types.String                        `tfsdk:"label_id"`
	QueryKey         types.String                        `tfsdk:"query_key"`
	LanguageCode     types.String                        `tfsdk:"language_code"`
	UseAdminAccess   types.Bool                          `tfsdk:"use_admin_access"`
}

func (fieldModel *gdriveLabelSelectionFieldResourceModel) toField() (field *drivelabels.GoogleAppsDriveLabelsV2Field) {
	field = &drivelabels.GoogleAppsDriveLabelsV2Field{
		SelectionOptions: &drivelabels.GoogleAppsDriveLabelsV2FieldSelectionOptions{},
	}
	if fieldModel.LifeCycle != nil {
		field.Lifecycle = fieldModel.LifeCycle.toLifecycle()
	}
	if fieldModel.Properties != nil {
		field.Properties = fieldModel.Properties.toProperties()
	}
	if fieldModel.SelectionOptions != nil && fieldModel.SelectionOptions.ListOptions != nil {
		field.SelectionOptions = &drivelabels.GoogleAppsDriveLabelsV2FieldSelectionOptions{
			ListOptions: &drivelabels.GoogleAppsDriveLabelsV2FieldListOptions{
				MaxEntries: fieldModel.SelectionOptions.ListOptions.MaxEntries.ValueInt64(),
			},
		}
		if field.SelectionOptions.ListOptions.MaxEntries == 0 {
			field.SelectionOptions.ListOptions.ForceSendFields = append(field.SelectionOptions.ListOptions.ForceSendFields, "MaxEntries")
		}
	}
	return
}

func (fieldModel *gdriveLabelSelectionFieldResourceModel) getLifeCycle() (lifecycle lifeCycleInterface) {
	return fieldModel.LifeCycle
}

func (fieldModel *gdriveLabelSelectionFieldResourceModel) setProperties(properties *gdriveLabelFieldPropertieseModel) {
	fieldModel.Properties = properties
}

func (fieldModel *gdriveLabelSelectionFieldResourceModel) setIds(labelId, fieldId, queryKey string) {
	fieldModel.Id = types.StringValue(combineId(labelId, fieldId))
	fieldModel.LabelId = types.StringValue(labelId)
	fieldModel.FieldId = types.StringValue(fieldId)
	fieldModel.QueryKey = types.StringValue(queryKey)
}

func (fieldModel *gdriveLabelSelectionFieldResourceModel) getProperties() *gdriveLabelFieldPropertieseModel {
	return fieldModel.Properties
}

func (fieldModel *gdriveLabelSelectionFieldResourceModel) getId() string {
	return fieldModel.Id.ValueString()
}

func (fieldModel *gdriveLabelSelectionFieldResourceModel) getFieldId() string {
	return fieldModel.FieldId.ValueString()
}

func (fieldModel *gdriveLabelSelectionFieldResourceModel) getLabelId() string {
	return fieldModel.LabelId.ValueString()
}

func (fieldModel *gdriveLabelSelectionFieldResourceModel) getLanguageCode() string {
	return fieldModel.LanguageCode.ValueString()
}

func (fieldModel *gdriveLabelSelectionFieldResourceModel) getUseAdminAccess() bool {
	return fieldModel.UseAdminAccess.ValueBool()
}

func (r *gdriveLabelSelectionFieldResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_label_selection_field"
}

func (r *gdriveLabelSelectionFieldResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Creates an Selection Field for a Drive Label",
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
				Description: `Set to true in order to use the user's admin credentials.
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
			"selection_options": schema.SingleNestedBlock{
				MarkdownDescription: `Selection field options.`,
				Blocks: map[string]schema.Block{
					"list_options": rsListOptions(),
				},
			},
		},
	}
}

func (r *gdriveLabelSelectionFieldResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *gdriveLabelSelectionFieldResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := &gdriveLabelSelectionFieldResourceModel{}
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

func (r *gdriveLabelSelectionFieldResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	state := &gdriveLabelSelectionFieldResourceModel{}
	resp.Diagnostics.Append(req.State.Get(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	field, err := populateField(state)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get label field, got error: %s", err))
		return
	}
	if field.SelectionOptions != nil && field.SelectionOptions.ListOptions != nil && field.SelectionOptions.ListOptions.MaxEntries != 0 {
		state.SelectionOptions = &gdriveLabelSelectionOptionsRSModel{
			ListOptions: &gdriveLabelListOptionsModel{
				MaxEntries: types.Int64Value(field.SelectionOptions.ListOptions.MaxEntries),
			},
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *gdriveLabelSelectionFieldResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := &gdriveLabelSelectionFieldResourceModel{}
	resp.Diagnostics.Append(req.Plan.Get(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state := &gdriveLabelSelectionFieldResourceModel{}
	resp.Diagnostics.Append(req.State.Get(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	updateLabelRequest := newUpdateFieldRequest(plan, state)
	planField := plan.toField()
	stateField := state.toField()
	fieldId := plan.getFieldId()
	if !reflect.DeepEqual(planField.SelectionOptions, stateField.SelectionOptions) {
		updateSelectionOptionsRequest := &drivelabels.GoogleAppsDriveLabelsV2DeltaUpdateLabelRequestRequest{
			UpdateFieldType: &drivelabels.GoogleAppsDriveLabelsV2DeltaUpdateLabelRequestUpdateFieldTypeRequest{
				Id:               fieldId,
				SelectionOptions: planField.SelectionOptions,
			},
		}
		updateLabelRequest.Requests = append(updateLabelRequest.Requests, updateSelectionOptionsRequest)
	}
	_, err := gsmdrivelabels.Delta(gsmhelpers.EnsurePrefix(plan.getLabelId(), "labels/"), "*", updateLabelRequest)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update selection field, got error: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *gdriveLabelSelectionFieldResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	state := &gdriveLabelSelectionFieldResourceModel{}
	resp.Diagnostics.Append(req.State.Get(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(deleteLabelField(state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *gdriveLabelSelectionFieldResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(importSplitId(ctx, req, resp, adminAttributeLabels, "label_id/field_id")...)
}
