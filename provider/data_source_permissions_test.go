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

func TestAccPermissions(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// 1 (File) - Create Files and assign Permissions
			{
				Config: testAccPermissionsDataSourceConfig("file", "1", "FIRST_USER", "SECOND_USER", "reader", "writer"),
				Check:  resource.ComposeAggregateTestCheckFunc(
				// TODO
				// resource.TestCheckResourceAttr("gdrive_permission_policy.policy", "role", "reader"),
				// resource.TestCheckResourceAttr("gdrive_permission_policy.policy", "email_address", os.Getenv("FIRST_USER")),
				),
			},
			// 2 (File) - ImportState testing
			{
				ResourceName:        "gdrive_permissions_policy.policy",
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: "false,",
			},
			// 3 (File) - Change first Role
			{
				Config: testAccPermissionsDataSourceConfig("file", "1", "FIRST_USER", "SECOND_USER", "writer", "writer"),
				Check:  resource.ComposeAggregateTestCheckFunc(
				// TODO
				// resource.TestCheckResourceAttr("gdrive_permission_policy.policy", "role", "reader"),
				// resource.TestCheckResourceAttr("gdrive_permission_policy.policy", "email_address", os.Getenv("FIRST_USER")),
				),
			},
			// 4 (File) - Remove Second User
			{
				Config: testAccPermissionsDataSourceConfig("file", "1", "FIRST_USER", "", "writer", "writer"),
				Check:  resource.ComposeAggregateTestCheckFunc(
				// TODO
				// resource.TestCheckResourceAttr("gdrive_permission_policy.policy", "role", "reader"),
				// resource.TestCheckResourceAttr("gdrive_permission_policy.policy", "email_address", os.Getenv("FIRST_USER")),
				),
			},
			// 5 (File) - Re-Add Second User with different permission
			{
				Config: testAccPermissionsDataSourceConfig("file", "1", "FIRST_USER", "", "writer", "reader"),
				Check:  resource.ComposeAggregateTestCheckFunc(
				// TODO
				// resource.TestCheckResourceAttr("gdrive_permission_policy.policy", "role", "reader"),
				// resource.TestCheckResourceAttr("gdrive_permission_policy.policy", "email_address", os.Getenv("FIRST_USER")),
				),
			},
			// 6 (Drive) - Create Files and assign Permissions
			{
				Config: testAccPermissionsDataSourceConfig("drive", "1", "FIRST_USER", "SECOND_USER", "reader", "writer"),
				Check:  resource.ComposeAggregateTestCheckFunc(
				// TODO
				// resource.TestCheckResourceAttr("gdrive_permission_policy.policy", "role", "reader"),
				// resource.TestCheckResourceAttr("gdrive_permission_policy.policy", "email_address", os.Getenv("FIRST_USER")),
				),
			},
			// 7 (Drive) - ImportState testing
			{
				ResourceName:        "gdrive_permissions_policy.policy",
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: "false,",
			},
			// 8 (Drive) - Change first Role
			{
				Config: testAccPermissionsDataSourceConfig("drive", "1", "FIRST_USER", "SECOND_USER", "writer", "writer"),
				Check:  resource.ComposeAggregateTestCheckFunc(
				// TODO
				// resource.TestCheckResourceAttr("gdrive_permission_policy.policy", "role", "reader"),
				// resource.TestCheckResourceAttr("gdrive_permission_policy.policy", "email_address", os.Getenv("FIRST_USER")),
				),
			},
			// 9 (Drive) -Remove Second User
			{
				Config: testAccPermissionsDataSourceConfig("drive", "1", "FIRST_USER", "", "writer", "writer"),
				Check:  resource.ComposeAggregateTestCheckFunc(
				// TODO
				// resource.TestCheckResourceAttr("gdrive_permission_policy.policy", "role", "reader"),
				// resource.TestCheckResourceAttr("gdrive_permission_policy.policy", "email_address", os.Getenv("FIRST_USER")),
				),
			},
			// 10 (Drive) - Re-Add Second User with different permission
			{
				Config: testAccPermissionsDataSourceConfig("drive", "1", "FIRST_USER", "", "writer", "reader"),
				Check:  resource.ComposeAggregateTestCheckFunc(
				// TODO
				// resource.TestCheckResourceAttr("gdrive_permission_policy.policy", "role", "reader"),
				// resource.TestCheckResourceAttr("gdrive_permission_policy.policy", "email_address", os.Getenv("FIRST_USER")),
				),
			},
			// 11 - Delete File
			{
				Config: testAccPermissionsDataSourceConfig("drive", "0", "FIRST_USER", "FIRST_USER", "writer", "reader"),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccPermissionsDataSourceConfig(fileDrive, count, firstUser, secondUser, roleFirstUser, roleSecondUser string) string {
	var permission string
	var fileId string
	var organizer string
	if fileDrive == "file" {
		fileId = "gdrive_file.folder.file_id"
	} else {
		fileId = "gdrive_drive.drive.drive_id"
		organizer = fmt.Sprintf(`
  {
    email_address = "%s"
    role          = "organizer"
    type          = "user"
  },`, os.Getenv("SUBJECT"))
	}
	if secondUser != "" {
		secondUser = fmt.Sprintf(`
  {
    email_address = "%s"
    role          = "%s"
    type          = "user"
  },`, os.Getenv(secondUser), roleSecondUser)
	}
	firstUser = fmt.Sprintf(`
  {
    email_address = "%s"
    role          = "%s"
    type          = "user"
  },`, os.Getenv(firstUser), roleFirstUser)
	if count == "1" {
		permission = fmt.Sprintf(`
data "gdrive_permissions" "some_file_permissions" {
  file_id = %s
}

resource "gdrive_file" "folder" {
  mime_type = "application/vnd.google-apps.folder"
  parent    = gdrive_drive.drive.drive_id
  drive_id  = gdrive_drive.drive.drive_id
  name      = "folder"
}

resource "gdrive_permissions_policy" "policy" {
  file_id                 = %s
  use_domain_admin_access = false
  permissions = [
    %s
    %s
    %s
  ]
}

`, fileId, fileId, organizer, firstUser, secondUser)
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
