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
	"google.golang.org/api/drive/v3"
)

func resourceDrive() *schema.Resource {
	return &schema.Resource{
		Description: "Creates a Shared Drive",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the Shared Drive",
			},
			"use_domain_admin_access": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Use domain admin access",
			},
			"restrictions": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"admin_managed_restrictions": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Whether administrative privileges on this shared drive are required to modify restrictions",
						},
						"copy_requires_writer_permission": {
							Type:     schema.TypeBool,
							Optional: true,
							Description: `Whether the options to copy, print, or download files inside this shared drive, should be disabled for readers and commenters.
When this restriction is set to true, it will override the similarly named field to true for any file inside this shared drive`,
						},
						"domain_users_only": {
							Type:     schema.TypeBool,
							Optional: true,
							Description: `Whether access to this shared drive and items inside this shared drive is restricted to users of the domain to which this shared drive belongs.
This restriction may be overridden by other sharing policies controlled outside of this shared drive`,
						},
						"drive_members_only": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Whether access to items inside this shared drive is restricted to its members",
						},
					},
				},
			},
		},
		Create: resourceCreateDrive,
		Read:   resourceReadDrive,
		Update: resourceUpdateDrive,
		Delete: resourceDeleteDrive,
		Exists: resourceExistsDrive,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func dataToDrive(d *schema.ResourceData, update bool) (*drive.Drive, error) {
	newDrive := &drive.Drive{}
	if d.HasChange("name") {
		newDrive.Name = d.Get("name").(string)
		if newDrive.Name == "" {
			newDrive.ForceSendFields = append(newDrive.ForceSendFields, "Name")
		}
	}
	if update {
		if d.HasChange("restrictions") {
			newDrive.Restrictions = &drive.DriveRestrictions{}
			restrictions := d.Get("restrictions").([]any)
			if len(restrictions) > 0 {
				if d.HasChange("restrictions.0.admin_managed_restrictions") {
					newDrive.Restrictions.AdminManagedRestrictions = d.Get("restrictions.0.admin_managed_restrictions").(bool)
					if !newDrive.Restrictions.AdminManagedRestrictions {
						newDrive.Restrictions.ForceSendFields = append(newDrive.Restrictions.ForceSendFields, "AdminManagedRestrictions")
					}
				}
				if d.HasChange("restrictions.0.copy_requires_writer_permission") {
					newDrive.Restrictions.CopyRequiresWriterPermission = d.Get("restrictions.0.copy_requires_writer_permission").(bool)
					if !newDrive.Restrictions.CopyRequiresWriterPermission {
						newDrive.Restrictions.ForceSendFields = append(newDrive.Restrictions.ForceSendFields, "CopyRequiresWriterPermission")
					}
				}
				if d.HasChange("restrictions.0.domain_users_only") {
					newDrive.Restrictions.DomainUsersOnly = d.Get("restrictions.0.domain_users_only").(bool)
					if !newDrive.Restrictions.DomainUsersOnly {
						newDrive.Restrictions.ForceSendFields = append(newDrive.Restrictions.ForceSendFields, "DomainUsersOnly")
					}
				}
				if d.HasChange("restrictions.0.drive_members_only") {
					newDrive.Restrictions.DriveMembersOnly = d.Get("restrictions.0.drive_members_only").(bool)
					if !newDrive.Restrictions.DriveMembersOnly {
						newDrive.Restrictions.ForceSendFields = append(newDrive.Restrictions.ForceSendFields, "DriveMembersOnly")
					}
				}
			} else {
				newDrive.Restrictions.ForceSendFields = append(newDrive.Restrictions.ForceSendFields, "AdminManagedRestrictions", "CopyRequiresWriterPermission", "DomainUsersOnly", "DriveMembersOnly")
			}
		}
	}
	return newDrive, nil
}

func resourceCreateDrive(d *schema.ResourceData, _ any) error {
	var driveResult *drive.Drive
	var err error
	driveRequest, err := dataToDrive(d, false)
	if err != nil {
		return err
	}
	driveResult, err = gsmdrive.CreateDrive(driveRequest, "*", true)
	if err != nil {
		return err
	}
	d.SetId(driveResult.Id)
	if d.HasChange("restrictions") {
		err = resourceUpdateDrive(d, nil)
		if err != nil {
			return err
		}
	} else {
		err = resourceReadDrive(d, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

func resourceReadDrive(d *schema.ResourceData, _ any) error {
	r, err := gsmdrive.GetDrive(d.Id(), "*", d.Get("use_domain_admin_access").(bool))
	if err != nil {
		return err
	}
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

func resourceUpdateDrive(d *schema.ResourceData, _ any) error {
	driveRequest, err := dataToDrive(d, true)
	if err != nil {
		return err
	}
	_, err = gsmdrive.UpdateDrive(d.Id(), "id", d.Get("use_domain_admin_access").(bool), driveRequest)
	if err != nil {
		return err
	}
	err = resourceReadDrive(d, nil)
	if err != nil {
		return err
	}
	return nil
}

func resourceDeleteDrive(d *schema.ResourceData, _ any) error {
	_, err := gsmdrive.DeleteDrive(d.Id())
	return err
}

func resourceExistsDrive(d *schema.ResourceData, _ any) (bool, error) {
	_, err := gsmdrive.GetDrive(d.Id(), "id", d.Get("use_domain_admin_access").(bool))
	if err != nil {
		return false, err
	}
	return true, nil
}
