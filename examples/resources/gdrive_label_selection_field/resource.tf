# Create a Label
resource "gdrive_label" "test" {
  label_type       = "ADMIN"
  use_admin_access = true
  properties {
    title = "My Label"
  }
}

# Create a simple Selection Field
resource "gdrive_label_selection_field" "field" {
  label_id         = gdrive_label.test.label_id
  use_admin_access = true
  properties {
    display_name = "My Field Name"
  }
}

# Create a second field before the first one that allows users to select multiple Choices
resource "gdrive_label_selection_field" "field_multi" {
  label_id         = gdrive_label.test.label_id
  use_admin_access = true
  properties {
    display_name        = "My Field Name"
    insert_before_field = gdrive_label_selection_field.field.field_id
  }
  selection_options {
    list_options {
      max_entries = 2
    }
  }
}
