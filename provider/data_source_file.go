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

// import (
// 	"github.com/hanneshayashi/gsm/gsmdrive"
// 	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
// )

// func dataSourceFile() *schema.Resource {
// 	return &schema.Resource{
// 		Description: "Gets a files metadata and optionally downloads / exports it to the local file system",
// 		Schema: map[string]*schema.Schema{
// 			"file_id": {
// 				Type:        schema.TypeString,
// 				Required:    true,
// 				Description: "ID of the file",
// 			},
// 			"parent": {
// 				Type:        schema.TypeString,
// 				Computed:    true,
// 				Description: "The ID of the file's parent",
// 			},
// 			"name": {
// 				Type:        schema.TypeString,
// 				Computed:    true,
// 				Description: "The name of the file",
// 			},
// 			"mime_type": {
// 				Type:        schema.TypeString,
// 				Computed:    true,
// 				Description: "Name MIME type of the file in Google Drive",
// 			},
// 			"drive_id": {
// 				Type:        schema.TypeString,
// 				Computed:    true,
// 				Description: "The driveId if the file is located in a Shared Drive",
// 			},
// 			"download_path": {
// 				Type:        schema.TypeString,
// 				Optional:    true,
// 				Description: "Use this to specify a local file path to download a (non-Google) file",
// 			},
// 			"export_path": {
// 				Type:          schema.TypeString,
// 				Optional:      true,
// 				Description:   "Use this to specify a local file path to export a Google file (sheet, doc, etc.)",
// 				ConflictsWith: []string{"download_path"},
// 				RequiredWith:  []string{"export_mime_type"},
// 			},
// 			"export_mime_type": {
// 				Type:          schema.TypeString,
// 				RequiredWith:  []string{"export_path"},
// 				ConflictsWith: []string{"download_path"},
// 				Optional:      true,
// 				Description: `Specify the target MIME type for the export.
// For a list of supported MIME types see https://developers.google.com/drive/api/v3/ref-export-formats`,
// 			},
// 			"local_file_path": {
// 				Type:        schema.TypeString,
// 				Computed:    true,
// 				Description: "The path where the local copy or export of the file was created",
// 			},
// 		},
// 		Read: dataSourceReadFile,
// 	}
// }

// func dataSourceReadFile(d *schema.ResourceData, _ any) error {
// 	fileID := d.Get("file_id").(string)
// 	r, err := gsmdrive.GetFile(fileID, "parents,mimeType,driveId,name", "")
// 	if err != nil {
// 		return err
// 	}
// 	d.SetId(fileID)
// 	d.Set("parent", getParent(r))
// 	d.Set("mime_type", r.MimeType)
// 	d.Set("drive_id", r.DriveId)
// 	d.Set("name", r.Name)
// 	downloadPath := d.Get("download_path").(string)
// 	if downloadPath != "" {
// 		filePath, err := gsmdrive.DownloadFile(fileID, downloadPath, false)
// 		if err != nil {
// 			return err
// 		}
// 		d.Set("local_file_path", filePath)
// 	}
// 	exportPath := d.Get("export_path").(string)
// 	if exportPath != "" {
// 		filePath, err := gsmdrive.ExportFile(fileID, d.Get("export_mime_type").(string), exportPath)
// 		if err != nil {
// 			return err
// 		}
// 		d.Set("local_file_path", filePath)
// 	}
// 	return nil
// }
