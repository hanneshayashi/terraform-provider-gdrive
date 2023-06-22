# Get all permissions set on a file
data "gdrive_permissions" "some_file_permissions" {
  file_id = "..."
}

# Set the exact same permissions on a different file
resource "gdrive_permissions_policy" "permissions_policy_target" {
  file_id = "..."
  permissions = [for p in data.gdrive_permissions.some_file_permissions.permissions : {
    role          = p.role
    type          = p.type
    email_address = p.email_address == "" ? null : p.email_address
    domain        = p.domain == "" ? null : p.domain
    }
  ]
}
