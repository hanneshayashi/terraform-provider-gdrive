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
	"os"

	"github.com/hanneshayashi/gsm/gsmdrive"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &gdriveFileResource{}
var _ resource.ResourceWithImportState = &gdriveFileResource{}

const fieldsFile = "parents,mimeType,driveId,name,id"

func newFile() resource.Resource {
	return &gdriveFileResource{}
}

// gdriveFileResource defines the resource implementation.
type gdriveFileResource struct {
	client *http.Client
}

// gdriveFileResourceModel describes the resource data model.
type gdriveFileResourceModel struct {
	Parent         types.String `tfsdk:"parent"`
	Name           types.String `tfsdk:"name"`
	MimeType       types.String `tfsdk:"mime_type"`
	Id             types.String `tfsdk:"id"`
	MimeTypeSource types.String `tfsdk:"mime_type_source"`
	DriveId        types.String `tfsdk:"drive_id"`
	Content        types.String `tfsdk:"content"`
}

func (r *gdriveFileResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_file"
}

func (r *gdriveFileResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Creates a file or folder with the given MIME type and optionally uploads a local file",
		Attributes: map[string]schema.Attribute{
			"parent": schema.StringAttribute{
				MarkdownDescription: "The fileId of the parent",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "name of the file / folder",
				Required:            true,
			},
			"mime_type": schema.StringAttribute{
				MarkdownDescription: "MIME type of the target file (in Google Drive)",
				Required:            true,
			},
			"mime_type_source": schema.StringAttribute{
				MarkdownDescription: "MIME type of the source file (on the local system)",
				Optional:            true,
			},
			"drive_id": schema.StringAttribute{
				MarkdownDescription: "ID of the Shared Drive",
				Optional:            true,
			},
			"content": schema.StringAttribute{
				MarkdownDescription: `Path to a file to upload.
The provider does not check the content of the file for updates.
If you need to upload a new version of a file, you need to supply a different file name.`,
				Optional: true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the file (fileId)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *gdriveFileResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *gdriveFileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := &gdriveFileResourceModel{}
	resp.Diagnostics.Append(req.Plan.Get(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	fileReq := plan.toRequest()
	fileReq.Parents = []string{plan.Parent.ValueString()}
	content, diags := plan.getContent()
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	f, err := gsmdrive.CreateFile(fileReq, content, false, false, false, "", "", plan.MimeTypeSource.ValueString(), fieldsFile)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create file, got error: %s", err))
		return
	}
	plan.Id = types.StringValue(f.Id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *gdriveFileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	state := &gdriveFileResourceModel{}
	resp.Diagnostics.Append(req.State.Get(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	f, err := gsmdrive.GetFile(state.Id.ValueString(), fieldsFile, "")
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get file, got error: %s", err))
		return
	}
	if len(f.Parents) != 0 {
		state.Parent = types.StringValue(f.Parents[0])
	}
	if f.DriveId != "" {
		state.DriveId = types.StringValue(f.DriveId)
	}
	state.Name = types.StringValue(f.Name)
	state.MimeType = types.StringValue(f.MimeType)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *gdriveFileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := &gdriveFileResourceModel{}
	resp.Diagnostics.Append(req.Plan.Get(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state := &gdriveFileResourceModel{}
	resp.Diagnostics.Append(req.State.Get(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var addParents string
	var removeParents string
	var err error
	fileReq := plan.toRequest()
	if !plan.Parent.Equal(state.Parent) {
		removeParents = state.Parent.ValueString()
		addParents = plan.Parent.ValueString()
	}
	var content *os.File
	if !plan.Content.Equal(state.Content) && !plan.Content.IsNull() {
		content, err = os.Open(plan.Content.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to open local file (content), got error: %s", err))
			return
		}
		defer content.Close()
	}
	_, err = gsmdrive.UpdateFile(plan.Id.ValueString(), addParents, removeParents, "", "", fieldsFile, fileReq, content, false, false)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update file, got error: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *gdriveFileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	plan := &gdriveFileResourceModel{}
	resp.Diagnostics.Append(req.State.Get(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, err := gsmdrive.DeleteFile(plan.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete file, got error: %s", err))
		return
	}
}

func (r *gdriveFileResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
