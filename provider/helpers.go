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

	"google.golang.org/api/drive/v3"
)

func contains(s string, slice []string) bool {
	for i := range slice {
		if s == slice[i] {
			return true
		}
	}
	return false
}

func getRestrictions(d *drive.Drive) (restrictions map[string]bool) {
	if d.Restrictions != nil {
		restrictions = map[string]bool{
			"admin_managed_restrictions":      d.Restrictions.AdminManagedRestrictions,
			"copy_requires_writer_permission": d.Restrictions.CopyRequiresWriterPermission,
			"domain_users_only":               d.Restrictions.DomainUsersOnly,
			"drive_members_only":              d.Restrictions.DriveMembersOnly,
		}
	}
	return
}

func combineId(a, b string) string {
	return fmt.Sprintf("%s/%s", a, b)
}

func getParent(file *drive.File) (parent string) {
	if file.Parents != nil {
		parent = file.Parents[0]
	}
	return
}

// func noDiff(_, _, _ string, _ *schema.ResourceData) bool {
// 	return true
// }
