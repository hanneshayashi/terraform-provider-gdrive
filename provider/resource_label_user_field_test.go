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

func TestAccUserField(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// 1 - Create and Read testing
			{
				Config: testAccUserFieldResourceConfig("UNPUBLISHED_DRAFT", "first field", "UNPUBLISHED_DRAFT", "UNPUBLISHED_DRAFT", "2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gdrive_label.test", "life_cycle.state", "UNPUBLISHED_DRAFT"),
					resource.TestCheckResourceAttr("gdrive_label_user_field.first_field", "life_cycle.state", "UNPUBLISHED_DRAFT"),
					resource.TestCheckResourceAttr("gdrive_label_user_field.first_field", "properties.display_name", "first field"),
					resource.TestCheckResourceAttr("gdrive_label_user_field.second_field", "life_cycle.state", "UNPUBLISHED_DRAFT"),
					resource.TestCheckResourceAttr("gdrive_label_user_field.second_field", "user_options.list_options.max_entries", "2"),
				),
			},
			// 2 - ImportState testing
			{
				ResourceName:            "gdrive_label_user_field.first_field",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdPrefix:     "true,",
				ImportStateVerifyIgnore: []string{"life_cycle", "properties.insert_before_field"},
			},
			// 3 - Rename
			{
				Config: testAccUserFieldResourceConfig("UNPUBLISHED_DRAFT", "renamed field", "UNPUBLISHED_DRAFT", "UNPUBLISHED_DRAFT", "3"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gdrive_label.test", "life_cycle.state", "UNPUBLISHED_DRAFT"),
					resource.TestCheckResourceAttr("gdrive_label_user_field.first_field", "life_cycle.state", "UNPUBLISHED_DRAFT"),
					resource.TestCheckResourceAttr("gdrive_label_user_field.first_field", "properties.display_name", "renamed field"),
					resource.TestCheckResourceAttr("gdrive_label_user_field.second_field", "life_cycle.state", "UNPUBLISHED_DRAFT"),
					resource.TestCheckResourceAttr("gdrive_label_user_field.second_field", "user_options.list_options.max_entries", "3"),
				),
			},
			// 4 - Publish
			{
				Config: testAccUserFieldResourceConfig("PUBLISHED", "renamed field", "PUBLISHED", "PUBLISHED", "3"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gdrive_label.test", "life_cycle.state", "PUBLISHED"),
					resource.TestCheckResourceAttr("gdrive_label_user_field.first_field", "life_cycle.state", "PUBLISHED"),
					resource.TestCheckResourceAttr("gdrive_label_user_field.second_field", "life_cycle.state", "PUBLISHED"),
				),
				ExpectNonEmptyPlan: true,
			},
			// 5 - Rename back
			{
				Config: testAccUserFieldResourceConfig("PUBLISHED", "first field", "PUBLISHED", "PUBLISHED", "3"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gdrive_label.test", "life_cycle.state", "PUBLISHED"),
					resource.TestCheckResourceAttr("gdrive_label_user_field.first_field", "life_cycle.state", "PUBLISHED"),
					resource.TestCheckResourceAttr("gdrive_label_user_field.first_field", "properties.display_name", "first field"),
					resource.TestCheckResourceAttr("gdrive_label_user_field.second_field", "life_cycle.state", "PUBLISHED"),
				),
				ExpectNonEmptyPlan: true,
			},
			// 6 - Publish again
			{
				Config: testAccUserFieldResourceConfig("PUBLISHED", "first field", "PUBLISHED", "PUBLISHED", "3"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gdrive_label.test", "life_cycle.state", "PUBLISHED"),
					resource.TestCheckResourceAttr("gdrive_label_user_field.first_field", "life_cycle.state", "PUBLISHED"),
					resource.TestCheckResourceAttr("gdrive_label_user_field.second_field", "life_cycle.state", "PUBLISHED"),
				),
			},
			// 7 - Disable first field
			{
				Config: testAccUserFieldResourceConfig("PUBLISHED", "first field", "DISABLED", "PUBLISHED", "3"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gdrive_label.test", "life_cycle.state", "PUBLISHED"),
					resource.TestCheckResourceAttr("gdrive_label_user_field.first_field", "life_cycle.state", "DISABLED"),
					resource.TestCheckResourceAttr("gdrive_label_user_field.second_field", "life_cycle.state", "PUBLISHED"),
				),
				ExpectNonEmptyPlan: true,
			},
			// 8 - Publish again
			{
				Config: testAccUserFieldResourceConfig("PUBLISHED", "first field", "DISABLED", "PUBLISHED", "3"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gdrive_label.test", "life_cycle.state", "PUBLISHED"),
					resource.TestCheckResourceAttr("gdrive_label_user_field.first_field", "life_cycle.state", "DISABLED"),
					resource.TestCheckResourceAttr("gdrive_label_user_field.second_field", "life_cycle.state", "PUBLISHED"),
				),
			},
			// 9 - Disable second field
			{
				Config: testAccUserFieldResourceConfig("PUBLISHED", "first field", "DISABLED", "DISABLED", "3"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gdrive_label.test", "life_cycle.state", "PUBLISHED"),
					resource.TestCheckResourceAttr("gdrive_label_user_field.first_field", "life_cycle.state", "DISABLED"),
					resource.TestCheckResourceAttr("gdrive_label_user_field.second_field", "life_cycle.state", "DISABLED"),
				),
				ExpectNonEmptyPlan: true,
			},
			// 10 - Publish again
			{
				Config: testAccUserFieldResourceConfig("PUBLISHED", "first field", "DISABLED", "DISABLED", "3"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gdrive_label.test", "life_cycle.state", "PUBLISHED"),
					resource.TestCheckResourceAttr("gdrive_label_user_field.first_field", "life_cycle.state", "DISABLED"),
					resource.TestCheckResourceAttr("gdrive_label_user_field.second_field", "life_cycle.state", "DISABLED"),
				),
			},
			// 11 - Disable field
			{
				Config: testAccUserFieldResourceConfig("DISABLED", "first field", "DISABLED", "DISABLED", "3"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gdrive_label.test", "life_cycle.state", "DISABLED"),
					resource.TestCheckResourceAttr("gdrive_label_user_field.first_field", "life_cycle.state", "DISABLED"),
					resource.TestCheckResourceAttr("gdrive_label_user_field.second_field", "life_cycle.state", "DISABLED"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccUserFieldResourceConfig(stateLabel, displayNameFirstField, stateFirstField, stateSecondField, maxEntries string) string {
	return fmt.Sprintf(`
resource "gdrive_label" "test" {
  label_type       = "ADMIN"
  use_admin_access = true
  properties {
    title       = "user field test"
  }
  life_cycle {
    state = "%s"
  }
}

resource "gdrive_label_user_field" "first_field" {
  label_id         = gdrive_label.test.label_id
  use_admin_access = true
  properties {
    display_name        = "%s"
    insert_before_field = gdrive_label_user_field.second_field.field_id
  }
  life_cycle {
    state = "%s"
  }
}

resource "gdrive_label_user_field" "second_field" {
  label_id         = gdrive_label.test.label_id
  use_admin_access = true
  properties {
    display_name = "second field"
  }
  life_cycle {
    state = "%s"
  }
  user_options {
	list_options {
		max_entries = %s
	}
  }
}
`, stateLabel, displayNameFirstField, stateFirstField, stateSecondField, maxEntries)
}
