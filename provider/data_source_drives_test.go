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

func TestAccDrives(t *testing.T) {
	name, err := uuid.GenerateUUID()
	if err != nil {
		panic(err)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// 1 - Create Drive
			{
				Config: testAccDrivesDataSourceConfig(name, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gdrive_drive.drive_restrictions", "name", name),
					resource.TestCheckResourceAttr("gdrive_drive.drive_restrictions", "restrictions.admin_managed_restrictions", "true"),
				),
			},
			// 2 - Rename and change restrictions
			{
				Config: testAccDrivesDataSourceConfig(name, "1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gdrive_drives.drives", "drives.0.name", name),
					resource.TestCheckResourceAttr("data.gdrive_drives.drives", "drives.0.restrictions.admin_managed_restrictions", "true"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccDrivesDataSourceConfig(name, createDS string) string {
	if createDS != "" {
		createDS = fmt.Sprintf(`
data "gdrive_drives" "drives" {
  query = "name contains '%s'"
}
`, name)
	}
	return fmt.Sprintf(`
%s

resource "gdrive_drive" "drive_restrictions" {
  name                    = "%s"
  use_domain_admin_access = true
  restrictions {
    admin_managed_restrictions = true
  }
}`, createDS, name)
}
