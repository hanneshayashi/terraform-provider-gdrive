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
)

var _ provider.Provider = &gdriveProvider{}

type gdriveProvider struct {
	version string
}

// gdriveProviderModel describes the provider data model.
type gdriveProviderModel struct {
	ServiceAccountKey types.String `tfsdk:"service_account_key"`
	ServiceAccount    types.String `tfsdk:"service_account"`
	Subject           types.String `tfsdk:"subject"`
	RetryOn           types.List   `tfsdk:"retry_on"`
	Scopes            types.List   `tfsdk:"scopes"`
}

func (p *gdriveProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "gdrive"
	resp.Version = p.version
}

func (p *gdriveProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "",
		Attributes: map[string]schema.Attribute{
			"service_account_key": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
				MarkdownDescription: `The path to or the content of a key file for your Service Account.
Leave empty if you want to use Application Default Credentials (ADC) (**recommended**).<br>
You can also use the "SERVICE_ACCOUNT_KEY" environment variable to store either the path to the key file or the key itself (in JSON format).`,
			},
			"service_account": schema.StringAttribute{
				Optional: true,
				MarkdownDescription: `The email address of the Service Account you want to impersonate with Application Default Credentials (ADC).
Leave empty if you want to use the Service Account of a GCP service (GCE, Cloud Run, Cloud Build, etc) directly.<br>
You can also use the "SERVICE_ACCOUNT" environment variable.`,
			},
			"subject": schema.StringAttribute{
				Optional: true,
				MarkdownDescription: `The email address of the Workspace user you want to impersonate with Domain Wide Delegation (DWD).<br>
You can also use the "SUBJECT" environment variable.`,
			},
			"retry_on": schema.ListAttribute{
				Optional: true,
				MarkdownDescription: `A list of HTTP error codes you want the provider to retry on.
If this is unset, the provider will retry on 404 and 502 using an exponential backoff strategy. If you DON'T want the provider to retry on any error, set this to an empty list.
The provider will ALWAYS retry on 403 errors that indicate a rate limiting / quota issue.`,
				ElementType: types.Int64Type,
			},
			"scopes": schema.ListAttribute{
				Optional: true,
				MarkdownDescription: `List of scopes that the provider will add to the API client.
If this is unset, the provider will use the following scopes that must be added to the Domain-Wide Delegation configuration in the Google Workspace Admin Console:
* https://www.googleapis.com/auth/drive
* https://www.googleapis.com/auth/drive.labels
* https://www.googleapis.com/auth/drive.admin.labels
* https://www.googleapis.com/auth/cloud-identity.orgunits`,
				ElementType: types.StringType,
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
	if serviceAccountKey == "" {
		serviceAccountKey = os.Getenv("SERVICE_ACCOUNT_KEY")
	}
	serviceAccount := data.ServiceAccount.ValueString()
	if serviceAccount == "" {
		serviceAccount = os.Getenv("SERVICE_ACCOUNT")
	}
	subject := data.Subject.ValueString()
	if subject == "" {
		subject = os.Getenv("SUBJECT")
	}
	if subject == "" {
		resp.Diagnostics.AddError("Configuration Error", fmt.Sprintf("Subject must be set"))
		return
	}
	var scopes []string
	if data.Scopes.IsNull() {
		scopes = []string{
			"https://www.googleapis.com/auth/cloud-identity.orgunits",
			"https://www.googleapis.com/auth/drive",
			"https://www.googleapis.com/auth/drive.labels",
			"https://www.googleapis.com/auth/drive.admin.labels",
		}
	} else {
		resp.Diagnostics.Append(data.Scopes.ElementsAs(ctx, &scopes, false)...)
		if resp.Diagnostics.HasError() {
			return
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
		client, err = gsmauth.GetClientADC(subject, serviceAccount, scopes...)
		if err != nil {
			resp.Diagnostics.AddError("Configuration Error", fmt.Sprintf("Unable to get ADC client, got error: %s", err))
			return
		}
	}
	gsmdrive.SetClient(client)
	gsmcibeta.SetClient(client)
	gsmdrivelabels.SetClient(client)
	var retryOn []int
	if data.RetryOn.IsNull() {
		retryOn = []int{404, 502}
	} else {
		resp.Diagnostics.Append(data.RetryOn.ElementsAs(ctx, &retryOn, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	for i := range retryOn {
		gsmhelpers.RetryOn = append(gsmhelpers.RetryOn, retryOn[i])
	}
	gsmhelpers.SetStandardRetrier(time.Duration(500*time.Millisecond), time.Duration(60*time.Second), time.Duration(3*time.Minute))
}

func (p *gdriveProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		newDrive,
		newFile,
		newPermission,
		newPermissionPolicy,
		newLabelAssignment,
		newLabelPolicy,
		newOrgUnitMembership,
		newLabel,
		newLabelTextField,
		newLabelIntegerField,
		newLabelDateField,
		newLabelUserField,
		newLabelSelectionField,
		newLabelSelectionChoice,
		newLabelPermission,
	}
}

func (p *gdriveProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		newDriveDataSource,
		newDrivesDataSource,
		newFileDataSource,
		newFilesDataSource,
		newLabelDataSource,
		newLabelsDataSource,
		newPermissionDataSource,
		newPermissionsDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &gdriveProvider{
			version: version,
		}
	}
}
