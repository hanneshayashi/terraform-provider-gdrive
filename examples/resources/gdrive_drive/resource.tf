# Create a Shared Drive
resource "gdrive_drive" "drive_simple" {
  name                    = "terraform-1"
  use_domain_admin_access = true
}

# Create a Shared Drive with restrictions
resource "gdrive_drive" "drive_restrictions" {
  name                    = "terraform-1"
  use_domain_admin_access = true
  restrictions {
    admin_managed_restrictions      = true
    drive_members_only              = false
    copy_requires_writer_permission = true
    domain_users_only               = true
  }
}
