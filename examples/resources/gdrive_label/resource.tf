# Create a Label
resource "gdrive_label" "test" {
  label_type       = "ADMIN"
  use_admin_access = true
  properties {
    title       = "..."
    description = "..."
  }
}

# Publish the Label
resource "gdrive_label" "test" {
  label_type       = "ADMIN"
  use_admin_access = true
  properties {
    title       = "..."
    description = "..."
  }
  life_cycle {
    state = "PUBLISHED"
  }
}
