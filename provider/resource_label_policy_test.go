package provider

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccLabelPolicy(t *testing.T) {
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
				Config: testAccLabelPolicyResourceConfig("UNPUBLISHED_DRAFT", "0", intBefore, textBefore),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gdrive_label.test", "life_cycle.state", "UNPUBLISHED_DRAFT"),
					resource.TestCheckResourceAttr("gdrive_label.test2", "life_cycle.state", "UNPUBLISHED_DRAFT"),
				),
			},
			// 2 - Publish Label
			{
				Config: testAccLabelPolicyResourceConfig("PUBLISHED", "0", intBefore, textBefore),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gdrive_label.test", "life_cycle.state", "PUBLISHED"),
					resource.TestCheckResourceAttr("gdrive_label.test2", "life_cycle.state", "PUBLISHED"),
				),
			},
			// 3 - Create Label Policy
			{
				Config: testAccLabelPolicyResourceConfig("PUBLISHED", "1", intBefore, textBefore),
				Check:  resource.ComposeAggregateTestCheckFunc(
				// TODO
				// resource.TestCheckResourceAttr("gdrive_label_policy.policy", "field[0].values[0]", dateBefore),
				// resource.TestCheckResourceAttr("gdrive_label_policy.policy", "field[1].values[0]", intBefore),
				// resource.TestCheckResourceAttr("gdrive_label_policy.policy", "field[3].values[0]", textBefore),
				),
			},
			// 4 - ImportState testing
			{
				ResourceName:      "gdrive_label_policy.policy",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// 5 - Change values
			{
				Config: testAccLabelPolicyResourceConfig("PUBLISHED", "1", intAfter, textAfter),
				Check:  resource.ComposeAggregateTestCheckFunc(
				// TODO
				// resource.TestCheckResourceAttr("gdrive_label_policy.policy", "field[0].values[0]", dateAfter),
				// resource.TestCheckResourceAttr("gdrive_label_policy.policy", "field[1].values[0]", intAfter),
				// resource.TestCheckResourceAttr("gdrive_label_policy.policy", "field[3].values[0]", textAfter),
				),
			},
			// 6 - Disable Label and Delete Files and Policy
			{
				Config: testAccLabelPolicyResourceConfig("DISABLED", "0", intAfter, textAfter),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gdrive_label.test", "life_cycle.state", "DISABLED"),
					resource.TestCheckResourceAttr("gdrive_label.test2", "life_cycle.state", "DISABLED"),
					resource.TestCheckResourceAttr("gdrive_label_integer_field.integer_field", "life_cycle.state", "DISABLED"),
					resource.TestCheckResourceAttr("gdrive_label_text_field.text_field", "life_cycle.state", "DISABLED"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccLabelPolicyResourceConfig(state, count, intValue, textValue string) string {
	var lifeCycle string
	if state == "DISABLED" {
		lifeCycle = `
life_cycle {
  state = "DISABLED"
}`
	}
	var policy string
	if count == "1" {
		policy = fmt.Sprintf(`
resource "gdrive_file" "folder" {
  mime_type = "application/vnd.google-apps.folder"
  parent    = gdrive_drive.drive.drive_id
  drive_id  = gdrive_drive.drive.drive_id
  name      = "folder"
}

resource "gdrive_file" "file" {
  name             = "file"
  mime_type        = "application/vnd.google-apps.spreadsheet"
  drive_id  	   = gdrive_drive.drive.drive_id
  parent           = gdrive_file.folder.file_id
}

resource "gdrive_label_policy" "policy" {
  file_id  = gdrive_file.file.file_id
  labels = [
    {
	  label_id = gdrive_label.test.label_id
      fields = [
        {
          field_id   = gdrive_label_text_field.text_field.field_id
          value_type = "text"
          values     = ["%s"]
        }
      ]
    },
    {
      label_id = gdrive_label.test2.label_id
      fields = [
        {
          field_id   =  gdrive_label_integer_field.integer_field.field_id
          value_type = "integer"
          values     = [%s]
        }
	  ]
    }
  ]
}`, textValue, intValue)
	}
	return strings.Join([]string{
		`resource "gdrive_drive" "drive" {
  name                    = "label_policy_test"
  use_domain_admin_access = true
}
`,
		fmt.Sprintf(`resource "gdrive_label" "test" {
  label_type       = "ADMIN"
  use_admin_access = true
  properties {
    title = "policy test"
  }
  life_cycle {
    state = "%s"
  }
}`, state),
		fmt.Sprintf(`resource "gdrive_label" "test2" {
  label_type       = "ADMIN"
  use_admin_access = true
  properties {
    title = "policy test"
  }
  life_cycle {
    state = "%s"
  }
}`, state),
		fmt.Sprintf(`
resource "gdrive_label_text_field" "text_field" {
  label_id         = gdrive_label.test.label_id
  use_admin_access = true
  properties {
    display_name = "text field"
  }
  %s
}`, lifeCycle),
		fmt.Sprintf(`
resource "gdrive_label_integer_field" "integer_field" {
  label_id         = gdrive_label.test2.label_id
  use_admin_access = true
  properties {
    display_name = "integer field"
  }
  %s
}`, lifeCycle),
		policy,
	}, "\n")
}
