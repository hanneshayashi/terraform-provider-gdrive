---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "gdrive_permission Data Source - terraform-provider-gdrive"
subcategory: ""
description: |-
  Returns the metadata of a permission on a file or Shared Drive.
---

# gdrive_permission (Data Source)

Returns the metadata of a permission on a file or Shared Drive.

## Example Usage

```terraform
data "gdrive_permission" "permission" {
  file_id       = "..."
  permission_id = "..."
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `file_id` (String) ID of the file or Shared Drive.
- `permission_id` (String) ID of the permission.

### Optional

- `use_domain_admin_access` (Boolean) Use domain admin access.

### Read-Only

- `domain` (String) The domain if the type of this permissions is 'domain'.
- `email_address` (String) The email address if the type of this permissions is 'user' or 'group'.
- `id` (String) The unique ID of this resource.
- `role` (String) The role that this trustee is granted.
- `type` (String) The type of the trustee. Can be 'user', 'domain', 'group' or 'anyone'.
