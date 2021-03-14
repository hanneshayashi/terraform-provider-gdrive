package provider

import (
	"github.com/hanneshayashi/gsm/gsmdrive"
	"github.com/hashicorp/terraform/helper/schema"
	"google.golang.org/api/drive/v3"
)

func resourceDrive() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Name of the Shared Drive",
				ValidateFunc: validateName,
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

func validateName(v interface{}, k string) (ws []string, es []error) {
	return nil, nil
}

func resourceCreateDrive(d *schema.ResourceData, m interface{}) error {
	dr, err := gsmdrive.CreateDrive(&drive.Drive{Name: d.Get("name").(string)}, "")
	if err != nil {
		return err
	}
	d.SetId(dr.Id)
	err = resourceReadDrive(d, m)
	if err != nil {
		return err
	}
	return nil
}

func resourceReadDrive(d *schema.ResourceData, m interface{}) error {
	dr, err := gsmdrive.GetDrive(d.Id(), "", false)
	if err != nil {
		return err
	}
	d.Set("name", dr.Name)
	return nil
}

func resourceUpdateDrive(d *schema.ResourceData, m interface{}) error {
	_, err := gsmdrive.UpdateDrive(d.Id(), "", false, &drive.Drive{Name: d.Get("name").(string)})
	return err
}

func resourceDeleteDrive(d *schema.ResourceData, m interface{}) error {
	_, err := gsmdrive.DeleteDrive(d.Id())
	return err
}

func resourceExistsDrive(d *schema.ResourceData, m interface{}) (bool, error) {
	_, err := gsmdrive.GetDrive(d.Id(), "", false)
	if err == nil {
		return true, nil
	}
	return false, err
}
