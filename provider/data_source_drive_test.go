package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDriveDS(t *testing.T) {
	name := "tftest"
	renamed := "testtest_renamed"
	restrictionsBefore := "false"
	restrictionsAfter := "true"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// 1 - Create and Read testing
			{
				Config: testAccDriveDataSourceConfig(name, restrictionsBefore),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gdrive_drive.drive_simple", "name", name),
					resource.TestCheckResourceAttr("data.gdrive_drive.drive_restrictions", "name", name),
				),
			},
			// 3 - Rename and change restrictions
			{
				Config: testAccDriveDataSourceConfig(renamed, restrictionsAfter),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gdrive_drive.drive_simple", "name", renamed),
					resource.TestCheckResourceAttr("data.gdrive_drive.drive_restrictions", "name", renamed),
					resource.TestCheckResourceAttr("data.gdrive_drive.drive_restrictions", "restrictions.admin_managed_restrictions", restrictionsAfter),
					resource.TestCheckResourceAttr("data.gdrive_drive.drive_restrictions", "restrictions.drive_members_only", restrictionsAfter),
					resource.TestCheckResourceAttr("data.gdrive_drive.drive_restrictions", "restrictions.copy_requires_writer_permission", restrictionsAfter),
					resource.TestCheckResourceAttr("data.gdrive_drive.drive_restrictions", "restrictions.domain_users_only", restrictionsAfter),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccDriveDataSourceConfig(name, restrictions string) string {
	return fmt.Sprintf(`
data "gdrive_drive" "drive_simple" {
  drive_id = gdrive_drive.drive_simple.drive_id
}

data "gdrive_drive" "drive_restrictions" {
  drive_id = gdrive_drive.drive_restrictions.drive_id
}

resource "gdrive_drive" "drive_simple" {
  name                    = "%s"
  use_domain_admin_access = true
}

resource "gdrive_drive" "drive_restrictions" {
  name                    = "%s"
  use_domain_admin_access = true
  restrictions {
    admin_managed_restrictions      = %s
    drive_members_only              = %s
    copy_requires_writer_permission = %s
    domain_users_only               = %s
  }
}`, name, name, restrictions, restrictions, restrictions, restrictions)
}
