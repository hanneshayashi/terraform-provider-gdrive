# terraform-provider-gdrive
A Terraform Provider for Google Drive

## Features
* Manage Shared Drives
* Manage Google Drive files (including file uploads)
* Manage Drive permissions

## Installation
1. Download the latest [release](https://github.com/hanneshayashi/terraform-provider-gdrive/releases).
2. Copy the binary to\
`~/.terraform.d/plugins/github.com/hanneshayashi/gdrive/0.2.0/linux_amd64/terraform-provider-gdrive_v0.2.0`

**or the equivalent of your OS** (I haven't actually tested the provider with anything other than Linux)

## Setup
First, you need a GCP Service Account with Domain Wide Delegation set up with the Google Drive scope.\
This provider uses [GSM](https://github.com/hanneshayashi/gsm)'s auth and drive packages.
You can take a look at the GSM [Setup Guide](https://gsm.hayashi-ke.online/setup), if you need help.\
TL;DR:
1. Create GCP Project
2. Enable Drive API
3. Create Service Account + DWD
4. Enter the Client ID of the Service Account Key file with the Drive scope (https://www.googleapis.com/auth/drive) in your Admin Console

You can authenticate in one of two ways:
1. Create a Service Account Key and configure the provider like so:
```terraform
provider "gdrive" {
  service_account_key = "/path/to/sa.json"  # This is the Key file for your Service Account
  subject             = "admin@example.com" # This is the user you want to impersonate with Domain Wide Delegation
}
```
2. Use Application Default Credentials:
Activate the [IAM Service Account Credentials API](https://console.developers.google.com/apis/api/iamcredentials.googleapis.com/overview) *in the project where the Service Account is located*

   a) Use `gcloud auth application-default login` on your local workstation

   **or**

   b) Use a Google Compute Engine instance

In **both** cases, the account needs the *Service Account Token Creator* role for the Service Account you set up for DWD (**even if your GCE instance is using the same account**).

You can then configure the provider like so:

```terraform
provider "gdrive" {
  service_account     = "email@my-project.iam.gserviceaccount.com"  # This is the email address of your Service Account. You can leave this empty on GCE, if you want to use the instance's account
  subject             = "admin@example.com"  # This is the user you want to impersonate with Domain Wide Delegation
}
```

## To Do
+ Port less used API features from GSM