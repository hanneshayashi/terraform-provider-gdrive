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
	"strconv"

	"github.com/hanneshayashi/gsm/gsmdrive"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	rsschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"google.golang.org/api/drive/v3"
)

func fieldsToMap(fields []*gdriveLabelFieldModel) map[string]*gdriveLabelFieldModel {
	m := map[string]*gdriveLabelFieldModel{}
	for i := range fields {
		m[fields[i].FieldId.ValueString()] = fields[i]
	}
	return m
}

func (labelModel *gdriveLabelPolicyLabelModel) toMap() map[string]*gdriveLabelFieldModel {
	return fieldsToMap(labelModel.Fields)
}

func (labelModel *gdriveLabelAssignmentResourceModel) toMap() map[string]*gdriveLabelFieldModel {
	return fieldsToMap(labelModel.Fields)
}

func (labelModel *gdriveLabelPolicyResourceModel) toMap() map[string]*gdriveLabelPolicyLabelModel {
	m := map[string]*gdriveLabelPolicyLabelModel{}
	for i := range labelModel.Labels {
		m[labelModel.Labels[i].LabelId.ValueString()] = labelModel.Labels[i]
	}
	return m
}

func (fieldModel *gdriveLabelFieldModel) toFieldModification(ctx context.Context) (*drive.LabelFieldModification, diag.Diagnostics) {
	var diags diag.Diagnostics
	fieldMod := &drive.LabelFieldModification{
		FieldId: fieldModel.FieldId.ValueString(),
	}
	if fieldModel.Values.IsNull() {
		fieldMod.UnsetValues = true
	} else {
		valueType := fieldModel.ValueType.ValueString()
		switch valueType {
		case "dateString":
			diags = fieldModel.Values.ElementsAs(ctx, &fieldMod.SetDateValues, false)
			if diags.HasError() {
				return nil, diags
			}
		case "integer":
			values := []string{}
			diags = fieldModel.Values.ElementsAs(ctx, &values, false)
			if diags.HasError() {
				return nil, diags
			}
			for v := range values {
				vi, err := strconv.ParseInt(values[v], 10, 64)
				if err != nil {
					diags.AddError("Configuration Error", fmt.Sprintf("Unable to use %s as a value for an integer field, got error: %s", values[v], err))
					return nil, diags
				}
				fieldMod.SetIntegerValues = append(fieldMod.SetIntegerValues, vi)
			}
		case "selection":
			diags = fieldModel.Values.ElementsAs(ctx, &fieldMod.SetSelectionValues, false)
			if diags.HasError() {
				return nil, diags
			}
		case "text":
			diags = fieldModel.Values.ElementsAs(ctx, &fieldMod.SetTextValues, false)
			if diags.HasError() {
				return nil, diags
			}
		case "user":
			diags = fieldModel.Values.ElementsAs(ctx, &fieldMod.SetUserValues, false)
			if diags.HasError() {
				return nil, diags
			}
		case "default":
			diags.AddError("Configuration Error", fmt.Sprintf("Unable to use %s as a value_type for field", valueType))
			return nil, diags
		}
	}
	return fieldMod, diags
}

func setFieldDiffs(plan, state *gdriveLabelAssignmentResourceModel, ctx context.Context) (diags diag.Diagnostics) {
	modLabelsReq := &drive.ModifyLabelsRequest{
		LabelModifications: []*drive.LabelModification{
			{
				LabelId:            plan.LabelId.ValueString(),
				FieldModifications: []*drive.LabelFieldModification{},
			},
		},
	}
	planMap := plan.toMap()
	stateMap := state.toMap()
	for i := range planMap {
		_, fieldAlreadySet := stateMap[i]
		if !fieldAlreadySet || (fieldAlreadySet && !planMap[i].Values.Equal(stateMap[i].Values)) {
			var fieldMod *drive.LabelFieldModification
			fieldMod, diags = planMap[i].toFieldModification(ctx)
			if diags.HasError() {
				return diags
			}
			modLabelsReq.LabelModifications[0].FieldModifications = append(modLabelsReq.LabelModifications[0].FieldModifications, fieldMod)
		}
	}
	for i := range stateMap {
		_, fieldStillExists := planMap[i]
		if !fieldStillExists {
			modLabelsReq.LabelModifications[0].FieldModifications = append(modLabelsReq.LabelModifications[0].FieldModifications, &drive.LabelFieldModification{
				FieldId:     i,
				UnsetValues: true,
			})
		}
	}
	_, err := gsmdrive.ModifyLabels(plan.FileId.ValueString(), "", modLabelsReq)
	if err != nil {
		diags.AddError("Configuration Error", fmt.Sprintf("Unable to update label assignment, got error: %s", err))
		return
	}
	return
}

func setLabelDiffs(plan, state *gdriveLabelPolicyResourceModel, ctx context.Context) (diags diag.Diagnostics) {
	planLabels := plan.toMap()
	stateLabels := state.toMap()
	modLabelsReq := &drive.ModifyLabelsRequest{
		LabelModifications: []*drive.LabelModification{},
	}
	for i := range planLabels {
		_, labelAlreadyExists := stateLabels[i]
		change := false
		labelMod := &drive.LabelModification{
			LabelId:            i,
			FieldModifications: []*drive.LabelFieldModification{},
		}
		if labelAlreadyExists {
			planFieldMap := planLabels[i].toMap()
			stateFieldMap := stateLabels[i].toMap()
			for f := range planFieldMap {
				_, fieldAlreadyExists := stateFieldMap[f]
				if !fieldAlreadyExists || (fieldAlreadyExists && !stateFieldMap[f].Values.Equal(planFieldMap[f].Values)) {
					var fieldMod *drive.LabelFieldModification
					fieldMod, diags = planFieldMap[f].toFieldModification(ctx)
					if diags.HasError() {
						return diags
					}
					labelMod.FieldModifications = append(labelMod.FieldModifications, fieldMod)
					change = true
				}
			}
			for f := range stateFieldMap {
				_, fieldShouldStillExist := planFieldMap[f]
				if !fieldShouldStillExist {
					labelMod.FieldModifications = append(labelMod.FieldModifications, &drive.LabelFieldModification{
						FieldId:     f,
						UnsetValues: true,
					})
					change = true
				}
			}
		} else {
			for f := range planLabels[i].Fields {
				var fieldMod *drive.LabelFieldModification
				fieldMod, diags = planLabels[i].Fields[f].toFieldModification(ctx)
				if diags.HasError() {
					return diags
				}
				labelMod.FieldModifications = append(labelMod.FieldModifications, fieldMod)
				change = true
			}
		}
		if change {
			modLabelsReq.LabelModifications = append(modLabelsReq.LabelModifications, labelMod)
		}
	}
	for i := range stateLabels {
		_, labelStillPlanned := planLabels[i]
		if !labelStillPlanned {
			modLabelsReq.LabelModifications = append(modLabelsReq.LabelModifications, &drive.LabelModification{
				LabelId:     stateLabels[i].LabelId.ValueString(),
				RemoveLabel: true,
			})
		}
	}
	_, err := gsmdrive.ModifyLabels(plan.FileId.ValueString(), "", modLabelsReq)
	if err != nil {
		diags.AddError("Configuration Error", fmt.Sprintf("Unable to update label assignment, got error: %s", err))
		return diags
	}
	return diags
}

func driveLabelFieldToFieldModel(field drive.LabelField, ctx context.Context) (fieldModel *gdriveLabelFieldModel, diags diag.Diagnostics) {
	fieldModel = &gdriveLabelFieldModel{
		ValueType: types.StringValue(field.ValueType),
		FieldId:   types.StringValue(field.Id),
	}
	switch field.ValueType {
	case "dateString":
		fieldModel.Values, diags = types.SetValueFrom(ctx, types.StringType, field.DateString)
		if diags.HasError() {
			return nil, diags
		}
	case "text":
		fieldModel.Values, diags = types.SetValueFrom(ctx, types.StringType, field.Text)
		if diags.HasError() {
			return nil, diags
		}
	case "user":
		values := []string{}
		for u := range field.User {
			values = append(values, field.User[u].EmailAddress)
		}
		fieldModel.Values, diags = types.SetValueFrom(ctx, types.StringType, values)
		if diags.HasError() {
			return nil, diags
		}
	case "selection":
		fieldModel.Values, diags = types.SetValueFrom(ctx, types.StringType, field.Selection)
		if diags.HasError() {
			return nil, diags
		}
	case "integer":
		values := []string{}
		for _, value := range field.Integer {
			values = append(values, strconv.FormatInt(value, 10))
		}
		fieldModel.Values, diags = types.SetValueFrom(ctx, types.StringType, values)
		if diags.HasError() {
			return nil, diags
		}
	}
	return fieldModel, diags
}

func (labelAssignmentModel *gdriveLabelAssignmentResourceModel) populate(ctx context.Context) (diags diag.Diagnostics) {
	fileID, labelID, e := splitId(labelAssignmentModel.Id.ValueString())
	if e != nil {
		diags.AddError("Config Error", fmt.Sprintf("Unable to use ID, got error: %s", e))
		return diags
	}
	labelAssignmentModel.Fields = []*gdriveLabelFieldModel{}
	currentLabels, err := gsmdrive.ListLabels(fileID, "", 1)
	for l := range currentLabels {
		if l.Id == labelID {
			for f := range l.Fields {
				var field *gdriveLabelFieldModel
				field, diags = driveLabelFieldToFieldModel(l.Fields[f], ctx)
				if diags.HasError() {
					return diags
				}
				labelAssignmentModel.Fields = append(labelAssignmentModel.Fields, field)
			}
		}
	}
	e = <-err
	if e != nil {
		diags.AddError("Client Error", fmt.Sprintf("Unable to list labels on file, got error: %s", e))
		return diags
	}
	labelAssignmentModel.FileId = types.StringValue(fileID)
	labelAssignmentModel.LabelId = types.StringValue(labelID)
	return diags
}

func (labelPolicyModel *gdriveLabelPolicyResourceModel) populate(ctx context.Context) diag.Diagnostics {
	var diags diag.Diagnostics
	labelPolicyModel.Labels = []*gdriveLabelPolicyLabelModel{}
	labelPolicyModel.FileId = labelPolicyModel.Id
	fileID := labelPolicyModel.FileId.ValueString()
	currentLabels, err := gsmdrive.ListLabels(fileID, "", 1)
	for l := range currentLabels {
		label := &gdriveLabelPolicyLabelModel{
			LabelId: types.StringValue(l.Id),
			Fields:  []*gdriveLabelFieldModel{},
		}
		for f := range l.Fields {
			field := &gdriveLabelFieldModel{
				ValueType: types.StringValue(l.Fields[f].ValueType),
				FieldId:   types.StringValue(l.Fields[f].Id),
			}
			switch l.Fields[f].ValueType {
			case "dateString":
				field.Values, diags = types.SetValueFrom(ctx, types.StringType, l.Fields[f].DateString)
				if diags.HasError() {
					return diags
				}
			case "text":
				field.Values, diags = types.SetValueFrom(ctx, types.StringType, l.Fields[f].Text)
				if diags.HasError() {
					return diags
				}
			case "user":
				values := []string{}
				for u := range l.Fields[f].User {
					values = append(values, l.Fields[f].User[u].EmailAddress)
				}
				field.Values, diags = types.SetValueFrom(ctx, types.StringType, values)
				if diags.HasError() {
					return diags
				}
			case "selection":
				field.Values, diags = types.SetValueFrom(ctx, types.StringType, l.Fields[f].Selection)
				if diags.HasError() {
					return diags
				}
			case "integer":
				values := []string{}
				for _, value := range l.Fields[f].Integer {
					values = append(values, strconv.FormatInt(value, 10))
				}
				field.Values, diags = types.SetValueFrom(ctx, types.StringType, values)
				if diags.HasError() {
					return diags
				}
			}
			label.Fields = append(label.Fields, field)
		}
		labelPolicyModel.Labels = append(labelPolicyModel.Labels, label)
	}
	e := <-err
	if e != nil {
		diags.AddError("Client Error", fmt.Sprintf("Unable to list labels on file, got error: %s", e))
		return diags
	}
	return diags
}

func labelAssignmentFields() rsschema.SetNestedAttribute {
	return rsschema.SetNestedAttribute{
		MarkdownDescription: `A Set of fields of the assigned label.`,
		Required:            true,
		NestedObject: rsschema.NestedAttributeObject{
			Attributes: map[string]rsschema.Attribute{
				"field_id": rsschema.StringAttribute{
					Required:            true,
					MarkdownDescription: "The identifier of this field.",
				},
				"value_type": rsschema.StringAttribute{
					Required: true,
					MarkdownDescription: `The field type.
While new values may be supported in the future, the following are currently allowed:
* dateString
* integer
* selection
* text
* user`,
				},
				"values": rsschema.SetAttribute{
					ElementType: types.StringType,
					Required:    true,
					MarkdownDescription: `The values that should be set.

Must be compatible with the specified value_type.`,
				},
			},
		},
	}
}
