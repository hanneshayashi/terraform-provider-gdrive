# Assign a label policy to a file
resource "gdrive_label_assignment" "label_assignment" {
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

# Assign a label to all files in a folder
data "gdrive_files" "files" {
  query = "'file_id of a folder' in parents"
}

resource "gdrive_label_assignment" "label_assignment" {
  for_each = toset([for file in data.gdrive_files.files.files : file.file_id])
  label_id = "..."
  field {
    field_id   = "..."
    value_type = "..."
    values     = ["..."]
  }
}
