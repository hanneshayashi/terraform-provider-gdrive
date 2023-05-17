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
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	rsschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/drivelabels/v2"
)

func fieldsToMap(fields []*gdriveLabelFieldModel) map[string]*gdriveLabelFieldModel {
	m := map[string]*gdriveLabelFieldModel{}
	for i := range fields {
		m[fields[i].FieldId.ValueString()] = fields[i]
	}
	return m
}

func (labelModel *gdriveLabelPolicyLabelModel) toMap() map[string]*gdriveLabelFieldModel {
	return fieldsToMap(labelModel.Field)
}

func (labelModel *gdriveLabelAssignmentResourceModel) toMap() map[string]*gdriveLabelFieldModel {
	return fieldsToMap(labelModel.Field)
}

func (labelModel *gdriveLabelPolicyResourceModel) toMap() map[string]*gdriveLabelPolicyLabelModel {
	m := map[string]*gdriveLabelPolicyLabelModel{}
	for i := range labelModel.Label {
		m[labelModel.Label[i].LabelId.ValueString()] = labelModel.Label[i]
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
		tflog.Debug(ctx, fmt.Sprintf("Planned Label: %s", i))
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
				tflog.Debug(ctx, fmt.Sprintf("Planned Label %s already exists, so checking if planned Field %s already exists and if it has changed", i, planFieldMap[f].FieldId.ValueString()))
				_, fieldAlreadyExists := stateFieldMap[f]
				tflog.Debug(ctx, fmt.Sprintf("Planned Label %s; Field %s exists: %v", i, planFieldMap[f].FieldId.ValueString(), fieldAlreadyExists))
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
				tflog.Debug(ctx, fmt.Sprintf("Planned Label %s already exists, so checking if state Field %s should still exists", i, stateFieldMap[f].FieldId.ValueString()))
				_, fieldShouldStillExist := planFieldMap[f]
				if !fieldShouldStillExist {
					tflog.Debug(ctx, fmt.Sprintf("State Field %s is no longer planned for Label %s", f, i))
					labelMod.FieldModifications = append(labelMod.FieldModifications, &drive.LabelFieldModification{
						FieldId:     f,
						UnsetValues: true,
					})
					change = true
				}
			}
		} else {
			tflog.Debug(ctx, fmt.Sprintf("Planned Label %s does not exist yet", i))
			for f := range planLabels[i].Field {
				tflog.Debug(ctx, fmt.Sprintf("Planned Label %s does not exist yet, so adding Field %s", i, planLabels[i].Field[f].FieldId.ValueString()))
				var fieldMod *drive.LabelFieldModification
				fieldMod, diags = planLabels[i].Field[f].toFieldModification(ctx)
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
		tflog.Debug(ctx, fmt.Sprintf("State Label %s", stateLabels[i].LabelId.ValueString()))
		_, labelStillPlanned := planLabels[i]
		if !labelStillPlanned {
			tflog.Debug(ctx, fmt.Sprintf("State Label %s is no longer planned, so removing", stateLabels[i].LabelId.ValueString()))
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
	fileID := labelAssignmentModel.FileId.ValueString()
	labelID := labelAssignmentModel.LabelId.ValueString()
	labelAssignmentModel.Field = []*gdriveLabelFieldModel{}
	currentLabels, err := gsmdrive.ListLabels(fileID, "", 1)
	for l := range currentLabels {
		if l.Id == labelID {
			for f := range l.Fields {
				var field *gdriveLabelFieldModel
				field, diags = driveLabelFieldToFieldModel(l.Fields[f], ctx)
				if diags.HasError() {
					return diags
				}
				labelAssignmentModel.Field = append(labelAssignmentModel.Field, field)
			}
		}
	}
	e := <-err
	if e != nil {
		diags.AddError("Client Error", fmt.Sprintf("Unable to list labels on file, got error: %s", e))
		return diags
	}
	return diags
}

func (labelPolicyModel *gdriveLabelPolicyResourceModel) populate(ctx context.Context) diag.Diagnostics {
	var diags diag.Diagnostics
	labelPolicyModel.Label = []*gdriveLabelPolicyLabelModel{}
	fileID := labelPolicyModel.FileId.ValueString()
	currentLabels, err := gsmdrive.ListLabels(fileID, "", 1)
	for l := range currentLabels {
		label := &gdriveLabelPolicyLabelModel{
			LabelId: types.StringValue(l.Id),
			Field:   []*gdriveLabelFieldModel{},
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
			label.Field = append(label.Field, field)
		}
		labelPolicyModel.Label = append(labelPolicyModel.Label, label)
	}
	e := <-err
	if e != nil {
		diags.AddError("Client Error", fmt.Sprintf("Unable to list labels on file, got error: %s", e))
		return diags
	}
	return diags
}

func dateFieldDS() dsschema.SingleNestedBlock {
	return dsschema.SingleNestedBlock{
		Attributes: map[string]dsschema.Attribute{
			"day": dsschema.Int64Attribute{
				Computed:    true,
				Description: `Day of a month.`,
			},
			"month": dsschema.Int64Attribute{
				Computed:    true,
				Description: "Month of a year.",
			},
			"year": dsschema.Int64Attribute{
				Computed:    true,
				Description: "Year of the date.",
			},
		},
		MarkdownDescription: "Maximum valid value (year, month, day).",
	}
}

func lifecycle() dsschema.SingleNestedBlock {
	return dsschema.SingleNestedBlock{
		MarkdownDescription: `The lifecycle state of an object, such as label, field, or choice.

The lifecycle enforces the following transitions:
UNPUBLISHED_DRAFT (starting state)
UNPUBLISHED_DRAFT -> PUBLISHED
UNPUBLISHED_DRAFT -> (Deleted)
PUBLISHED -> DISABLED
DISABLED -> PUBLISHED
DISABLED -> (Deleted)`,
		Attributes: map[string]dsschema.Attribute{
			"state": dsschema.StringAttribute{
				Description: "The state of the object associated with this lifecycle.",
				Computed:    true,
			},
		},
	}
}

func listOptions() dsschema.SingleNestedBlock {
	return dsschema.SingleNestedBlock{
		MarkdownDescription: "List options",
		Attributes: map[string]dsschema.Attribute{
			"max_entries": dsschema.Int64Attribute{
				Description: "Maximum number of entries permitted.",
				Computed:    true,
			},
		},
	}
}

func fieldsDS() dsschema.ListNestedBlock {
	return dsschema.ListNestedBlock{
		MarkdownDescription: "The fields of this label.",
		NestedObject: dsschema.NestedBlockObject{
			Attributes: map[string]dsschema.Attribute{
				"id": dsId(),
				"field_id": dsschema.StringAttribute{
					Computed: true,
					Description: `The key of a field, unique within a label or library.
							Use this when referencing a field somewhere.`,
				},
				"query_key": dsschema.StringAttribute{
					Computed: true,
					Description: `The key to use when constructing Drive search queries to find labels based on values defined for this field on labels.
						For example, "{queryKey} > 2001-01-01".`,
				},
				"value_type": dsschema.StringAttribute{
					Computed: true,
					Description: `The type of the field.
							Use this when setting the values for a field.`,
				},
			},
			Blocks: map[string]dsschema.Block{
				"life_cycle": lifecycle(),
				"date_options": dsschema.SingleNestedBlock{
					Description: "A set of restrictions that apply to this shared drive or items inside this shared drive.",
					Attributes: map[string]dsschema.Attribute{
						"date_format": dsschema.StringAttribute{
							Computed:    true,
							Description: "ICU date format.",
						},
						"date_format_type": dsschema.StringAttribute{
							Computed:    true,
							Description: "Localized date formatting option. Field values are rendered in this format according to their locale.",
						},
					},
					Blocks: map[string]dsschema.Block{
						"max_value": dateFieldDS(),
						"min_value": dateFieldDS(),
					},
				},
				"selection_options": dsschema.SingleNestedBlock{
					Description: "Options for the selection field type.",
					Blocks: map[string]dsschema.Block{
						"list_options": listOptions(),
						"choices": dsschema.SetNestedBlock{
							NestedObject: dsschema.NestedBlockObject{
								Attributes: map[string]dsschema.Attribute{
									"id": dsId(),
									"choice_id": dsschema.StringAttribute{
										Computed: true,
										Description: `The unique value of the choice.
											Use this when referencing / setting a choice.`,
									},
									"display_name": dsschema.StringAttribute{
										Computed:    true,
										Description: "The display text to show in the UI identifying this field.",
									},
								},
								Blocks: map[string]dsschema.Block{
									"life_cycle": lifecycle(),
								},
							},
						},
					},
				},
				"integer_options": dsschema.SingleNestedBlock{
					Description: "Options for the Integer field type.",
					Attributes: map[string]dsschema.Attribute{
						"max_value": dsschema.Int64Attribute{
							Computed:    true,
							Description: "The maximum valid value for the integer field.",
						},
						"min_value": dsschema.Int64Attribute{
							Computed:    true,
							Description: "The minimum valid value for the integer field.",
						},
					},
				},
				"text_options": dsschema.SingleNestedBlock{
					Description: "Options for the Text field type.",
					Attributes: map[string]dsschema.Attribute{
						"min_length": dsschema.Int64Attribute{
							Computed:    true,
							Description: "The minimum valid length of values for the text field.",
						},
						"max_length": dsschema.Int64Attribute{
							Computed:    true,
							Description: "The maximum valid length of values for the text field.",
						},
					},
				},
				"user_options": dsschema.SingleNestedBlock{
					Description: "Options for the user field type.",
					Blocks: map[string]dsschema.Block{
						"list_options": listOptions(),
					},
				},
				"properties": dsschema.SingleNestedBlock{
					Description: "The basic properties of the field.",
					Attributes: map[string]dsschema.Attribute{
						"display_name": dsschema.StringAttribute{
							Computed:    true,
							Description: "The display text to show in the UI identifying this field.",
						},
						"required": dsschema.BoolAttribute{
							Computed:    true,
							Description: "Whether the field should be marked as required.",
						},
					},
				},
			},
		},
	}
}

func fieldsToModel(fields []*drivelabels.GoogleAppsDriveLabelsV2Field) (model []*gdriveLabelDataSourceFieldsModel) {
	for i := range fields {
		field := &gdriveLabelDataSourceFieldsModel{
			Id:       types.StringValue(fields[i].Id),
			FieldId:  types.StringValue(fields[i].Id),
			QueryKey: types.StringValue(fields[i].QueryKey),
		}
		if fields[i].TextOptions != nil {
			field.ValueType = types.StringValue("text")
			field.TextOptions = &gdriveLabelTextOptionsModel{
				MinLength: types.Int64Value(fields[i].TextOptions.MinLength),
				MaxLength: types.Int64Value(fields[i].TextOptions.MaxLength),
			}
		} else if fields[i].IntegerOptions != nil {
			field.ValueType = types.StringValue("integer")
			field.IntegerOptions = &gdriveLabelIntegerOptionsModel{
				MinValue: types.Int64Value(fields[i].IntegerOptions.MinValue),
				MaxValue: types.Int64Value(fields[i].IntegerOptions.MaxValue),
			}
		} else if fields[i].UserOptions != nil {
			field.ValueType = types.StringValue("user")
			field.UserOptions = &gdriveLabelUserOptionseModel{}
			if fields[i].UserOptions.ListOptions != nil {
				field.UserOptions.ListOptions = &gdriveLabelListOptionsModel{
					MaxEntries: types.Int64Value(fields[i].UserOptions.ListOptions.MaxEntries),
				}
			}
		} else if fields[i].SelectionOptions != nil {
			field.ValueType = types.StringValue("selection")
			field.SelectionOptions = &gdriveLabelSelectionOptionsModel{}
			if fields[i].SelectionOptions.ListOptions != nil {
				field.SelectionOptions.ListOptions = &gdriveLabelListOptionsModel{
					MaxEntries: types.Int64Value(fields[i].SelectionOptions.ListOptions.MaxEntries),
				}
			}
			field.SelectionOptions.Choices = make([]*gdriveLabelChoicedModel, len(fields[i].SelectionOptions.Choices))
			for j := range fields[i].SelectionOptions.Choices {
				field.SelectionOptions.Choices[j] = &gdriveLabelChoicedModel{
					Id:       types.StringValue(fields[i].SelectionOptions.Choices[j].Id),
					ChoiceId: types.StringValue(fields[i].SelectionOptions.Choices[j].Id),
				}
				if fields[i].SelectionOptions.Choices[j].Properties != nil {
					field.SelectionOptions.Choices[j].DisplayName = types.StringValue(fields[i].SelectionOptions.Choices[j].Properties.DisplayName)
				}
				if fields[i].SelectionOptions.Choices[j].Lifecycle != nil {
					field.SelectionOptions.Choices[j].LifeCycle = &gdriveLabelLifeCycleModel{
						State: types.StringValue(fields[i].SelectionOptions.Choices[j].Lifecycle.State),
					}
				}
			}
		} else if fields[i].DateOptions != nil {
			field.ValueType = types.StringValue("dateString")
			field.DateOptions = &gdriveLabelDateOptionsModel{
				DateFormat:     types.StringValue(fields[i].DateOptions.DateFormat),
				DateFormatType: types.StringValue(fields[i].DateOptions.DateFormatType),
			}
			if fields[i].DateOptions.MinValue != nil {
				field.DateOptions.MinValue = &gdriveLabelDateFieldModel{
					Day:   types.Int64Value(fields[i].DateOptions.MinValue.Day),
					Month: types.Int64Value(fields[i].DateOptions.MinValue.Month),
					Year:  types.Int64Value(fields[i].DateOptions.MinValue.Year),
				}
			}
			if fields[i].DateOptions.MaxValue != nil {
				field.DateOptions.MaxValue = &gdriveLabelDateFieldModel{
					Day:   types.Int64Value(fields[i].DateOptions.MaxValue.Day),
					Month: types.Int64Value(fields[i].DateOptions.MaxValue.Month),
					Year:  types.Int64Value(fields[i].DateOptions.MaxValue.Year),
				}
			}
		}
		model = append(model, field)
	}
	return model
}

func labelAssignmentField() rsschema.SetNestedBlock {
	return rsschema.SetNestedBlock{
		MarkdownDescription: `A field of the assigned label.
This block may be used multiple times to set multiple fields of the same label.`,
		NestedObject: rsschema.NestedBlockObject{
			Attributes: map[string]rsschema.Attribute{
				"field_id": rsschema.StringAttribute{
					Required:    true,
					Description: "The identifier of this field.",
				},
				"value_type": rsschema.StringAttribute{
					Required: true,
					Description: `The field type.
While new values may be supported in the future, the following are currently allowed:
- dateString
- integer
- selection
- text
- user`,
				},
				"values": rsschema.SetAttribute{
					ElementType: types.StringType,
					Required:    true,
					Description: `The values that should be set.
Must be compatible with the specified valueType.`,
				},
			},
		},
	}
}
