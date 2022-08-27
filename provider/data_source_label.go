/*
Copyright © 2021-2022 Hannes Hayashi

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
	"github.com/hanneshayashi/gsm/gsmdrivelabels"
	"github.com/hanneshayashi/gsm/gsmhelpers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceLabel() *schema.Resource {
	return &schema.Resource{
		Description: "Gets a Shared Drive and returns its metadata",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				Description: `Label resource name.
May be any of:
- labels/{id} (equivalent to labels/{id}@latest)
- labels/{id}@latest
- labels/{id}@published
- labels/{id}@{revisionId}`,
			},
			"use_admin_access": {
				Type:     schema.TypeBool,
				Optional: true,
				Description: `Set to true in order to use the user's admin credentials.
The server verifies that the user is an admin for the label before allowing access.`,
			},
			"language_code": {
				Type:     schema.TypeString,
				Optional: true,
				Description: `The BCP-47 language code to use for evaluating localized field labels.
When not specified, values in the default configured language are used.`,
			},
			"revision": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: ``,
			},
			"label_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: ``,
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: ``,
			},
			"title": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: ``,
			},
			"fields": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"value_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"date_options": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"date_format": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"date_format_type": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"max_value": {
										Type:     schema.TypeList,
										Required: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"day": {
													Type:     schema.TypeInt,
													Computed: true,
												},
												"month": {
													Type:     schema.TypeInt,
													Computed: true,
												},
												"year": {
													Type:     schema.TypeInt,
													Computed: true,
												},
											},
										},
									},
									"min_value": {
										Type:     schema.TypeList,
										Required: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"day": {
													Type:     schema.TypeInt,
													Computed: true,
												},
												"month": {
													Type:     schema.TypeInt,
													Computed: true,
												},
												"year": {
													Type:     schema.TypeInt,
													Computed: true,
												},
											},
										},
									},
								},
							},
						},
						"selection_options": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"list_options": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"max_entries": {
													Type:     schema.TypeInt,
													Computed: true,
												},
											},
										},
									},
									"choices": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"id": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"state": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"display_name": {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},
								},
							},
						},
						"integer_options": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"max_value": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"min_value": {
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},
						"text_options": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"min_length": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"max_length": {
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},
						"user_options": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"list_options": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"max_entries": {
													Type:     schema.TypeInt,
													Computed: true,
												},
											},
										},
									},
								},
							},
						},
						"properties": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"display_name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"required": {
										Type:     schema.TypeBool,
										Computed: true,
									},
								},
							},
						},
						"query_key": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"life_cycle": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"state": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
		},
		Read: dataSourceReadLabel,
	}
}

func dataSourceReadLabel(d *schema.ResourceData, _ any) error {
	labelID := gsmhelpers.EnsurePrefix(d.Get("name").(string), "labels/")
	revision := d.Get("revision").(string)
	if revision != "" {
		labelID = labelID + "@" + revision
	}
	label, err := gsmdrivelabels.GetLabel(labelID, d.Get("language_code").(string), "LABEL_VIEW_FULL", "*", d.Get("use_admin_access").(bool))
	if err != nil {
		return err
	}
	d.SetId(label.Id)
	d.Set("description", label.Properties.Description)
	d.Set("label_type", label.LabelType)
	d.Set("title", label.Properties.Title)
	fields := make([]map[string]any, 0)
	for i := range label.Fields {
		field := make(map[string]any)
		field["id"] = label.Fields[i].Id
		field["query_key"] = label.Fields[i].QueryKey
		field["life_cycle"] = []map[string]any{
			{
				"state": label.Fields[i].Lifecycle.State,
			},
		}
		field["properties"] = []map[string]any{
			{
				"display_name": label.Fields[i].Properties.DisplayName,
				"required":     label.Fields[i].Properties.Required,
			},
		}
		if label.Fields[i].TextOptions != nil {
			field["value_type"] = "text"
			field["text_options"] = []map[string]int64{
				{
					"max_length": label.Fields[i].TextOptions.MaxLength,
					"min_length": label.Fields[i].TextOptions.MinLength,
				},
			}
		} else if label.Fields[i].IntegerOptions != nil {
			field["value_type"] = "integer"
			field["integer_options"] = []map[string]int64{
				{
					"max_value": label.Fields[i].IntegerOptions.MaxValue,
					"min_value": label.Fields[i].IntegerOptions.MinValue,
				},
			}
		} else if label.Fields[i].UserOptions != nil {
			field["value_type"] = "user"
			if label.Fields[i].UserOptions.ListOptions != nil {
				userOption := make(map[string]any)
				if label.Fields[i].UserOptions.ListOptions != nil {
					userOption["list_options"] = []map[string]int64{
						{
							"max_entries": label.Fields[i].UserOptions.ListOptions.MaxEntries,
						},
					}
				}
				field["user_options"] = []map[string]any{
					userOption,
				}
			}
		} else if label.Fields[i].SelectionOptions != nil {
			field["value_type"] = "selection"
			if label.Fields[i].SelectionOptions.ListOptions != nil || label.Fields[i].SelectionOptions.Choices != nil {
				selectionOptions := make(map[string]any)
				if label.Fields[i].SelectionOptions.ListOptions != nil {
					selectionOptions["list_options"] = []map[string]int64{
						{
							"max_entries": label.Fields[i].SelectionOptions.ListOptions.MaxEntries,
						},
					}
				}
				if label.Fields[i].SelectionOptions.Choices != nil {
					choices := make([]map[string]string, len(label.Fields[i].SelectionOptions.Choices))
					for c := range label.Fields[i].SelectionOptions.Choices {
						choices[c] = map[string]string{
							"display_name": label.Fields[i].SelectionOptions.Choices[c].Properties.DisplayName,
							"id":           label.Fields[i].SelectionOptions.Choices[c].Id,
							"state":        label.Fields[i].SelectionOptions.Choices[c].Lifecycle.State,
						}
					}
					selectionOptions["choices"] = choices
				}
				field["selection_options"] = []map[string]any{
					selectionOptions,
				}
			}
		} else if label.Fields[i].DateOptions != nil {
			field["value_type"] = "dateString"
			dateOptions := map[string]any{
				"date_format":      label.Fields[i].DateOptions.DateFormat,
				"date_format_type": label.Fields[i].DateOptions.DateFormatType,
			}
			if label.Fields[i].DateOptions.MinValue != nil {
				dateOptions["min_value"] = []map[string]any{
					{
						"day":   label.Fields[i].DateOptions.MinValue.Day,
						"month": label.Fields[i].DateOptions.MinValue.Month,
						"year":  label.Fields[i].DateOptions.MinValue.Year,
					},
				}
			}
			if label.Fields[i].DateOptions.MaxValue != nil {
				dateOptions["max_value"] = []map[string]any{
					{
						"day":   label.Fields[i].DateOptions.MaxValue.Day,
						"month": label.Fields[i].DateOptions.MaxValue.Month,
						"year":  label.Fields[i].DateOptions.MaxValue.Year,
					},
				}
			}
			field["date_options"] = []map[string]any{
				dateOptions,
			}
		}
		fields = append(fields, field)
	}
	d.Set("fields", fields)
	return nil
}
