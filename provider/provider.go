/*
Copyright © 2021-2023 Hannes Hayashi

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/

// Package provider implements the Terraform provider
package provider

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/hanneshayashi/gsm/gsmauth"
	"github.com/hanneshayashi/gsm/gsmcibeta"
	"github.com/hanneshayashi/gsm/gsmdrive"
	"github.com/hanneshayashi/gsm/gsmdrivelabels"
	"github.com/hanneshayashi/gsm/gsmhelpers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/api/drive/v3"
)

func init() {
	schema.DescriptionKind = schema.StringMarkdown
}

// Provider returns the Terraform provider
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"service_account_key": {
				Type:     schema.TypeString,
				Optional: true,
				Description: `The path to or the content of a key file for your Service Account.
Leave empty if you want to use Application Default Credentials (ADC).<br>
You can also use the "SERVICE_ACCOUNT_KEY" environment variable to store either the path to the key file or the key itself (in JSON format).`,
				DefaultFunc: schema.EnvDefaultFunc("SERVICE_ACCOUNT_KEY", ""),
			},
			"service_account": {
				Type:     schema.TypeString,
				Optional: true,
				Description: `The email address of the Service Account you want to impersonate with Application Default Credentials (ADC).
Leave empty if you want to use the Service Account of a GCP service (GCE, Cloud Run, Cloud Build, etc) directly.<br>
You can also use the "SERVICE_ACCOUNT" environment variable.`,
				DefaultFunc: schema.EnvDefaultFunc("SERVICE_ACCOUNT", ""),
			},
			"subject": {
				Type:     schema.TypeString,
				Required: true,
				Description: `The email address of the Workspace user you want to impersonate with Domain Wide Delegation (DWD).<br>
You can also use the "SUBJECT" environment variable.`,
				DefaultFunc: schema.EnvDefaultFunc("SUBJECT", ""),
			},
			"retry_on": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: `A list of HTTP error codes you want the provider to retry on (e.g. 404).`,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"use_cloud_identity_api": {
				Type:     schema.TypeBool,
				Optional: true,
				Description: `Set this to true if you want to manage Shared Drives in organizational units.
Adds the scope 'https://www.googleapis.com/auth/cloud-identity.orgunits' to the provider's http client.
This scope needs to be added to the Domain Wide Delegation configuration in the Admin Console in Google Workspace.
Can also be set with the environment variable "USE_CLOUD_IDENTITY_API"`,
				DefaultFunc: schema.EnvDefaultFunc("USE_CLOUD_IDENTITY_API", false),
			},
			"use_labels_api": {
				Type:     schema.TypeBool,
				Optional: true,
				Description: `Set this to true if you want to manage Drive labels.
Adds the scope 'https://www.googleapis.com/auth/drive.labels' to the provider's http client.
This scope needs to be added to the Domain Wide Delegation configuration in the Admin Console in Google Workspace.
Can also be set with the environment variable "USE_LABELS_API"`,
				DefaultFunc: schema.EnvDefaultFunc("USE_LABELS_API", false),
			},
			"use_labels_admin_scope": {
				Type:     schema.TypeBool,
				Optional: true,
				Description: `Set this to true if you want to manage Drive labels with the admin scope.
Only has effect if 'use_labels_api' is also set to 'true'.
Adds the scope 'https://www.googleapis.com/auth/drive.admin.labels' to the provider's http client.
This scope needs to be added to the Domain Wide Delegation configuration in the Admin Console in Google Workspace.
Can also be set with the environment variable "USE_LABELS_ADMIN_SCOPE"`,
				DefaultFunc: schema.EnvDefaultFunc("USE_LABELS_ADMIN_SCOPE", false),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"gdrive_drive":               resourceDrive(),
			"gdrive_permission":          resourcePermission(),
			"gdrive_permissions_policy":  resourcePermissionsPolicy(),
			"gdrive_file":                resourceFile(),
			"gdrive_drive_ou_membership": resourceDriveOuMembership(),
			"gdrive_label_assignment":    resourceLabelAssignment(),
			"gdrive_label_policy":        resourceLabelPolicy(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"gdrive_drive":       dataSourceDrive(),
			"gdrive_drives":      dataSourceDrives(),
			"gdrive_permission":  dataSourcePermission(),
			"gdrive_permissions": dataSourcePermissions(),
			"gdrive_file":        dataSourceFile(),
			"gdrive_files":       dataSourceFiles(),
			"gdrive_label":       dataSourceLabel(),
			"gdrive_labels":      dataSourceLabels(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (any, error) {
	serviceAccountKey := d.Get("service_account_key").(string)
	scopes := []string{drive.DriveScope}
	use_cloud_identity_api := d.Get("use_cloud_identity_api").(bool)
	if use_cloud_identity_api {
		scopes = append(scopes, "https://www.googleapis.com/auth/cloud-identity.orgunits")
	}
	use_labels_api := d.Get("use_labels_api").(bool)
	use_labels_admin_scope := d.Get("use_labels_admin_scope").(bool)
	if use_labels_api {
		scopes = append(scopes, "https://www.googleapis.com/auth/drive.labels")
		if use_labels_admin_scope {
			scopes = append(scopes, "https://www.googleapis.com/auth/drive.admin.labels")
		}
	}
	var client *http.Client
	var err error
	if serviceAccountKey != "" {
		var saKey []byte
		s := []byte(serviceAccountKey)
		if json.Valid(s) {
			saKey = s
		} else {
			var f *os.File
			f, err = os.Open(serviceAccountKey)
			if err != nil {
				return nil, err
			}
			saKey, err = io.ReadAll(f)
			if err != nil {
				return nil, err
			}
		}
		client, err = gsmauth.GetClient(d.Get("subject").(string), saKey, scopes...)
	} else {
		serviceAccount := d.Get("service_account").(string)
		client, err = gsmauth.GetClientADC(d.Get("subject").(string), serviceAccount, scopes...)
	}
	if err != nil {
		return nil, err
	}
	gsmdrive.SetClient(client)
	if use_cloud_identity_api {
		gsmcibeta.SetClient(client)
	}
	if use_labels_api {
		gsmdrivelabels.SetClient(client)
	}
	retryOn := d.Get("retry_on").([]any)
	if len(retryOn) > 0 {
		for i := range retryOn {
			gsmhelpers.RetryOn = append(gsmhelpers.RetryOn, retryOn[i].(int))
		}
	}
	gsmhelpers.SetStandardRetrier(time.Duration(500 * time.Millisecond))
	return nil, nil
}
