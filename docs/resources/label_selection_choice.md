---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "gdrive_label_selection_choice Resource - terraform-provider-gdrive"
subcategory: ""
description: |-
  Creates a Choice for a Selection Field.
  Changes made to a Choice must be published via the Choice's Selection Field's Label before they are available for files.
  Publishing can only be done via the Label resource, NOT the Choice or Field resources.
  This means that, if you have Labels and Choices / Fields in the same Terraform configuration and you make changes
  to the Choices / Fields you may have to apply twice in order to
  1. Apply the changes to the Choices / Fields.
  2. Publish the changes via the Label.
  A Choice must be deactivated before it can be deleted.
---

# gdrive_label_selection_choice (Resource)

Creates a Choice for a Selection Field.

Changes made to a Choice must be published via the Choice's Selection Field's Label before they are available for files.

Publishing can only be done via the Label resource, NOT the Choice or Field resources.

This means that, if you have Labels and Choices / Fields in the same Terraform configuration and you make changes
to the Choices / Fields you may have to apply twice in order to
1. Apply the changes to the Choices / Fields.
2. Publish the changes via the Label.

A Choice must be deactivated before it can be deleted.

## Example Usage

```terraform
# Create a Label
resource "gdrive_label" "test" {
  label_type       = "ADMIN"
  use_admin_access = true
  properties {
    title = "My Label"
  }
}

# Create a Selection Field for the Label
resource "gdrive_label_selection_field" "field" {
  label_id         = gdrive_label.test.label_id
  use_admin_access = true
  properties {
    display_name = "My Field"
  }
}

# Create a simple Choice for the Selection Field
resource "gdrive_label_selection_choice" "simple_choice" {
  label_id         = gdrive_label_selection_field.field.label_id
  field_id         = gdrive_label_selection_field.field.field_id
  use_admin_access = true
  properties {
    display_name = "My Choice"
  }
}

# Create a second Choice with a Badge Color before the other Choice
resource "gdrive_label_selection_choice" "second_choice" {
  label_id         = gdrive_label_selection_field.field.label_id
  field_id         = gdrive_label_selection_field.field.field_id
  use_admin_access = true
  properties {
    display_name         = "My Colorful Choice"
    insert_before_choice = gdrive_label_selection_choice.simple_choice.choice_id
    badge_config {
      color {
        blue  = 0.21960784
        green = 0.5019608
        red   = 0.09411765
      }
    }
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `field_id` (String) The ID of the field.
- `label_id` (String) The ID of the label.

### Optional

- `language_code` (String) The BCP-47 language code to use for evaluating localized field labels.

When not specified, values in the default configured language are used.
- `life_cycle` (Block, Optional) The lifecycle state of an object, such as label, field, or choice.

The lifecycle enforces the following transitions:
* UNPUBLISHED_DRAFT (starting state)
* UNPUBLISHED_DRAFT -> PUBLISHED
* UNPUBLISHED_DRAFT -> (Deleted)
* PUBLISHED -> DISABLED
* DISABLED -> PUBLISHED
* DISABLED -> (Deleted) (see [below for nested schema](#nestedblock--life_cycle))
- `properties` (Block, Optional) Basic properties of the choice. (see [below for nested schema](#nestedblock--properties))
- `use_admin_access` (Boolean) Set to true in order to use the user's admin credentials.

The server verifies that the user is an admin for the label before allowing access.

### Read-Only

- `choice_id` (String) The unique value of the choice. This ID is autogenerated. Matches the regex: ([a-zA-Z0-9_])+.
- `id` (String) The unique ID of this resource.

<a id="nestedblock--life_cycle"></a>
### Nested Schema for `life_cycle`

Optional:

- `disabled_policy` (Block, Optional) The policy that governs how to show a disabled label, field, or selection choice. (see [below for nested schema](#nestedblock--life_cycle--disabled_policy))
- `state` (String) The state of the object associated with this lifecycle.

<a id="nestedblock--life_cycle--disabled_policy"></a>
### Nested Schema for `life_cycle.disabled_policy`

Optional:

- `hide_in_search` (Boolean) Whether to hide this disabled object in the search menu for Drive items.

When false, the object is generally shown in the UI as disabled but it appears in the search results when searching for Drive items.
When true, the object is generally hidden in the UI when searching for Drive items.
- `show_in_apply` (Boolean) Whether to show this disabled object in the apply menu on Drive items.

When true, the object is generally shown in the UI as disabled and is unselectable.
When false, the object is generally hidden in the UI.



<a id="nestedblock--properties"></a>
### Nested Schema for `properties`

Required:

- `display_name` (String) The display text to show in the UI identifying this choice.

Optional:

- `badge_config` (Block, Optional) (see [below for nested schema](#nestedblock--properties--badge_config))
- `insert_before_choice` (String) Insert or move this choice before the indicated choice. If empty, the choice is placed at the end of the list.

<a id="nestedblock--properties--badge_config"></a>
### Nested Schema for `properties.badge_config`

Optional:

- `color` (Block, Optional) The color of the badge.
When not specified, no badge is rendered.
The background, foreground, and solo (light and dark mode) colors set here are changed in the Drive UI into the closest recommended supported color.

*After setting this property, the plan will likely show a difference, because the API automatically modifies the values.
It is recommended to change the Terraform configuration to match the values set by the API. (see [below for nested schema](#nestedblock--properties--badge_config--color))
- `priority_override` (Number) Override the default global priority of this badge.
When set to 0, the default priority heuristic is used.

<a id="nestedblock--properties--badge_config--color"></a>
### Nested Schema for `properties.badge_config.color`

Optional:

- `alpha` (Number) The alpha value for the badge color as a float (number between 1 and 0 - e.g. "0.5")
- `blue` (Number) The blue value for the badge color as a float (number between 1 and 0 - e.g. "0.5")
- `green` (Number) The green value for the badge color as a float (number between 1 and 0 - e.g. "0.5")
- `red` (Number) The red value for the badge color as a float (number between 1 and 0 - e.g. "0.5")

## Import

Import is supported using the following syntax:

```shell
# The ID for this resource is a combined ID that consistent of the label_id and the field_id.
# If you file's fileId is "abcdef" and your field_id is "12345", the ID of the resource would be:
# "abcdef/12345"
# In addition, the use_admin_access attribute must be specified during the import.
# Example: true,abcdef/12345
terraform import gdrive_label_selection_field.field [use_admin_access],[label_id]/[field_id]
```
