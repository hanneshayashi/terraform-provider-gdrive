# Set a permission policy on a file
resource "gdrive_permissions_policy" "permissions_policy_file" {
  file_id = "..."
  permissions {
    email_address = "user1@exaple.com"
    role          = "owner"
    type          = "user"
  }
  permissions {
    email_address = "user2@example.com"
    role          = "writer"
    type          = "user"
  }
}

# Set a permission policy on a Shared Drive
resource "gdrive_permissions_policy" "permissions_policy_shared_drive" {
  file_id                 = "..."
  use_domain_admin_access = true
  permissions {
    email_address           = "user1@exaple.com"
    role                    = "organizer"
    type                    = "user"
    send_notification_email = true
    email_message           = "Hi User 1"
  }
  permissions {
    email_address           = "user2@example.com"
    role                    = "fileOrganizer"
    type                    = "user"
    send_notification_email = true
    email_message           = "Hi User 2"
  }
}
