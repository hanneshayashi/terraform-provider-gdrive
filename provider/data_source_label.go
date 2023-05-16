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
	"github.com/hanneshayashi/gsm/gsmhelpers"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &labelDataSource{}

func newLabelDataSource() datasource.DataSource {
	return &labelDataSource{}
}

// labelDataSource defines the data source implementation.
type labelDataSource struct {
	client *http.Client
}

type gdriveLabelListOptionsModel struct {
	MaxEntries types.Int64 `tfsdk:"max_entries"`
}

type gdriveLabelChoicedModel struct {
	Id          types.String               `tfsdk:"id"`
	ChoiceId    types.String               `tfsdk:"choice_id"`
	DisplayName types.String               `tfsdk:"display_name"`
	LifeCycle   *gdriveLabelLifeCycleModel `tfsdk:"life_cycle"`
}

type gdriveLabelDateFieldModel struct {
	Day   types.Int64 `tfsdk:"day"`
	Month types.Int64 `tfsdk:"month"`
	Year  types.Int64 `tfsdk:"year"`
}

type gdriveLabelTextOptionsModel struct {
	MinLength types.Int64 `tfsdk:"min_length"`
	MaxLength types.Int64 `tfsdk:"max_length"`
}

type gdriveLabelIntegerOptionsModel struct {
	MinValue types.Int64 `tfsdk:"min_value"`
	MaxValue types.Int64 `tfsdk:"max_value"`
}

type gdriveLabelDateOptionsModel struct {
	DateFormat     types.String               `tfsdk:"date_format"`
	DateFormatType types.String               `tfsdk:"date_format_type"`
	MinValue       *gdriveLabelDateFieldModel `tfsdk:"min_value"`
	MaxValue       *gdriveLabelDateFieldModel `tfsdk:"max_value"`
}

type gdriveLabelUserOptionseModel struct {
	ListOptions *gdriveLabelListOptionsModel `tfsdk:"list_options"`
}

type gdriveLabelSelectionOptionsModel struct {
	ListOptions *gdriveLabelListOptionsModel `tfsdk:"list_options"`
	Choices     []*gdriveLabelChoicedModel   `tfsdk:"choices"`
}

type gdriveLabelLifeCycleModel struct {
	State types.String `tfsdk:"state"`
}

type gdriveLabelFieldPropertieseModel struct {
	DisplayName types.String `tfsdk:"display_name"`
	Required    types.Bool   `tfsdk:"required"`
}

type gdriveLabelDataSourceFieldsModel struct {
	Id               types.String                      `tfsdk:"id"`
	FieldId          types.String                      `tfsdk:"field_id"`
	QueryKey         types.String                      `tfsdk:"query_key"`
	ValueType        types.String                      `tfsdk:"value_type"`
	LifeCycle        *gdriveLabelLifeCycleModel        `tfsdk:"life_cycle"`
	DateOptions      *gdriveLabelDateOptionsModel      `tfsdk:"date_options"`
	SelectionOptions *gdriveLabelSelectionOptionsModel `tfsdk:"selection_options"`
	IntegerOptions   *gdriveLabelIntegerOptionsModel   `tfsdk:"integer_options"`
	TextOptions      *gdriveLabelTextOptionsModel      `tfsdk:"text_options"`
	UserOptions      *gdriveLabelUserOptionseModel     `tfsdk:"user_options"`
	Properties       *gdriveLabelFieldPropertieseModel `tfsdk:"properties"`
}

type gdriveLabelDataSourceModel struct {
	Id             types.String                        `tfsdk:"id"`
	LabelId        types.String                        `tfsdk:"label_id"`
	Name           types.String                        `tfsdk:"name"`
	LanguageCode   types.String                        `tfsdk:"language_code"`
	Revision       types.String                        `tfsdk:"revision"`
	LabelType      types.String                        `tfsdk:"label_type"`
	Description    types.String                        `tfsdk:"description"`
	Title          types.String                        `tfsdk:"title"`
	UseAdminAccess types.Bool                          `tfsdk:"use_admin_access"`
	LifeCycle      *gdriveLabelLifeCycleModel          `tfsdk:"life_cycle"`
	Fields         []*gdriveLabelDataSourceFieldsModel `tfsdk:"fields"`
}

func (d *labelDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_label"
}

func (d *labelDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `This resource can be used to get the fields and other metadata for a single label.
This resource requires additional setup:
1. Enable the Drive Labels API in your GCP project
2. Add 'https://www.googleapis.com/auth/drive.labels' as a scope to your Domain Wide Delegation config
3. Set 'use_labels_api' to 'true' in your provider configuration`,
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required: true,
				Description: `Label resource name.
May be any of:
- labels/{id} (equivalent to labels/{id}@latest)
- labels/{id}@latest
- labels/{id}@published
- labels/{id}@{revisionId}`,
			},
			"use_admin_access": schema.BoolAttribute{
				Optional: true,
				Description: `Set to true in order to use the user's admin credentials.
The server verifies that the user is an admin for the label before allowing access.
Requires setting the 'use_labels_admin_scope' property to 'true' in the provider config.`,
			},
			"language_code": schema.StringAttribute{
				Optional: true,
				Description: `The BCP-47 language code to use for evaluating localized field labels.
When not specified, values in the default configured language are used.`,
			},
			"revision": schema.StringAttribute{
				Optional: true,
				Description: `The revision of the label to retrieve.
Defaults to the latest revision.
Reading other revisions may require addtional permissions and / or setting the 'use_admin_access' flag.`,
			},
			"label_type": schema.StringAttribute{
				Computed:    true,
				Description: `The type of this label.`,
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: `The description of the label.`,
			},
			"title": schema.StringAttribute{
				Computed:    true,
				Description: `Title of the label.`,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the label",
				Computed:            true,
			},
			"label_id": schema.StringAttribute{
				MarkdownDescription: "ID of the label",
				Computed:            true,
			},
		},
		Blocks: map[string]schema.Block{
			"life_cycle": lifecycle(),
			"fields":     fieldsDS(),
		},
	}
}

func (d *labelDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (ds *labelDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	config := &gdriveLabelDataSourceModel{}
	resp.Diagnostics.Append(req.Config.Get(ctx, config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	labelID := gsmhelpers.EnsurePrefix(config.Name.ValueString(), "labels/")
	if !config.Revision.IsNull() {
		labelID += "@" + config.Revision.ValueString()
	}
	label, err := gsmdrivelabels.GetLabel(labelID, config.LanguageCode.ValueString(), "LABEL_VIEW_FULL", "*", config.UseAdminAccess.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get label, got error: %s", err))
		return
	}
	config.LabelId = types.StringValue(label.Id)
	config.Id = types.StringValue(label.Id)
	config.LabelType = types.StringValue(label.LabelType)
	config.Revision = types.StringValue(label.RevisionId)
	config.Fields = fieldsToModel(label.Fields)
	if label.Properties != nil {
		config.Title = types.StringValue(label.Properties.Title)
		config.Description = types.StringValue(label.Properties.Description)
	}
	if label.Lifecycle != nil {
		config.LifeCycle = &gdriveLabelLifeCycleModel{
			State: types.StringValue(label.Lifecycle.State),
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
