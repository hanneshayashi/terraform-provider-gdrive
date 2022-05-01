# Get all permissions set on a file
data "gdrive_permissions" "some_file_permissions" {
  file_id = "..."
}

# Set the exact same permissions on a different file
resource "gdrive_permissions_policy" "permissions_policy" {
  file_id = "..."
  dynamic "permissions" {
    for_each = data.gdrive_permissions.some_file_permissions.permissions
    content {
      role          = permissions.value.role
      email_address = permissions.value.email_address
      type          = permissions.value.type
      domain        = permissions.value.domain
    }
  }
}
