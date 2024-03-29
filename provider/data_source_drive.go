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

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &driveDataSource{}

func newDriveDataSource() datasource.DataSource {
	return &driveDataSource{}
}

// driveDataSource defines the data source implementation.
type driveDataSource struct {
	client *http.Client
}

func (d *driveDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_drive"
}

func (d *driveDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Gets a Shared Drive and returns its metadata.",
		Attributes: map[string]schema.Attribute{
			"id": dsId(),
			"drive_id": schema.StringAttribute{
				MarkdownDescription: "ID of the Shared Drive.",
				Required:            true,
			},
			"use_domain_admin_access": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Use domain admin access.",
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The name of this shared drive.",
			},
			"restrictions": dsDriveRestrictions(),
		},
	}
}

func (d *driveDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (ds *driveDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	config := &gdriveDriveResourceModelV1{}
	resp.Diagnostics.Append(req.Config.Get(ctx, config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	config.Id = config.DriveId
	resp.Diagnostics.Append(config.populate()...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
