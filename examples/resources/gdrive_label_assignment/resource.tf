# Create and publish a Label
resource "gdrive_label" "test" {
  label_type       = "ADMIN"
  use_admin_access = true
  properties {
    title       = "My Label"
    description = "My Description"
  }
  life_cycle {
    state = "PUBLISHED"
  }
}

# Create a Date Field for the Label
resource "gdrive_label_date_field" "field" {
  label_id         = gdrive_label.test.label_id
  use_admin_access = true
  properties {
    display_name = "My Date Field"
  }
}

# Create a Selection Field for the Label
resource "gdrive_label_selection_field" "field" {
  label_id         = gdrive_label.test.label_id
  use_admin_access = true
  properties {
    display_name = "My Selection Field"
  }
}

# Create a Choice for the Selection Field
resource "gdrive_label_selection_choice" "choice" {
  label_id         = gdrive_label_selection_field.field.label_id
  field_id         = gdrive_label_selection_field.field.field_id
  use_admin_access = true
  properties {
    display_name = "My Choice"
  }
}

# Create an empty Sheet
resource "gdrive_file" "empty_speadsheet" {
  name      = "my_sheet"
  mime_type = "application/vnd.google-apps.spreadsheet"
  parent    = "root"
}

# BE SURE TO CREATE AND PUBLISH(!) THE LABEL FIRST!
# Assign the Label with the Field and Choice to a File
resource "gdrive_label_assignment" "label_assignment" {
  file_id  = gdrive_file.empty_speadsheet.file_id
  label_id = gdrive_label.test.label_id
  fields = [
    {
      field_id   = gdrive_label_selection_field.field.field_id
      value_type = "selection"
      values     = [gdrive_label_selection_choice.choice.choice_id]
    },
    {
      field_id   = gdrive_label_date_field.field.field_id
      value_type = "dateString"
      values     = ["2023-06-22"] # YYYY-MM-DD
    }
  ]
}
