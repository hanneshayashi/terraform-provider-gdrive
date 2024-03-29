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
	Fields  []*gdriveLabelFieldModel `tfsdk:"fields"`
}

// gdriveLabelPolicyResourceModel describes the resource data model.
type gdriveLabelPolicyResourceModel struct {
	FileId types.String                   `tfsdk:"file_id"`
	Id     types.String                   `tfsdk:"id"`
	Labels []*gdriveLabelPolicyLabelModel `tfsdk:"labels"`
}

func (r *gdriveLabelPolicyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_label_policy"
}

func (r *gdriveLabelPolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Enforces a set of labels on a Drive object.",
		Attributes: map[string]schema.Attribute{
			"id": rsId(),
			"file_id": schema.StringAttribute{
				MarkdownDescription: "ID of the file to assign the label(s) to.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"labels": schema.SetNestedAttribute{
				MarkdownDescription: "The set of labels that should be applied to this file.",
				Required:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"label_id": schema.StringAttribute{
							MarkdownDescription: "The ID of the label.",
							Required:            true,
						},
						"fields": labelAssignmentFields(),
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
		Id:     plan.FileId,
	}
	resp.Diagnostics.Append(mockState.populate(ctx)...)
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
	resp.Diagnostics.Append(state.populate(ctx)...)
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
	for i := range state.Labels {
		modLabelsReq.LabelModifications = append(modLabelsReq.LabelModifications, &drive.LabelModification{
			LabelId:     state.Labels[i].LabelId.ValueString(),
			RemoveLabel: true,
		})
	}
	_, err := gsmdrive.ModifyLabels(state.Id.ValueString(), "", modLabelsReq)
	if err != nil {
		resp.Diagnostics.AddError("Configuration Error", fmt.Sprintf("Unable to remove label assignment(s), got error: %s", err))
		return
	}
}

func (r *gdriveLabelPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
