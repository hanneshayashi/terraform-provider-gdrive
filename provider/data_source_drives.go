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
var _ datasource.DataSource = &drivesDataSource{}

func newDrivesDataSource() datasource.DataSource {
	return &drivesDataSource{}
}

// drivesDataSource defines the data source implementation.
type drivesDataSource struct {
	client *http.Client
}

// gdriveDriveResourceModelV1 describes the resource data model V1.
type gdriveDrivesDataSourceDriveModel struct {
	Restrictions *driveRestrictionsModel `tfsdk:"restrictions"`
	Name         types.String            `tfsdk:"name"`
	Id           types.String            `tfsdk:"id"`
	DriveId      types.String            `tfsdk:"drive_id"`
}

// gdriveDriveResourceModelV1 describes the resource data model V1.
type gdriveDrivesDataSourceModel struct {
	Query                types.String                        `tfsdk:"query"`
	Name                 types.String                        `tfsdk:"name"`
	Id                   types.String                        `tfsdk:"id"`
	Drives               []*gdriveDrivesDataSourceDriveModel `tfsdk:"drives"`
	UseDomainAdminAccess types.Bool                          `tfsdk:"use_domain_admin_access"`
}

func (d *drivesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_drives"
}

func (d *drivesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Returns a list of Shared Drives that match the given query",
		Attributes: map[string]schema.Attribute{
			"id": dsId(),
			"query": schema.StringAttribute{
				Required: true,
				Description: `Query string for searching shared drives.
See the https://developers.google.com/drive/api/v3/search-shareddrives for supported syntax.`,
			},
			"use_domain_admin_access": schema.BoolAttribute{
				Optional:    true,
				Description: "Use domain admin access",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of this shared drive.",
			}, "drives": schema.SetNestedAttribute{
				Computed:            true,
				MarkdownDescription: "A set of Shared Drives that match the specified query.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": dsId(),
						"drive_id": schema.StringAttribute{
							MarkdownDescription: "ID of the Shared Drive",
							Required:            true,
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of this shared drive.",
						},
						"restrictions": dsDriveRestrictions(),
					},
				},
			},
		},
	}
}

func (d *drivesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (ds *drivesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	config := &gdriveDrivesDataSourceModel{}
	resp.Diagnostics.Append(req.Config.Get(ctx, config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	query := config.Query.ValueString()
	r, err := gsmdrive.ListDrives(query, fmt.Sprintf("drives(%s),nextPageToken", fieldsDrive), config.UseDomainAdminAccess.ValueBool(), 1)
	for d := range r {
		config.Drives = append(config.Drives, &gdriveDrivesDataSourceDriveModel{
			Name:    types.StringValue(d.Name),
			Id:      types.StringValue(d.Id),
			DriveId: types.StringValue(d.Id),
			Restrictions: &driveRestrictionsModel{
				AdminManagedRestrictions:     types.BoolValue(d.Restrictions.AdminManagedRestrictions),
				CopyRequiresWriterPermission: types.BoolValue(d.Restrictions.CopyRequiresWriterPermission),
				DomainUsersOnly:              types.BoolValue(d.Restrictions.DomainUsersOnly),
				DriveMembersOnly:             types.BoolValue(d.Restrictions.DriveMembersOnly),
			},
		})
	}
	e := <-err
	if e != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list Shared Drives, got error: %s", e))
		return
	}
	config.Id = config.Query
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
