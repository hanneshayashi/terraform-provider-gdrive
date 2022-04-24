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

func dataSourceFiles() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"query": {
				Type:     schema.TypeString,
				Required: true,
				Description: `A query for filtering the file results.
See the https://developers.google.com/drive/api/v3/search-files for the supported syntax.`,
			},
			"spaces": {
				Type:     schema.TypeString,
				Optional: true,
				Description: `A comma-separated list of spaces to query within the corpus.
Supported values are 'drive', 'appDataFolder' and 'photos'.`,
			},
			"corpora": {
				Type:     schema.TypeString,
				Optional: true,
				Description: `Groupings of files to which the query applies.
Supported groupings are:
'user' (files created by, opened by, or shared directly with the user)
'drive' (files in the specified shared drive as indicated by the 'driveId')
'domain' (files shared to the user's domain)
'allDrives' (A combination of 'user' and 'drive' for all drives where the user is a member).
When able, use 'user' or 'drive', instead of 'allDrives', for efficiency.`,
			},
			"drive_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: `ID of the shared drive.`,
			},
			"include_items_from_all_drives": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: `Whether both My Drive and shared drive items should be included in results.`,
			},
			"files": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"file_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"mime_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"drive_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"parent": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
		Read: dataSourceReadFiles,
	}
}

func dataSourceReadFiles(d *schema.ResourceData, _ any) error {
	query := d.Get("query").(string)
	files := make([]map[string]any, 0)
	results, err := gsmdrive.ListFiles(query, d.Get("drive_id").(string), d.Get("corpora").(string), "", "", d.Get("spaces").(string), "files(id,name,parents,mimeType,driveId),nextPageToken", d.Get("include_items_from_all_drives").(bool), 1)
	for f := range results {
		files = append(files, map[string]any{
			"file_id":   f.Id,
			"name":      f.Name,
			"mime_type": f.MimeType,
			"drive_id":  f.DriveId,
			"parent":    getParent(f),
		})
	}
	e := <-err
	if e != nil {
		return e
	}
	d.Set("files", files)
	d.SetId(query)
	return nil
}
