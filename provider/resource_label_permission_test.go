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
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccLabelPermission(t *testing.T) {
	roleBefore := "READER"
	roleAfter := "APPLIER"
	firstUser := os.Getenv("FIRST_USER")
	secondUser := os.Getenv("SECOND_USER")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// 1 - Create Files and assign Permissions
			{
				Config: testAccLabelPermissionResourceConfig(firstUser, roleBefore),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gdrive_label_permission.permission", "role", roleBefore),
					resource.TestCheckResourceAttr("gdrive_label_permission.permission", "email", firstUser),
					resource.TestCheckResourceAttr("gdrive_label_permission.permission_audience", "role", roleBefore),
					resource.TestCheckResourceAttr("gdrive_label_permission.permission_audience", "audience", "audiences/default"),
				),
			},
			// 2 - ImportState testing
			{
				ResourceName:            "gdrive_label_permission.permission",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdPrefix:     "true,",
				ImportStateVerifyIgnore: []string{"parent"},
			},
			// 3 - Change Role
			{
				Config: testAccLabelPermissionResourceConfig(firstUser, roleAfter),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gdrive_label_permission.permission", "role", roleAfter),
					resource.TestCheckResourceAttr("gdrive_label_permission.permission", "email", firstUser),
					resource.TestCheckResourceAttr("gdrive_label_permission.permission_audience", "role", roleAfter),
					resource.TestCheckResourceAttr("gdrive_label_permission.permission_audience", "audience", "audiences/default"),
				),
			},
			// 4 - Change User
			{
				Config: testAccLabelPermissionResourceConfig(secondUser, roleAfter),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gdrive_label_permission.permission", "role", roleAfter),
					resource.TestCheckResourceAttr("gdrive_label_permission.permission", "email", secondUser),
					resource.TestCheckResourceAttr("gdrive_label_permission.permission_audience", "role", roleAfter),
					resource.TestCheckResourceAttr("gdrive_label_permission.permission_audience", "audience", "audiences/default"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccLabelPermissionResourceConfig(email, role string) string {
	return fmt.Sprintf(`
resource "gdrive_label" "test" {
  label_type       = "ADMIN"
  use_admin_access = true
  properties {
    title = "label permission test"
  }
}

resource "gdrive_label_permission" "permission" {
  parent           = gdrive_label.test.label_id
  use_admin_access = true
  email            = "%s"
  role             = "%s"
}

resource "gdrive_label_permission" "permission_audience" {
  parent           = gdrive_label.test.label_id
  use_admin_access = true
  audience         = "audiences/default"
  role             = "%s"
}`, email, role, role)
}
