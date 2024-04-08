package provider

import (
	"context"
	"os"

	"github.com/circlefin/terraform-provider-quicknode/api/quicknode"

	"github.com/deepmap/oapi-codegen/pkg/securityprovider"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure ScaffoldingProvider satisfies various provider interfaces.
var _ provider.Provider = &QuickNodeProvider{}
var _ provider.ProviderWithFunctions = &QuickNodeProvider{}

// QuickNodeProvider defines the provider implementation.
type QuickNodeProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// QuickNodeProviderModel describes the provider data model.
type QuickNodeProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	ApiKey   types.String `tfsdk:"apikey"`
}

func (p *QuickNodeProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "quicknode"
	resp.Version = p.version
}

func (p *QuickNodeProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "QuickNode API Endpoint",
				Optional:            true,
			},
			"apikey": schema.StringAttribute{
				MarkdownDescription: "QuickNode API Key",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *QuickNodeProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data QuickNodeProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := "https://api.quicknode.com"
	if !data.Endpoint.IsNull() {
		endpoint = data.Endpoint.ValueString()
	}

	apiKey := os.Getenv("QUICKNODE_APIKEY")

	if !data.ApiKey.IsNull() {
		apiKey = data.ApiKey.ValueString()
	}

	if apiKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("apikey"),
			"Missing Quicknode API Key",
			"The provider cannot create the Quicknode API client as there is a missing or empty value for the Quicknode apikey. "+
				"Set the apikey value in the configuration or use the QUICKNODE_APIKEY environment variable."+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	bearerTokenProvider, _ := securityprovider.NewSecurityProviderBearerToken(apiKey)
	client, _ := quicknode.NewClientWithResponses(
		endpoint,
		quicknode.WithHTTPClient(retryablehttp.NewClient().StandardClient()),
		quicknode.WithRequestEditorFn(bearerTokenProvider.Intercept),
	)
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *QuickNodeProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewEndpointResource,
	}
}

func (p *QuickNodeProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return nil
}

func (p *QuickNodeProvider) Functions(ctx context.Context) []func() function.Function {
	return nil
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &QuickNodeProvider{
			version: version,
		}
	}
}
