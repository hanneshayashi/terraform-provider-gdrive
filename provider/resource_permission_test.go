package provider

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPermission(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// 1 - Create Files and assign Permissions
			{
				Config: testAccPermissionResourceConfig("1", "FIRST_USER", "reader"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gdrive_permission.permission", "role", "reader"),
					resource.TestCheckResourceAttr("gdrive_permission.permission", "email_address", os.Getenv("FIRST_USER")),
				),
			},
			// 2 - ImportState testing
			{
				ResourceName:        "gdrive_permission.permission",
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: "false,",
			},
			// 2 - Change Role
			{
				Config: testAccPermissionResourceConfig("1", "FIRST_USER", "writer"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gdrive_permission.permission", "role", "writer"),
					resource.TestCheckResourceAttr("gdrive_permission.permission", "email_address", os.Getenv("FIRST_USER")),
				),
			},
			// 3 - Change User
			{
				Config: testAccPermissionResourceConfig("1", "SECOND_USER", "writer"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gdrive_permission.permission", "role", "writer"),
					resource.TestCheckResourceAttr("gdrive_permission.permission", "email_address", os.Getenv("SECOND_USER")),
				),
			},
			// 4 - Delete File
			{
				Config: testAccPermissionResourceConfig("0", "SECOND_USER", "writer"),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccPermissionResourceConfig(count, user, role string) string {
	var permission string
	if count == "1" {
		permission = fmt.Sprintf(`
resource "gdrive_file" "folder" {
  mime_type = "application/vnd.google-apps.folder"
  parent    = gdrive_drive.drive.drive_id
  drive_id  = gdrive_drive.drive.drive_id
  name      = "folder"
}

resource "gdrive_permission" "permission" {
  file_id       = gdrive_file.folder.file_id
  email_address = "%s"
  role          = "%s"
  type          = "user"
}
`, os.Getenv(user), role)
	}
	return strings.Join([]string{`
resource "gdrive_drive" "drive" {
  name                    = "permission_test"
  use_domain_admin_access = true
}
`,
		permission,
	}, "\n")
}
