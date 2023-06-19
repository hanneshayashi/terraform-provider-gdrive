package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSelectionChoice(t *testing.T) {
	name := "choice"
	renamed := "choice_renamed"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// 1 - Create and Read testing
			{
				Config: testAccSelectionChoiceResourceConfig(name, "0.5", "UNPUBLISHED_DRAFT", "UNPUBLISHED_DRAFT"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gdrive_label_selection_choice.first_choice", "life_cycle.state", "UNPUBLISHED_DRAFT"),
					resource.TestCheckResourceAttr("gdrive_label_selection_choice.first_choice", "properties.badge_config.color.blue", "0.5"),
					resource.TestCheckResourceAttr("gdrive_label_selection_choice.second_choice", "life_cycle.state", "UNPUBLISHED_DRAFT"),
					resource.TestCheckResourceAttr("gdrive_label_selection_choice.first_choice", "properties.display_name", name),
					resource.TestCheckResourceAttr("gdrive_label_selection_choice.second_choice", "properties.display_name", name),
				),
			},
			// 2 - ImportState testing
			{
				ResourceName:            "gdrive_label_selection_choice.first_choice",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdPrefix:     "true,",
				ImportStateVerifyIgnore: []string{"life_cycle", "properties.insert_before_choice"},
			},
			// 3 - Rename and change color
			{
				Config: testAccSelectionChoiceResourceConfig(renamed, "1", "UNPUBLISHED_DRAFT", "UNPUBLISHED_DRAFT"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gdrive_label_selection_choice.first_choice", "life_cycle.state", "UNPUBLISHED_DRAFT"),
					resource.TestCheckResourceAttr("gdrive_label_selection_choice.first_choice", "properties.badge_config.color.blue", "1"),
					resource.TestCheckResourceAttr("gdrive_label_selection_choice.second_choice", "life_cycle.state", "UNPUBLISHED_DRAFT"),
					resource.TestCheckResourceAttr("gdrive_label_selection_choice.first_choice", "properties.display_name", renamed),
					resource.TestCheckResourceAttr("gdrive_label_selection_choice.second_choice", "properties.display_name", renamed),
				),
			},
			// 4 - Publish
			{
				Config: testAccSelectionChoiceResourceConfig(name, "1", "UNPUBLISHED_DRAFT", "PUBLISHED"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gdrive_label_selection_choice.first_choice", "life_cycle.state", "UNPUBLISHED_DRAFT"),
					resource.TestCheckResourceAttr("gdrive_label_selection_choice.second_choice", "life_cycle.state", "UNPUBLISHED_DRAFT"),
				),
				ExpectNonEmptyPlan: true,
			},
			// 5 - Disable Choices
			{
				Config: testAccSelectionChoiceResourceConfig(name, "1", "DISABLED", "DISABLED"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gdrive_label_selection_choice.first_choice", "life_cycle.state", "DISABLED"),
					resource.TestCheckResourceAttr("gdrive_label_selection_choice.second_choice", "life_cycle.state", "DISABLED"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccSelectionChoiceResourceConfig(name, blue, state, labelState string) string {
	return fmt.Sprintf(`
resource "gdrive_label" "test" {
  label_type       = "ADMIN"
  use_admin_access = true
  properties {
    title = "selection choice test"
  }
  life_cycle {
    state = "%s"
  }
}

resource "gdrive_label_selection_field" "field" {
  label_id         = gdrive_label.test.label_id
  use_admin_access = true
  properties {
    display_name = "field"
  }
  life_cycle {
    state = "%s"
  }
}

resource "gdrive_label_selection_choice" "first_choice" {
  label_id         = gdrive_label_selection_field.field.label_id
  field_id         = gdrive_label_selection_field.field.field_id
  use_admin_access = true
  properties {
    display_name = "%s"
    badge_config {
      color {
        blue = %s
      }
    }
    insert_before_choice = gdrive_label_selection_choice.second_choice.choice_id
  }
  life_cycle {
    state = "%s"
  }
}

resource "gdrive_label_selection_choice" "second_choice" {
  label_id         = gdrive_label_selection_field.field.label_id
  field_id         = gdrive_label_selection_field.field.field_id
  use_admin_access = true
  properties {
    display_name = "%s"
  }
  life_cycle {
    state = "%s"
  }
}
`, labelState, state, name, blue, state, name, state)
}
