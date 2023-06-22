# Create a Label
resource "gdrive_label" "test" {
  label_type       = "ADMIN"
  use_admin_access = true
  properties {
    title = "My Label"
  }
}

# Create a simple Date Field
resource "gdrive_label_date_field" "field" {
  label_id         = gdrive_label.test.label_id
  use_admin_access = true
  properties {
    display_name = "My Field Name"
  }
}

# Create a second field before the first one that uses a short format for the date
resource "gdrive_label_date_field" "field_multi" {
  label_id         = gdrive_label.test.label_id
  use_admin_access = true
  properties {
    display_name        = "My Field Name"
    insert_before_field = gdrive_label_date_field.field.field_id
  }
  date_options {
    date_format_type = "SHORT_DATE"
  }
}
