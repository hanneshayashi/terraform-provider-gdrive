# Search for all files having "test" in the file name in the user's Drive
data "gdrive_files" "files" {
  query = "name contains 'test'"
}

# Search for all files having "test" in the file name in all Drives including the user's
data "gdrive_files" "files" {
  query                         = "name contains 'test'"
  include_items_from_all_drives = true
  corpora                       = "allDrives"
}
