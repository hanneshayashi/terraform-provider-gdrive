# terraform-provider-gdrive
A Terraform Provider for Google Drive

## Why does this exist?
I don't know. I wanted to try to create a Terraform provider and this seemed like an easy option.

## What can it do?
Not much right now. You can manage (create and update and delete) Shared Drives with Terraform!

## Is this useful in any way?
Probably not.

## How do I use this?
Well, if your really want to...\
Copy the compiled binary to\
`~/.terraform.d/plugins/github.com/hanneshayashi/gdrive/0.1.0/linux_amd64`\
or\
`%APPDATA%\terraform.d\plugins\github.com\hanneshayashi\gdrive\0.1.0\windows_amd64` (probably, haven't actually tested it under Windows)

Also, you need a GCP Service Account with Domain Wide Delegation set up with the Google Drive scope.\
This Provider uses [GSM](https://github.com/hanneshayashi/gsm)'s auth and drive packages, because I was lazy and they work pretty well.\
You can take a look a the GSM [Setup Guide](https://gsm.hayashi-ke.online/setup) if you need help.\
TL;DR:
1. Create GCP Project
2. Enable Drive API
3. Create Service Account + Key and enable DWD
4. Enter the Client ID of the Service Account Key file with the Drive scope (https://www.googleapis.com/auth/drive) in your Admin Console
