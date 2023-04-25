package provider

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"encoding/json"

	"github.com/hanneshayashi/gsm/gsmauth"
	"github.com/hanneshayashi/gsm/gsmcibeta"
	"github.com/hanneshayashi/gsm/gsmdrive"
	"github.com/hanneshayashi/gsm/gsmdrivelabels"
	"github.com/hanneshayashi/gsm/gsmhelpers"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"google.golang.org/api/drive/v3"
)

var _ provider.Provider = &gdriveProvider{}

type gdriveProvider struct {
	version string
}

// gdriveProviderModel describes the provider data model.
type gdriveProviderModel struct {
	ServiceAccountKey   types.String `tfsdk:"service_account_key"`
	ServiceAccount      types.String `tfsdk:"service_account"`
	Subject             types.String `tfsdk:"subject"`
	RetryOn             types.List   `tfsdk:"retry_on"`
	UseCloudIdentityAPI types.Bool   `tfsdk:"use_cloud_identity_api"`
	UseLabelsAPI        types.Bool   `tfsdk:"use_labels_api"`
	UseLabelsAdminScope types.Bool   `tfsdk:"use_labels_admin_scope"`
}

func (p *gdriveProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "gdrive"
	resp.Version = p.version
}

func (p *gdriveProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"service_account_key": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
				MarkdownDescription: `The path to or the content of a key file for your Service Account.
Leave empty if you want to use Application Default Credentials (ADC).<br>
You can also use the "SERVICE_ACCOUNT_KEY" environment variable to store either the path to the key file or the key itself (in JSON format).`,
				// 				DefaultFunc: schema.EnvDefaultFunc("SERVICE_ACCOUNT_KEY", ""),
			},
			"service_account": schema.StringAttribute{
				Optional: true,
				MarkdownDescription: `The email address of the Service Account you want to impersonate with Application Default Credentials (ADC).
Leave empty if you want to use the Service Account of a GCP service (GCE, Cloud Run, Cloud Build, etc) directly.<br>
You can also use the "SERVICE_ACCOUNT" environment variable.`,
				// 				DefaultFunc: schema.EnvDefaultFunc("SERVICE_ACCOUNT", ""),
			},
			"subject": schema.StringAttribute{
				Required: true,
				MarkdownDescription: `The email address of the Workspace user you want to impersonate with Domain Wide Delegation (DWD).<br>
You can also use the "SUBJECT" environment variable.`,
				// 				DefaultFunc: schema.EnvDefaultFunc("SUBJECT", ""),
			},
			"retry_on": schema.ListAttribute{
				Optional:            true,
				MarkdownDescription: `A list of HTTP error codes you want the provider to retry on (e.g. 404).`,
				ElementType:         types.Int64Type,
			},
			"use_cloud_identity_api": schema.BoolAttribute{
				Optional: true,
				MarkdownDescription: `Set this to true if you want to manage Shared Drives in organizational units.
Adds the scope 'https://www.googleapis.com/auth/cloud-identity.orgunits' to the provider's http client.
This scope needs to be added to the Domain Wide Delegation configuration in the Admin Console in Google Workspace.
Can also be set with the environment variable "USE_CLOUD_IDENTITY_API"`,
				// 				DefaultFunc: schema.EnvDefaultFunc("USE_CLOUD_IDENTITY_API", false),
			},
			"use_labels_api": schema.BoolAttribute{
				Optional: true,
				MarkdownDescription: `Set this to true if you want to manage Drive labels.
Adds the scope 'https://www.googleapis.com/auth/drive.labels' to the provider's http client.
This scope needs to be added to the Domain Wide Delegation configuration in the Admin Console in Google Workspace.
Can also be set with the environment variable "USE_LABELS_API"`,
				// 				DefaultFunc: schema.EnvDefaultFunc("USE_LABELS_API", false),
			},
			"use_labels_admin_scope": schema.BoolAttribute{
				Optional: true,
				MarkdownDescription: `Set this to true if you want to manage Drive labels with the admin scope.
Only has effect if 'use_labels_api' is also set to 'true'.
Adds the scope 'https://www.googleapis.com/auth/drive.admin.labels' to the provider's http client.
This scope needs to be added to the Domain Wide Delegation configuration in the Admin Console in Google Workspace.
Can also be set with the environment variable "USE_LABELS_ADMIN_SCOPE"`,
				// 				DefaultFunc: schema.EnvDefaultFunc("USE_LABELS_ADMIN_SCOPE", false),
			},
		},
	}
}

func (p *gdriveProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data gdriveProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	serviceAccountKey := data.ServiceAccountKey.ValueString()
	subject := data.Subject.ValueString()
	scopes := []string{drive.DriveScope}
	useCloudIdentityAPI := data.UseCloudIdentityAPI.ValueBool()
	if useCloudIdentityAPI {
		scopes = append(scopes, "https://www.googleapis.com/auth/cloud-identity.orgunits")
	}
	useLabelsAPI := data.UseLabelsAPI.ValueBool()
	if useLabelsAPI {
		scopes = append(scopes, "https://www.googleapis.com/auth/drive.labels")
		if data.UseLabelsAdminScope.ValueBool() {
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
				resp.Diagnostics.AddError("Configuration Error", fmt.Sprintf("Unable to open Service Account Key file, got error: %s", err))
				return
			}
			saKey, err = io.ReadAll(f)
			if err != nil {
				resp.Diagnostics.AddError("Configuration Error", fmt.Sprintf("Unable to read Service Account Key file, got error: %s", err))
				return
			}
		}
		client, err = gsmauth.GetClient(subject, saKey, scopes...)
		if err != nil {
			resp.Diagnostics.AddError("Configuration Error", fmt.Sprintf("Unable to get client, got error: %s", err))
			return
		}
	} else {
		client, err = gsmauth.GetClientADC(subject, data.ServiceAccount.ValueString(), scopes...)
		if err != nil {
			resp.Diagnostics.AddError("Configuration Error", fmt.Sprintf("Unable to get ADC client, got error: %s", err))
			return
		}
	}
	gsmdrive.SetClient(client)
	if useCloudIdentityAPI {
		gsmcibeta.SetClient(client)
	}
	if useLabelsAPI {
		gsmdrivelabels.SetClient(client)
	}
	var retryOn []int
	diags := data.RetryOn.ElementsAs(ctx, &retryOn, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if len(retryOn) > 0 {
		for i := range retryOn {
			gsmhelpers.RetryOn = append(gsmhelpers.RetryOn, retryOn[i])
		}
	}
	gsmhelpers.SetStandardRetrier(time.Duration(500 * time.Millisecond))
}

func (p *gdriveProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		newDrive,
	}
}

func (p *gdriveProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// NewExampleDataSource,
	}
}

func New() provider.Provider {
	return &gdriveProvider{}
}

// // Provider returns the Terraform provider
// func Provider() *schema.Provider {
// 	return &schema.Provider{
// 		ResourcesMap: map[string]*schema.Resource{
// 			"gdrive_drive":               resourceDrive(),
// 			"gdrive_permission":          resourcePermission(),
// 			"gdrive_permissions_policy":  resourcePermissionsPolicy(),
// 			"gdrive_file":                resourceFile(),
// 			"gdrive_drive_ou_membership": resourceDriveOuMembership(),
// 			"gdrive_label_assignment":    resourceLabelAssignment(),
// 			"gdrive_label_policy":        resourceLabelPolicy(),
// 		},
// 		DataSourcesMap: map[string]*schema.Resource{
// 			"gdrive_drive":       dataSourceDrive(),
// 			"gdrive_drives":      dataSourceDrives(),
// 			"gdrive_permission":  dataSourcePermission(),
// 			"gdrive_permissions": dataSourcePermissions(),
// 			"gdrive_file":        dataSourceFile(),
// 			"gdrive_files":       dataSourceFiles(),
// 			"gdrive_label":       dataSourceLabel(),
// 			"gdrive_labels":      dataSourceLabels(),
// 		},
// 		ConfigureFunc: providerConfigure,
// 	}
// }

// func providerConfigure(d *schema.ResourceData) (any, error) {
// 	serviceAccountKey := d.Get("service_account_key").(string)
// 	scopes := []string{drive.DriveScope}
// 	use_cloud_identity_api := d.Get("use_cloud_identity_api").(bool)
// 	if use_cloud_identity_api {
// 		scopes = append(scopes, "https://www.googleapis.com/auth/cloud-identity.orgunits")
// 	}
// 	use_labels_api := d.Get("use_labels_api").(bool)
// 	use_labels_admin_scope := d.Get("use_labels_admin_scope").(bool)
// 	if use_labels_api {
// 		scopes = append(scopes, "https://www.googleapis.com/auth/drive.labels")
// 		if use_labels_admin_scope {
// 			scopes = append(scopes, "https://www.googleapis.com/auth/drive.admin.labels")
// 		}
// 	}
// 	var client *http.Client
// 	var err error
// 	if serviceAccountKey != "" {
// 		var saKey []byte
// 		s := []byte(serviceAccountKey)
// 		if json.Valid(s) {
// 			saKey = s
// 		} else {
// 			var f *os.File
// 			f, err = os.Open(serviceAccountKey)
// 			if err != nil {
// 				return nil, err
// 			}
// 			saKey, err = io.ReadAll(f)
// 			if err != nil {
// 				return nil, err
// 			}
// 		}
// 		client, err = gsmauth.GetClient(d.Get("subject").(string), saKey, scopes...)
// 	} else {
// 		serviceAccount := d.Get("service_account").(string)
// 		client, err = gsmauth.GetClientADC(d.Get("subject").(string), serviceAccount, scopes...)
// 	}
// 	if err != nil {
// 		return nil, err
// 	}
// 	gsmdrive.SetClient(client)
// 	if use_cloud_identity_api {
// 		gsmcibeta.SetClient(client)
// 	}
// 	if use_labels_api {
// 		gsmdrivelabels.SetClient(client)
// 	}
// 	retryOn := d.Get("retry_on").([]any)
// 	if len(retryOn) > 0 {
// 		for i := range retryOn {
// 			gsmhelpers.RetryOn = append(gsmhelpers.RetryOn, retryOn[i].(int))
// 		}
// 	}
// 	gsmhelpers.SetStandardRetrier(time.Duration(500 * time.Millisecond))
// 	return nil, nil
// }
