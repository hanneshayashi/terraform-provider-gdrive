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
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &fileDataSource{}

func newfileDataSource() datasource.DataSource {
	return &fileDataSource{}
}

// fileDataSource defines the data source implementation.
type fileDataSource struct {
	client *http.Client
}

type gdriveFileDataSourceModel struct {
	Id             types.String `tfsdk:"id"`
	FileId         types.String `tfsdk:"file_id"`
	Parent         types.String `tfsdk:"parent"`
	Name           types.String `tfsdk:"name"`
	MimeType       types.String `tfsdk:"mime_type"`
	DownloadPath   types.String `tfsdk:"download_path"`
	ExportPath     types.String `tfsdk:"export_path"`
	ExportMimeType types.String `tfsdk:"export_mime_type"`
	LocalFilePath  types.String `tfsdk:"local_file_path"`
	DriveId        types.String `tfsdk:"drive_id"`
}

func (d *fileDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_file"
}

func (d *fileDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Gets a Shared file and returns its metadata",
		Attributes: map[string]schema.Attribute{
			"file_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the file",
			},
			"drive_id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the Shared Drive the file is located it. Only present if the file is located in a Shared Drive.",
			},
			"parent": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the file's parent",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the file",
			},
			"mime_type": schema.StringAttribute{
				Computed:    true,
				Description: "Name MIME type of the file in Google file",
			},
			"download_path": schema.StringAttribute{
				Optional:    true,
				Description: "Use this to specify a local file path to download a (non-Google) file",
			},
			"export_path": schema.StringAttribute{
				Optional:    true,
				Description: "Use this to specify a local file path to export a Google file (sheet, doc, etc.)",
				// ConflictsWith: []string{"download_path"},
				// RequiredWith:  []string{"export_mime_type"},
			},
			"export_mime_type": schema.StringAttribute{
				// RequiredWith:  []string{"export_path"},
				// ConflictsWith: []string{"download_path"},
				Optional: true,
				Description: `Specify the target MIME type for the export.
For a list of supported MIME types see https://developers.google.com/file/api/v3/ref-export-formats`,
			},
			"local_file_path": schema.StringAttribute{
				Computed:    true,
				Description: "The path where the local copy or export of the file was created",
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the file",
				Computed:            true,
			},
		},
	}
}

func (d *fileDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*http.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (ds *fileDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	config := &gdriveFileDataSourceModel{}
	resp.Diagnostics.Append(req.Config.Get(ctx, config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	fileID := config.FileId.ValueString()
	r, err := gsmdrive.GetFile(fileID, fieldsFile, "")
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get file, got error: %s", err))
		return
	}
	config.Id = config.FileId
	config.Name = types.StringValue(r.Name)
	config.MimeType = types.StringValue(r.MimeType)
	if r.DriveId != "" {
		config.DriveId = types.StringValue(r.DriveId)
	}
	if len(r.Parents) > 0 {
		config.Parent = types.StringValue(r.Parents[0])
	}
	if !config.DownloadPath.IsNull() {
		filePath, err := gsmdrive.DownloadFile(fileID, config.DownloadPath.ValueString(), false)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to download file, got error: %s", err))
			return
		}
		config.LocalFilePath = types.StringValue(filePath)
	}
	if !config.ExportPath.IsNull() {
		filePath, err := gsmdrive.ExportFile(fileID, config.ExportMimeType.ValueString(), config.ExportPath.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to export file, got error: %s", err))
			return
		}
		config.LocalFilePath = types.StringValue(filePath)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
