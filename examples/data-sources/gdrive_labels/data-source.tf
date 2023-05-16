data "gdrive_labels" "labels" {
}

# Retrieve only published labels using admin access
# Requires setting the 'use_labels_admin_scope' property to 'true' in the provider config.
data "gdrive_labels" "label_published_only" {
  use_admin_access = true
  published_only   = true
}
