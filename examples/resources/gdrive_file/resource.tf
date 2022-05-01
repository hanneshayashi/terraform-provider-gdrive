# Create a folder inside the impersonated user's personal Drive
resource "gdrive_file" "folder" {
  mime_type = "application/vnd.google-apps.folder"
  parent    = "root"
  name      = "folder"
}

# Upload a file to the folder
resource "gdrive_file" "file_with_content" {
  mime_type = "text/plain"
  name      = "somefile"
  parent    = gdrive_file.folder.id
  content   = "/path/to/somefile"
}
