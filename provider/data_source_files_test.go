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
	"fmt"
	"testing"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccFiles(t *testing.T) {
	name, err := uuid.GenerateUUID()
	if err != nil {
		panic(err)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// 1 - Create Folder
			{
				Config: testAccFilesDataSourceConfig(name, "true", "false"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gdrive_file.folder", "name", name),
					resource.TestCheckResourceAttr("gdrive_file.folder", "mime_type", "application/vnd.google-apps.folder"),
				),
			},
			// 2 - Create Data Source
			{
				Config: testAccFilesDataSourceConfig(name, "true", "true"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gdrive_files.files", "files.0.name", name),
					resource.TestCheckResourceAttr("data.gdrive_files.files", "files.0.mime_type", "application/vnd.google-apps.folder"),
				),
			},
			// 3 - Delete files
			{
				Config: testAccFilesDataSourceConfig(name, "false", "true"),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccFilesDataSourceConfig(name, createFiles, createDS string) string {
	var files string
	var ds string
	if createDS == "true" {
		ds = fmt.Sprintf(`
data "gdrive_files" "files" {
  query                         = "name contains '%s'"
  include_items_from_all_drives = true
  corpora                       = "allDrives"
}`, name)
	}
	if createFiles == "true" {
		files = fmt.Sprintf(`
%s

resource "gdrive_file" "folder" {
  name      = "%s"
  mime_type = "application/vnd.google-apps.folder"
  parent    = gdrive_drive.drive.drive_id
  drive_id  = gdrive_drive.drive.drive_id
}
`, ds, name)
	}
	return fmt.Sprintf(`resource "gdrive_drive" "drive" {
  name                    = "file test"
  use_domain_admin_access = true
}
%s
`, files)
}
