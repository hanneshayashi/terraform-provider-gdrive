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
	"fmt"
	"reflect"

	"github.com/hanneshayashi/gsm/gsmdrivelabels"
	"github.com/hanneshayashi/gsm/gsmhelpers"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	rsschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"google.golang.org/api/drivelabels/v2"
)

type lifeCycleInterface interface {
	populate(*drivelabels.GoogleAppsDriveLabelsV2Lifecycle)
}

type fieldInterface interface {
	getLanguageCode() string
	getId() string
	getLabelId() string
	getFieldId() string
	getUseAdminAccess() bool
	getProperties() *gdriveLabelFieldPropertieseModel
	setProperties(*gdriveLabelFieldPropertieseModel)
	getLifeCycle() (lifecycle lifeCycleInterface)
	setIds(string, string, string)
	toField() *drivelabels.GoogleAppsDriveLabelsV2Field
}

func populateField(fieldModel fieldInterface) (field *drivelabels.GoogleAppsDriveLabelsV2Field, err error) {
	labelId, fieldId, err := splitId(fieldModel.getId())
	if err != nil {
		return nil, err
	}
	l, err := gsmdrivelabels.GetLabel(gsmhelpers.EnsurePrefix(labelId, "labels/"), fieldModel.getLanguageCode(), "LABEL_VIEW_FULL", "*", fieldModel.getUseAdminAccess())
	if err != nil {
		return nil, err
	}
	f := fieldModel.toField()
	for i := range l.Fields {
		if l.Fields[i].Id == fieldId && l.Fields[i].Properties != nil {
			properties := &gdriveLabelFieldPropertieseModel{
				DisplayName: types.StringValue(l.Fields[i].Properties.DisplayName),
				Required:    types.BoolValue(l.Fields[i].Properties.Required),
			}
			fieldModel.setIds(labelId, fieldId, l.Fields[i].QueryKey)
			if i < len(l.Fields)-1 && f.Properties != nil && f.Properties.InsertBeforeField != "" {
				properties.InsertBeforeField = types.StringValue(l.Fields[i+1].Id)
			}
			fieldModel.setProperties(properties)
			lifeCycle := fieldModel.getLifeCycle()
			if f.Lifecycle != nil && lifeCycle != nil {
				lifeCycle.populate(l.Fields[i].Lifecycle)
			}
			return l.Fields[i], nil
		}
	}
	return nil, fmt.Errorf("field not found")
}

func getUpdateFieldLifecycleRequest(id string, lifecycle *drivelabels.GoogleAppsDriveLabelsV2Lifecycle) (request *drivelabels.GoogleAppsDriveLabelsV2DeltaUpdateLabelRequestRequest) {
	switch lifecycle.State {
	case "PUBLISHED":
		return &drivelabels.GoogleAppsDriveLabelsV2DeltaUpdateLabelRequestRequest{
			EnableField: &drivelabels.GoogleAppsDriveLabelsV2DeltaUpdateLabelRequestEnableFieldRequest{
				Id: id,
			},
		}
	case "DISABLED":
		return &drivelabels.GoogleAppsDriveLabelsV2DeltaUpdateLabelRequestRequest{
			DisableField: &drivelabels.GoogleAppsDriveLabelsV2DeltaUpdateLabelRequestDisableFieldRequest{
				Id:             id,
				DisabledPolicy: lifecycle.DisabledPolicy,
			},
		}
	}
	return
}

func newUpdateFieldRequest(plan, state fieldInterface) *drivelabels.GoogleAppsDriveLabelsV2DeltaUpdateLabelRequest {
	updateLabelRequest := newUpdateLabelRequest(plan)
	fieldId := plan.getFieldId()
	planField := plan.toField()
	stateField := state.toField()
	if planField.Properties != nil && !reflect.DeepEqual(planField.Properties, stateField.Properties) {
		updateLabelRequest.Requests = append(updateLabelRequest.Requests, &drivelabels.GoogleAppsDriveLabelsV2DeltaUpdateLabelRequestRequest{
			UpdateField: &drivelabels.GoogleAppsDriveLabelsV2DeltaUpdateLabelRequestUpdateFieldPropertiesRequest{
				Id:         fieldId,
				Properties: plan.toField().Properties,
			},
		})
	}
	if planField.Lifecycle != nil && !reflect.DeepEqual(planField.Lifecycle, stateField.Lifecycle) {
		updateLifecycleRequest := getUpdateFieldLifecycleRequest(fieldId, plan.toField().Lifecycle)
		updateLabelRequest.Requests = append(updateLabelRequest.Requests, updateLifecycleRequest)
	}
	return updateLabelRequest
}

func createLabelField(plan fieldInterface) (diags diag.Diagnostics) {
	updateLabelRequest := newUpdateLabelRequest(plan)
	field := plan.toField()
	labelId := plan.getLabelId()
	updateLabelRequest.Requests = []*drivelabels.GoogleAppsDriveLabelsV2DeltaUpdateLabelRequestRequest{
		{
			CreateField: &drivelabels.GoogleAppsDriveLabelsV2DeltaUpdateLabelRequestCreateFieldRequest{
				Field: field,
			},
		},
	}
	updatedLabel, err := gsmdrivelabels.Delta(gsmhelpers.EnsurePrefix(labelId, "labels/"), "*", updateLabelRequest)
	if err != nil {
		diags.AddError("Client Error", fmt.Sprintf("Unable to create field, got error: %s", err))
		return diags
	}
	var newField *drivelabels.GoogleAppsDriveLabelsV2Field
	if updateLabelRequest.Requests[0].CreateField.Field.Properties.InsertBeforeField == "" {
		newField = updatedLabel.UpdatedLabel.Fields[len(updatedLabel.UpdatedLabel.Fields)-1]
	} else {
		for i := range updatedLabel.UpdatedLabel.Fields {
			if updatedLabel.UpdatedLabel.Fields[i].Id == updateLabelRequest.Requests[0].CreateField.Field.Properties.InsertBeforeField {
				newField = updatedLabel.UpdatedLabel.Fields[i-1]
				break
			}
		}
	}
	plan.setIds(labelId, newField.Id, newField.QueryKey)
	return
}

func deleteLabelField(state fieldInterface) (diags diag.Diagnostics) {
	updateLabelRequest := newUpdateLabelRequest(state)
	updateLabelRequest.Requests = []*drivelabels.GoogleAppsDriveLabelsV2DeltaUpdateLabelRequestRequest{
		{
			DeleteField: &drivelabels.GoogleAppsDriveLabelsV2DeltaUpdateLabelRequestDeleteFieldRequest{
				Id: state.getFieldId(),
			},
		},
	}
	_, err := gsmdrivelabels.Delta(gsmhelpers.EnsurePrefix(state.getLabelId(), "labels/"), "*", updateLabelRequest)
	if err != nil {
		diags.AddError("Client Error", fmt.Sprintf("Unable to delete text field, got error: %s", err))
	}
	return
}

func fieldProperties() rsschema.SingleNestedBlock {
	return rsschema.SingleNestedBlock{
		MarkdownDescription: "The basic properties of the field.",
		Attributes: map[string]rsschema.Attribute{
			"display_name": rsschema.StringAttribute{
				MarkdownDescription: "The display text to show in the UI identifying this field.",
				Required:            true,
			},
			"insert_before_field": rsschema.StringAttribute{
				MarkdownDescription: `Insert or move this field before the indicated field.
If empty, the field is placed at the end of the list.`,
				Optional: true,
			},
			"required": rsschema.BoolAttribute{
				MarkdownDescription: "Whether the field should be marked as required.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		},
	}
}

func rsListOptions() rsschema.SingleNestedBlock {
	return rsschema.SingleNestedBlock{
		MarkdownDescription: "Options for a multi-valued variant of an associated field type.",
		Attributes: map[string]rsschema.Attribute{
			"max_entries": rsschema.Int64Attribute{
				Optional: true,
			},
		},
	}
}
