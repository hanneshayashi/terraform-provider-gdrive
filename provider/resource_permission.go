package provider

import (
	"fmt"
	"strings"

	"github.com/hanneshayashi/gsm/gsmdrive"
	"github.com/hashicorp/terraform/helper/schema"
	"google.golang.org/api/drive/v3"
)

func resourcePermission() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"file_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the file or Shared Permission",
				ForceNew:    true,
			},
			"email_message": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "An optional email message that will be sent when the permission is created",
				// DiffSuppressFunc: noDiff,
			},
			"send_notification_email": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Wether to send a notfication email",
				// DiffSuppressFunc: noDiff,
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
		},
		Create: resourceCreatePermission,
		Read:   resourceReadPermission,
		Update: resourceUpdatePermission,
		Delete: resourceDeletePermission,
		Exists: resourceExistsPermission,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

var validPermissionTypes = []string{
	"user",
	"group",
	"domain",
	"anyone",
}

func splitCombinedPermissionId(id string) (fileID, permissionID string) {
	ids := strings.Split(id, "/")
	return ids[0], ids[1]
}

func validatePermissionType(v interface{}, _ string) (ws []string, es []error) {
	value := v.(string)
	if contains(value, validPermissionTypes) {
		return nil, nil
	}
	es = append(es, fmt.Errorf("%s is not a valid permission type", value))
	return nil, es
}

func resourceCreatePermission(d *schema.ResourceData, _ interface{}) error {
	fileId := d.Get("file_id").(string)
	p := &drive.Permission{
		Domain:       d.Get("domain").(string),
		EmailAddress: d.Get("email_address").(string),
		Role:         d.Get("role").(string),
		Type:         d.Get("type").(string),
	}
	r, err := gsmdrive.CreatePermission(fileId, d.Get("email_message").(string), "id", d.Get("use_domain_admin_access").(bool), d.Get("send_notification_email").(bool), false, false, p)
	if err != nil {
		return err
	}
	d.SetId(fmt.Sprintf("%s/%s", fileId, r.Id))
	err = resourceReadPermission(d, nil)
	if err != nil {
		return err
	}
	return nil
}

func resourceReadPermission(d *schema.ResourceData, _ interface{}) error {
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

func resourceUpdatePermission(d *schema.ResourceData, _ interface{}) error {
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

func resourceDeletePermission(d *schema.ResourceData, _ interface{}) error {
	fileID, permissionID := splitCombinedPermissionId(d.Id())
	_, err := gsmdrive.DeletePermission(fileID, permissionID, d.Get("use_domain_admin_access").(bool))
	return err
}

func resourceExistsPermission(d *schema.ResourceData, _ interface{}) (bool, error) {
	fileID, permissionID := splitCombinedPermissionId(d.Id())
	_, err := gsmdrive.GetPermission(fileID, permissionID, "", d.Get("use_domain_admin_access").(bool))
	if err != nil {
		return false, err
	}
	return true, nil
}
