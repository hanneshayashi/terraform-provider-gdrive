package provider

import (
	"os"

	"github.com/hanneshayashi/gsm/gsmdrive"
	"github.com/hashicorp/terraform/helper/schema"
	"google.golang.org/api/drive/v3"
)

func resourceFile() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"parent": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "fileId of the parent",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "name of the file / folder",
			},
			"mime_type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "MIME type",
			},
			"drive_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "driveId of the Shared Drive",
			},
			"content": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "path to a file to upload",
				ForceNew:    true,
			},
		},
		Create: resourceCreateFile,
		Read:   resourceReadFile,
		Update: resourceUpdateFile,
		Delete: resourceDeleteFile,
		Exists: resourceExistsFile,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func resourceCreateFile(d *schema.ResourceData, _ interface{}) error {
	var err error
	f := &drive.File{
		MimeType: d.Get("mime_type").(string),
		Name:     d.Get("name").(string),
		DriveId:  d.Get("drive_id").(string),
		Parents:  []string{d.Get("parent").(string)},
	}
	var content *os.File
	if d.HasChange("content") {
		content, err = os.Open(d.Get("content").(string))
		if err != nil {
			return err
		}
		defer content.Close()
	}
	r, err := gsmdrive.CreateFile(f, content, false, false, false, "", "", "id")
	if err != nil {
		return err
	}
	d.SetId(r.Id)
	err = resourceReadFile(d, nil)
	if err != nil {
		return err
	}
	return nil
}

func resourceReadFile(d *schema.ResourceData, _ interface{}) error {
	r, err := gsmdrive.GetFile(d.Id(), "parents,mimeType,driveId,name", "")
	if err != nil {
		return err
	}
	d.Set("parent", r.Parents[0])
	d.Set("mime_type", r.MimeType)
	d.Set("drive_id", r.DriveId)
	d.Set("name", r.Name)
	return nil
}

func resourceUpdateFile(d *schema.ResourceData, _ interface{}) error {
	var addParents string
	var removeParents string
	f := &drive.File{
		MimeType: d.Get("mime_type").(string),
		Name:     d.Get("name").(string),
		DriveId:  d.Get("drive_id").(string),
	}
	if d.HasChange("parent") {
		rp, ap := d.GetChange("parent")
		removeParents = rp.(string)
		addParents = ap.(string)
	}
	_, err := gsmdrive.UpdateFile(d.Id(), addParents, removeParents, "", "", "id", f, nil, false, false)
	if err != nil {
		return err
	}
	err = resourceReadFile(d, nil)
	if err != nil {
		return err
	}
	return nil
}

func resourceDeleteFile(d *schema.ResourceData, _ interface{}) error {
	_, err := gsmdrive.DeleteFile(d.Id())
	return err
}

func resourceExistsFile(d *schema.ResourceData, _ interface{}) (bool, error) {
	_, err := gsmdrive.GetFile(d.Id(), "", "")
	if err != nil {
		return false, err
	}
	return true, nil
}
