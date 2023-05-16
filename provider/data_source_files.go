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
var _ datasource.DataSource = &filesDataSource{}

func newFilesDataSource() datasource.DataSource {
	return &filesDataSource{}
}

// filesDataSource defines the data source implementation.
type filesDataSource struct {
	client *http.Client
}

// gdriveDriveResourceModelV1 describes the resource data model V1.
type gdriveFilesDataSourceFileModel struct {
	Name     types.String `tfsdk:"name"`
	Parent   types.String `tfsdk:"parent"`
	FileId   types.String `tfsdk:"file_id"`
	Id       types.String `tfsdk:"id"`
	DriveId  types.String `tfsdk:"drive_id"`
	MimeType types.String `tfsdk:"mime_type"`
}

// gdriveDriveResourceModelV1 describes the resource data model V1.
type gdriveFilesDataSourceModel struct {
	Id                        types.String                      `tfsdk:"id"`
	Query                     types.String                      `tfsdk:"query"`
	Spaces                    types.String                      `tfsdk:"spaces"`
	Corpora                   types.String                      `tfsdk:"corpora"`
	DriveId                   types.String                      `tfsdk:"drive_id"`
	IncludeItemsFromAllDrives types.Bool                        `tfsdk:"include_items_from_all_drives"`
	Files                     []*gdriveFilesDataSourceFileModel `tfsdk:"files"`
}

func (d *filesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_files"
}

func (d *filesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Returns a list of Shared Drives that match the given query",
		Attributes: map[string]schema.Attribute{
			"id": dsId(),
			"query": schema.StringAttribute{
				Required: true,
				Description: `A query for filtering the file results.
See the https://developers.google.com/drive/api/v3/search-files for the supported syntax.`,
			},
			"spaces": schema.StringAttribute{
				Optional: true,
				Description: `A comma-separated list of spaces to query within the corpus.
Supported values are 'drive', 'appDataFolder' and 'photos'.`,
			},
			"corpora": schema.StringAttribute{
				Optional: true,
				Description: `Groupings of files to which the query applies.
Supported groupings are:
'user' (files created by, opened by, or shared directly with the user)
'drive' (files in the specified shared drive as indicated by the 'driveId')
'domain' (files shared to the user's domain)
'allDrives' (A combination of 'user' and 'drive' for all drives where the user is a member).
When able, use 'user' or 'drive', instead of 'allDrives', for efficiency.`,
			},
			"drive_id": schema.StringAttribute{
				Optional:    true,
				Description: `ID of the shared drive.`,
			},
			"include_items_from_all_drives": schema.BoolAttribute{
				Optional:    true,
				Description: `Whether both My Drive and shared drive items should be included in results.`,
			},
		},
		Blocks: map[string]schema.Block{
			"files": schema.SetNestedBlock{
				MarkdownDescription: "A set of files that match the specified query.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": dsId(),
						"file_id": schema.StringAttribute{
							MarkdownDescription: "The ID of the file.",
							Computed:            true,
						},
						"drive_id": schema.StringAttribute{
							MarkdownDescription: "The ID of the Shared Drive the file is located in. Only present if the file is located in a Shared Drive.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "The name of the file.",
							Computed:            true,
						},
						"mime_type": schema.StringAttribute{
							MarkdownDescription: "The MIME type of the file",
							Computed:            true,
						},
						"parent": schema.StringAttribute{
							MarkdownDescription: "The ID of the file's parent.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *filesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (ds *filesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	config := &gdriveFilesDataSourceModel{}
	resp.Diagnostics.Append(req.Config.Get(ctx, config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	query := config.Query.ValueString()
	r, err := gsmdrive.ListFiles(query, config.DriveId.ValueString(), config.Corpora.ValueString(), "", "", config.Spaces.ValueString(), fmt.Sprintf("files(%s),nextPageToken", fieldsFile), config.IncludeItemsFromAllDrives.ValueBool(), 1)
	for f := range r {
		fileModel := &gdriveFilesDataSourceFileModel{
			Name:     types.StringValue(f.Name),
			Id:       types.StringValue(f.Id),
			FileId:   types.StringValue(f.Id),
			MimeType: types.StringValue(f.MimeType),
		}
		if len(f.Parents) > 0 {
			fileModel.Parent = types.StringValue(f.Parents[0])
		}
		if f.DriveId != "" {
			fileModel.DriveId = types.StringValue(f.DriveId)
		}
		config.Files = append(config.Files, fileModel)
	}
	e := <-err
	if e != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list files, got error: %s", e))
		return
	}
	config.Id = config.Query
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
