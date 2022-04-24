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

func dataSourceDrive() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"drive_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the Shared Drive",
			},
			"use_domain_admin_access": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Use domain admin access",
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"restrictions": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"admin_managed_restrictions": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"copy_requires_writer_permission": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"domain_users_only": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"drive_members_only": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
		},
		Read: dataSourceReadDrive,
	}
}

func dataSourceReadDrive(d *schema.ResourceData, _ any) error {
	driveID := d.Get("drive_id").(string)
	r, err := gsmdrive.GetDrive(driveID, "*", d.Get("use_domain_admin_access").(bool))
	if err != nil {
		return err
	}
	d.SetId(driveID)
	d.Set("name", r.Name)
	restrictions := make([]map[string]bool, 1)
	if r.Restrictions != nil {
		restrictions[0] = map[string]bool{
			"admin_managed_restrictions":      r.Restrictions.AdminManagedRestrictions,
			"copy_requires_writer_permission": r.Restrictions.CopyRequiresWriterPermission,
			"domain_users_only":               r.Restrictions.DomainUsersOnly,
			"drive_members_only":              r.Restrictions.DriveMembersOnly,
		}
	}
	d.Set("restrictions", restrictions)
	return nil
}
