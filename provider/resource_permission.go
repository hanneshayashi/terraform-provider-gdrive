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
	"google.golang.org/api/drive/v3"
)

func resourcePermission() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"file_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the file or Shared Drive",
				ForceNew:    true,
			},
			"email_message": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "An optional email message that will be sent when the permission is created",
			},
			"send_notification_email": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Wether to send a notfication email",
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The type of the trustee. Can be 'user', 'domain', 'group' or 'anyone'",
				ValidateFunc: validatePermissionType,
				ForceNew:     true,
			},
			"domain": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "The domain that should be granted access",
				ConflictsWith: []string{"email_address"},
				ForceNew:      true,
			},
			"email_address": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "The email address of the trustee",
				ConflictsWith: []string{"domain"},
				ForceNew:      true,
			},
			"role": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The role",
			},
			"use_domain_admin_access": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Use domain admin access",
			},
			"transfer_ownership": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether to transfer ownership to the specified user",
			},
			"move_to_new_owners_root": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "This parameter only takes effect if the item is not in a shared drive and the request is attempting to transfer the ownership of the item.",
			},
		},
		Create: resourceCreatePermission,
		Read:   resourceReadPermission,
		Update: resourceUpdatePermission,
		Delete: resourceDeletePermission,
		Exists: resourceExistsPermission,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

var validPermissionTypes = []string{
	"user",
	"group",
	"domain",
	"anyone",
}

func resourceCreatePermission(d *schema.ResourceData, _ any) error {
	fileID := d.Get("file_id").(string)
	p := &drive.Permission{
		Domain:       d.Get("domain").(string),
		EmailAddress: d.Get("email_address").(string),
		Role:         d.Get("role").(string),
		Type:         d.Get("type").(string),
	}
	r, err := gsmdrive.CreatePermission(fileID, d.Get("email_message").(string), "id", d.Get("use_domain_admin_access").(bool), d.Get("send_notification_email").(bool), d.Get("transfer_ownership").(bool), d.Get("move_to_new_owners_root").(bool), p)
	if err != nil {
		return err
	}
	d.SetId(fmt.Sprintf("%s/%s", fileID, r.Id))
	err = resourceReadPermission(d, nil)
	if err != nil {
		return err
	}
	return nil
}

func resourceReadPermission(d *schema.ResourceData, _ any) error {
	fileID, permissionID := splitCombinedPermissionId(d.Id())
	r, err := gsmdrive.GetPermission(fileID, permissionID, "emailAddress,domain,role,type", d.Get("use_domain_admin_access").(bool))
	if err != nil {
		return err
	}
	d.Set("email_address", r.EmailAddress)
	d.Set("domain", r.Domain)
	d.Set("role", r.Role)
	d.Set("type", r.Type)
	return nil
}

func resourceUpdatePermission(d *schema.ResourceData, _ any) error {
	fileID, permissionID := splitCombinedPermissionId(d.Id())
	p := &drive.Permission{Role: d.Get("role").(string)}
	_, err := gsmdrive.UpdatePermission(fileID, permissionID, "id", d.Get("use_domain_admin_access").(bool), false, p)
	if err != nil {
		return err
	}
	err = resourceReadPermission(d, nil)
	if err != nil {
		return err
	}
	return nil
}

func resourceDeletePermission(d *schema.ResourceData, _ any) error {
	fileID, permissionID := splitCombinedPermissionId(d.Id())
	_, err := gsmdrive.DeletePermission(fileID, permissionID, d.Get("use_domain_admin_access").(bool))
	return err
}

func resourceExistsPermission(d *schema.ResourceData, _ any) (bool, error) {
	fileID, permissionID := splitCombinedPermissionId(d.Id())
	_, err := gsmdrive.GetPermission(fileID, permissionID, "", d.Get("use_domain_admin_access").(bool))
	if err != nil {
		return false, err
	}
	return true, nil
}
