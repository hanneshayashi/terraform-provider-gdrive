# Create a Shared Drive
resource "gdrive_drive" "drive" {
  name                    = "terraform-1"
  use_domain_admin_access = true
}

#Move Shared Drive to OU
resource "gdrive_drive_ou_membership" "membership" {
  drive_id = gdrive_drive.drive.id
  parent   = "04ghsuhd1u5ll4p"
}

provider "gdrive" {
  service_account_key    = "/home/hannes_siefert_gmail_com/.config/gsm/demo.json"
  subject                = "hannes.hayashi@demo.wabion.cloud"
  use_cloud_identity_api = true
}

terraform {
  required_providers {
    gdrive = {
      source  = "github.com/hanneshayashi/gdrive"
      version = "0.8.0"
    }
  }
}
