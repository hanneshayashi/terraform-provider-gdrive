---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "gdrive_label Data Source - terraform-provider-gdrive"
subcategory: ""
description: |-
  This resource can be used to get the fields and other metadata for a single label
---

# gdrive_label (Data Source)

This resource can be used to get the fields and other metadata for a single label

## Example Usage

```terraform
data "gdrive_label" "label" {
  name = "..."
}

# Retrieve a specific revision of a label using admin access.
# Requires setting the 'use_labels_admin_scope' property to 'true' in the provider config.
data "gdrive_label" "label_revision" {
  name             = "..."
  revision         = "1"
  use_admin_access = true
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Label resource name.
May be any of:
- labels/{id} (equivalent to labels/{id}@latest)
- labels/{id}@latest
- labels/{id}@published
- labels/{id}@{revisionId}

### Optional

- `language_code` (String) The BCP-47 language code to use for evaluating localized field labels.
When not specified, values in the default configured language are used.
- `revision` (String) The revision of the label to retrieve.
Defaults to the latest revision.
Reading other revisions may require addtional permissions and / or setting the 'use_admin_access' flag.
- `use_admin_access` (Boolean) Set to true in order to use the user's admin credentials.
The server verifies that the user is an admin for the label before allowing access.
Requires setting the 'use_labels_admin_scope' property to 'true' in the provider config.

### Read-Only

- `description` (String) The description of the label.
- `fields` (List of Object) (see [below for nested schema](#nestedatt--fields))
- `id` (String) The ID of this resource.
- `label_type` (String) The type of this label.
- `title` (String) Title of the label.

<a id="nestedatt--fields"></a>
### Nested Schema for `fields`

Read-Only:

- `date_options` (List of Object) (see [below for nested schema](#nestedobjatt--fields--date_options))
- `id` (String)
- `integer_options` (List of Object) (see [below for nested schema](#nestedobjatt--fields--integer_options))
- `life_cycle` (List of Object) (see [below for nested schema](#nestedobjatt--fields--life_cycle))
- `properties` (List of Object) (see [below for nested schema](#nestedobjatt--fields--properties))
- `query_key` (String)
- `selection_options` (List of Object) (see [below for nested schema](#nestedobjatt--fields--selection_options))
- `text_options` (List of Object) (see [below for nested schema](#nestedobjatt--fields--text_options))
- `user_options` (List of Object) (see [below for nested schema](#nestedobjatt--fields--user_options))
- `value_type` (String)

<a id="nestedobjatt--fields--date_options"></a>
### Nested Schema for `fields.date_options`

Read-Only:

- `date_format` (String)
- `date_format_type` (String)
- `max_value` (List of Object) (see [below for nested schema](#nestedobjatt--fields--date_options--max_value))
- `min_value` (List of Object) (see [below for nested schema](#nestedobjatt--fields--date_options--min_value))

<a id="nestedobjatt--fields--date_options--max_value"></a>
### Nested Schema for `fields.date_options.max_value`

Read-Only:

- `day` (Number)
- `month` (Number)
- `year` (Number)


<a id="nestedobjatt--fields--date_options--min_value"></a>
### Nested Schema for `fields.date_options.min_value`

Read-Only:

- `day` (Number)
- `month` (Number)
- `year` (Number)



<a id="nestedobjatt--fields--integer_options"></a>
### Nested Schema for `fields.integer_options`

Read-Only:

- `max_value` (Number)
- `min_value` (Number)


<a id="nestedobjatt--fields--life_cycle"></a>
### Nested Schema for `fields.life_cycle`

Read-Only:

- `state` (String)


<a id="nestedobjatt--fields--properties"></a>
### Nested Schema for `fields.properties`

Read-Only:

- `display_name` (String)
- `required` (Boolean)


<a id="nestedobjatt--fields--selection_options"></a>
### Nested Schema for `fields.selection_options`

Read-Only:

- `choices` (List of Object) (see [below for nested schema](#nestedobjatt--fields--selection_options--choices))
- `list_options` (List of Object) (see [below for nested schema](#nestedobjatt--fields--selection_options--list_options))

<a id="nestedobjatt--fields--selection_options--choices"></a>
### Nested Schema for `fields.selection_options.choices`

Read-Only:

- `display_name` (String)
- `id` (String)
- `life_cycle` (List of Object) (see [below for nested schema](#nestedobjatt--fields--selection_options--choices--life_cycle))

<a id="nestedobjatt--fields--selection_options--choices--life_cycle"></a>
### Nested Schema for `fields.selection_options.choices.life_cycle`

Read-Only:

- `state` (String)



<a id="nestedobjatt--fields--selection_options--list_options"></a>
### Nested Schema for `fields.selection_options.list_options`

Read-Only:

- `max_entries` (Number)



<a id="nestedobjatt--fields--text_options"></a>
### Nested Schema for `fields.text_options`

Read-Only:

- `max_length` (Number)
- `min_length` (Number)


<a id="nestedobjatt--fields--user_options"></a>
### Nested Schema for `fields.user_options`

Read-Only:

- `list_options` (List of Object) (see [below for nested schema](#nestedobjatt--fields--user_options--list_options))

<a id="nestedobjatt--fields--user_options--list_options"></a>
### Nested Schema for `fields.user_options.list_options`

Read-Only:

- `max_entries` (Number)