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
	"fmt"
	"strconv"

	"github.com/hanneshayashi/gsm/gsmdrive"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/api/drive/v3"
)

func resourceLabelAssignment() *schema.Resource {
	return &schema.Resource{
		Description: "Sets the labels on a Drive object",
		Schema: map[string]*schema.Schema{
			"file_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the file to assign the label to",
			},
			"label_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "",
			},
			"field": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: ``,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"field_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "",
						},
						"value_type": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "",
						},
						"values": {
							Type:        schema.TypeSet,
							Required:    true,
							Description: "",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
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
	old, new := d.GetChange("field")
	oL := old.(*schema.Set).List()
	nL := new.(*schema.Set).List()
	nM := make(map[string]bool)
	oM := make(map[string]bool)
	for _, n := range nL {
		nf := n.(map[string]any)
		nM[nf["field_id"].(string)] = true
	}
	for _, o := range oL {
		of := o.(map[string]any)
		fid := of["field_id"].(string)
		if _, ok := nM[fid]; !ok {
			oM[fid] = true
		}
	}
	req := &drive.ModifyLabelsRequest{
		LabelModifications: []*drive.LabelModification{
			{
				LabelId: labelID,
			},
		},
	}
	fields := d.Get("field").(*schema.Set)
	req.LabelModifications[0].FieldModifications = make([]*drive.LabelFieldModification, fields.Len())
	for i, f := range fields.List() {
		field := f.(map[string]any)
		valueType := field["value_type"]
		values := field["values"].(*schema.Set).List()
		req.LabelModifications[0].FieldModifications[i] = &drive.LabelFieldModification{
			FieldId: field["field_id"].(string),
		}
		fmt.Println("[DEBUG] UPDATE VALUES ON", req.LabelModifications[0].FieldModifications[i].FieldId)
		switch valueType {
		case "text":
			for v := range values {
				req.LabelModifications[0].FieldModifications[i].SetTextValues = append(req.LabelModifications[0].FieldModifications[i].SetTextValues, values[v].(string))
			}
		case "dateString":
			for v := range values {
				req.LabelModifications[0].FieldModifications[i].SetDateValues = append(req.LabelModifications[0].FieldModifications[i].SetDateValues, values[v].(string))
			}
		case "selection":
			for v := range values {
				req.LabelModifications[0].FieldModifications[i].SetSelectionValues = append(req.LabelModifications[0].FieldModifications[i].SetSelectionValues, values[v].(string))
			}
		case "user":
			for v := range values {
				req.LabelModifications[0].FieldModifications[i].SetUserValues = append(req.LabelModifications[0].FieldModifications[i].SetUserValues, values[v].(string))
			}
		case "integer":
			for v := range values {
				value, err := strconv.ParseInt(values[v].(string), 10, 64)
				if err != nil {
					return err
				}
				req.LabelModifications[0].FieldModifications[i].SetIntegerValues = append(req.LabelModifications[0].FieldModifications[i].SetIntegerValues, value)
			}
		default:
			return fmt.Errorf("unknown value_type %s", valueType)
		}
	}
	for o := range oM {
		req.LabelModifications[0].FieldModifications = append(req.LabelModifications[0].FieldModifications, &drive.LabelFieldModification{FieldId: o, UnsetValues: true})
	}
	_, err := gsmdrive.ModifyLabels(fileID, "", req)
	if err != nil {
		return err
	}
	d.SetId(combineId(fileID, labelID))
	return nil
}

func resourceReadLabelAssignment(d *schema.ResourceData, _ any) error {
	fileID := d.Get("file_id").(string)
	labelID := d.Get("label_id").(string)
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
	_, err := gsmdrive.ModifyLabels(d.Get("file_id").(string), "", &drive.ModifyLabelsRequest{
		LabelModifications: []*drive.LabelModification{
			{
				LabelId:     d.Get("label_id").(string),
				RemoveLabel: true,
			},
		},
	})
	return err
}

func resourceExistsLabelAssignment(d *schema.ResourceData, _ any) (bool, error) {
	labelID := d.Get("label_id").(string)
	r, err := gsmdrive.ListLabels(d.Get("file_id").(string), "", 1)
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
