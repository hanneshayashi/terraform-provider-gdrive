/*
Copyright © 2021-2022 Hannes Hayashi

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
	"github.com/hanneshayashi/gsm/gsmdrive"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourcePermissions() *schema.Resource {
	return &schema.Resource{
		Description: "Returns a list of all permissions set on a file or Shared Drive",
		Schema: map[string]*schema.Schema{
			"file_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the file or Shared Drive",
			},
			"use_domain_admin_access": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Use domain admin access",
			},
			"permissions": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"permission_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"display_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"domain": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"deleted": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"email_address": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"expiration_time": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"role": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"type": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
		Read: dataSourceReadPermissions,
	}
}

func dataSourceReadPermissions(d *schema.ResourceData, _ any) error {
	fileID := d.Get("file_id").(string)
	permissions := make([]map[string]any, 0)
	r, err := gsmdrive.ListPermissions(fileID, "", "permissions(id,displayName,domain,deleted,emailAddress,expirationTime,role,type),nextPageToken", d.Get("use_domain_admin_access").(bool), 1)
	for p := range r {
		permissions = append(permissions, map[string]any{
			"permission_id":   p.Id,
			"display_name":    p.DisplayName,
			"domain":          p.Domain,
			"deleted":         p.Deleted,
			"email_address":   p.EmailAddress,
			"expiration_time": p.ExpirationTime,
			"role":            p.Role,
			"type":            p.Type,
		})
	}
	e := <-err
	if e != nil {
		return e
	}
	d.SetId(fileID)
	d.Set("permissions", permissions)
	return nil
}
