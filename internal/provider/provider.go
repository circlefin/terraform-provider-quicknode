// Copyright 2024 Circle Internet Group, Inc.  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/circlefin/terraform-provider-quicknode/api/quicknode"
	"github.com/circlefin/terraform-provider-quicknode/internal/client/transport"
	"github.com/circlefin/terraform-provider-quicknode/internal/utils"

	"github.com/deepmap/oapi-codegen/pkg/securityprovider"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	quicknodeEndpointDefault          = "https://api.quicknode.com"
	quicknodeRequestsPerSecondDefault = 5
)

// Ensure ScaffoldingProvider satisfies various provider interfaces.
var _ provider.Provider = &QuickNodeProvider{}
var _ provider.ProviderWithFunctions = &QuickNodeProvider{}

// QuickNodeData is provided in the DataSourceData and ResourceData to be made accessible by data and resources.
type QuickNodeData struct {
	Client quicknode.ClientWithResponsesInterface
	Chains []quicknode.Chain
}

// QuickNodeProvider defines the provider implementation.
type QuickNodeProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// QuickNodeProviderModel describes the provider data model.
type QuickNodeProviderModel struct {
	Endpoint          types.String `tfsdk:"endpoint"`
	ApiKey            types.String `tfsdk:"apikey"`
	RequestsPerSecond types.Int64  `tfsdk:"requests_per_second"`
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
			"requests_per_second": schema.Int64Attribute{
				MarkdownDescription: "Maximum requests per second to limit requests to quicknode api",
				Optional:            true,
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

	endpoint := quicknodeEndpointDefault
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

	requestsPerSecond := quicknodeRequestsPerSecondDefault
	if !data.RequestsPerSecond.IsNull() {
		requestsPerSecond = int(data.RequestsPerSecond.ValueInt64())
	}

	if resp.Diagnostics.HasError() {
		return
	}

	bearerTokenProvider, _ := securityprovider.NewSecurityProviderBearerToken(apiKey)
	client, _ := quicknode.NewClientWithResponses(
		endpoint,
		quicknode.WithHTTPClient(transport.NewRetryableThrottledClient(requestsPerSecond)),
		quicknode.WithRequestEditorFn(bearerTokenProvider.Intercept),
	)

	chainsResponse, err := client.GetV0ChainsWithResponse(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("%s - configuring provider", utils.ClientErrorSummary),
			utils.BuildClientErrorMessage(err),
		)

		return
	}

	if chainsResponse.StatusCode() != 200 {
		m, err := utils.BuildRequestErrorMessage(chainsResponse.Status(), chainsResponse.Body)
		if err != nil {
			resp.Diagnostics.AddWarning(fmt.Sprintf("%s - configuring provider", utils.InternalErrorSummary), utils.BuildInternalErrorMessage(err))
		}

		resp.Diagnostics.AddError(
			fmt.Sprintf("%s - configuring provider", utils.RequestErrorSummary),
			m,
		)

		return
	}

	chains := chainsResponse.JSON200.Data

	qnd := QuickNodeData{
		Client: client,
		Chains: chains,
	}

	resp.DataSourceData = qnd
	resp.ResourceData = qnd
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
