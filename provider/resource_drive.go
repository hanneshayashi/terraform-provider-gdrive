package provider

import (
	"time"

	"github.com/hanneshayashi/gsm/gsmdrive"
	"github.com/hashicorp/terraform/helper/schema"
	"google.golang.org/api/drive/v3"
)

func resourceDrive() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the Shared Drive",
				// ValidateFunc: validateName,
			},
			"use_domain_admin_access": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Use domain admin access",
			},
		},
		Create: resourceCreateDrive,
		Read:   resourceReadDrive,
		Update: resourceUpdateDrive,
		Delete: resourceDeleteDrive,
		Exists: resourceExistsDrive,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

// func validateName(v interface{}, k string) (ws []string, es []error) {
// 	return nil, nil
// }

func resourceCreateDrive(d *schema.ResourceData, _ interface{}) error {
	r, err := gsmdrive.CreateDrive(&drive.Drive{Name: d.Get("name").(string)}, "")
	if err != nil {
		return err
	}
	d.SetId(r.Id)
	time.Sleep(time.Second * 10)
	err = resourceReadDrive(d, nil)
	if err != nil {
		return err
	}
	return nil
}

func resourceReadDrive(d *schema.ResourceData, _ interface{}) error {
	r, err := gsmdrive.GetDrive(d.Id(), "name", d.Get("use_domain_admin_access").(bool))
	if err != nil {
		return err
	}
	d.Set("name", r.Name)
	return nil
}

func resourceUpdateDrive(d *schema.ResourceData, _ interface{}) error {
	_, err := gsmdrive.UpdateDrive(d.Id(), "id", d.Get("use_domain_admin_access").(bool), &drive.Drive{Name: d.Get("name").(string)})
	if err != nil {
		return err
	}
	err = resourceReadDrive(d, nil)
	if err != nil {
		return err
	}
	return nil
}

func resourceDeleteDrive(d *schema.ResourceData, _ interface{}) error {
	_, err := gsmdrive.DeleteDrive(d.Id())
	return err
}

func resourceExistsDrive(d *schema.ResourceData, _ interface{}) (bool, error) {
	_, err := gsmdrive.GetDrive(d.Id(), "id", d.Get("use_domain_admin_access").(bool))
	if err != nil {
		return false, err
	}
	return true, nil
}
