package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PanfactumProvider struct {
}

type PanfactumProviderModel struct {
	Environment  types.String `tfsdk:"environment"`
	Region       types.String `tfsdk:"region"`
	RootModule   types.String `tfsdk:"root_module"`
	StackVersion types.String `tfsdk:"stack_version"`
	StackCommit  types.String `tfsdk:"stack_commit"`
	IsLocal      types.Bool   `tfsdk:"is_local"`
	ExtraTags    types.Map    `tfsdk:"extra_tags"`
}

func New() provider.Provider {
	return &PanfactumProvider{}
}

func (p *PanfactumProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "pf"
}

func (p *PanfactumProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"environment": schema.StringAttribute{
				Required:            true,
				Description:         "The name of the environment that you are currently deploying infrastructure to",
				MarkdownDescription: "The name of the environment that you are currently deploying infrastructure to",
			},
			"region": schema.StringAttribute{
				Required:            true,
				Description:         "The name of the region that you are currently deploying infrastructure to",
				MarkdownDescription: "The name of the region that you are currently deploying infrastructure to",
			},
			"root_module": schema.StringAttribute{
				Required:            true,
				Description:         "The name of the root / top-level module that you are currently deploying infrastructure with",
				MarkdownDescription: "The name of the root / top-level module that you are currently deploying infrastructure with",
			},
			"stack_version": schema.StringAttribute{
				Required:            true,
				Description:         "The version of the Panfactum Stack that you are currently using",
				MarkdownDescription: "The version of the Panfactum Stack that you are currently using",
			},
			"stack_commit": schema.StringAttribute{
				Required:            true,
				Description:         "The commit hash of the Panfactum Stack that you are currently using",
				MarkdownDescription: "The commit hash of the Panfactum Stack that you are currently using",
			},
			"is_local": schema.BoolAttribute{
				Optional:            true,
				Description:         "Whether the provider is being used a part of a local development deployment",
				MarkdownDescription: "Whether the provider is being used a part of a local development deployment",
			},
			"extra_tags": schema.MapAttribute{
				Optional:            true,
				Description:         "Extra tags to apply to all resources",
				MarkdownDescription: "Extra tags to apply to all resources",
				ElementType:         types.StringType,
			},
		},
	}
}

func (p *PanfactumProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data PanfactumProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	resp.DataSourceData = &data
}

func (p *PanfactumProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
}

func (p *PanfactumProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewKubeLabelsDataSource,
		NewAWSTagsDataSource,
	}
}

func (p *PanfactumProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{
		NewLowercasedFunction,
	}
}
