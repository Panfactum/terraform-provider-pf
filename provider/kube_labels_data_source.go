// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"regexp"
	"strings"
	"unicode"
)

/**************************************************************
  Provider Definition
 **************************************************************/

var _ datasource.DataSource = &kubeLabelsDataSource{}

func NewKubeLabelsDataSource() datasource.DataSource {
	return &kubeLabelsDataSource{}
}

type kubeLabelsDataSource struct {
	ProviderData *PanfactumProviderModel
}

type kubeLabelsDataSourceModel struct {
	Module types.String `tfsdk:"module"`
	Labels types.Map    `tfsdk:"labels"`
}

func (d *kubeLabelsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kube_labels"
}

func (d *kubeLabelsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Example data source",

		Attributes: map[string]schema.Attribute{
			"labels": schema.MapAttribute{
				Description:         "Labels to apply to Kubernetes resources",
				MarkdownDescription: "Labels to apply to Kubernetes resources",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"module": schema.StringAttribute{
				Description:         "The module within which this data source is called",
				MarkdownDescription: "The module within which this data source is called",
				Required:            true,
			},
		},
	}
}

func (d *kubeLabelsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	data, ok := req.ProviderData.(*PanfactumProviderModel)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected PanfactumProviderModel, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.ProviderData = data
}

func (d *kubeLabelsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var data kubeLabelsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	labels := map[string]attr.Value{
		"panfactum.com/environment":   sanitizeKubeLabelValue(d.ProviderData.Environment),
		"panfactum.com/region":        sanitizeKubeLabelValue(d.ProviderData.Region),
		"panfactum.com/stack-version": sanitizeKubeLabelValue(d.ProviderData.StackVersion),
		"panfactum.com/stack-commit":  sanitizeKubeLabelValue(d.ProviderData.StackCommit),
		"panfactum.com/local":         types.StringValue(d.ProviderData.IsLocal.String()),
		"panfactum.com/root-module":   sanitizeKubeLabelValue(d.ProviderData.RootModule),
		"panfactum.com/module":        sanitizeKubeLabelValue(data.Module),
	}

	for key, value := range (d.ProviderData.ExtraTags).Elements() {
		strValue, ok := value.(types.String)
		if ok {
			labels[sanitizeKubeLabelKey(key)] = sanitizeKubeLabelValue(strValue)
		} else {
			resp.Diagnostics.AddError(
				"Invalid type found",
				fmt.Sprintf("Failed to convert value for key '%s' to string.", key),
			)
			return
		}
	}

	data.Labels, _ = types.MapValue(types.StringType, labels)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

/**************************************************************
  Utility Functions
 **************************************************************/

// sanitizeKubeLabelValue performs the required sanitization steps:
// 1. Replaces any non-alphanumeric, '.', '_', or '-' characters with '.'
// 2. Ensures the string starts and ends with an alphanumeric character
func sanitizeKubeLabelValue(input types.String) types.String {
	// Replace any non-alphanumeric, '.', '_', or '-' characters with '.'
	re := regexp.MustCompile(`[^a-zA-Z0-9._-]`)
	sanitized := re.ReplaceAllString(input.ValueString(), ".")

	// Trim any leading or trailing non-alphanumeric characters
	sanitized = strings.TrimFunc(sanitized, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})

	return types.StringValue(sanitized)
}

// sanitizeKubeLabelKey performs the required sanitization steps:
// 1. Replaces any non-alphanumeric, '.', '_', '-', or '/' characters with '.'
// 2. Ensures the string starts and ends with an alphanumeric character
func sanitizeKubeLabelKey(input string) string {
	// Replace any non-alphanumeric, '.', '_', '-', or '/' characters with '.'
	re := regexp.MustCompile(`[^a-zA-Z0-9._/-]`)
	sanitized := re.ReplaceAllString(input, ".")

	// Trim any leading or trailing non-alphanumeric characters
	sanitized = strings.TrimFunc(sanitized, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})

	return sanitized
}
