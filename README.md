# terraform-provider-gdrive
A Terraform Provider for Google Drive

[![Go Report Card](https://goreportcard.com/badge/github.com/hanneshayashi/terraform-provider-gdrive)](https://goreportcard.com/report/github.com/hanneshayashi/terraform-provider-gdrive)

[View on Terraform Registry](https://registry.terraform.io/providers/hanneshayashi/gdrive/latest).

## Features
* Manage Shared Drives
* Manage Google Drive files (including file uploads)
* Manage Drive permissions

## Installation
To install this provider, copy and paste this code into your Terraform configuration. Then, run terraform init.

```terraform
terraform {
  required_providers {
    gdrive = {
      source = "hanneshayashi/gdrive"
      version = ">= 0.5.0"
    }
  }
}
```

## Setup
First, you need a GCP Service Account with Domain Wide Delegation set up with the Google Drive scope.

This provider uses [GSM](https://github.com/hanneshayashi/gsm)'s auth and drive packages.
You can take a look at the GSM [Setup Guide](https://gsm.hayashi-ke.online/setup), if you need help.

The basic steps are:
1. Create GCP Project
2. Enable Drive API
3. Create Service Account + DWD
4. Enter the Client ID of the Service Account Key file with the Drive scope (https://www.googleapis.com/auth/drive) in your Admin Console

You can authenticate in one of two ways:
1. Create a Service Account Key and configure the provider like so:
```terraform
provider "gdrive" {
  service_account_key = "/path/to/sa.json"  # This is the path to your Service Account Key file or its content in JSON format
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