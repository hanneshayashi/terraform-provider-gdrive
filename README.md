# terraform-provider-gdrive
A Terraform Provider for Google Drive

[![Go Report Card](https://goreportcard.com/badge/github.com/hanneshayashi/terraform-provider-gdrive)](https://goreportcard.com/report/github.com/hanneshayashi/terraform-provider-gdrive)

[View on Terraform Registry](https://registry.terraform.io/providers/hanneshayashi/gdrive/latest).

The Terraform provider for Google Drive can be used to manage Google Drive objects like files and folders, Shared Drives and Labels.

It can also be used to manage permissions to any of these objects, as well as import and export files to and from Google Drive.

Using Terraform and a source code management solution to manage your Google Drive environment can help you estabilsh secure processes
that require approval from multiple people before changes are deployed to production. You can also build your own modules to estabilsh
standards across your orgnization like naming conventions, default permissions or Label assignments.

## Features

* Manage Shared Drives and organize them into organizational units
* Manage Google Drive files (including file uploads, downloads and exports)
* Manage Google Drive permissions
* Manage Google Drive Labels, fields, assignments to files and permissions

## Installation

To install this provider, copy and paste this code into your Terraform configuration. Then, run `terraform init`.

```terraform
terraform {
  required_providers {
    gdrive = {
      source = "hanneshayashi/gdrive"
      version = "~> 1.0"
    }
  }
}
```

## Upgrade from 0.x

Please see the [Upgrade Guide](https://registry.terraform.io/providers/hanneshayashi/gdrive/latest/docs/guides/version_1_upgrade) and make sure you have a backup of your state file before upgrading.

## Setup

1. Create GCP Project (or use an existing one)
2. Enable the following APIs:
    * Drive API
    * Drive Labels API
    * Cloud Identity API
3. Create a Service Account + Enable Domain Wide Delegation
    * See [Perform Google Workspace Domain-Wide Delegation of Authority](https://developers.google.com/admin-sdk/directory/v1/guides/delegation)
    * **You *don't* need the Service Account Key if you want to use [Application Default Credential](https://cloud.google.com/iam/docs/best-practices-for-using-and-managing-service-accounts#use-attached-service-accounts)**
4. Enter the Client ID of the Service Account with the following scopes in your Admin Console:
    *	`https://www.googleapis.com/auth/drive`
    *	`https://www.googleapis.com/auth/drive.labels`
    *	`https://www.googleapis.com/auth/drive.admin.labels`
    * `https://www.googleapis.com/auth/cloud-identity.orgunits`

You can authenticate in one of two ways:
1. Use Application Default Credentials (**recommended**):
  Activate the [IAM Service Account Credentials API](https://console.developers.google.com/apis/api/iamcredentials.googleapis.com/overview) *in the project where the Service Account is located*

   a) Use a Google Compute Engine instance or [any service that supports attaching a Service Account in GCP](https://cloud.google.com/iam/docs/impersonating-service-accounts#attaching-new-resource)

   **or**

   b) Use `gcloud auth application-default login --impersonate-service-account` on your local workstation

In **both** cases, the account needs the *[Service Account Token Creator](https://cloud.google.com/iam/docs/service-accounts#token-creator-role)* role for the Service Account you set up for DWD (**even if your GCP service is using the same account**).

You can then configure the provider like so:

```terraform
provider "gdrive" {
  service_account     = "email@my-project.iam.gserviceaccount.com"  # This is the email address of your Service Account. You can leave this empty on GCP, if you want to use the service's account
  subject             = "admin@example.com"                         # This is the user you want to impersonate with Domain Wide Delegation
}
```

2. Create a Service Account Key and configure the provider like so:
```terraform
provider "gdrive" {
  service_account_key = "/path/to/sa.json"  # This is the path to your Service Account Key file or its content in JSON format
  subject             = "admin@example.com" # This is the user you want to impersonate with Domain Wide Delegation
}
```

You can also set the `SERVICE_ACCOUNT_KEY` environment variable to store either the path to the Key file or the JSON contents directly.

This provider uses [GSM](https://github.com/hanneshayashi/gsm) for authentication and API access.
You can take a look at the GSM [Setup Guide](https://gsm.hayashi-ke.online/setup), if you need help.
