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
	"strconv"

	"github.com/hanneshayashi/gsm/gsmdrive"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"google.golang.org/api/drive/v3"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &gdriveLabelAssignmentResource{}
var _ resource.ResourceWithImportState = &gdriveLabelAssignmentResource{}

func newLabelAssignment() resource.Resource {
	return &gdriveLabelAssignmentResource{}
}

// gdriveLabelAssignmentResource defines the resource implementation.
type gdriveLabelAssignmentResource struct {
	client *http.Client
}

type gdriveLabelAssignmentFieldModel struct {
	FieldId   types.String `tfsdk:"field_id"`
	ValueType types.String `tfsdk:"value_type"`
	Values    types.Set    `tfsdk:"values"`
}

// gdriveLabelAssignmentResourceModel describes the resource data model.
type gdriveLabelAssignmentResourceModel struct {
	FileId  types.String                       `tfsdk:"file_id"`
	LabelId types.String                       `tfsdk:"label_id"`
	Id      types.String                       `tfsdk:"id"`
	Field   []*gdriveLabelAssignmentFieldModel `tfsdk:"field"`
}

func (r *gdriveLabelAssignmentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_label_assignment"
}

func (r *gdriveLabelAssignmentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Sets a label on a Drive object",
		Attributes: map[string]schema.Attribute{
			"file_id": schema.StringAttribute{
				MarkdownDescription: "ID of the file to assign the label to.",
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
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the label assignment (fileId/labelId)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"field": schema.SetNestedBlock{
				MarkdownDescription: `A field of the assigned label.
This block may be used multiple times to set multiple fields of the same label.`,
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"field_id": schema.StringAttribute{
							Required:    true,
							Description: "The identifier of this field.",
						},
						"value_type": schema.StringAttribute{
							Required: true,
							Description: `The field type.
While new values may be supported in the future, the following are currently allowed:
- dateString
- integer
- selection
- text
- user`,
						},
						"values": schema.SetAttribute{
							ElementType: types.StringType,
							Required:    true,
							Description: `The values that should be set.
Must be compatible with the specified valueType.`,
						},
					},
				},
			},
		},
	}
}

func (r *gdriveLabelAssignmentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *gdriveLabelAssignmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := &gdriveLabelAssignmentResourceModel{}
	diags := req.Plan.Get(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	labelID := plan.LabelId.ValueString()
	fileID := plan.FileId.ValueString()
	modLabelsReq := &drive.ModifyLabelsRequest{
		LabelModifications: []*drive.LabelModification{
			{
				LabelId:            labelID,
				FieldModifications: []*drive.LabelFieldModification{},
			},
		},
	}
	for i := range plan.Field {
		valueType := plan.Field[i].ValueType.ValueString()
		fieldMod := &drive.LabelFieldModification{
			FieldId: plan.Field[i].FieldId.ValueString(),
		}
		switch valueType {
		case "dateString":
			diags = plan.Field[i].Values.ElementsAs(ctx, &fieldMod.SetDateValues, false)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
		case "integer":
			values := []string{}
			diags = plan.Field[i].Values.ElementsAs(ctx, &values, false)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			for v := range values {
				vi, err := strconv.ParseInt(values[v], 10, 64)
				if err != nil {
					resp.Diagnostics.AddError("Configuration Error", fmt.Sprintf("Unable to use %s as a value for an integer field, got error: %s", values[v], err))
					return
				}
				fieldMod.SetIntegerValues = append(fieldMod.SetIntegerValues, vi)
			}
		case "selection":
			diags = plan.Field[i].Values.ElementsAs(ctx, &fieldMod.SetSelectionValues, false)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
		case "text":
			diags = plan.Field[i].Values.ElementsAs(ctx, &fieldMod.SetTextValues, false)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
		case "user":
			diags = plan.Field[i].Values.ElementsAs(ctx, &fieldMod.SetUserValues, false)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
		case "default":
			resp.Diagnostics.AddError("Configuration Error", fmt.Sprintf("Unable to use %s as a value_type for field", valueType))
			return
		}
		modLabelsReq.LabelModifications[0].FieldModifications = append(modLabelsReq.LabelModifications[0].FieldModifications, fieldMod)
	}
	_, err := gsmdrive.ModifyLabels(fileID, "", modLabelsReq)
	if err != nil {
		resp.Diagnostics.AddError("Configuration Error", fmt.Sprintf("Unable to create label assignment, got error: %s", err))
		return
	}
	plan.Id = types.StringValue(combineId(fileID, labelID))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *gdriveLabelAssignmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	state := &gdriveLabelAssignmentResourceModel{}
	diags := req.State.Get(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	fileID := state.FileId.ValueString()
	labelID := state.LabelId.ValueString()
	state.Field = []*gdriveLabelAssignmentFieldModel{}
	currentLabels, err := gsmdrive.ListLabels(fileID, "", 1)
	for l := range currentLabels {
		if l.Id == labelID {
			for f := range l.Fields {
				field := &gdriveLabelAssignmentFieldModel{
					ValueType: types.StringValue(l.Fields[f].ValueType),
					FieldId:   types.StringValue(l.Fields[f].Id),
				}
				switch l.Fields[f].ValueType {
				case "dateString":
					field.Values, diags = types.SetValueFrom(ctx, types.StringType, l.Fields[f].DateString)
					resp.Diagnostics.Append(diags...)
					if resp.Diagnostics.HasError() {
						return
					}
				case "text":
					field.Values, diags = types.SetValueFrom(ctx, types.StringType, l.Fields[f].Text)
					resp.Diagnostics.Append(diags...)
					if resp.Diagnostics.HasError() {
						return
					}
				case "user":
					values := []string{}
					for u := range l.Fields[f].User {
						values = append(values, l.Fields[f].User[u].EmailAddress)
					}
					field.Values, diags = types.SetValueFrom(ctx, types.StringType, values)
					resp.Diagnostics.Append(diags...)
					if resp.Diagnostics.HasError() {
						return
					}
				case "selection":

					field.Values, diags = types.SetValueFrom(ctx, types.StringType, l.Fields[f].Selection)
					resp.Diagnostics.Append(diags...)
					if resp.Diagnostics.HasError() {
						return
					}
				case "integer":
					values := []string{}
					for _, value := range l.Fields[f].Integer {
						values = append(values, strconv.FormatInt(value, 10))
					}
					field.Values, diags = types.SetValueFrom(ctx, types.StringType, values)
					resp.Diagnostics.Append(diags...)
					if resp.Diagnostics.HasError() {
						return
					}
				}
				state.Field = append(state.Field, field)
			}
		}
	}
	e := <-err
	if e != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to use list labels on file, got error: %s", e))
		return
	}
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (label *gdriveLabelAssignmentResourceModel) toMap() map[string]*gdriveLabelAssignmentFieldModel {
	m := map[string]*gdriveLabelAssignmentFieldModel{}
	for i := range label.Field {
		m[label.Field[i].FieldId.ValueString()] = label.Field[i]
	}
	return m
}

func (r *gdriveLabelAssignmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var diags diag.Diagnostics
	plan := &gdriveLabelAssignmentResourceModel{}
	resp.Diagnostics.Append(req.Plan.Get(ctx, plan)...)
	state := &gdriveLabelAssignmentResourceModel{}
	resp.Diagnostics.Append(req.State.Get(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
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
			fieldMod := &drive.LabelFieldModification{
				FieldId: i,
			}
			if planMap[i].Values.IsNull() {
				fieldMod.UnsetValues = true
			} else {
				valueType := planMap[i].ValueType.ValueString()
				switch valueType {
				case "dateString":
					diags = planMap[i].Values.ElementsAs(ctx, &fieldMod.SetDateValues, false)
					resp.Diagnostics.Append(diags...)
					if resp.Diagnostics.HasError() {
						return
					}
				case "integer":
					values := []string{}
					diags = planMap[i].Values.ElementsAs(ctx, &values, false)
					resp.Diagnostics.Append(diags...)
					if resp.Diagnostics.HasError() {
						return
					}
					for v := range values {
						vi, err := strconv.ParseInt(values[v], 10, 64)
						if err != nil {
							resp.Diagnostics.AddError("Configuration Error", fmt.Sprintf("Unable to use %s as a value for an integer field, got error: %s", values[v], err))
							return
						}
						fieldMod.SetIntegerValues = append(fieldMod.SetIntegerValues, vi)
					}
				case "selection":
					diags = planMap[i].Values.ElementsAs(ctx, &fieldMod.SetSelectionValues, false)
					resp.Diagnostics.Append(diags...)
					if resp.Diagnostics.HasError() {
						return
					}
				case "text":
					diags = planMap[i].Values.ElementsAs(ctx, &fieldMod.SetTextValues, false)
					resp.Diagnostics.Append(diags...)
					if resp.Diagnostics.HasError() {
						return
					}
				case "user":
					diags = planMap[i].Values.ElementsAs(ctx, &fieldMod.SetUserValues, false)
					resp.Diagnostics.Append(diags...)
					if resp.Diagnostics.HasError() {
						return
					}
				case "default":
					resp.Diagnostics.AddError("Configuration Error", fmt.Sprintf("Unable to use %s as a value_type for field", valueType))
					return
				}
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
		resp.Diagnostics.AddError("Configuration Error", fmt.Sprintf("Unable to update label assignment, got error: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *gdriveLabelAssignmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	state := &gdriveLabelAssignmentResourceModel{}
	resp.Diagnostics.Append(req.State.Get(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, err := gsmdrive.ModifyLabels(state.FileId.ValueString(), "", &drive.ModifyLabelsRequest{
		LabelModifications: []*drive.LabelModification{
			{
				LabelId:     state.LabelId.ValueString(),
				RemoveLabel: true,
			}},
	})
	if err != nil {
		resp.Diagnostics.AddError("Configuration Error", fmt.Sprintf("Unable to remove label assignment, got error: %s", err))
		return
	}
}

func (r *gdriveLabelAssignmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
