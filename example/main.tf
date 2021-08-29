terraform {
  required_providers {
    gdrive = {
      source  = "hanneshayashi/gdrive"
      version = ">= 0.6.0"
    }
  }
}

provider "gdrive" {
  service_account_key = "/path/to/sa.json"  # This is the path to your Service Account Key file or its content in JSON format
  # service_account     = "email@my-project.iam.gserviceaccount.com"  # This is the email address of your Service Account. You can leave this empty on GCE, if you want to use the instance's account
  subject             = "admin@example.com" # This is the user you want to impersonate with Domain Wide Delegation
  # retry_on = [ 404 ] # Retry on specific HTTP error codes such as 404
}

resource "gdrive_drive" "example_drive" {
  name = "terraform-1"

  # restrictions {
  #   admin_managed_restrictions      = true
  #   drive_members_only              = true
  #   copy_requires_writer_permission = true
  #   domain_users_only               = true
  # }
}

resource "gdrive_file" "folder_1" {
  mime_type = "application/vnd.google-apps.folder"
  drive_id  = gdrive_drive.example_drive.id
  name      = "folder-1"
  parent    = gdrive_drive.example_drive.id
}

resource "gdrive_file" "folder_2" {
  mime_type = "application/vnd.google-apps.folder"
  drive_id  = gdrive_drive.example_drive.id
  name      = "folder-2"
  parent    = gdrive_file.folder_1.id
}

resource "gdrive_file" "content-1" {
  mime_type = "text/plain"
  drive_id  = gdrive_drive.example_drive.id
  name      = "somefile"
  parent    = gdrive_file.folder_2.id
  content   = "/path/to/somefile"
}

resource "gdrive_permission" "permissions_1" {
  file_id                 = gdrive_drive.example_drive.id
  email_address           = "user@example.com"
  role                    = "reader"
  type                    = "user"
  send_notification_email = true
  email_message           = "Example message"
}

resource "gdrive_permission" "permissions_2" {
  file_id       = gdrive_file.folder_1.id
  email_address = "user2@example.com"
  role          = "reader"
  type          = "user"
}
