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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"google.golang.org/api/drivelabels/v2"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &gdriveLabelSelectionChoiceResource{}
var _ resource.ResourceWithImportState = &gdriveLabelSelectionChoiceResource{}

func newLabelSelectionChoice() resource.Resource {
	return &gdriveLabelSelectionChoiceResource{}
}

// gdriveLabelSelectionChoiceResource defines the resource implementation.
type gdriveLabelSelectionChoiceResource struct {
	client *http.Client
}

type gdriveLabelChoiceBadgeColorConfigRSModel struct {
	Red   types.Float64 `tfsdk:"red"`
	Green types.Float64 `tfsdk:"green"`
	Blue  types.Float64 `tfsdk:"blue"`
	Alpha types.Float64 `tfsdk:"alpha"`
}

type gdriveLabelChoiceBadgeConfigRSModel struct {
	Color            *gdriveLabelChoiceBadgeColorConfigRSModel `tfsdk:"color"`
	PriorityOverride types.Int64                               `tfsdk:"priority_override"`
}

type gdriveLabelChoicePropertiesRSModel struct {
	BadgeConfig        *gdriveLabelChoiceBadgeConfigRSModel `tfsdk:"badge_config"`
	DisplayName        types.String                         `tfsdk:"display_name"`
	InsertBeforeChoice types.String                         `tfsdk:"insert_before_choice"`
}

type gdriveLabelSelectionChoiceResourceModel struct {
	Properties     *gdriveLabelChoicePropertiesRSModel `tfsdk:"properties"`
	Id             types.String                        `tfsdk:"id"`
	ChoiceId       types.String                        `tfsdk:"choice_id"`
	FieldId        types.String                        `tfsdk:"field_id"`
	LabelId        types.String                        `tfsdk:"label_id"`
	LanguageCode   types.String                        `tfsdk:"language_code"`
	UseAdminAccess types.Bool                          `tfsdk:"use_admin_access"`
}

func (choiceModel *gdriveLabelSelectionChoiceResourceModel) toChoice() (choice *drivelabels.GoogleAppsDriveLabelsV2FieldSelectionOptionsChoice) {
	choice = &drivelabels.GoogleAppsDriveLabelsV2FieldSelectionOptionsChoice{
		Properties: &drivelabels.GoogleAppsDriveLabelsV2FieldSelectionOptionsChoiceProperties{
			DisplayName:        choiceModel.Properties.DisplayName.ValueString(),
			InsertBeforeChoice: choiceModel.Properties.InsertBeforeChoice.ValueString(),
		},
	}
	if choiceModel.Properties.BadgeConfig != nil {
		choice.Properties.BadgeConfig = &drivelabels.GoogleAppsDriveLabelsV2BadgeConfig{
			PriorityOverride: choiceModel.Properties.BadgeConfig.PriorityOverride.ValueInt64(),
		}
		if choice.Properties.BadgeConfig.PriorityOverride == 0 {
			choice.Properties.BadgeConfig.ForceSendFields = append(choice.Properties.BadgeConfig.ForceSendFields, "PriorityOverride")
		}
		if choiceModel.Properties.BadgeConfig.Color != nil {
			choice.Properties.BadgeConfig.Color = &drivelabels.GoogleTypeColor{
				Red:   choiceModel.Properties.BadgeConfig.Color.Red.ValueFloat64(),
				Green: choiceModel.Properties.BadgeConfig.Color.Green.ValueFloat64(),
				Blue:  choiceModel.Properties.BadgeConfig.Color.Blue.ValueFloat64(),
				Alpha: choiceModel.Properties.BadgeConfig.Color.Alpha.ValueFloat64(),
			}
			if choice.Properties.BadgeConfig.Color.Red == 0 {
				choice.Properties.BadgeConfig.Color.ForceSendFields = append(choice.Properties.BadgeConfig.Color.ForceSendFields, "Red")
			}
			if choice.Properties.BadgeConfig.Color.Green == 0 {
				choice.Properties.BadgeConfig.Color.ForceSendFields = append(choice.Properties.BadgeConfig.Color.ForceSendFields, "Green")
			}
			if choice.Properties.BadgeConfig.Color.Blue == 0 {
				choice.Properties.BadgeConfig.Color.ForceSendFields = append(choice.Properties.BadgeConfig.Color.ForceSendFields, "Blue")
			}
			if choice.Properties.BadgeConfig.Color.Alpha == 0 {
				choice.Properties.BadgeConfig.Color.ForceSendFields = append(choice.Properties.BadgeConfig.Color.ForceSendFields, "Alpha")
			}
		}
	}
	return
}

func (choiceModel *gdriveLabelSelectionChoiceResourceModel) getLanguageCode() string {
	return choiceModel.LanguageCode.ValueString()
}

func (choiceModel *gdriveLabelSelectionChoiceResourceModel) getUseAdminAccess() bool {
	return choiceModel.UseAdminAccess.ValueBool()
}

func (r *gdriveLabelSelectionChoiceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_label_selection_choice"
}

func (r *gdriveLabelSelectionChoiceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Creates an Selection Field for a Drive Label",
		Attributes: map[string]schema.Attribute{
			"id": rsId(),
			"choice_id": schema.StringAttribute{
				MarkdownDescription: "The unique value of the choice. This ID is autogenerated. Matches the regex: ([a-zA-Z0-9_])+.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"field_id": schema.StringAttribute{
				MarkdownDescription: `The ID of the field.`,
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"label_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the label.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
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
		},
		Blocks: map[string]schema.Block{
			"properties": schema.SingleNestedBlock{
				MarkdownDescription: "Basic properties of the choice.",
				Attributes: map[string]schema.Attribute{
					"display_name": schema.StringAttribute{
						MarkdownDescription: "The display text to show in the UI identifying this choice.",
						Required:            true,
					},
					"insert_before_choice": schema.StringAttribute{
						MarkdownDescription: "Insert or move this choice before the indicated choice. If empty, the choice is placed at the end of the list.",
						Optional:            true,
					},
				},
				Blocks: map[string]schema.Block{
					"badge_config": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"priority_override": schema.Int64Attribute{
								MarkdownDescription: `Override the default global priority of this badge.
When set to 0, the default priority heuristic is used.`,
								Optional: true,
								Computed: true,
								Default:  int64default.StaticInt64(0),
							},
						},
						Blocks: map[string]schema.Block{
							"color": schema.SingleNestedBlock{
								MarkdownDescription: `The color of the badge.
When not specified, no badge is rendered.
The background, foreground, and solo (light and dark mode) colors set here are changed in the Drive UI into the closest recommended supported color.

*After setting this property, the plan will likely show a difference, because the API automatically modifies the values.
It is recommended to change the Terraform configuration to match the values set by the API.`,
								Attributes: map[string]schema.Attribute{
									"alpha": schema.Float64Attribute{
										MarkdownDescription: `The alpha value for the badge color as a float (number between 1 and 0 - e.g. "0.5")`,
										Optional:            true,
										Computed:            true,
										Default:             float64default.StaticFloat64(0),
									},
									"blue": schema.Float64Attribute{
										MarkdownDescription: `The blue value for the badge color as a float (number between 1 and 0 - e.g. "0.5")`,
										Optional:            true,
										Computed:            true,
										Default:             float64default.StaticFloat64(0),
									},
									"green": schema.Float64Attribute{
										MarkdownDescription: `The green value for the badge color as a float (number between 1 and 0 - e.g. "0.5")`,
										Optional:            true,
										Computed:            true,
										Default:             float64default.StaticFloat64(0),
									},
									"red": schema.Float64Attribute{
										MarkdownDescription: `The red value for the badge color as a float (number between 1 and 0 - e.g. "0.5")`,
										Optional:            true,
										Computed:            true,
										Default:             float64default.StaticFloat64(0),
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *gdriveLabelSelectionChoiceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *gdriveLabelSelectionChoiceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := &gdriveLabelSelectionChoiceResourceModel{}
	resp.Diagnostics.Append(req.Plan.Get(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	updateLabelRequest := newUpdateLabelRequest(plan)
	updateLabelRequest.Requests = []*drivelabels.GoogleAppsDriveLabelsV2DeltaUpdateLabelRequestRequest{
		{
			CreateSelectionChoice: &drivelabels.GoogleAppsDriveLabelsV2DeltaUpdateLabelRequestCreateSelectionChoiceRequest{
				FieldId: plan.FieldId.ValueString(),
				Choice:  plan.toChoice(),
			},
		},
	}
	updatedLabel, err := gsmdrivelabels.Delta(gsmhelpers.EnsurePrefix(plan.LabelId.ValueString(), "labels/"), "*", updateLabelRequest)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create selection choice, got error: %s", err))
		return
	}
	var newChoice *drivelabels.GoogleAppsDriveLabelsV2FieldSelectionOptionsChoice
	for i := range updatedLabel.UpdatedLabel.Fields {
		if updatedLabel.UpdatedLabel.Fields[i].Id == updateLabelRequest.Requests[0].CreateSelectionChoice.FieldId {
			if updateLabelRequest.Requests[0].CreateSelectionChoice.Choice.Properties.InsertBeforeChoice == "" {
				newChoice = updatedLabel.UpdatedLabel.Fields[i].SelectionOptions.Choices[len(updatedLabel.UpdatedLabel.Fields[i].SelectionOptions.Choices)-1]
			} else {
				for j := range updatedLabel.UpdatedLabel.Fields[i].SelectionOptions.Choices {
					if updatedLabel.UpdatedLabel.Fields[i].SelectionOptions.Choices[j].Id == updateLabelRequest.Requests[0].CreateSelectionChoice.Choice.Properties.InsertBeforeChoice {
						newChoice = updatedLabel.UpdatedLabel.Fields[i].SelectionOptions.Choices[j-1]
						break
					}
				}
			}
		}
	}
	plan.Id = types.StringValue(newChoice.Id)
	plan.ChoiceId = plan.Id
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *gdriveLabelSelectionChoiceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	state := &gdriveLabelSelectionChoiceResourceModel{}
	resp.Diagnostics.Append(req.State.Get(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	l, err := gsmdrivelabels.GetLabel(gsmhelpers.EnsurePrefix(state.LabelId.ValueString(), "labels/"), state.LanguageCode.ValueString(), "LABEL_VIEW_FULL", "*", state.UseAdminAccess.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get label field, got error: %s", err))
		return
	}
	fieldId := state.FieldId.ValueString()
	choiceId := state.ChoiceId.ValueString()
	for i := range l.Fields {
		if l.Fields[i].Id == fieldId {
			if l.Fields[i].SelectionOptions != nil {
				for j := range l.Fields[i].SelectionOptions.Choices {
					if l.Fields[i].SelectionOptions.Choices[j].Id == choiceId {
						if l.Fields[i].SelectionOptions.Choices[j].Properties != nil {
							state.Properties = &gdriveLabelChoicePropertiesRSModel{
								DisplayName: types.StringValue(l.Fields[i].SelectionOptions.Choices[j].Properties.DisplayName),
							}
							if l.Fields[i].SelectionOptions.Choices[j].Properties.BadgeConfig != nil {
								state.Properties.BadgeConfig = &gdriveLabelChoiceBadgeConfigRSModel{
									PriorityOverride: types.Int64Value(l.Fields[i].SelectionOptions.Choices[j].Properties.BadgeConfig.PriorityOverride),
								}
								if l.Fields[i].SelectionOptions.Choices[j].Properties.BadgeConfig.Color != nil {
									state.Properties.BadgeConfig.Color = &gdriveLabelChoiceBadgeColorConfigRSModel{
										Red:   types.Float64Value(l.Fields[i].SelectionOptions.Choices[j].Properties.BadgeConfig.Color.Red),
										Green: types.Float64Value(l.Fields[i].SelectionOptions.Choices[j].Properties.BadgeConfig.Color.Green),
										Blue:  types.Float64Value(l.Fields[i].SelectionOptions.Choices[j].Properties.BadgeConfig.Color.Blue),
										Alpha: types.Float64Value(l.Fields[i].SelectionOptions.Choices[j].Properties.BadgeConfig.Color.Alpha),
									}
								}
							}
							if j < len(l.Fields[i].SelectionOptions.Choices)-1 {
								state.Properties.InsertBeforeChoice = types.StringValue(l.Fields[i].SelectionOptions.Choices[j+1].Id)
							}
						}
					}
				}
			}
			break
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *gdriveLabelSelectionChoiceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := &gdriveLabelSelectionChoiceResourceModel{}
	resp.Diagnostics.Append(req.Plan.Get(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state := &gdriveLabelSelectionChoiceResourceModel{}
	resp.Diagnostics.Append(req.State.Get(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	updateLabelRequest := newUpdateLabelRequest(plan)
	updateLabelRequest.Requests = []*drivelabels.GoogleAppsDriveLabelsV2DeltaUpdateLabelRequestRequest{
		{
			UpdateSelectionChoiceProperties: &drivelabels.GoogleAppsDriveLabelsV2DeltaUpdateLabelRequestUpdateSelectionChoicePropertiesRequest{
				Id:         plan.Id.ValueString(),
				FieldId:    plan.FieldId.ValueString(),
				Properties: plan.toChoice().Properties,
			},
		},
	}
	_, err := gsmdrivelabels.Delta(gsmhelpers.EnsurePrefix(plan.LabelId.ValueString(), "labels/"), "*", updateLabelRequest)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update selection choice, got error: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *gdriveLabelSelectionChoiceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	state := &gdriveLabelSelectionChoiceResourceModel{}
	resp.Diagnostics.Append(req.State.Get(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	updateLabelRequest := newUpdateLabelRequest(state)
	updateLabelRequest.Requests = []*drivelabels.GoogleAppsDriveLabelsV2DeltaUpdateLabelRequestRequest{
		{
			DeleteSelectionChoice: &drivelabels.GoogleAppsDriveLabelsV2DeltaUpdateLabelRequestDeleteSelectionChoiceRequest{
				Id:      state.Id.ValueString(),
				FieldId: state.FieldId.ValueString(),
			},
		},
	}
	_, err := gsmdrivelabels.Delta(gsmhelpers.EnsurePrefix(state.LabelId.ValueString(), "labels/"), "*", updateLabelRequest)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete selection choice, got error: %s", err))
		return
	}
}

func (r *gdriveLabelSelectionChoiceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
