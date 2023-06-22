# The ID for this resource is a combined ID that consistent of the file_id and the label_id.
# If you file's file_id is "abcdef" and your label_id is "xzy", the ID of the resource would be:
# "abcdef/xzy"
terraform import gdrive_label_assignment.label_assignment [file_id]/[label_id]
