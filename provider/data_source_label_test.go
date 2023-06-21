package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccLabelDS(t *testing.T) {
	// renamed := "choice_renamed"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// 1 - Create and Read testing
			{
				Config: testAccLabelDataSourceConfig(""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gdrive_label_selection_field.field", "properties.display_name", "field"),
				),
			},
			// 2 - Create Data Source
			{
				Config: testAccLabelDataSourceConfig("1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gdrive_label.label", "properties.title", "selection choice test"),
					resource.TestCheckResourceAttr("data.gdrive_label.label", "fields.0.properties.display_name", "field"),
					resource.TestCheckResourceAttr("data.gdrive_label.label", "fields.0.selection_options.choices.0.properties.badge_config.color.blue", "0.5"),
					resource.TestCheckResourceAttr("data.gdrive_label.label", "fields.0.selection_options.choices.0.properties.display_name", "foo"),
					resource.TestCheckResourceAttr("data.gdrive_label.label", "fields.0.selection_options.choices.1.properties.display_name", "bar"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccLabelDataSourceConfig(ds string) string {
	if ds != "" {
		ds = `
data "gdrive_label" "label" {
  name             = gdrive_label.test.name
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
