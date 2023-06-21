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

	"github.com/hanneshayashi/gsm/gsmdrivelabels"
	"github.com/hanneshayashi/gsm/gsmhelpers"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	rsschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"google.golang.org/api/drivelabels/v2"
)

type gdriveLabelFieldPropertieseModel struct {
	DisplayName       types.String `tfsdk:"display_name"`
	InsertBeforeField types.String `tfsdk:"insert_before_field"`
	Required          types.Bool   `tfsdk:"required"`
}

func (disablePolicyModel *gdriveLabelLifeCycleDisabledPolicyModel) populate(disablePolicy *drivelabels.GoogleAppsDriveLabelsV2LifecycleDisabledPolicy) {
	if disablePolicy != nil {
		disablePolicyModel.HideInSearch = types.BoolValue(disablePolicy.HideInSearch)
		disablePolicyModel.ShowInApply = types.BoolValue(disablePolicy.ShowInApply)
	}
}

func (lifeCycleModel *gdriveLabelLifeCycleModel) populate(lifecycle *drivelabels.GoogleAppsDriveLabelsV2Lifecycle) {
	if lifecycle != nil {
		var state string
		if lifecycle.State == "PUBLISHED" && lifecycle.HasUnpublishedChanges {
			state = "PUBLISHED_WITH_PENDING_CHANGES"
		} else {
			state = lifecycle.State
		}
		lifeCycleModel.State = types.StringValue(state)
		if lifeCycleModel.DisabledPolicy != nil {
			lifeCycleModel.DisabledPolicy.populate(lifecycle.DisabledPolicy)
		}
	}
}

func (lifeCycleModel *gdriveLabelLifeCycleDSModel) populate(lifecycle *drivelabels.GoogleAppsDriveLabelsV2Lifecycle) {
	if lifecycle != nil {
		lifeCycleModel.State = types.StringValue(lifecycle.State)
		lifeCycleModel.HasUnpublishedChanges = types.BoolValue(lifecycle.HasUnpublishedChanges)
		if lifeCycleModel.DisabledPolicy != nil {
			lifeCycleModel.DisabledPolicy.populate(lifecycle.DisabledPolicy)
		}
	}
}

func (lifecycleModel *gdriveLabelLifeCycleDSModel) toLifecycle() (lifecycle *drivelabels.GoogleAppsDriveLabelsV2Lifecycle) {
	lifecycle = &drivelabels.GoogleAppsDriveLabelsV2Lifecycle{
		State:                 lifecycleModel.State.ValueString(),
		HasUnpublishedChanges: lifecycleModel.HasUnpublishedChanges.ValueBool(),
	}
	if lifecycleModel.DisabledPolicy != nil {
		lifecycle.DisabledPolicy = &drivelabels.GoogleAppsDriveLabelsV2LifecycleDisabledPolicy{
			HideInSearch: lifecycleModel.DisabledPolicy.HideInSearch.ValueBool(),
			ShowInApply:  lifecycleModel.DisabledPolicy.ShowInApply.ValueBool(),
		}
		if !lifecycle.DisabledPolicy.HideInSearch {
			lifecycle.DisabledPolicy.ForceSendFields = append(lifecycle.DisabledPolicy.ForceSendFields, "HideInSearch")
		}
		if !lifecycle.DisabledPolicy.ShowInApply {
			lifecycle.DisabledPolicy.ForceSendFields = append(lifecycle.DisabledPolicy.ForceSendFields, "ShowInApply")
		}
	}
	return
}

func (lifecycleModel *gdriveLabelLifeCycleModel) toLifecycle() (lifecycle *drivelabels.GoogleAppsDriveLabelsV2Lifecycle) {
	lifecycle = &drivelabels.GoogleAppsDriveLabelsV2Lifecycle{
		State: lifecycleModel.State.ValueString(),
	}
	if lifecycleModel.DisabledPolicy != nil {
		lifecycle.DisabledPolicy = &drivelabels.GoogleAppsDriveLabelsV2LifecycleDisabledPolicy{
			HideInSearch: lifecycleModel.DisabledPolicy.HideInSearch.ValueBool(),
			ShowInApply:  lifecycleModel.DisabledPolicy.ShowInApply.ValueBool(),
		}
		if !lifecycle.DisabledPolicy.HideInSearch {
			lifecycle.DisabledPolicy.ForceSendFields = append(lifecycle.DisabledPolicy.ForceSendFields, "HideInSearch")
		}
		if !lifecycle.DisabledPolicy.ShowInApply {
			lifecycle.DisabledPolicy.ForceSendFields = append(lifecycle.DisabledPolicy.ForceSendFields, "ShowInApply")
		}
	}
	return
}

type labelInterface interface {
	getLanguageCode() string
	getUseAdminAccess() bool
}

func (labelModel *gdriveLabelResourceModel) getLanguageCode() string {
	return labelModel.LanguageCode.ValueString()
}

func (LabelModel *gdriveLabelResourceModel) getUseAdminAccess() bool {
	return LabelModel.UseAdminAccess.ValueBool()
}

func (labelModel *gdriveLabelResourceModel) populate(ctx context.Context) (diags diag.Diagnostics) {
	l, err := gsmdrivelabels.GetLabel(gsmhelpers.EnsurePrefix(labelModel.Id.ValueString(), "labels/"), labelModel.LanguageCode.ValueString(), "LABEL_VIEW_FULL", "*", labelModel.UseAdminAccess.ValueBool())
	if err != nil {
		diags.AddError("Client Error", fmt.Sprintf("Unable to get label, got error: %s", err))
		return
	}
	labelModel.LabelId = types.StringValue(l.Id)
	labelModel.Name = types.StringValue(l.Name)
	labelModel.LabelType = types.StringValue(l.LabelType)
	labelModel.Properties = &gdriveLabelResourcePropertiesModel{}
	labelModel.Properties.populate(l.Properties)
	if labelModel.LifeCycle != nil {
		labelModel.LifeCycle.populate(l.Lifecycle)
	}
	return
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

func lifeCycleRS() rsschema.SingleNestedBlock {
	return rsschema.SingleNestedBlock{
		MarkdownDescription: `The lifecycle state of an object, such as label, field, or choice.

The lifecycle enforces the following transitions:
UNPUBLISHED_DRAFT (starting state)
UNPUBLISHED_DRAFT -> PUBLISHED
UNPUBLISHED_DRAFT -> (Deleted)
PUBLISHED -> DISABLED
DISABLED -> PUBLISHED
DISABLED -> (Deleted)`,
		Attributes: map[string]rsschema.Attribute{
			"state": rsschema.StringAttribute{
				MarkdownDescription: "The state of the object associated with this lifecycle.",
				Optional:            true,
			},
		},
		Blocks: map[string]rsschema.Block{
			"disabled_policy": rsschema.SingleNestedBlock{
				MarkdownDescription: "The policy that governs how to show a disabled label, field, or selection choice.",
				Attributes: map[string]rsschema.Attribute{
					"hide_in_search": rsschema.BoolAttribute{
						MarkdownDescription: `Whether to hide this disabled object in the search menu for Drive items.

When false, the object is generally shown in the UI as disabled but it appears in the search results when searching for Drive items.
When true, the object is generally hidden in the UI when searching for Drive items.`,
						Optional: true,
						Computed: true,
						Default:  booldefault.StaticBool(false),
					},
					"show_in_apply": rsschema.BoolAttribute{
						MarkdownDescription: `Whether to show this disabled object in the apply menu on Drive items.

When true, the object is generally shown in the UI as disabled and is unselectable.
When false, the object is generally hidden in the UI.`,
						Optional: true,
						Computed: true,
						Default:  booldefault.StaticBool(true),
					},
				},
			},
		},
	}
}

func lifecycleDS() dsschema.SingleNestedBlock {
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
				MarkdownDescription: "The state of the object associated with this lifecycle.",
				Computed:            true,
			},
			"has_unpublished_changes": dsschema.BoolAttribute{
				MarkdownDescription: "Whether the object associated with this lifecycle has unpublished changes.",
				Computed:            true,
			},
		},
		Blocks: map[string]dsschema.Block{
			"disabled_policy": dsschema.SingleNestedBlock{
				MarkdownDescription: "The policy that governs how to show a disabled label, field, or selection choice.",
				Attributes: map[string]dsschema.Attribute{
					"hide_in_search": dsschema.BoolAttribute{
						MarkdownDescription: `Whether to hide this disabled object in the search menu for Drive items.

When false, the object is generally shown in the UI as disabled but it appears in the search results when searching for Drive items.
When true, the object is generally hidden in the UI when searching for Drive items.`,
						Computed: true,
					},
					"show_in_apply": dsschema.BoolAttribute{
						MarkdownDescription: `Whether to show this disabled object in the apply menu on Drive items.

When true, the object is generally shown in the UI as disabled and is unselectable.
When false, the object is generally hidden in the UI.`,
						Computed: true,
					},
				},
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
				"life_cycle": lifecycleDS(),
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
						"choices": dsschema.ListNestedBlock{
							NestedObject: dsschema.NestedBlockObject{
								Attributes: map[string]dsschema.Attribute{
									"id": dsId(),
									"choice_id": dsschema.StringAttribute{
										Computed: true,
										Description: `The unique value of the choice.
											Use this when referencing / setting a choice.`,
									},
								},
								Blocks: map[string]dsschema.Block{
									"life_cycle": lifecycleDS(),
									"properties": dsschema.SingleNestedBlock{
										MarkdownDescription: "Basic properties of the choice.",
										Attributes: map[string]dsschema.Attribute{
											"display_name": dsschema.StringAttribute{
												MarkdownDescription: "The display text to show in the UI identifying this choice.",
												Required:            true,
											},
										},
										Blocks: map[string]dsschema.Block{
											"badge_config": dsschema.SingleNestedBlock{
												Attributes: map[string]dsschema.Attribute{
													"priority_override": dsschema.Int64Attribute{
														MarkdownDescription: `Override the default global priority of this badge.
When set to 0, the default priority heuristic is used.`,
														Computed: true,
													},
												},
												Blocks: map[string]dsschema.Block{
													"color": dsschema.SingleNestedBlock{
														MarkdownDescription: `The color of the badge.
When not specified, no badge is rendered.
The background, foreground, and solo (light and dark mode) colors set here are changed in the Drive UI into the closest recommended supported color.

*After setting this property, the plan will likely show a difference, because the API automatically modifies the values.
It is recommended to change the Terraform configuration to match the values set by the API.`,
														Attributes: map[string]dsschema.Attribute{
															"alpha": dsschema.Float64Attribute{
																MarkdownDescription: `The alpha value for the badge color as a float (number between 1 and 0 - e.g. "0.5")`,
																Computed:            true,
															},
															"blue": dsschema.Float64Attribute{
																MarkdownDescription: `The blue value for the badge color as a float (number between 1 and 0 - e.g. "0.5")`,
																Computed:            true,
															},
															"green": dsschema.Float64Attribute{
																MarkdownDescription: `The green value for the badge color as a float (number between 1 and 0 - e.g. "0.5")`,
																Computed:            true,
															},
															"red": dsschema.Float64Attribute{
																MarkdownDescription: `The red value for the badge color as a float (number between 1 and 0 - e.g. "0.5")`,
																Computed:            true,
															},
														},
													},
												},
											},
										},
									},
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
			Id:        types.StringValue(fields[i].Id),
			FieldId:   types.StringValue(fields[i].Id),
			QueryKey:  types.StringValue(fields[i].QueryKey),
			LifeCycle: &gdriveLabelLifeCycleDSModel{},
		}
		if fields[i].Properties != nil {
			field.Properties = &gdriveLabelDataSourceFieldPropertieseModel{
				DisplayName: types.StringValue(fields[i].Properties.DisplayName),
				Required:    types.BoolValue(fields[i].Properties.Required),
			}
		}
		field.LifeCycle.populate(fields[i].Lifecycle)
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
			field.SelectionOptions.Choices = make([]*gdriveLabelChoiceModel, len(fields[i].SelectionOptions.Choices))
			for j := range fields[i].SelectionOptions.Choices {
				field.SelectionOptions.Choices[j] = &gdriveLabelChoiceModel{
					Id:        types.StringValue(fields[i].SelectionOptions.Choices[j].Id),
					ChoiceId:  types.StringValue(fields[i].SelectionOptions.Choices[j].Id),
					LifeCycle: &gdriveLabelLifeCycleDSModel{},
				}
				if fields[i].SelectionOptions.Choices[j].Properties != nil {
					field.SelectionOptions.Choices[j].Properties = &gdriveLabelChoicePropertiesModel{
						DisplayName: types.StringValue(fields[i].SelectionOptions.Choices[j].Properties.DisplayName),
					}
					if fields[i].SelectionOptions.Choices[j].Properties.BadgeConfig != nil {
						field.SelectionOptions.Choices[j].Properties.BadgeConfig = &gdriveLabelChoiceBadgeConfigModel{}
						field.SelectionOptions.Choices[j].Properties.BadgeConfig.populate(fields[i].SelectionOptions.Choices[j].Properties.BadgeConfig)
					}
				}
				field.SelectionOptions.Choices[j].LifeCycle.populate(fields[i].SelectionOptions.Choices[j].Lifecycle)
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

func (propertiesModel *gdriveLabelFieldPropertieseModel) toProperties() *drivelabels.GoogleAppsDriveLabelsV2FieldProperties {
	properties := &drivelabels.GoogleAppsDriveLabelsV2FieldProperties{
		DisplayName:       propertiesModel.DisplayName.ValueString(),
		InsertBeforeField: propertiesModel.InsertBeforeField.ValueString(),
		Required:          propertiesModel.Required.ValueBool(),
	}
	if properties.DisplayName == "" {
		properties.ForceSendFields = append(properties.ForceSendFields, "DisplayName")
	}
	if properties.InsertBeforeField == "" {
		properties.ForceSendFields = append(properties.ForceSendFields, "InsertBeforeField")
	}
	if !properties.Required {
		properties.ForceSendFields = append(properties.ForceSendFields, "Required")
	}
	return properties
}

func (plan *gdriveLabelDateFieldResourceModel) toUpateRequest(state *gdriveLabelDateFieldResourceModel) (request *drivelabels.GoogleAppsDriveLabelsV2DeltaUpdateLabelRequestRequest) {
	if !plan.DateOptions.DateFormatType.Equal(state.DateOptions.DateFormatType) {
		request = &drivelabels.GoogleAppsDriveLabelsV2DeltaUpdateLabelRequestRequest{
			UpdateFieldType: &drivelabels.GoogleAppsDriveLabelsV2DeltaUpdateLabelRequestUpdateFieldTypeRequest{
				Id: plan.FieldId.ValueString(),
				DateOptions: &drivelabels.GoogleAppsDriveLabelsV2FieldDateOptions{
					DateFormatType: plan.DateOptions.DateFormatType.ValueString(),
				},
			},
		}
	}
	return
}

func newUpdateLabelRequest(plan labelInterface) *drivelabels.GoogleAppsDriveLabelsV2DeltaUpdateLabelRequest {
	return &drivelabels.GoogleAppsDriveLabelsV2DeltaUpdateLabelRequest{
		LanguageCode:   plan.getLanguageCode(),
		UseAdminAccess: plan.getUseAdminAccess(),
		View:           "LABEL_VIEW_FULL",
	}
}
