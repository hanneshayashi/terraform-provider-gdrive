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
	"google.golang.org/api/drive/v3"
)

func (restrictionsModel *restrictionsModel) toDriveRestrictions() *drive.DriveRestrictions {
	restrictions := &drive.DriveRestrictions{}
	if !restrictionsModel.AdminManagedRestrictions.IsNull() {
		restrictions.AdminManagedRestrictions = restrictionsModel.AdminManagedRestrictions.ValueBool()
		if !restrictions.AdminManagedRestrictions {
			restrictions.ForceSendFields = append(restrictions.ForceSendFields, "AdminManagedRestrictions")
		}
	}
	if !restrictionsModel.CopyRequiresWriterPermission.IsNull() {
		restrictions.CopyRequiresWriterPermission = restrictionsModel.CopyRequiresWriterPermission.ValueBool()
		if !restrictions.CopyRequiresWriterPermission {
			restrictions.ForceSendFields = append(restrictions.ForceSendFields, "CopyRequiresWriterPermission")
		}
	}
	if !restrictionsModel.DomainUsersOnly.IsNull() {
		restrictions.DomainUsersOnly = restrictionsModel.DomainUsersOnly.ValueBool()
		if !restrictions.DomainUsersOnly {
			restrictions.ForceSendFields = append(restrictions.ForceSendFields, "DomainUsersOnly")
		}
	}
	if !restrictionsModel.DriveMembersOnly.IsNull() {
		restrictions.DriveMembersOnly = restrictionsModel.DriveMembersOnly.ValueBool()
		if !restrictions.DriveMembersOnly {
			restrictions.ForceSendFields = append(restrictions.ForceSendFields, "DriveMembersOnly")
		}
	}
	return restrictions
}
