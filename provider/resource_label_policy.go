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

	"github.com/hanneshayashi/gsm/gsmdrive"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"google.golang.org/api/drive/v3"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &gdriveLabelPolicyResource{}
var _ resource.ResourceWithImportState = &gdriveLabelPolicyResource{}

func newLabelPolicy() resource.Resource {
	return &gdriveLabelPolicyResource{}
}

// gdriveLabelPolicyResource defines the resource implementation.
type gdriveLabelPolicyResource struct {
	client *http.Client
}

type gdriveLabelPolicyLabelModel struct {
	LabelId types.String             `tfsdk:"label_id"`
	Field   []*gdriveLabelFieldModel `tfsdk:"field"`
}

// gdriveLabelPolicyResourceModel describes the resource data model.
type gdriveLabelPolicyResourceModel struct {
	FileId types.String                   `tfsdk:"file_id"`
	Id     types.String                   `tfsdk:"id"`
	Label  []*gdriveLabelPolicyLabelModel `tfsdk:"label"`
}

func (r *gdriveLabelPolicyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_label_policy"
}

func (r *gdriveLabelPolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the label policy (fileId/labelId)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"label": schema.SetNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"label_id": schema.StringAttribute{
							MarkdownDescription: "The ID of the label.",
							Required:            true,
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
				},
			},
		},
	}
}

func (r *gdriveLabelPolicyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *gdriveLabelPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := &gdriveLabelPolicyResourceModel{}
	resp.Diagnostics.Append(req.Plan.Get(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	mockState := &gdriveLabelPolicyResourceModel{
		FileId: plan.FileId,
	}
	resp.Diagnostics.Append(mockState.getCurrentLabels(ctx)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(setLabelDiffs(plan, mockState, ctx)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Id = plan.FileId
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *gdriveLabelPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	state := &gdriveLabelPolicyResourceModel{}
	resp.Diagnostics.Append(req.State.Get(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(state.getCurrentLabels(ctx)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *gdriveLabelPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := &gdriveLabelPolicyResourceModel{}
	resp.Diagnostics.Append(req.Plan.Get(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state := &gdriveLabelPolicyResourceModel{}
	resp.Diagnostics.Append(req.State.Get(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(setLabelDiffs(plan, state, ctx)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *gdriveLabelPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	state := &gdriveLabelPolicyResourceModel{}
	resp.Diagnostics.Append(req.State.Get(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	modLabelsReq := &drive.ModifyLabelsRequest{
		LabelModifications: []*drive.LabelModification{},
	}
	for i := range state.Label {
		modLabelsReq.LabelModifications = append(modLabelsReq.LabelModifications, &drive.LabelModification{
			LabelId:     state.Label[i].LabelId.ValueString(),
			RemoveLabel: true,
		})
	}
	tflog.Debug(ctx, fmt.Sprintf("Removing all Labels from %s", state.FileId.ValueString()))
	_, err := gsmdrive.ModifyLabels(state.FileId.ValueString(), "", modLabelsReq)
	if err != nil {
		resp.Diagnostics.AddError("Configuration Error", fmt.Sprintf("Unable to remove label assignment(s), got error: %s", err))
		return
	}
}

func (r *gdriveLabelPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// import (
// 	"fmt"

// 	"github.com/hanneshayashi/gsm/gsmdrive"
// 	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
// 	"google.golang.org/api/drive/v3"
// )

// func resourceLabelPolicy() *schema.Resource {
// 	return &schema.Resource{
// 		Description: "Sets the labels on a Drive object",
// 		Schema: map[string]*schema.Schema{
// 			"file_id": {
// 				Type:        schema.TypeString,
// 				Required:    true,
// 				Description: "ID of the file to assign the label policy to.",
// 			},
// 			"label": {
// 				Type:     schema.TypeSet,
// 				Optional: true,
// 				Description: `Represents a single label configuration.
// May be used multiple times to assign multiple labels to the file.
// All labels not configured here will be removed.
// If no labels are defined, all labels will be removed!`,
// 				Elem: &schema.Resource{
// 					Schema: map[string]*schema.Schema{
// 						"label_id": {
// 							Type:        schema.TypeString,
// 							Required:    true,
// 							Description: "The ID of the label.",
// 						},
// 						"field": labelFieldsR(),
// 					},
// 				},
// 			},
// 		},
// 		Create: resourceUpdateLabelPolicy,
// 		Read:   resourceReadLabelPolicy,
// 		Update: resourceUpdateLabelPolicy,
// 		Delete: resourceDeleteLabelPolicy,
// 		Exists: resourceExistsFile,
// 		Importer: &schema.ResourceImporter{
// 			StateContext: schema.ImportStatePassthroughContext,
// 		},
// 	}
// }

// func resourceUpdateLabelPolicy(d *schema.ResourceData, _ any) error {
// 	fileID := d.Get("file_id").(string)
// 	labels := d.Get("label").(*schema.Set).List()
// 	labelsToRemove := getRemovedItemsFromSet(d, "label", "label_id")
// 	req := &drive.ModifyLabelsRequest{
// 		LabelModifications: make([]*drive.LabelModification, len(labels)),
// 	}
// 	for l := range labels {
// 		label := labels[l].(map[string]any)
// 		labelModification, err := getLabelModification(label["label_id"].(string), getRemovedItemsFromSet(d, fmt.Sprintf("label.%d.field", l), "field_id"), label["field"].(*schema.Set))
// 		if err != nil {
// 			return err
// 		}
// 		req.LabelModifications[l] = labelModification
// 	}
// 	for o := range labelsToRemove {
// 		req.LabelModifications = append(req.LabelModifications, &drive.LabelModification{
// 			LabelId:     o,
// 			RemoveLabel: true,
// 		})
// 	}
// 	_, err := gsmdrive.ModifyLabels(fileID, "", req)
// 	if err != nil {
// 		return err
// 	}
// 	d.SetId(fileID)
// 	return nil
// }

// func resourceReadLabelPolicy(d *schema.ResourceData, _ any) error {
// 	labels := make([]map[string]any, 0)
// 	r, err := gsmdrive.ListLabels(d.Id(), "", 1)
// 	for l := range r {
// 		labels = append(labels, map[string]any{
// 			"label_id": l.Id,
// 			"field":    getFields(l.Fields),
// 		})
// 	}
// 	e := <-err
// 	if e != nil {
// 		return e
// 	}
// 	d.Set("label", labels)
// 	return nil
// }

// func resourceDeleteLabelPolicy(d *schema.ResourceData, _ any) error {
// 	labels := d.Get("label").(*schema.Set).List()
// 	req := &drive.ModifyLabelsRequest{
// 		LabelModifications: make([]*drive.LabelModification, len(labels)),
// 	}
// 	for l := range labels {
// 		label := labels[l].(map[string]any)
// 		req.LabelModifications[l] = &drive.LabelModification{
// 			LabelId:     label["label_id"].(string),
// 			RemoveLabel: true,
// 		}
// 	}
// 	_, err := gsmdrive.ModifyLabels(d.Get("file_id").(string), "", req)
// 	return err
// }
