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

func TestAccLabelsDS(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// 1 - Create and Read testing
			{
				Config: testAccLabelsDataSourceConfig(""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gdrive_label_selection_field.field", "properties.display_name", "field"),
				),
			},
			// 2 - Create Data Source
			{
				Config: testAccLabelsDataSourceConfig("1"),
				Check:  resource.ComposeAggregateTestCheckFunc(),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccLabelsDataSourceConfig(ds string) string {
	if ds != "" {
		ds = `
data "gdrive_labels" "labels" {
  use_admin_access = true
}
`
	}
	return fmt.Sprintf(`
%s

resource "gdrive_label" "test" {
  label_type       = "ADMIN"
  use_admin_access = true
  properties {
    title = "selection choice test"
  }
}

resource "gdrive_label_selection_field" "field" {
  label_id         = gdrive_label.test.label_id
  use_admin_access = true
  properties {
    display_name = "field"
  }
}

resource "gdrive_label_selection_choice" "first_choice" {
  label_id         = gdrive_label_selection_field.field.label_id
  field_id         = gdrive_label_selection_field.field.field_id
  use_admin_access = true
  properties {
    display_name = "foo"
    badge_config {
      color {
        blue = 0.5
      }
    }
    insert_before_choice = gdrive_label_selection_choice.second_choice.choice_id
  }
}

resource "gdrive_label_selection_choice" "second_choice" {
  label_id         = gdrive_label_selection_field.field.label_id
  field_id         = gdrive_label_selection_field.field.field_id
  use_admin_access = true
  properties {
    display_name = "bar"
  }
}
`, ds)
}
