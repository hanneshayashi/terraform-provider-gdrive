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

func (permissionPolicyModel *gdrivePermissionPolicyResourceModel) getCurrentPermissions(ctx context.Context) ([]*gdrivePermissionPolicyPermissionResourceModel, error) {
	currentP, err := gsmdrive.ListPermissions(permissionPolicyModel.FileId.ValueString(), "", fmt.Sprintf("permissions(%s),nextPageToken", fieldsPermission), permissionPolicyModel.UseDomainAdminAccess.ValueBool(), 1)
	permissions := []*gdrivePermissionPolicyPermissionResourceModel{}
	for i := range currentP {
		if i.PermissionDetails != nil && i.PermissionDetails[0].Inherited {
			continue
		}
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

func (permissionPolicyModel *gdrivePermissionPolicyResourceModel) setPermissionPolicy(ctx context.Context) error {
	fileID := permissionPolicyModel.FileId.ValueString()
	useDomAccess := permissionPolicyModel.UseDomainAdminAccess.ValueBool()
	currentP, err := permissionPolicyModel.getCurrentPermissions(ctx)
	if err != nil {
		return err
	}
	currentPMap := permissionsToMap(currentP)
	plannedPMap := map[string]*gdrivePermissionPolicyPermissionResourceModel{}
	for i := range permissionPolicyModel.Permissions {
		role := permissionPolicyModel.Permissions[i].Role.ValueString()
		mapID := combineId(permissionPolicyModel.Permissions[i].Domain.ValueString(), permissionPolicyModel.Permissions[i].EmailAddress.ValueString())
		plannedPMap[mapID] = permissionPolicyModel.Permissions[i]
		_, ok := currentPMap[mapID]
		if ok {
			tflog.Debug(ctx, fmt.Sprintf("ZZZ found %s on %s", permissionPolicyModel.Permissions[i].PermissionId.ValueString(), permissionPolicyModel.FileId.ValueString()))
			if !permissionPolicyModel.Permissions[i].Role.Equal(currentPMap[mapID].Role) {
				tflog.Debug(ctx, fmt.Sprintf("QQQ I will %s on %s from %s to %s", permissionPolicyModel.Permissions[i].PermissionId.ValueString(), permissionPolicyModel.FileId.ValueString(), currentPMap[mapID].Role.ValueString(), permissionPolicyModel.Permissions[i].PermissionId.ValueString()))
				permissionReq := &drive.Permission{
					Role: role,
				}
				p, err := gsmdrive.UpdatePermission(fileID, currentPMap[mapID].PermissionId.ValueString(), fieldsPermission, useDomAccess, false, permissionReq)
				if err != nil {
					return fmt.Errorf("Unable to update permission on file, got error: %s", err)
				}
				permissionPolicyModel.Permissions[i].PermissionId = types.StringValue(p.Id)
			} else {
				permissionPolicyModel.Permissions[i].PermissionId = currentPMap[mapID].PermissionId
			}
			delete(currentPMap, mapID)
		} else {
			permissionReq := &drive.Permission{
				Domain:       permissionPolicyModel.Permissions[i].Domain.ValueString(),
				EmailAddress: permissionPolicyModel.Permissions[i].EmailAddress.ValueString(),
				Role:         role,
				Type:         permissionPolicyModel.Permissions[i].Type.ValueString(),
			}
			p, err := gsmdrive.CreatePermission(fileID, permissionPolicyModel.Permissions[i].EmailMessage.ValueString(), fieldsPermission, useDomAccess, permissionPolicyModel.Permissions[i].SendNotificationEmail.ValueBool(), permissionPolicyModel.Permissions[i].TransferOwnership.ValueBool(), permissionPolicyModel.Permissions[i].MoveToNewOwnersRoot.ValueBool(), permissionReq)
			if err != nil {
				return fmt.Errorf("Unable to set permission on file, got error: %s", err)
			}
			permissionPolicyModel.Permissions[i].PermissionId = types.StringValue(p.Id)
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
	permissionPolicyModel.Id = types.StringValue(fileID)
	return nil
}
