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
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("SERVICE_ACCOUNT_KEY", ""),
			},
			"subject": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("SUBJECT", ""),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"gdrive_drive": resourceDrive(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	f, err := os.Open(d.Get("service_account_key").(string))
	if err != nil {
		return nil, err
	}
	saKey, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	client := gsmauth.GetClient(d.Get("subject").(string), saKey, drive.DriveScope)
	gsmdrive.SetClient(client)
	return nil, nil
}
