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
	"encoding/json"
	"fmt"

	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	rsschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"google.golang.org/api/drive/v3"

	"github.com/hanneshayashi/gsm/gsmcibeta"
	cibeta "google.golang.org/api/cloudidentity/v1beta1"
)

func combineId(a, b string) string {
	return fmt.Sprintf("%s/%s", a, b)
}

func (fileModel *gdriveFileResourceModel) toRequest() *drive.File {
	return &drive.File{
		MimeType: fileModel.MimeType.ValueString(),
		Name:     fileModel.Name.ValueString(),
		DriveId:  fileModel.DriveId.ValueString(),
	}
}

func (membershipModel *gdriveOrgUnitMembershipResourceModel) move() (diags diag.Diagnostics) {
	moveOrgMembershipRequest := &cibeta.MoveOrgMembershipRequest{
		Customer:           "customers/my_customer",
		DestinationOrgUnit: "orgUnits/" + membershipModel.Parent.ValueString(),
	}
	membership, err := gsmcibeta.MoveOrgUnitMemberships("orgUnits/-/memberships/shared_drive;"+membershipModel.DriveId.ValueString(), "", moveOrgMembershipRequest)
	if err != nil {
		diags.AddError("Client Error", fmt.Sprintf("Unable to move Shared Drive to Org Unit, got error: %s", err))
		return
	}
	var m map[string]string
	j, err := membership.MarshalJSON()
	json.Unmarshal(j, &m)
	if err != nil {
		diags.AddError("Client Error", fmt.Sprintf("Unable to parse response from Org Unit move operation, got error: %s", err))
		return
	}
	membershipModel.Id = types.StringValue(m["name"])
	membershipModel.OrgUnitId = membershipModel.Id
	return diags
}

func dsId() dsschema.StringAttribute {
	return dsschema.StringAttribute{
		Computed:    true,
		Description: "The unique ID of this resource.",
	}
}

func rsId() rsschema.StringAttribute {
	id := rsschema.StringAttribute{
		Computed:    true,
		Description: "The unique ID of this resource.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}
	return id
}
