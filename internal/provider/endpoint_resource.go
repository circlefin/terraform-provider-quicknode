// Copyright 2024 Circle Internet Financial, LTD.  All rights reserved.
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
	"net/url"
	"strings"

	"github.com/circlefin/terraform-provider-quicknode/api/quicknode"
	"github.com/circlefin/terraform-provider-quicknode/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &EndpointResource{}
	_ resource.ResourceWithImportState = &EndpointResource{}
	_ resource.ResourceWithModifyPlan  = &EndpointResource{}

	securityAttributes = map[string]attr.Type{
		"tokens": basetypes.ListType{
			ElemType: basetypes.ObjectType{
				AttrTypes: tokensAttributes,
			},
		},
	}
	tokensAttributes = map[string]attr.Type{
		"id":    types.StringType,
		"token": types.StringType,
	}
)

func NewEndpointResource() resource.Resource {
	return &EndpointResource{}
}

// EndpointResource defines the resource implementation.
type EndpointResource struct {
	client quicknode.ClientWithResponsesInterface
}

// EndpointResourceModel describes the resource data model.
type EndpointResourceModel struct {
	Label    types.String `tfsdk:"label"`
	Chain    types.String `tfsdk:"chain"`
	Network  types.String `tfsdk:"network"`
	Url      types.String `tfsdk:"url"`
	Id       types.String `tfsdk:"id"`
	Security types.Object `tfsdk:"security"`
}

type EndpointResourceSecurityToken struct {
	Id    types.String
	Token types.String
}

func (r *EndpointResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_endpoint"
}

func (r *EndpointResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Endpoint resource",
		Attributes: map[string]schema.Attribute{
			"chain": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Chain to configure an endpoint for",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"network": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Network to configure an endpoint for",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"label": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Label to decorate an endpoint with",
			},
			"url": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Endpoint URL that was created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "ID of the endpoint",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"security": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"tokens": schema.ListNestedAttribute{
						Computed:            true,
						MarkdownDescription: "Tokens used to authenticate with the endpoint",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"id": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "The ID of the Security Token",
								},
								"token": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "The Security Token",
									Sensitive:           true,
								},
							},
						},
						PlanModifiers: []planmodifier.List{
							listplanmodifier.UseStateForUnknown(),
						},
					},
				},
				Computed:            true,
				MarkdownDescription: "Security Configuration of the endpoint",
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *EndpointResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// If the entire plan is null, the resource is planned for destruction and we need no validation.
	if !req.Plan.Raw.IsNull() {
		chainsResponse, err := r.client.GetV0ChainsWithResponse(ctx)
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

		var data EndpointResourceModel

		resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
		if resp.Diagnostics.HasError() {
			return
		}

		var validChainSlugs []string
		var validNetworkSlugs []string
		for _, chain := range chains {
			validChainSlugs = append(validChainSlugs, strings.ToLower(*chain.Slug))

			if strings.EqualFold(*chain.Slug, data.Chain.ValueString()) {
				for _, network := range *chain.Networks {
					validNetworkSlugs = append(validNetworkSlugs, strings.ToLower(*network.Slug))
					if strings.EqualFold(*network.Slug, data.Network.ValueString()) {
						return
					}
				}
			}
		}

		// If this is empty, then we never matched a chain slug.
		if len(validNetworkSlugs) == 0 {
			resp.Diagnostics.AddAttributeError(
				path.Root("chain"),
				"Invalid Chain",
				fmt.Sprintf("Expected chain to be one of %v, but was %s", validChainSlugs, data.Chain.ValueString()),
			)

			return
		}

		resp.Diagnostics.AddAttributeError(
			path.Root("network"),
			"Invalid Network",
			fmt.Sprintf("Expected network to be one of %v for chain %s, but was %s", validNetworkSlugs, data.Chain.ValueString(), data.Network.ValueString()),
		)
	}
}

func (r *EndpointResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(quicknode.ClientWithResponsesInterface)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected quicknode.ClientWithResponsesInterface, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *EndpointResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data EndpointResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	endpointResp, err := r.client.PostV0EndpointsWithResponse(
		ctx,
		quicknode.PostV0EndpointsJSONRequestBody{
			Chain:   data.Chain.ValueStringPointer(),
			Network: data.Network.ValueStringPointer(),
		},
	)

	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("%s - Creating Endpoint", utils.ClientErrorSummary),
			utils.BuildClientErrorMessage(err),
		)
		return
	}

	if endpointResp.StatusCode() != 200 {
		m, err := utils.BuildRequestErrorMessage(endpointResp.Status(), endpointResp.Body)
		if err != nil {
			resp.Diagnostics.AddWarning(fmt.Sprintf("%s - Creating Endpoint", utils.InternalErrorSummary), utils.BuildInternalErrorMessage(err))
		}

		resp.Diagnostics.AddError(
			fmt.Sprintf("%s - Creating Endpoint", utils.RequestErrorSummary),
			m,
		)
		return
	}

	endpoint := endpointResp.JSON200.Data
	data.Id = types.StringValue(endpoint.Id)
	u, _ := url.Parse(endpoint.HttpUrl)
	data.Url = types.StringValue(fmt.Sprintf("%s://%s", u.Scheme, u.Host))
	data.Security = types.ObjectNull(securityAttributes)
	if endpoint.Security.Tokens != nil {
		var tokens []basetypes.ObjectValuable
		for _, token := range *endpoint.Security.Tokens {
			tokenValue, diags := types.ObjectValue(tokensAttributes, map[string]attr.Value{
				"id":    types.StringValue(*token.Id),
				"token": types.StringValue(*token.Token),
			})

			resp.Diagnostics.Append(diags...)
			tokens = append(tokens, tokenValue)
		}

		tokensValueList, diags := types.ListValueFrom(ctx, basetypes.ObjectType{AttrTypes: tokensAttributes}, tokens)

		resp.Diagnostics.Append(diags...)
		securityValueObject, diags := types.ObjectValue(securityAttributes, map[string]attr.Value{
			"tokens": tokensValueList,
		})

		resp.Diagnostics.Append(diags...)
		data.Security = securityValueObject
	}

	l := data.Label.ValueString()
	if l != "" {
		endpointUpdateResp, err := r.client.PatchV0EndpointsIdWithResponse(
			ctx,
			data.Id.ValueString(),
			quicknode.PatchV0EndpointsIdJSONRequestBody{
				Label: &l,
			},
		)
		if err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("%s - Patching Endpoint Label", utils.ClientErrorSummary),
				utils.BuildClientErrorMessage(err),
			)
		}

		if endpointUpdateResp.StatusCode() != 200 {
			m, err := utils.BuildRequestErrorMessage(endpointResp.Status(), endpointResp.Body)
			if err != nil {
				resp.Diagnostics.AddWarning(fmt.Sprintf("%s - Patching Endpoint Label", utils.InternalErrorSummary), utils.BuildInternalErrorMessage(err))
			}

			resp.Diagnostics.AddError(
				fmt.Sprintf("%s - Patching Endpoint Label", utils.RequestErrorSummary),
				m,
			)
		}
	}

	tflog.Trace(ctx, "created a resource")
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EndpointResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data EndpointResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	endpointResp, err := r.client.GetV0EndpointsIdWithResponse(
		ctx,
		data.Id.ValueString(),
	)

	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("%s - Reading Endpoint", utils.ClientErrorSummary),
			utils.BuildClientErrorMessage(err),
		)
		return
	}

	if endpointResp.StatusCode() != 200 {
		m, err := utils.BuildRequestErrorMessage(endpointResp.Status(), endpointResp.Body)
		if err != nil {
			resp.Diagnostics.AddWarning(fmt.Sprintf("%s - Reading Endpoint", utils.InternalErrorSummary), utils.BuildInternalErrorMessage(err))
		}

		resp.Diagnostics.AddError(
			fmt.Sprintf("%s - Reading Endpoint", utils.RequestErrorSummary),
			m,
		)
		return
	}

	endpoint := endpointResp.JSON200.Data
	data.Chain = types.StringValue(endpoint.Chain)
	data.Network = types.StringValue(endpoint.Network)
	data.Label = types.StringNull()
	if endpoint.Label != nil && *endpoint.Label != "" {
		data.Label = types.StringPointerValue(endpoint.Label)
	}
	u, _ := url.Parse(endpoint.HttpUrl)
	data.Url = types.StringValue(fmt.Sprintf("%s://%s", u.Scheme, u.Host))
	data.Security = types.ObjectNull(securityAttributes)
	if endpoint.Security.Tokens != nil {
		var tokens []basetypes.ObjectValuable
		for _, token := range *endpoint.Security.Tokens {
			tokenValue, diags := types.ObjectValue(tokensAttributes, map[string]attr.Value{
				"id":    types.StringValue(*token.Id),
				"token": types.StringValue(*token.Token),
			})

			resp.Diagnostics.Append(diags...)
			tokens = append(tokens, tokenValue)
		}

		tokensValueList, diags := types.ListValueFrom(ctx, basetypes.ObjectType{AttrTypes: tokensAttributes}, tokens)

		resp.Diagnostics.Append(diags...)
		securityValueObject, diags := types.ObjectValue(securityAttributes, map[string]attr.Value{
			"tokens": tokensValueList,
		})

		resp.Diagnostics.Append(diags...)
		data.Security = securityValueObject
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EndpointResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data EndpointResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	l := data.Label.ValueString()

	endpointResp, err := r.client.PatchV0EndpointsIdWithResponse(
		ctx,
		data.Id.ValueString(),
		quicknode.PatchV0EndpointsIdJSONRequestBody{
			Label: &l,
		},
	)

	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("%s - Patching Endpoint", utils.ClientErrorSummary),
			utils.BuildClientErrorMessage(err),
		)
		return
	}

	if endpointResp.StatusCode() != 200 {
		m, err := utils.BuildRequestErrorMessage(endpointResp.Status(), endpointResp.Body)
		if err != nil {
			resp.Diagnostics.AddWarning(fmt.Sprintf("%s - Patching Endpoint", utils.InternalErrorSummary), utils.BuildInternalErrorMessage(err))
		}

		resp.Diagnostics.AddError(
			fmt.Sprintf("%s - Patching Endpoint", utils.RequestErrorSummary),
			m,
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EndpointResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data EndpointResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	endpointResp, err := r.client.DeleteV0EndpointsIdWithResponse(
		ctx,
		data.Id.ValueString(),
	)

	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("%s - Deleting Endpoint", utils.ClientErrorSummary),
			utils.BuildClientErrorMessage(err),
		)
		return
	}

	if endpointResp.StatusCode() != 200 {
		m, err := utils.BuildRequestErrorMessage(endpointResp.Status(), endpointResp.Body)
		if err != nil {
			resp.Diagnostics.AddWarning(fmt.Sprintf("%s - Deleting Endpoint", utils.InternalErrorSummary), utils.BuildInternalErrorMessage(err))
		}

		resp.Diagnostics.AddError(
			fmt.Sprintf("%s - Deleting Endpoint", utils.RequestErrorSummary),
			m,
		)
		return
	}
}

func (r *EndpointResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
