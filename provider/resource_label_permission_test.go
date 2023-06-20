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
