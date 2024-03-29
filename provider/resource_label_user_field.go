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
var _ resource.Resource = &gdriveLabelUserFieldResource{}
var _ resource.ResourceWithImportState = &gdriveLabelUserFieldResource{}

func newLabelUserField() resource.Resource {
	return &gdriveLabelUserFieldResource{}
}

// gdriveLabelUserFieldResource defines the resource implementation.
type gdriveLabelUserFieldResource struct {
	client *http.Client
}

type gdriveLabelUserOptionsRSModel struct {
	ListOptions *gdriveLabelListOptionsModel `tfsdk:"list_options"`
}

type gdriveLabelUserFieldResourceModel struct {
	LifeCycle      *gdriveLabelLifeCycleModel        `tfsdk:"life_cycle"`
	Properties     *gdriveLabelFieldPropertieseModel `tfsdk:"properties"`
	UserOptions    *gdriveLabelUserOptionsRSModel    `tfsdk:"user_options"`
	Id             types.String                      `tfsdk:"id"`
	FieldId        types.String                      `tfsdk:"field_id"`
	LabelId        types.String                      `tfsdk:"label_id"`
	QueryKey       types.String                      `tfsdk:"query_key"`
	LanguageCode   types.String                      `tfsdk:"language_code"`
	UseAdminAccess types.Bool                        `tfsdk:"use_admin_access"`
}

func (fieldModel *gdriveLabelUserFieldResourceModel) toField() (field *drivelabels.GoogleAppsDriveLabelsV2Field) {
	field = &drivelabels.GoogleAppsDriveLabelsV2Field{
		UserOptions: &drivelabels.GoogleAppsDriveLabelsV2FieldUserOptions{},
	}
	if fieldModel.UserOptions != nil && fieldModel.UserOptions.ListOptions != nil {
		field.UserOptions = &drivelabels.GoogleAppsDriveLabelsV2FieldUserOptions{
			ListOptions: &drivelabels.GoogleAppsDriveLabelsV2FieldListOptions{
				MaxEntries: fieldModel.UserOptions.ListOptions.MaxEntries.ValueInt64(),
			},
		}
		if field.UserOptions.ListOptions.MaxEntries == 0 {
			field.UserOptions.ListOptions.ForceSendFields = append(field.UserOptions.ListOptions.ForceSendFields, "MaxEntries")
		}
	}
	if fieldModel.LifeCycle != nil {
		field.Lifecycle = fieldModel.LifeCycle.toLifecycle()
	}
	if fieldModel.Properties != nil {
		field.Properties = fieldModel.Properties.toProperties()
	}
	return
}

func (fieldModel *gdriveLabelUserFieldResourceModel) getLifeCycle() (lifecycle lifeCycleInterface) {
	return fieldModel.LifeCycle
}

func (fieldModel *gdriveLabelUserFieldResourceModel) setProperties(properties *gdriveLabelFieldPropertieseModel) {
	fieldModel.Properties = properties
}

func (fieldModel *gdriveLabelUserFieldResourceModel) setIds(labelId, fieldId, queryKey string) {
	fieldModel.Id = types.StringValue(combineId(labelId, fieldId))
	fieldModel.LabelId = types.StringValue(labelId)
	fieldModel.FieldId = types.StringValue(fieldId)
	fieldModel.QueryKey = types.StringValue(queryKey)
}

func (fieldModel *gdriveLabelUserFieldResourceModel) getProperties() *gdriveLabelFieldPropertieseModel {
	return fieldModel.Properties
}

func (fieldModel *gdriveLabelUserFieldResourceModel) getId() string {
	return fieldModel.Id.ValueString()
}

func (fieldModel *gdriveLabelUserFieldResourceModel) getFieldId() string {
	return fieldModel.FieldId.ValueString()
}

func (fieldModel *gdriveLabelUserFieldResourceModel) getLabelId() string {
	return fieldModel.LabelId.ValueString()
}

func (fieldModel *gdriveLabelUserFieldResourceModel) getLanguageCode() string {
	return fieldModel.LanguageCode.ValueString()
}

func (fieldModel *gdriveLabelUserFieldResourceModel) getUseAdminAccess() bool {
	return fieldModel.UseAdminAccess.ValueBool()
}

func (r *gdriveLabelUserFieldResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_label_user_field"
}

func (r *gdriveLabelUserFieldResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Creates a User Field for a Drive Label.

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
			"user_options": schema.SingleNestedBlock{
				MarkdownDescription: `User field options.`,
				Blocks: map[string]schema.Block{
					"list_options": rsListOptions(),
				},
			},
		},
	}
}

func (r *gdriveLabelUserFieldResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *gdriveLabelUserFieldResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := &gdriveLabelUserFieldResourceModel{}
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

func (r *gdriveLabelUserFieldResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	state := &gdriveLabelUserFieldResourceModel{}
	resp.Diagnostics.Append(req.State.Get(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	field, err := populateField(state)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get label field, got error: %s", err))
		return
	}
	if field.UserOptions != nil && field.UserOptions.ListOptions != nil && field.UserOptions.ListOptions.MaxEntries != 0 {
		state.UserOptions = &gdriveLabelUserOptionsRSModel{
			ListOptions: &gdriveLabelListOptionsModel{
				MaxEntries: types.Int64Value(field.UserOptions.ListOptions.MaxEntries),
			},
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *gdriveLabelUserFieldResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := &gdriveLabelUserFieldResourceModel{}
	resp.Diagnostics.Append(req.Plan.Get(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state := &gdriveLabelUserFieldResourceModel{}
	resp.Diagnostics.Append(req.State.Get(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	updateLabelRequest := newUpdateFieldRequest(plan, state)
	planField := plan.toField()
	stateField := state.toField()
	fieldId := plan.getFieldId()
	if !reflect.DeepEqual(planField.UserOptions, stateField.UserOptions) {
		updateUserOptionsRequest := &drivelabels.GoogleAppsDriveLabelsV2DeltaUpdateLabelRequestRequest{
			UpdateFieldType: &drivelabels.GoogleAppsDriveLabelsV2DeltaUpdateLabelRequestUpdateFieldTypeRequest{
				Id:          fieldId,
				UserOptions: planField.UserOptions,
			},
		}
		updateLabelRequest.Requests = append(updateLabelRequest.Requests, updateUserOptionsRequest)
	}
	if len(updateLabelRequest.Requests) > 0 {
		_, err := gsmdrivelabels.Delta(gsmhelpers.EnsurePrefix(plan.getLabelId(), "labels/"), "*", updateLabelRequest)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update user field, got error: %s", err))
			return
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *gdriveLabelUserFieldResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	state := &gdriveLabelUserFieldResourceModel{}
	resp.Diagnostics.Append(req.State.Get(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(deleteLabelField(state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *gdriveLabelUserFieldResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(importSplitId(ctx, req, resp, adminAttributeLabels, "label_id/field_id")...)
}
