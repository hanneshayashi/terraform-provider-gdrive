terraform {
  required_providers {
    gdrive = {
      source = "github.com/hanneshayashi/gdrive"
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
