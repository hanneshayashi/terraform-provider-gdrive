---
page_title: "Terraform Google Drive Provider 1.0.0 Upgrade Guide"
description: |-
  Terraform Google Drive Provider 1.0.0 Upgrade Guide
---

# Terraform Google Drive Provider 1.0.0 Upgrade Guide

**Please make sure you have a backup of your state file before upgrading!**

Version `1.0.0` of the Terraform provider for Google Drive is major rewrite that transitions
the provider from the "old" Terraform Plugin SDK to the new Terraform Plugin Framework.

The Plugin Framework brings new capabilities that the provider takes advantage of to eliminate
some workarounds that were necessary with the SDK. This results in some slight differences
in the schema and configuration of some existing resources and data sources.

# Changes to Provider Configuration

Previous versions of the provider used boolean attributes to activate specific API scopes:

```terraform
provider "gdrive" {
  # ...
  use_labels_api         = true
  use_labels_admin_scope = true
  use_cloud_identity_api = true
}
```

This has been replaced by a single list attribute that can be used to specify the exact
scopes you want to use:

```terraform
provider "gdrive" {
  # ...
  scopes = [
    # ...
  ]
}
```

If the attribute is not set, all of the following scopes are used:

* `https://www.googleapis.com/auth/drive`
* `https://www.googleapis.com/auth/drive.labels`
* `https://www.googleapis.com/auth/drive.admin.labels`
* `https://www.googleapis.com/auth/cloud-identity.orgunits`

# Resources with Changed State Representation

## gdrive_drive

The `gdrive_drive` resource used to manage Shared Drives has a different state representation
for the `restrictions` attributed. In previous versions of the provider, the restrictions
were represented as a list of objects with a maximum number of 1:

```json
"restrictions": [
    {
        "admin_managed_restrictions": false,
        "copy_requires_writer_permission": false,
        "domain_users_only": false,
        "drive_members_only": false
    }
],
```

The new provider saves the restrictions as a single block:

```json
"restrictions": {
    "admin_managed_restrictions": false,
    "copy_requires_writer_permission": false,
    "domain_users_only": false,
    "drive_members_only": false
}
```

The provider will automatically upgrade the state representation if you run a `terraform apply`.

**Once the new state format has been saved, the state can no longer be managed by the old provider!**

# Resources with a changed Configuration Schema

## label -> labels

The `label` property of the `gdrive_label_policy` resource used to be configured by specifying multiple blocks:

```terraform
resource "gdrive_label_policy" "policy" {
  label {
    # ...
  }
  label {
    # ...
  }
  # ...
}
```

This was quite verbose and had the additional problem of being confusing when trying to access a specific label:

```terraform
gdrive_label_policy.policy.label[0]
```

The property has therefore been renamed to `fields` (plural) and changed to a nested attribute, which is more concise:
```terraform
resource "gdrive_label_policy" "policy" {
  labels = [
    {
        # ...
    },
    {
        # ...
    }
  ]
  # ...
}
```

A state upgrade is not necessary for this change, but your configuration files must be updated. Any references must also
be changed from

```terraform
gdrive_label_policy.policy.label[0]
```

to


```terraform
gdrive_label_policy.policy.labels[0]
```

## field -> fields

Similarly to the `label` property of the `gdrive_label_policy` resource, the `field` property of both the `gdrive_label_policy`
and the `gdrive_label_assignment` resources has also been renamed to `fields` and changed to a nested attribute.

Old configuration:

```terraform
field {
  # ...
}
field {
  # ...
}
```

New configuration:

```terraform
fields = [
  {
    # ...
  },
  {
    # ...
  }
]
```

A state upgrade is not necessary for this change, but your configuration files must be updated. Any references must also
be changed from

```terraform
...field[0]
```

to


```terraform
...fields[0]
```

## Permissions

The `permissions` property of the `gdrive_permissions_policy` resources has also been changed from multiple blocks to
a single nested attribute set:

Old configuration:

```terraform
resource "gdrive_permissions_policy" "policy" {
  permissions {
    email_address = "user1@example.com"
    role          = "owner"
    type          = "user"
  }
  permissions {
    email_address = "user2@example.com"
    role          = "writer"
    type          = "user"
  }
  # ...
}
```

New configuration:

```terraform
resource "gdrive_permissions_policy" "policy" {
  permissions = [
    {
      email_address = "user1@example.com"
      role          = "owner"
      type          = "user"
    },
    {
      email_address = "user2@example.com"
      role          = "writer"
      type          = "user"
    }
  ]
  # ...
}
```

The state representation has not changed, but the Terraform configuration must be updated as shown above.

## wait_after_create

Starting with `1.0.0`, the provider will now retry on HTTP error codes `404` and `502` by default, making the workaround
that was introduced in `0.9.1` to deal with the Drive API not being strongly consistent (i.e., the API will return
the Drive object immediately, even though other API endpoints will not find it yet and return `404`) unnecessary.

The attribute `wait_after_create` needs to be *removed* from the Terraform configuration of `gdrive_drive` resources.

Old configuration:

```terraform
resource "gdrive_drive" "drive" {
  wait_after_create = 30
  # ...
}
```

New configuration:

```terraform
resource "gdrive_drive" "drive" {
  # ...
}
```

# Changes to Data Sources

## Labels

The `gdrive_labels` data source now implements the `properties` attribute to align with the related resources
and data sources. This removes the `description` and `title` attributes from the root and moves them to the
new `properties` attribute. If you previously accessed the title attribute like this:

```terraform
data.gdrive_labels.labels.title
```

You will now have to access it like this:

```terraform
data.gdrive_labels.labels.properties.title
```

The data source will now also return the `life_cycle` and IDs for the labels, fields and choices.

## List to Block

The following properties where previously implemented as lists with a single block:

* `gdrive_drive`
  * `restrictions`
* `gdrive_label` and `gdrive_labels`
  * `life_cycle`
  * `list_options`
  * `date_options`
  * `max_value`
  * `min_value`
  * `selection_options`
  * `integer_options`
  * `text_options`
  * `user_options`
  * `properties`

They are now implemented as a nested block and you must no longer specify an index when accessing these properties. Example:

Before:

```terraform
data.gdrive_drive.drive.restrictions[0].admin_managed_restrictions
```

After


```terraform
data.gdrive_drive.drive.restrictions.admin_managed_restrictions
```

# ID Fields

Every resources in Terraform needs to implement the `id` attribute. For most resources, this is simply the ID as
represented by the respective API (e.g., the `file_id` for files or the `drive_id` for Shared Drives).
For some other resources, this may not be the case and the `id` attribute actually contains a combination of IDs
(e.g., the `id` attribute of the `gdrive_permission` resource is a combination of the `file_id` and the `permission_id`).
This is because both are required to uniquely represent the resource and read it from the API.

To give users more direct access to the IDs that they need for referencing between resources, every resource and nested
block now implements separate attributes for every ID (e.g., the `gdrive_permission` resource now also implements
an attribute for the `permission_id`). Users are encouraged to use these IDs instead of the more technical
`id` attribute that the provider implements. Refer to the help pages of the respective resource or data sources for details.

# Imports

The Terraform Plugin Framework can no longer reference attributes in the Terraform configuration during an import.
This means that some attributes that might be necessary to read the resource after the import, need to be specified
with the import command.

The following resources now require specifying the `use_domain_admin_access` attribute during the import:
* `gdrive_drive`
* `gdrive_permission`
* `gdrive_permissions_policy`

The following *new* resources introduced with `1.0.0` require specifying the `use_admin_access` attribute:
* `gdrive_label`
* `gdrive_label_date_field`
* `gdrive_label_integer_field`
* `gdrive_label_selection_field`
* `gdrive_label_selection_field`
* `gdrive_label_text_field`
* `gdrive_label_user_field`

Both attributes are bools and must be set to either `true` or `false`:

```shell
terraform import [resource_address] [true|false],[resource_id]
```

For specific instructions see the import section for the respective resources.
