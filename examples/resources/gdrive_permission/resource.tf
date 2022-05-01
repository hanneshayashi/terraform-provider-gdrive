# Grant a user read access to a file
resource "gdrive_permission" "permissions_simple" {
  file_id       = "..."
  email_address = "user@example.com"
  role          = "reader"
  type          = "user"
}

# Make a user the owner of a file and move it to the new owner's root
resource "gdrive_permission" "permissions_owner_transfer" {
  file_id                 = "..."
  email_address           = "user@example.com"
  role                    = "owner"
  type                    = "user"
  send_notification_email = true
  transfer_ownership      = true
  move_to_new_owners_root = true
  email_message           = "Tag, you're it!"
}
