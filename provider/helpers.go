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

	"github.com/hanneshayashi/gsm/gsmdrive"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"google.golang.org/api/drive/v3"
)

// import (
// 	"fmt"
// 	"strconv"
// 	"strings"

// 	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
// 	"google.golang.org/api/drive/v3"
// 	"google.golang.org/api/drivelabels/v2"
// )

// func contains(s string, slice []string) bool {
// 	for i := range slice {
// 		if s == slice[i] {
// 			return true
// 		}
// 	}
// 	return false
// }

// func getRestrictions(d *drive.Drive) (restrictions map[string]bool) {
// 	if d.Restrictions != nil {
// 		restrictions = map[string]bool{
// 			"admin_managed_restrictions":      d.Restrictions.AdminManagedRestrictions,
// 			"copy_requires_writer_permission": d.Restrictions.CopyRequiresWriterPermission,
// 			"domain_users_only":               d.Restrictions.DomainUsersOnly,
// 			"drive_members_only":              d.Restrictions.DriveMembersOnly,
// 		}
// 	}
// 	return
// }

// func splitId(id string) (fileID, secondaryID string) {
// 	ids := strings.Split(id, "/")
// 	return ids[0], ids[1]
// }

func combineId(a, b string) string {
	return fmt.Sprintf("%s/%s", a, b)
}

// func permissionToSet(i any) int {
// 	m := i.(map[string]any)
// 	id, _ := strconv.Atoi(m["permission_id"].(string))
// 	return id
// }

// func validatePermissionType(v any, _ string) (ws []string, es []error) {
// 	value := v.(string)
// 	if contains(value, validPermissionTypes) {
// 		return nil, nil
// 	}
// 	es = append(es, fmt.Errorf("%s is not a valid permission type", value))
// 	return nil, es
// }

// func getParent(file *drive.File) (parent string) {
// 	if file.Parents != nil {
// 		parent = file.Parents[0]
// 	}
// 	return
// }

// func getLabelFields(labelFields []*drivelabels.GoogleAppsDriveLabelsV2Field) []map[string]any {
// 	fields := make([]map[string]any, 0)
// 	for i := range labelFields {
// 		field := make(map[string]any)
// 		field["id"] = labelFields[i].Id
// 		field["query_key"] = labelFields[i].QueryKey
// 		field["life_cycle"] = []map[string]any{
// 			{
// 				"state": labelFields[i].Lifecycle.State,
// 			},
// 		}
// 		field["properties"] = []map[string]any{
// 			{
// 				"display_name": labelFields[i].Properties.DisplayName,
// 				"required":     labelFields[i].Properties.Required,
// 			},
// 		}
// 		if labelFields[i].TextOptions != nil {
// 			field["value_type"] = "text"
// 			field["text_options"] = []map[string]int64{
// 				{
// 					"max_length": labelFields[i].TextOptions.MaxLength,
// 					"min_length": labelFields[i].TextOptions.MinLength,
// 				},
// 			}
// 		} else if labelFields[i].IntegerOptions != nil {
// 			field["value_type"] = "integer"
// 			field["integer_options"] = []map[string]int64{
// 				{
// 					"max_value": labelFields[i].IntegerOptions.MaxValue,
// 					"min_value": labelFields[i].IntegerOptions.MinValue,
// 				},
// 			}
// 		} else if labelFields[i].UserOptions != nil {
// 			field["value_type"] = "user"
// 			if labelFields[i].UserOptions.ListOptions != nil {
// 				userOption := make(map[string]any)
// 				if labelFields[i].UserOptions.ListOptions != nil {
// 					userOption["list_options"] = []map[string]int64{
// 						{
// 							"max_entries": labelFields[i].UserOptions.ListOptions.MaxEntries,
// 						},
// 					}
// 				}
// 				field["user_options"] = []map[string]any{
// 					userOption,
// 				}
// 			}
// 		} else if labelFields[i].SelectionOptions != nil {
// 			field["value_type"] = "selection"
// 			if labelFields[i].SelectionOptions.ListOptions != nil || labelFields[i].SelectionOptions.Choices != nil {
// 				selectionOptions := make(map[string]any)
// 				if labelFields[i].SelectionOptions.ListOptions != nil {
// 					selectionOptions["list_options"] = []map[string]int64{
// 						{
// 							"max_entries": labelFields[i].SelectionOptions.ListOptions.MaxEntries,
// 						},
// 					}
// 				}
// 				if labelFields[i].SelectionOptions.Choices != nil {
// 					choices := make([]map[string]any, len(labelFields[i].SelectionOptions.Choices))
// 					for c := range labelFields[i].SelectionOptions.Choices {
// 						choices[c] = map[string]any{
// 							"display_name": labelFields[i].SelectionOptions.Choices[c].Properties.DisplayName,
// 							"id":           labelFields[i].SelectionOptions.Choices[c].Id,
// 							"life_cycle": []map[string]any{
// 								{
// 									"state": labelFields[i].SelectionOptions.Choices[c].Lifecycle.State,
// 								},
// 							},
// 						}
// 					}
// 					selectionOptions["choices"] = choices
// 				}
// 				field["selection_options"] = []map[string]any{
// 					selectionOptions,
// 				}
// 			}
// 		} else if labelFields[i].DateOptions != nil {
// 			field["value_type"] = "dateString"
// 			dateOptions := map[string]any{
// 				"date_format":      labelFields[i].DateOptions.DateFormat,
// 				"date_format_type": labelFields[i].DateOptions.DateFormatType,
// 			}
// 			if labelFields[i].DateOptions.MinValue != nil {
// 				dateOptions["min_value"] = []map[string]any{
// 					{
// 						"day":   labelFields[i].DateOptions.MinValue.Day,
// 						"month": labelFields[i].DateOptions.MinValue.Month,
// 						"year":  labelFields[i].DateOptions.MinValue.Year,
// 					},
// 				}
// 			}
// 			if labelFields[i].DateOptions.MaxValue != nil {
// 				dateOptions["max_value"] = []map[string]any{
// 					{
// 						"day":   labelFields[i].DateOptions.MaxValue.Day,
// 						"month": labelFields[i].DateOptions.MaxValue.Month,
// 						"year":  labelFields[i].DateOptions.MaxValue.Year,
// 					},
// 				}
// 			}
// 			field["date_options"] = []map[string]any{
// 				dateOptions,
// 			}
// 		}
// 		fields = append(fields, field)
// 	}
// 	return fields
// }

// func getFields(labelFields map[string]drive.LabelField) []map[string]any {
// 	fields := make([]map[string]any, 0)
// 	for f := range labelFields {
// 		field := map[string]any{
// 			"field_id":   labelFields[f].Id,
// 			"value_type": labelFields[f].ValueType,
// 		}
// 		switch labelFields[f].ValueType {
// 		case "dateString":
// 			field["values"] = labelFields[f].DateString
// 		case "text":
// 			field["values"] = labelFields[f].Text
// 		case "user":
// 			values := make([]string, len(labelFields[f].User))
// 			for u := range labelFields[f].User {
// 				values[u] = labelFields[f].User[u].EmailAddress
// 			}
// 			field["values"] = values
// 		case "selection":
// 			field["values"] = labelFields[f].Selection
// 		case "integer":
// 			values := []string{}
// 			for _, value := range labelFields[f].Integer {
// 				values = append(values, strconv.FormatInt(value, 10))
// 			}
// 			field["values"] = values
// 		}
// 		fields = append(fields, field)
// 	}
// 	return fields
// }

// func getRemovedItemsFromSet(d *schema.ResourceData, fieldName, id string) map[string]bool {
// 	old, new := d.GetChange(fieldName)
// 	oL := old.(*schema.Set).List()
// 	nL := new.(*schema.Set).List()
// 	nM := make(map[string]bool)
// 	oM := make(map[string]bool)
// 	for _, n := range nL {
// 		nf := n.(map[string]any)
// 		nM[nf[id].(string)] = true
// 	}
// 	for _, o := range oL {
// 		of := o.(map[string]any)
// 		fid := of[id].(string)
// 		if _, ok := nM[fid]; !ok {
// 			oM[fid] = true
// 		}
// 	}
// 	return oM
// }

// func getLabelModification(labelID string, fieldsToRemove map[string]bool, fields *schema.Set) (*drive.LabelModification, error) {
// 	labelModification := &drive.LabelModification{
// 		LabelId: labelID,
// 	}
// 	labelModification.FieldModifications = make([]*drive.LabelFieldModification, fields.Len())
// 	for i, f := range fields.List() {
// 		field := f.(map[string]any)
// 		valueType := field["value_type"]
// 		values := field["values"].(*schema.Set).List()
// 		labelModification.FieldModifications[i] = &drive.LabelFieldModification{
// 			FieldId: field["field_id"].(string),
// 		}
// 		switch valueType {
// 		case "text":
// 			for v := range values {
// 				labelModification.FieldModifications[i].SetTextValues = append(labelModification.FieldModifications[i].SetTextValues, values[v].(string))
// 			}
// 		case "dateString":
// 			for v := range values {
// 				labelModification.FieldModifications[i].SetDateValues = append(labelModification.FieldModifications[i].SetDateValues, values[v].(string))
// 			}
// 		case "selection":
// 			for v := range values {
// 				labelModification.FieldModifications[i].SetSelectionValues = append(labelModification.FieldModifications[i].SetSelectionValues, values[v].(string))
// 			}
// 		case "user":
// 			for v := range values {
// 				labelModification.FieldModifications[i].SetUserValues = append(labelModification.FieldModifications[i].SetUserValues, values[v].(string))
// 			}
// 		case "integer":
// 			for v := range values {
// 				value, err := strconv.ParseInt(values[v].(string), 10, 64)
// 				if err != nil {
// 					return nil, err
// 				}
// 				labelModification.FieldModifications[i].SetIntegerValues = append(labelModification.FieldModifications[i].SetIntegerValues, value)
// 			}
// 		default:
// 			return nil, fmt.Errorf("unknown value_type %s", valueType)
// 		}
// 	}
// 	for o := range fieldsToRemove {
// 		labelModification.FieldModifications = append(labelModification.FieldModifications, &drive.LabelFieldModification{
// 			FieldId:     o,
// 			UnsetValues: true,
// 		})
// 	}
// 	return labelModification, nil
// }

func (plan *gdrivePermissionPolicyResourceModel) getCurrentPermissions(ctx context.Context) ([]*gdrivePermissionPolicyPermissionResourceModel, error) {
	currentP, err := gsmdrive.ListPermissions(plan.FileId.ValueString(), "", fmt.Sprintf("permissions(%s),nextPageToken", fieldsPermission), plan.UseDomainAdminAccess.ValueBool(), 1)
	permissions := []*gdrivePermissionPolicyPermissionResourceModel{}
	for i := range currentP {
		if i.PermissionDetails != nil && i.PermissionDetails[0].Inherited {
			tflog.Debug(ctx, fmt.Sprintf("XXX Skipping permission %s on %s", i.Id, plan.FileId.ValueString()))
			continue
		}
		tflog.Debug(ctx, fmt.Sprintf("YYY ADDING permission %s on %s", i.Id, plan.FileId.ValueString()))
		permissions = append(permissions, &gdrivePermissionPolicyPermissionResourceModel{
			PermissionId: types.StringValue(i.Id),
			Type:         types.StringValue(i.Type),
			Domain:       types.StringValue(i.Domain),
			EmailAddress: types.StringValue(i.EmailAddress),
			Role:         types.StringValue(i.Role),
		})
	}
	e := <-err
	if e != nil {
		return nil, fmt.Errorf("Unable to read current permissions on file, got error: %s", e)
	}
	return permissions, nil
}

func permissionsToMap(permissions []*gdrivePermissionPolicyPermissionResourceModel) map[string]*gdrivePermissionPolicyPermissionResourceModel {
	m := map[string]*gdrivePermissionPolicyPermissionResourceModel{}
	for i := range permissions {
		m[combineId(permissions[i].Domain.ValueString(), permissions[i].EmailAddress.ValueString())] = permissions[i]
	}
	return m
}

func (plan *gdrivePermissionPolicyResourceModel) setPermissionPolicy(ctx context.Context) error {
	fileID := plan.FileId.ValueString()
	useDomAccess := plan.UseDomainAdminAccess.ValueBool()
	currentP, err := plan.getCurrentPermissions(ctx)
	if err != nil {
		return err
	}
	currentPMap := permissionsToMap(currentP)
	plannedPMap := map[string]*gdrivePermissionPolicyPermissionResourceModel{}
	for i := range plan.Permissions {
		role := plan.Permissions[i].Role.ValueString()
		mapID := combineId(plan.Permissions[i].Domain.ValueString(), plan.Permissions[i].EmailAddress.ValueString())
		plannedPMap[mapID] = plan.Permissions[i]
		_, ok := currentPMap[mapID]
		if ok {
			tflog.Debug(ctx, fmt.Sprintf("ZZZ found %s on %s", plan.Permissions[i].PermissionId.ValueString(), plan.FileId.ValueString()))
			if !plan.Permissions[i].Role.Equal(currentPMap[mapID].Role) {
				tflog.Debug(ctx, fmt.Sprintf("QQQ I will %s on %s from %s to %s", plan.Permissions[i].PermissionId.ValueString(), plan.FileId.ValueString(), currentPMap[mapID].Role.ValueString(), plan.Permissions[i].PermissionId.ValueString()))
				permissionReq := &drive.Permission{
					Role: role,
				}
				p, err := gsmdrive.UpdatePermission(fileID, currentPMap[mapID].PermissionId.ValueString(), fieldsPermission, useDomAccess, false, permissionReq)
				if err != nil {
					return fmt.Errorf("Unable to update permission on file, got error: %s", err)
				}
				plan.Permissions[i].PermissionId = types.StringValue(p.Id)
			} else {
				plan.Permissions[i].PermissionId = currentPMap[mapID].PermissionId
			}
			delete(currentPMap, mapID)
		} else {
			permissionReq := &drive.Permission{
				Domain:       plan.Permissions[i].Domain.ValueString(),
				EmailAddress: plan.Permissions[i].EmailAddress.ValueString(),
				Role:         role,
				Type:         plan.Permissions[i].Type.ValueString(),
			}
			p, err := gsmdrive.CreatePermission(fileID, plan.Permissions[i].EmailMessage.ValueString(), fieldsPermission, useDomAccess, plan.Permissions[i].SendNotificationEmail.ValueBool(), plan.Permissions[i].TransferOwnership.ValueBool(), plan.Permissions[i].MoveToNewOwnersRoot.ValueBool(), permissionReq)
			if err != nil {
				return fmt.Errorf("Unable to set permission on file, got error: %s", err)
			}
			plan.Permissions[i].PermissionId = types.StringValue(p.Id)
		}
	}
	for i := range currentPMap {
		_, ok := plannedPMap[i]
		if !ok {
			_, err := gsmdrive.DeletePermission(fileID, currentPMap[i].PermissionId.ValueString(), useDomAccess)
			if err != nil {
				return fmt.Errorf("Unable to remove permission from file, got error: %s", err)
			}
		}
	}
	plan.Id = types.StringValue(fileID)
	return nil
}
