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
	"fmt"

	"github.com/hanneshayashi/gsm/gsmdrive"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/api/drive/v3"
)

func resourceLabelPolicy() *schema.Resource {
	return &schema.Resource{
		Description: "Sets the labels on a Drive object",
		Schema: map[string]*schema.Schema{
			"file_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the file to assign the label policy to.",
			},
			"label": {
				Type:     schema.TypeSet,
				Optional: true,
				Description: `Represents a single label configuration.
May be used multiple times to assign multiple labels to the file.
All labels not configured here will be removed.
If no labels are defined, all labels will be removed!`,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"label_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The ID of the label.",
						},
						"field": labelFieldsR(),
					},
				},
			},
		},
		Create: resourceUpdateLabelPolicy,
		Read:   resourceReadLabelPolicy,
		Update: resourceUpdateLabelPolicy,
		Delete: resourceDeleteLabelPolicy,
		Exists: resourceExistsFile,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceUpdateLabelPolicy(d *schema.ResourceData, _ any) error {
	fileID := d.Get("file_id").(string)
	labels := d.Get("label").(*schema.Set).List()
	labelsToRemove := getRemovedItemsFromSet(d, "label", "label_id")
	req := &drive.ModifyLabelsRequest{
		LabelModifications: make([]*drive.LabelModification, len(labels)),
	}
	for l := range labels {
		label := labels[l].(map[string]any)
		labelModification, err := getLabelModification(label["label_id"].(string), getRemovedItemsFromSet(d, fmt.Sprintf("label.%d.field", l), "field_id"), label["field"].(*schema.Set))
		if err != nil {
			return err
		}
		req.LabelModifications[l] = labelModification
	}
	for o := range labelsToRemove {
		req.LabelModifications = append(req.LabelModifications, &drive.LabelModification{
			LabelId:     o,
			RemoveLabel: true,
		})
	}
	_, err := gsmdrive.ModifyLabels(fileID, "", req)
	if err != nil {
		return err
	}
	d.SetId(fileID)
	return nil
}

func resourceReadLabelPolicy(d *schema.ResourceData, _ any) error {
	labels := make([]map[string]any, 0)
	r, err := gsmdrive.ListLabels(d.Id(), "", 1)
	for l := range r {
		labels = append(labels, map[string]any{
			"label_id": l.Id,
			"field":    getFields(l.Fields),
		})
	}
	e := <-err
	if e != nil {
		return e
	}
	d.Set("label", labels)
	return nil
}

func resourceDeleteLabelPolicy(d *schema.ResourceData, _ any) error {
	labels := d.Get("label").(*schema.Set).List()
	req := &drive.ModifyLabelsRequest{
		LabelModifications: make([]*drive.LabelModification, len(labels)),
	}
	for l := range labels {
		label := labels[l].(map[string]any)
		req.LabelModifications[l] = &drive.LabelModification{
			LabelId:     label["label_id"].(string),
			RemoveLabel: true,
		}
	}
	_, err := gsmdrive.ModifyLabels(d.Get("file_id").(string), "", req)
	return err
}
