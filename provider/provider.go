/*
Copyright © 2021 Hannes Hayashi

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
	"io"
	"os"

	"github.com/hanneshayashi/gsm/gsmauth"
	"github.com/hanneshayashi/gsm/gsmdrive"
	"github.com/hanneshayashi/gsm/gsmhelpers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/api/drive/v3"
)

// Provider returns the Terraform provider
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"service_account_key": {
				Type:     schema.TypeString,
				Optional: true,
				Description: `The path to or the content of a key file for your Service Account.
Leave empty if you want to use Application Default Credentials (ADC).`,
				DefaultFunc: schema.EnvDefaultFunc("SERVICE_ACCOUNT_KEY", ""),
			},
			"service_account": {
				Type:     schema.TypeString,
				Optional: true,
				Description: `The email address of the Service Account you want to impersonate with Application Default Credentials (ADC).
Leave empty if you want to use the Service Account of a GCE instance directly.`,
				DefaultFunc: schema.EnvDefaultFunc("SERVICE_ACCOUNT", ""),
			},
			"subject": {
				Type:        schema.TypeString,
				Required:    true,
				Description: `The email address of the Workspace user you want to impersonate with Domain Wide Delegation (DWD)`,
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
		},
		ResourcesMap: map[string]*schema.Resource{
			"gdrive_drive":      resourceDrive(),
			"gdrive_permission": resourcePermission(),
			"gdrive_file":       resourceFile(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"gdrive_drive":      dataSourceDrive(),
			"gdrive_permission": dataSourcePermission(),
			"gdrive_file":       dataSourceFile(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	serviceAccountKey := d.Get("service_account_key").(string)
	if serviceAccountKey != "" {
		saKey := []byte(serviceAccountKey)
		if f, err := os.Open(serviceAccountKey); err == nil {
			saKey, err = io.ReadAll(f)
			if err != nil {
				return nil, err
			}
		}
		client := gsmauth.GetClient(d.Get("subject").(string), saKey, drive.DriveScope)
		gsmdrive.SetClient(client)
	} else {
		serviceAccount := d.Get("service_account").(string)
		client := gsmauth.GetClientADC(d.Get("subject").(string), serviceAccount, drive.DriveScope)
		gsmdrive.SetClient(client)
	}
	retryOn := d.Get("retry_on").([]interface{})
	if len(retryOn) > 0 {
		for i := range retryOn {
			gsmhelpers.RetryOn = append(gsmhelpers.RetryOn, retryOn[i].(int))
		}
	}

	return nil, nil
}
