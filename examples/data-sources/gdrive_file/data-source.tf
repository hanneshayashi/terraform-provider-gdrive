# Access the metadata of a file or folder
data "gdrive_file" "file_metadata" {
  file_id = "..."
}

# Download a file
data "gdrive_file" "file_download" {
  file_id       = "..."
  download_path = "./my-file"
}

# Export a Google Docs file to MS Word and download it
data "gdrive_file" "file_export" {
  file_id          = "..."
  export_path      = "./test.docx"
  export_mime_type = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
}
