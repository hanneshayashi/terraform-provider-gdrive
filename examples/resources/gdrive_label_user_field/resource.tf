# Create a Label
resource "gdrive_label" "test" {
  label_type       = "ADMIN"
  use_admin_access = true
  properties {
    title = "My Label"
  }
}

# Create a simple User Field
resource "gdrive_label_user_field" "field" {
  label_id         = gdrive_label.test.label_id
  use_admin_access = true
  properties {
    display_name = "My Field Name"
  }
}

# Create a second field before the first one that allows multiple selections
resource "gdrive_label_user_field" "field_multi" {
  label_id         = gdrive_label.test.label_id
  use_admin_access = true
  properties {
    display_name        = "My Field Name"
    insert_before_field = gdrive_label_user_field.field.field_id
  }
  user_options {
    list_options {
      max_entries = 2
    }
  }
}
