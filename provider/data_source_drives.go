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
	"github.com/hanneshayashi/gsm/gsmdrive"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDrives() *schema.Resource {
	return &schema.Resource{
		Description: "Returns a list of Shared Drives that match the given query",
		Schema: map[string]*schema.Schema{
			"query": {
				Type:     schema.TypeString,
				Required: true,
				Description: `Query string for searching shared drives.
See the https://developers.google.com/drive/api/v3/search-shareddrives for supported syntax.`,
			},
			"use_domain_admin_access": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Use domain admin access",
			},
			"drives": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"drive_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"restrictions": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeBool,
							},
						},
					},
				},
			},
		},
		Read: dataSourceReadDrives,
	}
}

func dataSourceReadDrives(d *schema.ResourceData, _ any) error {
	query := d.Get("query").(string)
	drives := make([]map[string]any, 0)
	r, err := gsmdrive.ListDrives(query, "drives(name,id,restrictions),nextPageToken", d.Get("use_domain_admin_access").(bool), 1)
	for d := range r {
		drives = append(drives, map[string]any{
			"drive_id":     d.Id,
			"name":         d.Name,
			"restrictions": getRestrictions(d),
		})
	}
	d.Set("drives", drives)
	d.SetId(query)
	e := <-err
	if e != nil {
		return e
	}
	return nil
}
