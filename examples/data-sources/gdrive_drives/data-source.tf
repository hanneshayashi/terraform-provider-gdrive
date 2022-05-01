# Shared Drives that have "test" in their name and where the user is a member
data "gdrive_drives" "drives_test" {
  query = "name contains 'test'"
}

# All Shared Drives that have no members
data "gdrive_drives" "drives_no_members" {
  query                   = "memberCount = 0"
  use_domain_admin_access = true
}
