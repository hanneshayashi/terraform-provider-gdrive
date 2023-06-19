package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccFile(t *testing.T) {
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
				Config: testAccFileResourceConfig(name, testFile, "true"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gdrive_file.folder", "name", name),
					resource.TestCheckResourceAttr("gdrive_file.folder", "mime_type", "application/vnd.google-apps.folder"),
					resource.TestCheckResourceAttr("gdrive_file.file_with_content", "name", name),
					resource.TestCheckResourceAttr("gdrive_file.file_with_content", "mime_type", "text/plain"),
					resource.TestCheckResourceAttr("gdrive_file.import_csv", "name", name),
					resource.TestCheckResourceAttr("gdrive_file.import_csv", "mime_type", "application/vnd.google-apps.spreadsheet"),
					resource.TestCheckResourceAttr("gdrive_file.empty_spreadsheet", "name", name),
					resource.TestCheckResourceAttr("gdrive_file.empty_spreadsheet", "mime_type", "application/vnd.google-apps.spreadsheet"),
				),
			},
			// 2 - ImportState testing
			{
				ResourceName:      "gdrive_file.folder",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// 3 - Rename and change restrictions
			{
				Config: testAccFileResourceConfig(renamed, testFile, "true"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gdrive_file.folder", "name", renamed),
					resource.TestCheckResourceAttr("gdrive_file.folder", "mime_type", "application/vnd.google-apps.folder"),
					resource.TestCheckResourceAttr("gdrive_file.file_with_content", "name", renamed),
					resource.TestCheckResourceAttr("gdrive_file.file_with_content", "mime_type", "text/plain"),
					resource.TestCheckResourceAttr("gdrive_file.import_csv", "name", renamed),
					resource.TestCheckResourceAttr("gdrive_file.import_csv", "mime_type", "application/vnd.google-apps.spreadsheet"),
					resource.TestCheckResourceAttr("gdrive_file.empty_spreadsheet", "name", renamed),
					resource.TestCheckResourceAttr("gdrive_file.empty_spreadsheet", "mime_type", "application/vnd.google-apps.spreadsheet"),
				),
			},
			// 4 - Delete files
			{
				Config: testAccFileResourceConfig(renamed, testFile, "false"),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccFileResourceConfig(name, testFile, createFiles string) string {
	var files string
	if createFiles == "true" {
		files = fmt.Sprintf(`
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
`, name, name, name, testFile, name, testFile)
	}
	return fmt.Sprintf(`resource "gdrive_drive" "drive" {
  name                    = "file test"
  use_domain_admin_access = true
}
%s
`, files)
}
