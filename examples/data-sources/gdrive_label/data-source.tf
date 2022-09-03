data "gdrive_label" "label" {
  name = "..."
}

# Retrieve a specific revision of a label using admin access.
# Requires setting the 'use_labels_admin_scope' property to 'true' in the provider config.
data "gdrive_label" "label_revision" {
  name             = "..."
  revision         = "1"
  use_admin_access = true
}

# Find all files with a specific field value
data "gdrive_files" "files_with_label" {
  query = "${data.gdrive_label.label.fields[0].query_key} = 'my value'"
}
