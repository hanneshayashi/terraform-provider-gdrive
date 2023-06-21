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

type gdriveLabelChoicePropertiesModel struct {
	BadgeConfig *gdriveLabelChoiceBadgeConfigModel `tfsdk:"badge_config"`
	DisplayName types.String                       `tfsdk:"display_name"`
}

type gdriveLabelChoiceModel struct {
	LifeCycle  *gdriveLabelLifeCycleDSModel      `tfsdk:"life_cycle"`
	Properties *gdriveLabelChoicePropertiesModel `tfsdk:"properties"`
	Id         types.String                      `tfsdk:"id"`
	ChoiceId   types.String                      `tfsdk:"choice_id"`
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
	MinValue       *gdriveLabelDateFieldModel `tfsdk:"min_value"`
	MaxValue       *gdriveLabelDateFieldModel `tfsdk:"max_value"`
	DateFormat     types.String               `tfsdk:"date_format"`
	DateFormatType types.String               `tfsdk:"date_format_type"`
}

type gdriveLabelUserOptionseModel struct {
	ListOptions *gdriveLabelListOptionsModel `tfsdk:"list_options"`
}

type gdriveLabelSelectionOptionsModel struct {
	ListOptions *gdriveLabelListOptionsModel `tfsdk:"list_options"`
	Choices     []*gdriveLabelChoiceModel    `tfsdk:"choices"`
}

type gdriveLabelLifeCycleDisabledPolicyModel struct {
	HideInSearch types.Bool `tfsdk:"hide_in_search"`
	ShowInApply  types.Bool `tfsdk:"show_in_apply"`
}

type gdriveLabelLifeCycleDSModel struct {
	DisabledPolicy        *gdriveLabelLifeCycleDisabledPolicyModel `tfsdk:"disabled_policy"`
	State                 types.String                             `tfsdk:"state"`
	HasUnpublishedChanges types.Bool                               `tfsdk:"has_unpublished_changes"`
}

type gdriveLabelLifeCycleModel struct {
	DisabledPolicy *gdriveLabelLifeCycleDisabledPolicyModel `tfsdk:"disabled_policy"`
	State          types.String                             `tfsdk:"state"`
}

type gdriveLabelDataSourceFieldPropertieseModel struct {
	DisplayName types.String `tfsdk:"display_name"`
	Required    types.Bool   `tfsdk:"required"`
}

type gdriveLabelDataSourceFieldsModel struct {
	LifeCycle        *gdriveLabelLifeCycleDSModel                `tfsdk:"life_cycle"`
	DateOptions      *gdriveLabelDateOptionsModel                `tfsdk:"date_options"`
	SelectionOptions *gdriveLabelSelectionOptionsModel           `tfsdk:"selection_options"`
	IntegerOptions   *gdriveLabelIntegerOptionsModel             `tfsdk:"integer_options"`
	TextOptions      *gdriveLabelTextOptionsModel                `tfsdk:"text_options"`
	UserOptions      *gdriveLabelUserOptionseModel               `tfsdk:"user_options"`
	Properties       *gdriveLabelDataSourceFieldPropertieseModel `tfsdk:"properties"`
	Id               types.String                                `tfsdk:"id"`
	FieldId          types.String                                `tfsdk:"field_id"`
	QueryKey         types.String                                `tfsdk:"query_key"`
	ValueType        types.String                                `tfsdk:"value_type"`
}

type gdriveLabelDataSourceModel struct {
	LifeCycle      *gdriveLabelLifeCycleDSModel        `tfsdk:"life_cycle"`
	Id             types.String                        `tfsdk:"id"`
	LabelId        types.String                        `tfsdk:"label_id"`
	Name           types.String                        `tfsdk:"name"`
	LanguageCode   types.String                        `tfsdk:"language_code"`
	Revision       types.String                        `tfsdk:"revision"`
	LabelType      types.String                        `tfsdk:"label_type"`
	Properties     *gdriveLabelResourcePropertiesModel `tfsdk:"properties"`
	Fields         []*gdriveLabelDataSourceFieldsModel `tfsdk:"fields"`
	UseAdminAccess types.Bool                          `tfsdk:"use_admin_access"`
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
			"id": dsId(),
			"label_id": schema.StringAttribute{
				MarkdownDescription: "ID of the label",
				Computed:            true,
			},
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
The server verifies that the user is an admin for the label before allowing access.`,
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
		},
		Blocks: map[string]schema.Block{
			"life_cycle": lifecycleDS(),
			"fields":     fieldsDS(),
			"properties": schema.SingleNestedBlock{
				MarkdownDescription: "Basic properties of the label.",
				Attributes: map[string]schema.Attribute{
					"title": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Title of the label.",
					},
					"description": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The description of the label.",
					},
				},
			},
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
	l, err := gsmdrivelabels.GetLabel(labelID, config.LanguageCode.ValueString(), "LABEL_VIEW_FULL", "*", config.UseAdminAccess.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get label, got error: %s", err))
		return
	}
	config.LabelId = types.StringValue(l.Id)
	config.Id = types.StringValue(l.Id)
	config.LabelType = types.StringValue(l.LabelType)
	config.Revision = types.StringValue(l.RevisionId)
	config.Fields = fieldsToModel(l.Fields)
	config.LifeCycle = &gdriveLabelLifeCycleDSModel{}
	config.LifeCycle.populate(l.Lifecycle)
	if l.Properties != nil {
		config.Properties = &gdriveLabelResourcePropertiesModel{
			Title:       types.StringValue(l.Properties.Title),
			Description: types.StringValue(l.Properties.Description),
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
