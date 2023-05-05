package provider

import (
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"google.golang.org/api/drive/v3"
)

func (fileModel *gdriveFileResourceModel) toRequest() *drive.File {
	return &drive.File{
		MimeType: fileModel.MimeType.ValueString(),
		Name:     fileModel.Name.ValueString(),
		DriveId:  fileModel.DriveId.ValueString(),
	}
}

func (fileModel *gdriveFileResourceModel) getContent() (content *os.File, diags diag.Diagnostics) {
	var err error
	if !fileModel.Content.IsNull() {
		content, err = os.Open(fileModel.Content.ValueString())
		if err != nil {
			diags.AddError("Client Error", fmt.Sprintf("Unable to open local file (content), got error: %s", err))
			return
		}
		defer content.Close()
	}
	return content, diags
}
