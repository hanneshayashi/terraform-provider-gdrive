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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"google.golang.org/api/drive/v3"
)

func (permissionPolicyModel *gdrivePermissionPolicyResourceModel) populate(ctx context.Context) (diags diag.Diagnostics) {
	currentP, err := gsmdrive.ListPermissions(permissionPolicyModel.FileId.ValueString(), "", fmt.Sprintf("permissions(%s),nextPageToken", fieldsPermission), permissionPolicyModel.UseDomainAdminAccess.ValueBool(), 1)
	for i := range currentP {
		if i.PermissionDetails != nil && i.PermissionDetails[0].Inherited {
			continue
		}
		permissionPolicyModel.Permissions = append(permissionPolicyModel.Permissions, &gdrivePermissionPolicyPermissionResourceModel{
			PermissionId: types.StringValue(i.Id),
			Id:           types.StringValue(i.Id),
			Type:         types.StringValue(i.Type),
			Domain:       types.StringValue(i.Domain),
			EmailAddress: types.StringValue(i.EmailAddress),
			Role:         types.StringValue(i.Role),
		})
	}
	e := <-err
	if e != nil {
		diags.AddError("Client Error", fmt.Sprintf("Unable to list permissions on file, got error: %s", e))
	}
	return diags
}

func (permissionsPolicyModel *gdrivePermissionPolicyResourceModel) toMap() map[string]*gdrivePermissionPolicyPermissionResourceModel {
	m := map[string]*gdrivePermissionPolicyPermissionResourceModel{}
	for i := range permissionsPolicyModel.Permissions {
		m[combineId(permissionsPolicyModel.Permissions[i].Domain.ValueString(), permissionsPolicyModel.Permissions[i].EmailAddress.ValueString())] = permissionsPolicyModel.Permissions[i]
	}
	return m
}

func (permissionModel *gdrivePermissionPolicyPermissionResourceModel) toRequest() *drive.Permission {
	return &drive.Permission{
		Domain:       permissionModel.Domain.ValueString(),
		EmailAddress: permissionModel.EmailAddress.ValueString(),
		Role:         permissionModel.Role.ValueString(),
		Type:         permissionModel.Type.ValueString(),
	}
}

func setPermissionDiffs(plan, state *gdrivePermissionPolicyResourceModel, ctx context.Context) (diags diag.Diagnostics) {
	fileId := plan.FileId.ValueString()
	useDomAccess := plan.UseDomainAdminAccess.ValueBool()
	planPermissions := plan.toMap()
	statePermissions := state.toMap()
	for i := range planPermissions {
		_, permissionAlreadyExists := statePermissions[i]
		if permissionAlreadyExists && !planPermissions[i].Role.Equal(statePermissions[i].Role) {
			permissionID := statePermissions[i].PermissionId.ValueString()
			tflog.Debug(ctx, fmt.Sprintf("Permission with ID %s found on file %s, but roles differ. Updating...", permissionID, fileId))
			_, err := gsmdrive.UpdatePermission(fileId, permissionID, fieldsPermission, useDomAccess, false, planPermissions[i].toRequest())
			if err != nil {
				diags.AddError("Client Error", fmt.Sprintf("Unable to update permission on file, got error: %s", err))
				return
			}
		} else if !permissionAlreadyExists {
			tflog.Debug(ctx, fmt.Sprintf("No permission found on file %s for %s. Creating...", fileId, i))
			_, err := gsmdrive.CreatePermission(fileId, planPermissions[i].EmailMessage.ValueString(), fieldsPermission, useDomAccess, planPermissions[i].SendNotificationEmail.ValueBool(), planPermissions[i].TransferOwnership.ValueBool(), planPermissions[i].MoveToNewOwnersRoot.ValueBool(), planPermissions[i].toRequest())
			if err != nil {
				diags.AddError("Client Error", fmt.Sprintf("Unable to create permission on file, got error: %s", err))
				return
			}
		}
	}
	for i := range statePermissions {
		_, permissionStillPlanned := statePermissions[i]
		if !permissionStillPlanned {
			tflog.Debug(ctx, fmt.Sprintf("Permission found on file %s for %s, but no longer in plan. Deleting...", fileId, i))
			_, err := gsmdrive.DeletePermission(fileId, statePermissions[i].PermissionId.ValueString(), useDomAccess)
			if err != nil {
				diags.AddError("Client Error", fmt.Sprintf("Unable to delete permission from file, got error: %s", err))
				return
			}
		}
	}
	return diags
}
