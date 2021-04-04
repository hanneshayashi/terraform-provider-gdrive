/*
Package provider implements the Terraform provider
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
package provider

import (
	"io"
	"os"

	"github.com/hanneshayashi/gsm/gsmauth"
	"github.com/hanneshayashi/gsm/gsmdrive"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"google.golang.org/api/drive/v3"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"service_account_key": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("SERVICE_ACCOUNT_KEY", ""),
			},
			"service_account": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("SERVICE_ACCOUNT", ""),
			},
			"subject": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("SUBJECT", ""),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"gdrive_drive":      resourceDrive(),
			"gdrive_permission": resourcePermission(),
			"gdrive_file":       resourceFile(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	service_account_key := d.Get("service_account_key").(string)
	if service_account_key != "" {
		f, err := os.Open(service_account_key)
		if err != nil {
			return nil, err
		}
		saKey, err := io.ReadAll(f)
		if err != nil {
			return nil, err
		}
		client := gsmauth.GetClient(d.Get("subject").(string), saKey, drive.DriveScope)
		gsmdrive.SetClient(client)
	} else {
		serviceAccount := d.Get("service_account").(string)
		client := gsmauth.GetClientADC(d.Get("subject").(string), serviceAccount, drive.DriveScope)
		gsmdrive.SetClient(client)
	}
	return nil, nil
}
