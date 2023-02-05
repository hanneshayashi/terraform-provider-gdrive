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
	"google.golang.org/api/drive/v3"
)

type gDrivePermission struct {
	Permission            *drive.Permission
	EmailMessage          string
	SendNotificationEmail bool
	TransferOwnership     bool
	MoveToNewOwnersRoot   bool
}

func resourcePermissionsPolicy() *schema.Resource {
	return &schema.Resource{
		Description: `Creates an authoratative permissions policy on a file or Shared Drive.

**Warning: This resource will set exactly the defined permissions and remove everything else!**

It is HIGHLY recommended that you import the resource and make sure that the owner is properly set before applying it!

You can import the resource using the file's or Shared Drive's id like so:

terraform import [resource address] [fileId]

**Important**: On a *destroy*, this resource will preserve the owner and organizer permissions!`,
		Schema: map[string]*schema.Schema{
			"file_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the file or Shared Drive",
				ForceNew:    true,
			},
			"permissions": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"permission_id": {
							Type:        schema.TypeString,
							Description: "The permission ID",
							Computed:    true,
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
						},
						"domain": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The domain that should be granted access",
						},
						"email_address": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The email address of the trustee",
						},
						"role": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The role",
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
				},
			},
			"use_domain_admin_access": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Use domain admin access",
			},
		},
		Create: resourceCreatePermissionsPolicy,
		Read:   resourceReadPermissionsPolicy,
		Update: resourceUpdatePermissionsPolicy,
		Delete: resourceDeletePermissionsPolicy,
		Exists: resourceExistsFile,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceCreatePermissionsPolicy(d *schema.ResourceData, _ any) error {
	fileID := d.Get("file_id").(string)
	useDomAccess := d.Get("use_domain_admin_access").(bool)
	permissionsD := d.Get("permissions").(*schema.Set).List()
	for i := range permissionsD {
		p := permissionsD[i].(map[string]any)
		permission := mapToPermission(p)
		r, err := gsmdrive.CreatePermission(fileID, permission.EmailMessage, "id", useDomAccess, permission.SendNotificationEmail, false, false, permission.Permission)
		if err != nil {
			return err
		}
		p["permission_id"] = r.Id
		permissionsD[i] = p
	}
	d.SetId(fileID)
	d.Set("permissions", permissionsD)
	err := resourceReadPermissionsPolicy(d, nil)
	if err != nil {
		return err
	}
	return nil
}

func resourceReadPermissionsPolicy(d *schema.ResourceData, _ any) error {
	fileID := d.Id()
	permissionsD := d.Get("permissions").(*schema.Set).List()
	permissionsM := make(map[string]*gDrivePermission, 0)
	for _, o := range permissionsD {
		m := o.(map[string]any)
		permissionsM[mapToTrusteeId(m)] = mapToPermission(m)
	}
	permissionsSet := schema.NewSet(permissionToSet, nil)
	r, err := gsmdrive.ListPermissions(fileID, "", "permissions(id,displayName,domain,deleted,emailAddress,expirationTime,role,type),nextPageToken", d.Get("use_domain_admin_access").(bool), 1)
	for p := range r {
		trustee := combineId(p.Domain, p.EmailAddress)
		permission := map[string]any{
			"permission_id": p.Id,
			"domain":        p.Domain,
			"email_address": p.EmailAddress,
			"role":          p.Role,
			"type":          p.Type,
		}
		if permissionsM[trustee] != nil {
			permission["email_message"] = permissionsM[trustee].EmailMessage
			permission["send_notification_email"] = permissionsM[trustee].SendNotificationEmail
			permission["transfer_ownership"] = permissionsM[trustee].TransferOwnership
			permission["move_to_new_owners_root"] = permissionsM[trustee].MoveToNewOwnersRoot
		}
		permissionsSet.Add(permission)
	}
	e := <-err
	if e != nil {
		return e
	}
	d.SetId(fileID)
	d.Set("file_id", fileID)
	d.Set("permissions", permissionsSet)
	return nil
}

func mapToPermission(m map[string]any) *gDrivePermission {
	return &gDrivePermission{
		Permission: &drive.Permission{
			Type:         m["type"].(string),
			Role:         m["role"].(string),
			EmailAddress: m["email_address"].(string),
			Domain:       m["domain"].(string),
			Id:           m["permission_id"].(string),
		},
		EmailMessage:          m["email_message"].(string),
		SendNotificationEmail: m["send_notification_email"].(bool),
		TransferOwnership:     m["transfer_ownership"].(bool),
		MoveToNewOwnersRoot:   m["move_to_new_owners_root"].(bool),
	}
}

func mapToTrusteeId(m map[string]any) string {
	return combineId(m["domain"].(string), m["email_address"].(string))
}

func resourceUpdatePermissionsPolicy(d *schema.ResourceData, _ any) error {
	fileID := d.Id()
	old, new := d.GetChange("permissions")
	useDomAccess := d.Get("use_domain_admin_access").(bool)
	permissionsOld := make(map[string]*gDrivePermission, 0)
	permissionsNew := make(map[string]*gDrivePermission, 0)
	remove := make([]string, 0)
	create := make([]*gDrivePermission, 0)
	update := make(map[string]*drive.Permission, 0)
	for _, o := range old.(*schema.Set).List() {
		m := o.(map[string]any)
		permissionsOld[mapToTrusteeId(m)] = mapToPermission(m)
	}
	for _, n := range new.(*schema.Set).List() {
		m := n.(map[string]any)
		permissionsNew[mapToTrusteeId(m)] = mapToPermission(m)
	}
	for t := range permissionsOld {
		if permissionsNew[t] == nil {
			remove = append(remove, permissionsOld[t].Permission.Id)
		}
	}
	for t := range permissionsNew {
		if permissionsOld[t] == nil {
			create = append(create, permissionsNew[t])
		} else if permissionsNew[t].Permission.Role != permissionsOld[t].Permission.Role {
			if permissionsNew[t].Permission.Role == "owner" {
				_, err := gsmdrive.UpdatePermission(fileID, permissionsOld[t].Permission.Id, "id", useDomAccess, false, &drive.Permission{Role: permissionsNew[t].Permission.Role})
				if err != nil {
					return err
				}
			} else {
				update[permissionsOld[t].Permission.Id] = &drive.Permission{Role: permissionsNew[t].Permission.Role}
			}
		}
	}
	for i := range create {
		_, err := gsmdrive.CreatePermission(fileID, create[i].EmailMessage, "id", useDomAccess, create[i].SendNotificationEmail, create[i].TransferOwnership, create[i].MoveToNewOwnersRoot, create[i].Permission)
		if err != nil {
			return err
		}
	}
	for i := range update {
		_, err := gsmdrive.UpdatePermission(fileID, i, "id", useDomAccess, false, update[i])
		if err != nil {
			return err
		}
	}
	for i := range remove {
		_, err := gsmdrive.DeletePermission(fileID, remove[i], useDomAccess)
		if err != nil {
			return err
		}
	}
	err := resourceReadPermissionsPolicy(d, nil)
	if err != nil {
		return err
	}
	return nil
}

func resourceDeletePermissionsPolicy(d *schema.ResourceData, _ any) error {
	fileID := d.Id()
	useDomAccess := d.Get("use_domain_admin_access").(bool)
	permissions := d.Get("permissions").(*schema.Set).List()
	for p := range permissions {
		permission := permissions[p].(map[string]any)
		if permission["role"].(string) != "owner" && permission["role"].(string) != "organizer" {
			_, err := gsmdrive.DeletePermission(fileID, permission["permission_id"].(string), useDomAccess)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
