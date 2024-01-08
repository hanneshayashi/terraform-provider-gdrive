# Create a folder inside the impersonated user's personal Drive
resource "gdrive_file" "folder" {
  mime_type = "application/vnd.google-apps.folder"
  parent    = "root" # This will cause a planned changed during the next refresh because the provider will read the actual FileID of the user's root folder
  name      = "folder"
}

# Upload a file to the folder
resource "gdrive_file" "file_with_content" {
  mime_type = "text/plain"
  name      = "somefile"
  parent    = gdrive_file.folder.id
  content   = "/path/to/somefile"
}

# Upload a CSV file and import it as a Google Sheet
resource "gdrive_file" "import_csv" {
  name             = "my_sheet"
  mime_type        = "application/vnd.google-apps.spreadsheet"
  mime_type_source = "text/csv"
  content          = "./test.csv"
  parent           = gdrive_file.folder.id
}
