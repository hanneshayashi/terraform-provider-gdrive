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
var _ datasource.DataSource = &permissionsDataSource{}

func newPermissionsDataSource() datasource.DataSource {
	return &permissionsDataSource{}
}

// permissionsDataSource defines the data source implementation.
type permissionsDataSource struct {
	client *http.Client
}

type gdrivePermissionsDataSourcePermissionModel struct {
	PermissionId   types.String `tfsdk:"permission_id"`
	DisplayName    types.String `tfsdk:"display_name"`
	Domain         types.String `tfsdk:"domain"`
	EmailAddress   types.String `tfsdk:"email_address"`
	ExpirationTime types.String `tfsdk:"expiration_time"`
	Role           types.String `tfsdk:"role"`
	Type           types.String `tfsdk:"type"`
	Deleted        types.Bool   `tfsdk:"deleted"`
}

type gdrivePermissionsDataSourceModel struct {
	Id                   types.String                                  `tfsdk:"id"`
	FileId               types.String                                  `tfsdk:"file_id"`
	UseDomainAdminAccess types.Bool                                    `tfsdk:"use_domain_admin_access"`
	Permissions          []*gdrivePermissionsDataSourcePermissionModel `tfsdk:"permissions"`
}

func (d *permissionsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_permissions"
}

func (d *permissionsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Returns the metadata of a permission on a file or Shared Drive",
		Attributes: map[string]schema.Attribute{
			"id": dsId(),
			"file_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the file or Shared Drive",
			},
			"use_domain_admin_access": schema.BoolAttribute{
				Optional:    true,
				Description: "Use domain admin access",
			},
		},
		Blocks: map[string]schema.Block{
			"permissions": schema.SetNestedBlock{
				MarkdownDescription: "The list of permissions set on this file or Shared Drive",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"permission_id": schema.StringAttribute{
							Computed:    true,
							Description: "ID of the permission",
						},
						"display_name": schema.StringAttribute{
							Computed: true,
							Description: `The "pretty" name of the value of the permission.
The following is a list of examples for each type of permission:
- user    - User's full name, as defined for their Google account, such as "Joe Smith."
- group   - Name of the Google Group, such as "The Company Administrators."
- domain  - String domain name, such as "thecompany.com."
- anyone  - No displayName is present.`,
						},
						"domain": schema.StringAttribute{
							Computed:    true,
							Description: "The domain if the type of this permissions is 'domain'",
						},
						"deleted": schema.BoolAttribute{
							Computed: true,
							Description: `Whether the account associated with this permission has been deleted.
This field only pertains to user and group permissions.`,
						},
						"email_address": schema.StringAttribute{
							Computed:    true,
							Description: "The email address if the type of this permissions is 'user' or 'group'",
						},
						"expiration_time": schema.StringAttribute{
							Computed:    true,
							Description: "The time at which this permission will expire (RFC 3339 date-time)",
						},
						"role": schema.StringAttribute{
							Computed:    true,
							Description: "The role that this trustee is granted",
						},
						"type": schema.StringAttribute{
							Computed:    true,
							Description: "The type of the trustee. Can be 'user', 'domain', 'group' or 'anyone'",
						},
					},
				},
			},
		},
	}
}

func (d *permissionsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (ds *permissionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	config := &gdrivePermissionsDataSourceModel{}
	resp.Diagnostics.Append(req.Config.Get(ctx, config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	fileID := config.FileId.ValueString()
	r, err := gsmdrive.ListPermissions(fileID, "", "permissions(id,displayName,domain,deleted,emailAddress,expirationTime,role,type),nextPageToken", config.UseDomainAdminAccess.ValueBool(), 1)
	for p := range r {
		config.Permissions = append(config.Permissions, &gdrivePermissionsDataSourcePermissionModel{
			PermissionId:   types.StringValue(p.Id),
			DisplayName:    types.StringValue(p.DisplayName),
			Domain:         types.StringValue(p.Domain),
			EmailAddress:   types.StringValue(p.EmailAddress),
			ExpirationTime: types.StringValue(p.ExpirationTime),
			Role:           types.StringValue(p.Role),
			Type:           types.StringValue(p.Type),
			Deleted:        types.BoolValue(p.Deleted),
		})
	}
	e := <-err
	if e != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list permissions, got error: %s", e))
		return
	}
	config.Id = config.FileId
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
