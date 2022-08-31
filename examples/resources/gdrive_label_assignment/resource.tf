resource "gdrive_label_assignment" "label_assignment" {
  name     = "..."
  file_id  = "..."
  label_id = "..."
  field {
    field_id   = "..."
    value_type = "text"
    values     = ["one"]
  }
  field {
    field_id   = "..."
    value_type = "selection"
    values     = ["foo", "bar"]
  }
}
