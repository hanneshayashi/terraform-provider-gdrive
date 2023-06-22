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
