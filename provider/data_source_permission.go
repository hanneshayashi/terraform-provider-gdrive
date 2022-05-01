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
	"fmt"

	"github.com/hanneshayashi/gsm/gsmdrive"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourcePermission() *schema.Resource {
	return &schema.Resource{
		Description: "Returns the metadata of a permission on a file or Shared Drive",
		Schema: map[string]*schema.Schema{
			"permission_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the permission",
			},
			"file_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the file or Shared Drive",
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"email_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"role": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"use_domain_admin_access": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Use domain admin access",
			},
		},
		Read: dataSourceReadPermission,
	}
}

func dataSourceReadPermission(d *schema.ResourceData, _ any) error {
	fileID := d.Get("file_id").(string)
	permissionID := d.Get("permission_id").(string)
	r, err := gsmdrive.GetPermission(fileID, permissionID, "emailAddress,domain,role,type", d.Get("use_domain_admin_access").(bool))
	if err != nil {
		return err
	}
	d.SetId(fmt.Sprintf("%s/%s", fileID, r.Id))
	d.Set("email_address", r.EmailAddress)
	d.Set("domain", r.Domain)
	d.Set("role", r.Role)
	d.Set("type", r.Type)
	return nil
}
