package provider

import (
	"google.golang.org/api/drive/v3"
)

func (fileModel *gdriveFileResourceModel) toRequest() *drive.File {
	return &drive.File{
		MimeType: fileModel.MimeType.ValueString(),
		Name:     fileModel.Name.ValueString(),
		DriveId:  fileModel.DriveId.ValueString(),
	}
}
