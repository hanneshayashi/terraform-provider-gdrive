# The ID for this resource is a combined ID that consistent of the fileId and the permissionId.
# If you file's fileId is "abcdef" and your permissionId is "12345", the ID of the resource would be:
# "abcdef/12345"
terraform import gdrive_permission.permission [fileId/permisssionId]
