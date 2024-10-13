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
)

/**************************************************************
  Provider Definition
 **************************************************************/

var _ datasource.DataSource = &awsTagsDataSource{}

func NewAWSTagsDataSource() datasource.DataSource {
	return &awsTagsDataSource{}
}

type awsTagsDataSource struct {
	ProviderData *PanfactumProviderModel
}

type awsLabelsDataSourceModel struct {
	Module         types.String `tfsdk:"module"`
	Tags           types.Map    `tfsdk:"tags"`
	RegionOverride types.String `tfsdk:"region_override"`
}

func (d *awsTagsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_aws_tags"
}

func (d *awsTagsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Example data source",

		Attributes: map[string]schema.Attribute{
			"tags": schema.MapAttribute{
				Description:         "Tags to apply to AWS resources",
				MarkdownDescription: "Tags to apply to AWS resources",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"module": schema.StringAttribute{
				Description:         "The module within which this data source is called",
				MarkdownDescription: "The module within which this data source is called",
				Required:            true,
			},
			"region_override": schema.StringAttribute{
				Description:         "Overrides the default region tag of the provider",
				MarkdownDescription: "Overrides the default region tag of the provider",
				Optional:            true,
			},
		},
	}
}

func (d *awsTagsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *awsTagsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var data awsLabelsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	labels := map[string]attr.Value{
		"panfactum.com/environment":   sanitizeAWSTagValue(d.ProviderData.Environment),
		"panfactum.com/stack-version": sanitizeAWSTagValue(d.ProviderData.StackVersion),
		"panfactum.com/stack-commit":  sanitizeAWSTagValue(d.ProviderData.StackCommit),
		"panfactum.com/local":         types.StringValue(d.ProviderData.IsLocal.String()),
		"panfactum.com/root-module":   sanitizeAWSTagValue(d.ProviderData.RootModule),
		"panfactum.com/module":        sanitizeAWSTagValue(data.Module),
	}

	// Allow the region to be overridden
	trueRegion := d.ProviderData.Region
	if !data.RegionOverride.IsNull() && !data.RegionOverride.IsUnknown() {
		trueRegion = data.RegionOverride
	}
	labels["panfactum.com/region"] = sanitizeAWSTagValue(trueRegion)

	for key, value := range (d.ProviderData.ExtraTags).Elements() {
		strValue, ok := value.(types.String)
		if ok {
			labels[sanitizeAWSTagKey(key)] = sanitizeAWSTagValue(strValue)
		} else {
			resp.Diagnostics.AddError(
				"Invalid type found",
				fmt.Sprintf("Failed to convert value for key '%s' to string.", key),
			)
			return
		}
	}

	data.Tags, _ = types.MapValue(types.StringType, labels)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

/**************************************************************
  Utility Functions
 **************************************************************/

func sanitizeAWSTagKey(input string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9.:/_@+=-]`)
	return re.ReplaceAllString(input, ".")
}

func sanitizeAWSTagValue(input types.String) types.String {
	re := regexp.MustCompile(`[^a-zA-Z0-9.:/_@+=-]`)
	return types.StringValue(re.ReplaceAllString(input.ValueString(), "."))
}
