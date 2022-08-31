/*
Copyright © 2021-2022 Hannes Hayashi

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
	"github.com/hanneshayashi/gsm/gsmdrivelabels"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceLabels() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves all matching labels.",
		Schema: map[string]*schema.Schema{
			"use_admin_access": {
				Type:     schema.TypeBool,
				Optional: true,
				Description: `Set to true in order to use the user's admin credentials.
The server verifies that the user is an admin for the label before allowing access.
Requires setting the 'use_labels_admin_scope' property to 'true' in the provider config.`,
			},
			"published_only": {
				Type:     schema.TypeBool,
				Optional: true,
				Description: `Whether to include only published labels in the results.

When true, only the current published label revisions are returned.
Disabled labels are included.
Returned label resource names reference the published revision (labels/{id}/{revisionId}).

When false, the current label revisions are returned, which might not be published.
Returned label resource names don't reference a specific revision (labels/{id}).`,
			},
			"language_code": {
				Type:     schema.TypeString,
				Optional: true,
				Description: `The BCP-47 language code to use for evaluating localized field labels.
When not specified, values in the default configured language are used.`,
			},
			"minimum_role": {
				Type:     schema.TypeString,
				Optional: true,
				Description: `Specifies the level of access the user must have on the returned Labels.
The minimum role a user must have on a label.
Defaults to READER.
[READER|APPLIER|ORGANIZER|EDITOR]
READER     - A reader can read the label and associated metadata applied to Drive items.
APPLIER    - An applier can write associated metadata on Drive items in which they also have write access to. Implies READER.`,
			},
			"labels": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: `The id of this label.`,
						},
						"label_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: `The type of this label.`,
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: `The description of the label.`,
						},
						"title": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: `Title of the label.`,
						},
						"fields": labelFieldsDS(),
					},
				},
			},
		},
		Read: dataSourceReadLabels,
	}
}

func dataSourceReadLabels(d *schema.ResourceData, _ any) error {
	labels := make([]map[string]any, 0)
	r, err := gsmdrivelabels.ListLabels(d.Get("language_code").(string), "LABEL_VIEW_FULL", d.Get("minimum_role").(string), "*", d.Get("use_admin_access").(bool), d.Get("published_only").(bool), 1)
	for l := range r {
		labels = append(labels, map[string]any{
			"id":          l.Id,
			"description": l.Properties.Description,
			"label_type":  l.LabelType,
			"title":       l.Properties.Title,
			"fields":      getLabelFields(l.Fields),
		})
	}
	e := <-err
	if e != nil {
		return e
	}
	d.SetId("1")
	d.Set("labels", labels)
	return nil
}
