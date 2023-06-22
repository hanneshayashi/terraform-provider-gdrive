# Create a Label
resource "gdrive_label" "test" {
  label_type       = "ADMIN"
  use_admin_access = true
  properties {
    title = "My Label"
  }
}

# Allow a specific User to apply the Label
resource "gdrive_label_permission" "permission" {
  parent           = gdrive_label.test.label_id
  use_admin_access = true
  email            = "user1@example.com"
  role             = "APPLIER"
}

# Allow everyone in the Org to apply the Label
resource "gdrive_label_permission" "permission_audience" {
  parent           = gdrive_label.test.label_id
  use_admin_access = true
  audience         = "audiences/default"
  role             = "APPLIER"
}
