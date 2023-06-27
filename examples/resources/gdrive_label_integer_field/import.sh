# The ID for this resource is a combined ID that consistent of the label_id and the field_id.
# If you file's fileId is "abcdef" and your field_id is "12345", the ID of the resource would be:
# "abcdef/12345"
# In addition, the use_admin_access attribute must be specified during the import.
# Example: true,abcdef/12345
terraform import gdrive_label_integer_field.field [use_admin_access],[label_id]/[field_id]
