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
	"strings"

	"github.com/hanneshayashi/gsm/gsmcibeta"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	cibeta "google.golang.org/api/cloudidentity/v1beta1"
)

func resourceDriveOuMembership() *schema.Resource {
	return &schema.Resource{
		Description: `BEWARE! THE API AND THIS RESOURCE ARE IN BETA AND MAY BREAK WITHOUT WARNING!
Sets the membership of a Shared Drive in an organizational unit.
This resource requires additional setup:
1. Enable the Cloud Identity API in your GCP project
2. Add 'https://www.googleapis.com/auth/cloud-identity.orgunits' as a scope to your Domain Wide Delegation config
3. Set 'use_cloud_identity_api' to 'true' in your provider configuration

The resource will move the Shared Drive to the specified OU in your Admin Console.

Some things to note:
- You need to specify the ID of the OU (NOT THE PATH!)
  - You can find the ID via the Admin SDK (or https://gsm.hayashi-ke.online/gsm/orgunits/list/)
- If you move the Shared Drive outside of Terraform, the resource will be re-created
- A destroy of this resource will not do anything`,
		Schema: map[string]*schema.Schema{
			"drive_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "driveId of the Shared Drive",
			},
			"parent": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the organizational unit (NOT the path!)",
			},
		},
		Create: resourceCreateDriveOuMembership,
		Read:   resourceReadDriveOuMembership,
		Update: resourceCreateDriveOuMembership,
		Delete: resourceDeleteDriveOuMembership,
		Exists: resourceExistsDriveOuMembership,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceCreateDriveOuMembership(d *schema.ResourceData, _ any) error {
	moveOrgMembershipRequest := &cibeta.MoveOrgMembershipRequest{
		Customer:           "customers/my_customer",
		DestinationOrgUnit: "orgUnits/" + d.Get("parent").(string),
	}
	r, err := gsmcibeta.MoveOrgUnitMemberships("orgUnits/-/memberships/shared_drive;"+d.Get("drive_id").(string), "", moveOrgMembershipRequest)
	if err != nil {
		return err
	}
	var m map[string]string
	j, err := r.MarshalJSON()
	json.Unmarshal(j, &m)
	if err != nil {
		return err
	}
	d.SetId(m["name"])
	return nil
}

func resourceReadDriveOuMembership(d *schema.ResourceData, _ any) error {
	_, err := resourceExistsDriveOuMembership(d, nil)
	if err != nil {
		return err
	}
	return nil
}

func resourceDeleteDriveOuMembership(d *schema.ResourceData, _ any) error {
	return nil
}

func resourceExistsDriveOuMembership(d *schema.ResourceData, _ any) (bool, error) {
	name := d.Id()[0:strings.Index(d.Id(), "/memberships")]
	r, err := gsmcibeta.ListOrgUnitMemberships(name, "customers/my_customer", "", "", 1)
	for i := range r {
		if i.Name == d.Id() {
			return true, nil
		}
	}
	e := <-err
	if e != nil {
		return false, e
	}
	return false, nil
}
