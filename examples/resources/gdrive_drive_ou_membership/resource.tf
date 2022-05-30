# Create a Shared Drive
resource "gdrive_drive" "drive" {
  name                    = "terraform-1"
  use_domain_admin_access = true
}

# Move Shared Drive to OU
resource "gdrive_drive_ou_membership" "membership" {
  drive_id = gdrive_drive.drive.id
  parent   = "my-org-unit-id"
}
