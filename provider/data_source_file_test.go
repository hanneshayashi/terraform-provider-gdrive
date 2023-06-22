package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccFileDS(t *testing.T) {
	name := "tftest"
	renamed := "testtest_renamed"
	content := `
h1,h2
c1,c2
c3,c4
`
	testFile := fmt.Sprintf("%s/test.csv", t.TempDir())
	d1 := []byte(content)
	err := os.WriteFile(testFile, d1, 0644)
	if err != nil {
		panic(err)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// 1 - Create and Read testing
			{
				Config: testAccFileDataSourceConfig(name, testFile, "true"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gdrive_file.file_metadata", "name", name),
					resource.TestCheckResourceAttr("data.gdrive_file.file_download", "name", name),
					resource.TestCheckResourceAttr("data.gdrive_file.file_export", "name", name),
					resource.TestCheckResourceAttr("data.gdrive_file.file_metadata", "mime_type", "application/vnd.google-apps.folder"),
					resource.TestCheckResourceAttr("data.gdrive_file.file_download", "mime_type", "text/plain"),
					resource.TestCheckResourceAttr("data.gdrive_file.file_export", "mime_type", "application/vnd.google-apps.spreadsheet"),
				),
			},
			// 2 - Rename
			{
				Config: testAccFileDataSourceConfig(renamed, testFile, "true"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gdrive_file.file_metadata", "name", renamed),
					resource.TestCheckResourceAttr("data.gdrive_file.file_download", "name", renamed),
					resource.TestCheckResourceAttr("data.gdrive_file.file_export", "name", renamed),
				),
			},
			// 4 - Delete files
			{
				Config: testAccFileDataSourceConfig(renamed, testFile, "false"),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccFileDataSourceConfig(name, testFile, createFiles string) string {
	var files string
	if createFiles == "true" {
		files = fmt.Sprintf(`
data "gdrive_file" "file_metadata" {
  file_id = gdrive_file.folder.file_id
}

data "gdrive_file" "file_download" {
  file_id       = gdrive_file.file_with_content.file_id
  download_path = "%s_"
}

data "gdrive_file" "file_export" {
  file_id          = gdrive_file.import_csv.file_id
  export_path      =  "%s__"
  export_mime_type = "text/csv"
}

resource "gdrive_file" "folder" {
  name      = "%s"
  mime_type = "application/vnd.google-apps.folder"
  parent    = gdrive_drive.drive.drive_id
  drive_id  = gdrive_drive.drive.drive_id
}

resource "gdrive_file" "empty_spreadsheet" {
  name      = "%s"
  mime_type = "application/vnd.google-apps.spreadsheet"
  drive_id  = gdrive_file.folder.drive_id
  parent    = gdrive_file.folder.file_id
}

resource "gdrive_file" "file_with_content" {
  name      = "%s"
  mime_type = "text/plain"
  drive_id  = gdrive_file.folder.drive_id
  parent    = gdrive_file.folder.file_id
  content   = "%s"
}

resource "gdrive_file" "import_csv" {
  name             = "%s"
  mime_type        = "application/vnd.google-apps.spreadsheet"
  mime_type_source = "text/csv"
  content          = "%s"
  drive_id  	   = gdrive_file.folder.drive_id
  parent           = gdrive_file.folder.file_id
}
`, name, name, name, name, name, testFile, name, testFile)
	}
	return fmt.Sprintf(`resource "gdrive_drive" "drive" {
  name                    = "file test"
  use_domain_admin_access = true
}
%s
`, files)
}
