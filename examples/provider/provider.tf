# Use a Service Account Key from a JSON file
provider "gdrive" {
  service_account_key = "/path/to/sa.json"
  subject             = "admin@example.com"
}

# Use Application Default Credentials
provider "gdrive" {
  subject = "admin@example.com"
}

# Use Application Default Credentials to impersonate a different Service Account
provider "gdrive" {
  service_account = "email@my-project.iam.gserviceaccount.com"
  subject         = "admin@example.com"
}
