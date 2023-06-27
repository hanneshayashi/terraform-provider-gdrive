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

func TestAccLabelAssignment(t *testing.T) {
	dateBefore := "2023-01-31"
	dateAfter := "2023-02-28"
	intBefore := "12"
	intAfter := "13"
	textBefore := "foo"
	textAfter := "bar"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// 1 - Create Label and Fields as Draft
			{
				Config: testAccLabelAssignmentResourceConfig("UNPUBLISHED_DRAFT", "0", dateBefore, intBefore, textBefore),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gdrive_label.test", "life_cycle.state", "UNPUBLISHED_DRAFT"),
				),
			},
			// 2 - Publish Label
			{
				Config: testAccLabelAssignmentResourceConfig("PUBLISHED", "0", dateBefore, intBefore, textBefore),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gdrive_label.test", "life_cycle.state", "PUBLISHED"),
				),
			},
			// 3 - Create Label Assignment
			{
				Config: testAccLabelAssignmentResourceConfig("PUBLISHED", "1", dateBefore, intBefore, textBefore),
				Check:  resource.ComposeAggregateTestCheckFunc(
				// TODO
				// resource.TestCheckResourceAttr("gdrive_label_assignment.assignment", "field[0].values[0]", dateBefore),
				// resource.TestCheckResourceAttr("gdrive_label_assignment.assignment", "field[1].values[0]", intBefore),
				// resource.TestCheckResourceAttr("gdrive_label_assignment.assignment", "field[3].values[0]", textBefore),
				),
			},
			// 4 - ImportState testing
			{
				ResourceName:      "gdrive_label_assignment.assignment",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// 5 - Change values
			{
				Config: testAccLabelAssignmentResourceConfig("PUBLISHED", "1", dateAfter, intAfter, textAfter),
				Check:  resource.ComposeAggregateTestCheckFunc(
				// TODO
				// resource.TestCheckResourceAttr("gdrive_label_assignment.assignment", "field[0].values[0]", dateAfter),
				// resource.TestCheckResourceAttr("gdrive_label_assignment.assignment", "field[1].values[0]", intAfter),
				// resource.TestCheckResourceAttr("gdrive_label_assignment.assignment", "field[3].values[0]", textAfter),
				),
			},
			// 6 - Disable Label and Delete Files and Assignment
			{
				Config: testAccLabelAssignmentResourceConfig("DISABLED", "0", dateAfter, intAfter, textAfter),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gdrive_label.test", "life_cycle.state", "DISABLED"),
					resource.TestCheckResourceAttr("gdrive_label_date_field.date_field", "life_cycle.state", "DISABLED"),
					resource.TestCheckResourceAttr("gdrive_label_integer_field.integer_field", "life_cycle.state", "DISABLED"),
					resource.TestCheckResourceAttr("gdrive_label_selection_field.selection_field", "life_cycle.state", "DISABLED"),
					resource.TestCheckResourceAttr("gdrive_label_selection_choice.first_choice", "life_cycle.state", "DISABLED"),
					resource.TestCheckResourceAttr("gdrive_label_selection_choice.second_choice", "life_cycle.state", "DISABLED"),
					resource.TestCheckResourceAttr("gdrive_label_text_field.text_field", "life_cycle.state", "DISABLED"),
					resource.TestCheckResourceAttr("gdrive_label_user_field.user_field", "life_cycle.state", "DISABLED"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccLabelAssignmentResourceConfig(state, count, dateValue, intValue, textValue string) string {
	var lifeCycle string
	if state == "DISABLED" {
		lifeCycle = `
life_cycle {
  state = "DISABLED"
}`
	}
	var assignment string
	if count == "1" {
		assignment = fmt.Sprintf(`
resource "gdrive_file" "folder" {
  mime_type = "application/vnd.google-apps.folder"
  parent    = gdrive_drive.drive.drive_id
  drive_id  = gdrive_drive.drive.drive_id
  name      = "folder"
}

resource "gdrive_file" "file" {
  name      = "file"
  mime_type = "application/vnd.google-apps.spreadsheet"
  drive_id  = gdrive_drive.drive.drive_id
  parent    = gdrive_file.folder.file_id
}

resource "gdrive_label_assignment" "assignment" {
  file_id  = gdrive_file.file.file_id
  label_id = gdrive_label.test.label_id
  fields = [
    {
      field_id   = gdrive_label_date_field.date_field.field_id
      value_type = "dateString"
      values     = ["%s"]
    },
    {
      field_id   =  gdrive_label_integer_field.integer_field.field_id
      value_type = "integer"
      values     = [%s]
    },
    {
      field_id   =  gdrive_label_selection_field.selection_field.field_id
      value_type = "selection"
      values     = [
        gdrive_label_selection_choice.first_choice.choice_id,
        gdrive_label_selection_choice.second_choice.choice_id
      ]
    },
    {
      field_id   =  gdrive_label_text_field.text_field.field_id
      value_type = "text"
      values     = ["%s"]
    },
    {
      field_id   =  gdrive_label_user_field.user_field.field_id
      value_type = "user"
      values     = [
        "%s",
	    "%s"
      ]
    }
  ]
}`, dateValue, intValue, textValue, os.Getenv("FIRST_USER"), os.Getenv("SECOND_USER"))
	}
	return strings.Join([]string{
		`resource "gdrive_drive" "drive" {
  name                    = "label_assignment_test"
  use_domain_admin_access = true
}
`,
		fmt.Sprintf(`resource "gdrive_label" "test" {
  label_type       = "ADMIN"
  use_admin_access = true
  properties {
    title = "assignment test"
  }
  life_cycle {
    state = "%s"
  }
}`, state),
		fmt.Sprintf(`resource "gdrive_label_date_field" "date_field" {
		  label_id         = gdrive_label.test.label_id
		  use_admin_access = true
		  properties {
		    display_name = "date field"
		  }
		  %s
		}`, lifeCycle),
		fmt.Sprintf(`resource "gdrive_label_integer_field" "integer_field" {
  label_id         = gdrive_label.test.label_id
  use_admin_access = true
  properties {
    display_name = "integer field"
  }
  %s
}`, lifeCycle),
		fmt.Sprintf(`resource "gdrive_label_selection_field" "selection_field" {
  label_id         = gdrive_label.test.label_id
  use_admin_access = true
  properties {
    display_name = "selection field"
  }
  selection_options {
    list_options {
      max_entries = 2
    }
  }
  %s
}`, lifeCycle),
		fmt.Sprintf(`resource "gdrive_label_selection_choice" "first_choice" {
  label_id         = gdrive_label_selection_field.selection_field.label_id
  field_id         = gdrive_label_selection_field.selection_field.field_id
  use_admin_access = true
  properties {
    display_name = "first choice"
  }
  %s
}`, lifeCycle),
		fmt.Sprintf(`resource "gdrive_label_selection_choice" "second_choice" {
  label_id         = gdrive_label_selection_field.selection_field.label_id
  field_id         = gdrive_label_selection_field.selection_field.field_id
  use_admin_access = true
  properties {
    display_name = "second choice"
  }
  %s
}`, lifeCycle),
		fmt.Sprintf(`resource "gdrive_label_text_field" "text_field" {
  label_id         = gdrive_label.test.label_id
  use_admin_access = true
  properties {
    display_name = "text field"
  }
  %s
}`, lifeCycle),
		fmt.Sprintf(`resource "gdrive_label_user_field" "user_field" {
  label_id         = gdrive_label.test.label_id
  use_admin_access = true
  properties {
    display_name = "user field"
  }
  user_options {
    list_options {
      max_entries = 2
    }
  }
  %s
}`, lifeCycle),
		assignment,
	}, "\n")
}
