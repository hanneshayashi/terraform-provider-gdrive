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
	"net/http"

	"github.com/hanneshayashi/gsm/gsmdrivelabels"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &labelsDataSource{}

func newLabelsDataSource() datasource.DataSource {
	return &labelsDataSource{}
}

// labelsDataSource defines the data source implementation.
type labelsDataSource struct {
	client *http.Client
}

type gdriveLabelsDataSourceLabelModel struct {
	Id         types.String                        `tfsdk:"id"`
	LabelId    types.String                        `tfsdk:"label_id"`
	LabelType  types.String                        `tfsdk:"label_type"`
	Properties *gdriveLabelResourcePropertiesModel `tfsdk:"properties"`
	LifeCycle  *gdriveLabelLifeCycleDSModel        `tfsdk:"life_cycle"`
	Fields     []*gdriveLabelDataSourceFieldsModel `tfsdk:"fields"`
}

type gdriveLabelsDataSourceModel struct {
	Id             types.String                        `tfsdk:"id"`
	LanguageCode   types.String                        `tfsdk:"language_code"`
	MinumumRole    types.String                        `tfsdk:"minimum_role"`
	Labels         []*gdriveLabelsDataSourceLabelModel `tfsdk:"labels"`
	UseAdminAccess types.Bool                          `tfsdk:"use_admin_access"`
	PublishedOnly  types.Bool                          `tfsdk:"published_only"`
}

func (d *labelsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_labels"
}

func (d *labelsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `This resource can be used to get the fields and other metadata for all labels that match the specified restrictions.`,
		Attributes: map[string]schema.Attribute{
			"id": dsId(),
			"use_admin_access": schema.BoolAttribute{
				Optional: true,
				MarkdownDescription: `Set to true in order to use the user's admin credentials.

The server verifies that the user is an admin for the label before allowing access.`,
			},
			"published_only": schema.BoolAttribute{
				Optional: true,
				MarkdownDescription: `Whether to include only published labels in the results.

When true, only the current published label revisions are returned. Disabled labels are included.
Returned label resource names reference the published revision (labels/{id}/{revisionId}).

When false, the current label revisions are returned, which might not be published.
Returned label resource names don't reference a specific revision (labels/{id}).`,
			},
			"language_code": schema.StringAttribute{
				Optional: true,
				MarkdownDescription: `The BCP-47 language code to use for evaluating localized field labels.

When not specified, values in the default configured language are used.`,
			},
			"minimum_role": schema.StringAttribute{
				Optional: true,
				MarkdownDescription: `Specifies the level of access the user must have on the returned Labels.
The minimum role a user must have on a label.
Defaults to READER.
READER     - A reader can read the label and associated metadata applied to Drive items.
APPLIER    - An applier can write associated metadata on Drive items in which they also have write access to. Implies READER.
ORGANIZER  - An organizer can pin this label in shared drives they manage and add new appliers to the label.
EDITOR     - Editors can make any update including deleting the label which also deletes the associated Drive item metadata. Implies APPLIER.`,
			},
			"labels": schema.SetNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The labels that were found.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": dsId(),
						"label_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: `The id of this label.`,
						},
						"label_type": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: `The type of this label.`,
						},
						"life_cycle": lifecycleDS(),
						"fields":     fieldsDS(),
						"properties": labelPropertiesDS(),
					},
				},
			},
		},
	}
}

func (d *labelsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*http.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (ds *labelsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	config := &gdriveLabelsDataSourceModel{}
	resp.Diagnostics.Append(req.Config.Get(ctx, config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r, err := gsmdrivelabels.ListLabels(config.LanguageCode.ValueString(), "LABEL_VIEW_FULL", config.MinumumRole.ValueString(), "*", config.UseAdminAccess.ValueBool(), config.PublishedOnly.ValueBool(), 1)
	for l := range r {
		label := &gdriveLabelsDataSourceLabelModel{
			Id:         types.StringValue(l.Id),
			LabelId:    types.StringValue(l.Id),
			LabelType:  types.StringValue(l.LabelType),
			Fields:     fieldsToModel(l.Fields),
			LifeCycle:  &gdriveLabelLifeCycleDSModel{},
			Properties: &gdriveLabelResourcePropertiesModel{},
		}
		label.LifeCycle.populate(l.Lifecycle)
		label.Properties.populate(l.Properties)
		config.Labels = append(config.Labels, label)
	}
	e := <-err
	if e != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list labels, got error: %s", e))
		return
	}
	config.Id = types.StringValue("1")
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
