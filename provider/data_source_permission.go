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
var _ datasource.DataSource = &permissionDataSource{}

func newPermissionDataSource() datasource.DataSource {
	return &permissionDataSource{}
}

// permissionDataSource defines the data source implementation.
type permissionDataSource struct {
	client *http.Client
}

type gdrivePermissionDataSourceModel struct {
	Id                   types.String `tfsdk:"id"`
	PermissionId         types.String `tfsdk:"permission_id"`
	FileId               types.String `tfsdk:"file_id"`
	Type                 types.String `tfsdk:"type"`
	Domain               types.String `tfsdk:"domain"`
	EmailAddress         types.String `tfsdk:"email_address"`
	Role                 types.String `tfsdk:"role"`
	UseDomainAdminAccess types.Bool   `tfsdk:"use_domain_admin_access"`
}

func (d *permissionDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_permission"
}

func (d *permissionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Returns the metadata of a permission on a file or Shared Drive",
		Attributes: map[string]schema.Attribute{
			"id": dsId(),
			"permission_id": schema.StringAttribute{
				MarkdownDescription: "ID of the permission",
				Required:            true,
			},
			"file_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ID of the file or Shared Drive",
			},
			"type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The type of the trustee. Can be 'user', 'domain', 'group' or 'anyone'",
			},
			"domain": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The domain if the type of this permissions is 'domain'",
			},
			"email_address": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The email address if the type of this permissions is 'user' or 'group'",
			},
			"role": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The role that this trustee is granted",
			},
			"use_domain_admin_access": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Use domain admin access",
			},
		},
	}
}

func (d *permissionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (ds *permissionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	config := &gdrivePermissionDataSourceModel{}
	resp.Diagnostics.Append(req.Config.Get(ctx, config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	fileID := config.FileId.ValueString()
	permissionID := config.PermissionId.ValueString()
	r, err := gsmdrive.GetPermission(fileID, permissionID, "emailAddress,domain,role,type,id", config.UseDomainAdminAccess.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get permission, got error: %s", err))
		return
	}
	config.Id = types.StringValue(combineId(fileID, r.Id))
	config.EmailAddress = types.StringValue(r.EmailAddress)
	config.Domain = types.StringValue(r.Domain)
	config.Role = types.StringValue(r.Role)
	config.Type = types.StringValue(r.Type)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
