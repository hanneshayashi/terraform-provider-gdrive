package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccIntegerField(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// 1 - Create and Read testing
			{
				Config: testAccIntegerFieldResourceConfig("UNPUBLISHED_DRAFT", "first field", "UNPUBLISHED_DRAFT", "UNPUBLISHED_DRAFT"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gdrive_label.test", "life_cycle.state", "UNPUBLISHED_DRAFT"),
					resource.TestCheckResourceAttr("gdrive_label_integer_field.first_field", "life_cycle.state", "UNPUBLISHED_DRAFT"),
					resource.TestCheckResourceAttr("gdrive_label_integer_field.first_field", "properties.display_name", "first field"),
					resource.TestCheckResourceAttr("gdrive_label_integer_field.second_field", "life_cycle.state", "UNPUBLISHED_DRAFT"),
				),
			},
			// 2 - ImportState testing
			{
				ResourceName:            "gdrive_label_integer_field.first_field",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdPrefix:     "true,",
				ImportStateVerifyIgnore: []string{"life_cycle", "properties.insert_before_field"},
			},
			// 3 - Rename
			{
				Config: testAccIntegerFieldResourceConfig("UNPUBLISHED_DRAFT", "renamed field", "UNPUBLISHED_DRAFT", "UNPUBLISHED_DRAFT"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gdrive_label.test", "life_cycle.state", "UNPUBLISHED_DRAFT"),
					resource.TestCheckResourceAttr("gdrive_label_integer_field.first_field", "life_cycle.state", "UNPUBLISHED_DRAFT"),
					resource.TestCheckResourceAttr("gdrive_label_integer_field.first_field", "properties.display_name", "renamed field"),
					resource.TestCheckResourceAttr("gdrive_label_integer_field.second_field", "life_cycle.state", "UNPUBLISHED_DRAFT"),
				),
			},
			// 4 - Publish
			{
				Config: testAccIntegerFieldResourceConfig("PUBLISHED", "renamed field", "PUBLISHED", "PUBLISHED"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gdrive_label.test", "life_cycle.state", "PUBLISHED"),
					resource.TestCheckResourceAttr("gdrive_label_integer_field.first_field", "life_cycle.state", "PUBLISHED"),
					resource.TestCheckResourceAttr("gdrive_label_integer_field.second_field", "life_cycle.state", "PUBLISHED"),
				),
				ExpectNonEmptyPlan: true,
			},
			// 5 - Rename back
			{
				Config: testAccIntegerFieldResourceConfig("PUBLISHED", "first field", "PUBLISHED", "PUBLISHED"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gdrive_label.test", "life_cycle.state", "PUBLISHED"),
					resource.TestCheckResourceAttr("gdrive_label_integer_field.first_field", "life_cycle.state", "PUBLISHED"),
					resource.TestCheckResourceAttr("gdrive_label_integer_field.first_field", "properties.display_name", "first field"),
					resource.TestCheckResourceAttr("gdrive_label_integer_field.second_field", "life_cycle.state", "PUBLISHED"),
				),
				ExpectNonEmptyPlan: true,
			},
			// 6 - Publish again
			{
				Config: testAccIntegerFieldResourceConfig("PUBLISHED", "first field", "PUBLISHED", "PUBLISHED"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gdrive_label.test", "life_cycle.state", "PUBLISHED"),
					resource.TestCheckResourceAttr("gdrive_label_integer_field.first_field", "life_cycle.state", "PUBLISHED"),
					resource.TestCheckResourceAttr("gdrive_label_integer_field.second_field", "life_cycle.state", "PUBLISHED"),
				),
			},
			// 7 - Disable first field
			{
				Config: testAccIntegerFieldResourceConfig("PUBLISHED", "first field", "DISABLED", "PUBLISHED"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gdrive_label.test", "life_cycle.state", "PUBLISHED"),
					resource.TestCheckResourceAttr("gdrive_label_integer_field.first_field", "life_cycle.state", "DISABLED"),
					resource.TestCheckResourceAttr("gdrive_label_integer_field.second_field", "life_cycle.state", "PUBLISHED"),
				),
				ExpectNonEmptyPlan: true,
			},
			// 8 - Publish again
			{
				Config: testAccIntegerFieldResourceConfig("PUBLISHED", "first field", "DISABLED", "PUBLISHED"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gdrive_label.test", "life_cycle.state", "PUBLISHED"),
					resource.TestCheckResourceAttr("gdrive_label_integer_field.first_field", "life_cycle.state", "DISABLED"),
					resource.TestCheckResourceAttr("gdrive_label_integer_field.second_field", "life_cycle.state", "PUBLISHED"),
				),
			},
			// 9 - Disable second field
			{
				Config: testAccIntegerFieldResourceConfig("PUBLISHED", "first field", "DISABLED", "DISABLED"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gdrive_label.test", "life_cycle.state", "PUBLISHED"),
					resource.TestCheckResourceAttr("gdrive_label_integer_field.first_field", "life_cycle.state", "DISABLED"),
					resource.TestCheckResourceAttr("gdrive_label_integer_field.second_field", "life_cycle.state", "DISABLED"),
				),
				ExpectNonEmptyPlan: true,
			},
			// 10 - Publish again
			{
				Config: testAccIntegerFieldResourceConfig("PUBLISHED", "first field", "DISABLED", "DISABLED"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gdrive_label.test", "life_cycle.state", "PUBLISHED"),
					resource.TestCheckResourceAttr("gdrive_label_integer_field.first_field", "life_cycle.state", "DISABLED"),
					resource.TestCheckResourceAttr("gdrive_label_integer_field.second_field", "life_cycle.state", "DISABLED"),
				),
			},
			// 11 - Disable field
			{
				Config: testAccIntegerFieldResourceConfig("DISABLED", "first field", "DISABLED", "DISABLED"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gdrive_label.test", "life_cycle.state", "DISABLED"),
					resource.TestCheckResourceAttr("gdrive_label_integer_field.first_field", "life_cycle.state", "DISABLED"),
					resource.TestCheckResourceAttr("gdrive_label_integer_field.second_field", "life_cycle.state", "DISABLED"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccIntegerFieldResourceConfig(stateLabel, displayNameFirstField, stateFirstField, stateSecondField string) string {
	return fmt.Sprintf(`
resource "gdrive_label" "test" {
  label_type       = "ADMIN"
  use_admin_access = true
  properties {
    title       = "integer field test"
  }
  life_cycle {
    state = "%s"
  }
}

resource "gdrive_label_integer_field" "first_field" {
  label_id         = gdrive_label.test.label_id
  use_admin_access = true
  properties {
    display_name        = "%s"
    insert_before_field = gdrive_label_integer_field.second_field.field_id
  }
  life_cycle {
    state = "%s"
  }
}

resource "gdrive_label_integer_field" "second_field" {
  label_id         = gdrive_label.test.label_id
  use_admin_access = true
  properties {
    display_name = "second field"
  }
  life_cycle {
    state = "%s"
  }
}
`, stateLabel, displayNameFirstField, stateFirstField, stateSecondField)
}
