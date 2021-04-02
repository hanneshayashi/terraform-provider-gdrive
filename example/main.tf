terraform {
  required_providers {
    gdrive = {
      source  = "github.com/hanneshayashi/gdrive"
      version = "0.2.0"
    }
  }
}

provider "gdrive" {
  service_account_key = "/path/to/sa.json"  # This is the Key file for your Service Account
  subject             = "admin@example.com" # This is the user you want to impersonate with Domain Wide Delegation
}

resource "gdrive_drive" "test_drive" {
  name = "terraform-1"
}

resource "gdrive_permission" "test_permission" {
  file_id                 = gdrive_drive.test_drive.id
  email_address           = "user@example.com"
  role                    = "reader"
  type                    = "user"
  send_notification_email = true
  email_message           = "Test"
}

resource "gdrive_file" "folder-1" {
  mime_type = "application/vnd.google-apps.folder"
  drive_id  = gdrive_drive.test_drive.id
  name      = "terraform-test"
  parent    = gdrive_file.new-parent.id
}

resource "gdrive_file" "new-parent" {
  mime_type = "application/vnd.google-apps.folder"
  drive_id  = gdrive_drive.test_drive.id
  name      = "root"
  parent    = gdrive_drive.test_drive.id
}

resource "gdrive_permission" "test_permission-2" {
  file_id       = gdrive_file.folder-1.id
  email_address = "user2@example.com"
  role          = "reader"
  type          = "user"
}

resource "gdrive_file" "content-1" {
  mime_type = "text/plain"
  drive_id  = gdrive_drive.test_drive.id
  name      = "somefile"
  parent    = gdrive_drive.test_drive.id
  content   = "/path/to/somefile"
}
