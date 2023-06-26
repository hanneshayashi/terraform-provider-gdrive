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

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDrive(t *testing.T) {
	name := "tftest"
	renamed := "testtest_renamed"
	restrictionsBefore := "false"
	restrictionsAfter := "true"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// 1 - Create and Read testing
			{
				Config: testAccDriveResourceConfig(name, restrictionsBefore),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gdrive_drive.drive_simple", "name", name),
					resource.TestCheckResourceAttr("gdrive_drive.drive_restrictions", "name", name),
					resource.TestCheckResourceAttr("gdrive_drive.drive_restrictions", "restrictions.admin_managed_restrictions", restrictionsBefore),
					resource.TestCheckResourceAttr("gdrive_drive.drive_restrictions", "restrictions.drive_members_only", restrictionsBefore),
					resource.TestCheckResourceAttr("gdrive_drive.drive_restrictions", "restrictions.copy_requires_writer_permission", restrictionsBefore),
					resource.TestCheckResourceAttr("gdrive_drive.drive_restrictions", "restrictions.domain_users_only", restrictionsBefore),
				),
			},
			// 2 - ImportState testing
			{
				ResourceName:        "gdrive_drive.drive_simple",
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: "true,",
			},
			// 3 - Rename and change restrictions
			{
				Config: testAccDriveResourceConfig(renamed, restrictionsAfter),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gdrive_drive.drive_simple", "name", renamed),
					resource.TestCheckResourceAttr("gdrive_drive.drive_restrictions", "name", renamed),
					resource.TestCheckResourceAttr("gdrive_drive.drive_restrictions", "restrictions.admin_managed_restrictions", restrictionsAfter),
					resource.TestCheckResourceAttr("gdrive_drive.drive_restrictions", "restrictions.drive_members_only", restrictionsAfter),
					resource.TestCheckResourceAttr("gdrive_drive.drive_restrictions", "restrictions.copy_requires_writer_permission", restrictionsAfter),
					resource.TestCheckResourceAttr("gdrive_drive.drive_restrictions", "restrictions.domain_users_only", restrictionsAfter),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccDriveResourceConfig(name, restrictions string) string {
	return fmt.Sprintf(`
resource "gdrive_drive" "drive_simple" {
  name                    = "%s"
  use_domain_admin_access = true
}

resource "gdrive_drive" "drive_restrictions" {
  name                    = "%s"
  use_domain_admin_access = true
  restrictions {
    admin_managed_restrictions      = %s
    drive_members_only              = %s
    copy_requires_writer_permission = %s
    domain_users_only               = %s
  }
}`, name, name, restrictions, restrictions, restrictions, restrictions)
}
