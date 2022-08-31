resource "gdrive_label_policy" "policy" {
  file_id = "..."
  label {
    label_id = "..."
    field {
      field_id   = "..."
      value_type = "..."
      values     = ["..."]
    }
  }
  label {
    label_id = "..."
    field {
      field_id   = "..."
      value_type = "..."
      values     = ["..."]
    }
    field {
      field_id   = "..."
      value_type = "..."
      values     = ["..."]
    }
  }
}
