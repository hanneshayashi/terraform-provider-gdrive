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
	"strconv"

	"github.com/hanneshayashi/gsm/gsmdrive"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/api/drive/v3"
)

func labelFieldsR() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeSet,
		Required:    true,
		Description: ``,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"field_id": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The identifier of this field.",
				},
				"value_type": {
					Type:     schema.TypeString,
					Required: true,
					Description: `The field type.
While new values may be supported in the future, the following are currently allowed:
- dateString
- integer
- selection
- text
- user`,
				},
				"values": {
					Type:     schema.TypeSet,
					Required: true,
					Description: `The values that should be set.
Must be compatible with the specified valueType.`,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
			},
		},
	}
}

func resourceLabelAssignment() *schema.Resource {
	return &schema.Resource{
		Description: "Sets the labels on a Drive object",
		Schema: map[string]*schema.Schema{
			"file_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the file to assign the label to.",
			},
			"label_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the label.",
			},
			"field": labelFieldsR(),
		},
		Create: resourceUpdateLabelAssignment,
		Read:   resourceReadLabelAssignment,
		Update: resourceUpdateLabelAssignment,
		Delete: resourceDeleteLabelAssignment,
		Exists: resourceExistsLabelAssignment,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceUpdateLabelAssignment(d *schema.ResourceData, _ any) error {
	labelID := d.Get("label_id").(string)
	fileID := d.Get("file_id").(string)
	labelModification, err := getLabelModification(labelID, getRemovedItemsFromSet(d, "field", "field_id"), d.Get("field").(*schema.Set))
	if err != nil {
		return err
	}
	req := &drive.ModifyLabelsRequest{
		LabelModifications: []*drive.LabelModification{
			labelModification,
		},
	}
	_, err = gsmdrive.ModifyLabels(fileID, "", req)
	if err != nil {
		return err
	}
	d.SetId(combineId(fileID, labelID))
	return nil
}

func resourceReadLabelAssignment(d *schema.ResourceData, _ any) error {
	fileID, labelID := splitId(d.Id())
	fields := make([]map[string]any, 0)
	r, err := gsmdrive.ListLabels(fileID, "", 1)
	for l := range r {
		if l.Id == labelID {
			for f := range l.Fields {
				field := map[string]any{
					"field_id":   l.Fields[f].Id,
					"value_type": l.Fields[f].ValueType,
				}
				switch l.Fields[f].ValueType {
				case "dateString":
					field["values"] = l.Fields[f].DateString
				case "text":
					field["values"] = l.Fields[f].Text
				case "user":
					values := make([]string, len(l.Fields[f].User))
					for u := range l.Fields[f].User {
						values[u] = l.Fields[f].User[u].EmailAddress
					}
					field["values"] = values
				case "selection":
					field["values"] = l.Fields[f].Selection
				case "integer":
					values := []string{}
					for _, value := range l.Fields[f].Integer {
						values = append(values, strconv.FormatInt(value, 10))
					}
					field["values"] = values
				}
				fields = append(fields, field)
			}
		}
	}
	e := <-err
	if e != nil {
		return e
	}
	d.Set("field", fields)
	return nil
}

func resourceDeleteLabelAssignment(d *schema.ResourceData, _ any) error {
	fileID, labelID := splitId(d.Id())
	_, err := gsmdrive.ModifyLabels(fileID, "", &drive.ModifyLabelsRequest{
		LabelModifications: []*drive.LabelModification{
			{
				LabelId:     labelID,
				RemoveLabel: true,
			},
		},
	})
	return err
}

func resourceExistsLabelAssignment(d *schema.ResourceData, _ any) (bool, error) {
	fileID, labelID := splitId(d.Id())
	r, err := gsmdrive.ListLabels(fileID, "", 1)
	for l := range r {
		if l.Id == labelID {
			return true, nil
		}
	}
	e := <-err
	if e != nil {
		return false, e
	}
	return false, nil
}
