resource "gdrive_label_policy" "policy" {
  file_id = "..."
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
  label {
    label_id = "..."
    field {
      field_id   = "..."
      value_type = "..."
      values     = ["..."]
    }
  }
}

# Assign a label policy to all files in a folder
data "gdrive_files" "files" {
  query = "'file_id of a folder' in parents"
}

resource "gdrive_label_policy" "policy" {
  for_each = toset([for file in data.gdrive_files.files.files : file.file_id])
  file_id  = each.value
  label {
    label_id = "..."
    field {
      field_id   = "..."
      value_type = "..."
      values     = ["..."]
    }
  }
}
