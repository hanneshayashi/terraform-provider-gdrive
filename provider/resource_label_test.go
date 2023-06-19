package provider

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccLabel(t *testing.T) {
	title := "foo"
	description := "bar"
	changedDescription := "baz"
	renamed := "renamed"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// 1 - Create and Read testing
			{
				Config: testAccLabelResourceConfig(title, description, "", "PUBLISHED", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gdrive_label.created_without_life_cycle", "properties.title", title),
					resource.TestCheckResourceAttr("gdrive_label.created_without_life_cycle", "properties.description", description),
					resource.TestCheckResourceAttr("gdrive_label.created_as_published", "life_cycle.state", "PUBLISHED"),
					resource.TestCheckResourceAttr("gdrive_label.created_as_disabled", "life_cycle.state", "DISABLED"),
				),
			},
			// 2 - ImportState no life_cycle
			{
				ResourceName:        "gdrive_label.created_without_life_cycle",
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: "true,",
			},
			// 3 - ImportState published
			{
				ResourceName:            "gdrive_label.created_as_published",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdPrefix:     "true,",
				ImportStateVerifyIgnore: []string{"life_cycle"},
			},
			// 4 - Rename and change description
			{
				Config: testAccLabelResourceConfig(renamed, changedDescription, "UNPUBLISHED_DRAFT", "PUBLISHED", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gdrive_label.created_without_life_cycle", "properties.title", renamed),
					resource.TestCheckResourceAttr("gdrive_label.created_without_life_cycle", "properties.description", changedDescription),
					resource.TestCheckResourceAttr("gdrive_label.created_without_life_cycle", "life_cycle.state", "UNPUBLISHED_DRAFT"),
					resource.TestCheckResourceAttr("gdrive_label.created_as_published", "properties.title", renamed),
					resource.TestCheckResourceAttr("gdrive_label.created_as_disabled", "properties.title", renamed),
					resource.TestCheckResourceAttr("gdrive_label.created_as_published", "life_cycle.state", "PUBLISHED"),
					resource.TestCheckResourceAttr("gdrive_label.created_as_disabled", "life_cycle.state", "DISABLED"),
				),
			},
			// 5 - Publish and change DisabledPolicy
			{
				Config: testAccLabelResourceConfig(renamed, changedDescription, "PUBLISHED", "PUBLISHED", "true"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gdrive_label.created_without_life_cycle", "life_cycle.state", "PUBLISHED"),
					resource.TestCheckResourceAttr("gdrive_label.created_as_published", "life_cycle.state", "PUBLISHED"),
					resource.TestCheckResourceAttr("gdrive_label.created_as_disabled", "life_cycle.state", "DISABLED"),
					resource.TestCheckResourceAttr("gdrive_label.created_as_disabled", "life_cycle.disabled_policy.hide_in_search", "true"),
					resource.TestCheckResourceAttr("gdrive_label.created_as_disabled", "life_cycle.disabled_policy.show_in_apply", "true"),
				),
			},
			// 6 - Rename back and don't publish
			{
				Config: testAccLabelResourceConfig(title, description, "PUBLISHED_WITH_PENDING_CHANGES", "PUBLISHED_WITH_PENDING_CHANGES", "true"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gdrive_label.created_without_life_cycle", "life_cycle.state", "PUBLISHED_WITH_PENDING_CHANGES"),
					resource.TestCheckResourceAttr("gdrive_label.created_as_published", "life_cycle.state", "PUBLISHED_WITH_PENDING_CHANGES"),
					resource.TestCheckResourceAttr("gdrive_label.created_as_disabled", "life_cycle.state", "DISABLED"),
					resource.TestCheckResourceAttr("gdrive_label.created_as_disabled", "life_cycle.disabled_policy.hide_in_search", "true"),
					resource.TestCheckResourceAttr("gdrive_label.created_as_disabled", "life_cycle.disabled_policy.show_in_apply", "true"),
				),
			},
			// 7 - Disable and change DisablePolicy
			{
				Config: testAccLabelResourceConfig(title, description, "DISABLED", "DISABLED", "false"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gdrive_label.created_without_life_cycle", "life_cycle.state", "DISABLED"),
					resource.TestCheckResourceAttr("gdrive_label.created_as_published", "life_cycle.state", "DISABLED"),
					resource.TestCheckResourceAttr("gdrive_label.created_as_disabled", "life_cycle.state", "DISABLED"),
					resource.TestCheckResourceAttr("gdrive_label.created_as_disabled", "life_cycle.disabled_policy.hide_in_search", "false"),
					resource.TestCheckResourceAttr("gdrive_label.created_as_disabled", "life_cycle.disabled_policy.show_in_apply", "false"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccLabelResourceConfig(title, description, stateNone, statePublished, dPolicy string) string {
	var lifeCycle string
	var disabledPolicy string
	if stateNone != "" {
		lifeCycle = fmt.Sprintf(`
life_cycle {
  state = "%s"
}`, stateNone)
	}
	if dPolicy != "" {
		disabledPolicy = fmt.Sprintf(`
disabled_policy {
  hide_in_search = %s
  show_in_apply  = %s
}`, dPolicy, dPolicy)
	}
	return strings.Join([]string{
		fmt.Sprintf(`
		resource "gdrive_label" "created_without_life_cycle" {
		  label_type       = "ADMIN"
		  use_admin_access = true
		  properties {
		    title       = "%s"
		    description = "%s"
		  }
		  %s
		}`, title, description, lifeCycle),
		fmt.Sprintf(`
		resource "gdrive_label" "created_as_published" {
		  label_type       = "ADMIN"
		  use_admin_access = true
		  properties {
		    title       =  "%s"
		  }
		  life_cycle {
		    state = "%s"
		  }
		}`, title, statePublished),
		fmt.Sprintf(`
		resource "gdrive_label" "created_as_disabled" {
		  label_type       = "ADMIN"
		  use_admin_access = true
		  properties {
		    title       =  "%s"
		  }
		  life_cycle {
		    state = "DISABLED"
			%s
		  }
		}`, title, disabledPolicy),
	}, "\n")
}
