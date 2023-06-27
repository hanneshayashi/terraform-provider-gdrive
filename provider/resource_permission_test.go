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
			// 3 - Change Role
			{
				Config: testAccPermissionResourceConfig("1", "FIRST_USER", "writer"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gdrive_permission.permission", "role", "writer"),
					resource.TestCheckResourceAttr("gdrive_permission.permission", "email_address", os.Getenv("FIRST_USER")),
				),
			},
			// 4 - Change User
			{
				Config: testAccPermissionResourceConfig("1", "SECOND_USER", "writer"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gdrive_permission.permission", "role", "writer"),
					resource.TestCheckResourceAttr("gdrive_permission.permission", "email_address", os.Getenv("SECOND_USER")),
				),
			},
			// 5 - Delete File
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
  use_domain_admin_access = false
  file_id                 = gdrive_file.folder.file_id
  email_address           = "%s"
  role                    = "%s"
  type                    = "user"
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
