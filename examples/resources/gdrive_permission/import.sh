# The ID for this resource is a combined ID that consistent of the file_id and the permission_id.
# If you file's file_id is "abcdef" and your permission_id is "12345", the ID of the resource would be:
# "abcdef/12345"
# In addition, the use_domain_admin_access attribute must be specified during the import.
# Example: false,abcdef/12345
terraform import gdrive_permission.permission [use_domain_admin_access],[file_id]/[permission_id]
