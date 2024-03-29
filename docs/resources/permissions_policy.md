---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "gdrive_permissions_policy Resource - terraform-provider-gdrive"
subcategory: ""
description: |-
  Creates an authoratative permissions policy on a file or Shared Drive.
  Warning: This resource will set exactly the defined permissions and remove everything else!
  It is HIGHLY recommended that you import the resource and make sure that the owner is properly set before applying it!
  Important: On a destroy, this resource will preserve the owner and organizer permissions!
---

# gdrive_permissions_policy (Resource)

Creates an authoratative permissions policy on a file or Shared Drive.

**Warning: This resource will set exactly the defined permissions and remove everything else!**

It is HIGHLY recommended that you import the resource and make sure that the owner is properly set before applying it!

**Important**: On a *destroy*, this resource will preserve the owner and organizer permissions!

## Example Usage

```terraform
# Set a permission policy on a file
resource "gdrive_permissions_policy" "permissions_policy_file" {
  file_id = "..."
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
}

# Set a permission policy on a Shared Drive
resource "gdrive_permissions_policy" "permissions_policy_shared_drive" {
  file_id                 = "..."
  use_domain_admin_access = true
  permissions = [
    {
      email_address           = "user1@example.com"
      role                    = "organizer"
      type                    = "user"
      send_notification_email = true
      email_message           = "Hi User 1"
    },
    {
      email_address           = "user2@example.com"
      role                    = "fileOrganizer"
      type                    = "user"
      send_notification_email = true
      email_message           = "Hi User 2"
    }
  ]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `file_id` (String) ID of the file or Shared Drive.
- `permissions` (Attributes Set) Defines the set of permissions to set on the file or Shared Drive. (see [below for nested schema](#nestedatt--permissions))

### Optional

- `use_domain_admin_access` (Boolean) Use domain admin access.

### Read-Only

- `id` (String) The unique ID of this resource.

<a id="nestedatt--permissions"></a>
### Nested Schema for `permissions`

Required:

- `role` (String) The role.

Optional:

- `domain` (String) The domain that should be granted access.
- `email_address` (String) The email address of the trustee.
- `email_message` (String) An optional email message that will be sent when the permission is created.
- `move_to_new_owners_root` (Boolean) This parameter only takes effect if the item is not in a shared drive and the request is attempting to transfer the ownership of the item.
- `send_notification_email` (Boolean) Wether to send a notfication email.
- `transfer_ownership` (Boolean) Whether to transfer ownership to the specified user.
- `type` (String) The type of the trustee. Can be 'user', 'domain', 'group' or 'anyone'.

Read-Only:

- `permission_id` (String) PermissionID of the trustee.

## Import

Import is supported using the following syntax:

```shell
terraform import gdrive_permissions_policy.policy [use_domain_admin_access],[fileId]
```
