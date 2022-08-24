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
	"os"

	"github.com/hanneshayashi/gsm/gsmdrive"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/api/drive/v3"
)

func resourceFile() *schema.Resource {
	return &schema.Resource{
		Description: "Creates a file or folder with the given MIME type and optionally uploads a local file",
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
				Description: "MIME type of the target file (in Google Drive)",
			},
			"mime_type_source": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "MIME type of the source file (on the local system)",
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
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceCreateFile(d *schema.ResourceData, _ any) error {
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
	r, err := gsmdrive.CreateFile(f, content, false, false, false, "", "", d.Get("mime_type_source").(string), "id")
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

func resourceReadFile(d *schema.ResourceData, _ any) error {
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

func resourceUpdateFile(d *schema.ResourceData, _ any) error {
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

func resourceDeleteFile(d *schema.ResourceData, _ any) error {
	_, err := gsmdrive.DeleteFile(d.Id())
	return err
}

func resourceExistsFile(d *schema.ResourceData, _ any) (bool, error) {
	_, err := gsmdrive.GetFile(d.Id(), "", "")
	if err != nil {
		return false, err
	}
	return true, nil
}
