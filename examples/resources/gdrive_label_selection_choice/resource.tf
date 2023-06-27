# Create a Label
resource "gdrive_label" "test" {
  label_type       = "ADMIN"
  use_admin_access = true
  properties {
    title = "My Label"
  }
}

# Create a Selection Field for the Label
resource "gdrive_label_selection_field" "field" {
  label_id         = gdrive_label.test.label_id
  use_admin_access = true
  properties {
    display_name = "My Field"
  }
}

# Create a simple Choice for the Selection Field
resource "gdrive_label_selection_choice" "simple_choice" {
  label_id         = gdrive_label_selection_field.field.label_id
  field_id         = gdrive_label_selection_field.field.field_id
  use_admin_access = true
  properties {
    display_name = "My Choice"
  }
}

# Create a second Choice with a Badge Color before the other Choice
resource "gdrive_label_selection_choice" "second_choice" {
  label_id         = gdrive_label_selection_field.field.label_id
  field_id         = gdrive_label_selection_field.field.field_id
  use_admin_access = true
  properties {
    display_name         = "My Colorful Choice"
    insert_before_choice = gdrive_label_selection_choice.simple_choice.choice_id
    badge_config {
      color {
        blue  = 0.21960784
        green = 0.5019608
        red   = 0.09411765
      }
    }
  }
}
