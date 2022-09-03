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
		Description: "Gets a Shared Drive and returns its metadata",
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
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of this shared drive.",
			},
			"restrictions": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "A set of restrictions that apply to this shared drive or items inside this shared drive.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"admin_managed_restrictions": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether administrative privileges on this shared drive are required to modify restrictions.",
						},
						"copy_requires_writer_permission": {
							Type:     schema.TypeBool,
							Computed: true,
							Description: `Whether the options to copy, print, or download files inside this shared drive, should be disabled for readers and commenters.
When this restriction is set to true, it will override the similarly named field to true for any file inside this shared drive.`,
						},
						"domain_users_only": {
							Type:     schema.TypeBool,
							Computed: true,
							Description: `Whether access to this shared drive and items inside this shared drive is restricted to users of the domain to which this shared drive belongs.
This restriction may be overridden by other sharing policies controlled outside of this shared drive.	`,
						},
						"drive_members_only": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether access to items inside this shared drive is restricted to its members.	",
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
		restrictions[0] = getRestrictions(r)
	}
	d.Set("restrictions", restrictions)
	return nil
}
