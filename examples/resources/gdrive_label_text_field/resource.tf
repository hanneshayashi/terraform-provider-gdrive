# Create a Label
resource "gdrive_label" "test" {
  label_type       = "ADMIN"
  use_admin_access = true
  properties {
    title = "My Label"
  }
}

# Create a simple Text Field
resource "gdrive_label_text_field" "field" {
  label_id         = gdrive_label.test.label_id
  use_admin_access = true
  properties {
    display_name = "My Field Name"
  }
}
