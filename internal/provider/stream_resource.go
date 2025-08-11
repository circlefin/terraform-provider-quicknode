// Copyright 2025 Circle Internet Group, Inc.  All rights reserved.
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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/circlefin/terraform-provider-quicknode/api/streams"
	"github.com/circlefin/terraform-provider-quicknode/internal/utils"
	"github.com/circlefin/terraform-provider-quicknode/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &StreamResource{}
	_ resource.ResourceWithImportState = &StreamResource{}
)

var (
	networkValidator             = validators.NetworkValidator
	datasetValidator             = validators.DatasetValidator
	metadataValidator            = validators.MetadataValidator
	destinationValidator         = validators.DestinationValidator
	statusValidator              = validators.StatusValidator
	regionValidator              = validators.RegionValidator
	compressionValidator         = validators.CompressionValidator
	fileCompressionValidator     = validators.FileCompressionValidator
	fileTypeValidator            = validators.FileTypeValidator
	sslmodeValidator             = validators.SslmodeValidator
	emailValidator               = validators.EmailValidator
	startRangeValidator          = validators.StartRangeValidator
	endRangeValidator            = validators.EndRangeValidator
	datasetBatchSizeValidator    = validators.DatasetBatchSizeValidator
	fixBlockReorgsValidator      = validators.FixBlockReorgsValidator
	keepDistanceFromTipValidator = validators.KeepDistanceFromTipValidator
	maxRetryValidator            = validators.MaxRetryValidator
	retryIntervalSecValidator    = validators.RetryIntervalSecValidator
	postTimeoutSecValidator      = validators.PostTimeoutSecValidator
	portValidator                = validators.PortValidator
)

// StreamResourceModel represents the Terraform state structure.
type StreamResourceModel struct {
	Id                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	Network               types.String `tfsdk:"network"`
	Dataset               types.String `tfsdk:"dataset"`
	StartRange            types.Int64  `tfsdk:"start_range"`
	EndRange              types.Int64  `tfsdk:"end_range"`
	DatasetBatchSize      types.Int64  `tfsdk:"dataset_batch_size"`
	IncludeStreamMetadata types.String `tfsdk:"include_stream_metadata"`
	Destination           types.String `tfsdk:"destination"`
	Status                types.String `tfsdk:"status"`
	ElasticBatchEnabled   types.Bool   `tfsdk:"elastic_batch_enabled"`
	Region                types.String `tfsdk:"region"`
	FixBlockReorgs        types.Int64  `tfsdk:"fix_block_reorgs"`
	KeepDistanceFromTip   types.Int64  `tfsdk:"keep_distance_from_tip"`
	NotificationEmail     types.String `tfsdk:"notification_email"`
	DestinationAttributes types.Object `tfsdk:"destination_attributes"`
	FilterFunction        types.String `tfsdk:"filter_function"`
}

func NewStreamResource() resource.Resource {
	return &StreamResource{}
}

type StreamResource struct {
	client streams.ClientWithResponsesInterface
}

func (r *StreamResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	qnd, ok := req.ProviderData.(QuickNodeData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected ProviderData type",
			fmt.Sprintf("Expected QuickNodeData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = qnd.StreamsClient
}

func (r *StreamResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stream"
}

func (r *StreamResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Stream resource for QuickNode Streams API",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},

			"name": schema.StringAttribute{
				Required: true,
			},

			"network": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					networkValidator,
				},
			},

			"dataset": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					datasetValidator,
				},
			},

			"start_range": schema.Int64Attribute{
				Required: true,
				Validators: []validator.Int64{
					startRangeValidator,
				},
			},

			"end_range": schema.Int64Attribute{
				Optional: true,
				Validators: []validator.Int64{
					endRangeValidator,
				},
			},

			"dataset_batch_size": schema.Int64Attribute{
				Required: true,
				Validators: []validator.Int64{
					datasetBatchSizeValidator,
				},
			},

			"include_stream_metadata": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					metadataValidator,
				},
			},

			"destination": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					destinationValidator,
				},
			},

			"status": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					statusValidator,
				},
			},

			"elastic_batch_enabled": schema.BoolAttribute{
				Required: true,
			},

			"region": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					regionValidator,
				},
			},

			"fix_block_reorgs": schema.Int64Attribute{
				Optional: true,
				Validators: []validator.Int64{
					fixBlockReorgsValidator,
				},
			},

			"keep_distance_from_tip": schema.Int64Attribute{
				Optional: true,
				Validators: []validator.Int64{
					keepDistanceFromTipValidator,
				},
			},

			"notification_email": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					emailValidator,
				},
			},

			"filter_function": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "JavaScript function to filter and modify stream data. Must be base64 encoded.",
			},

			"destination_attributes": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"url": schema.StringAttribute{
						Optional: true,
					},

					"compression": schema.StringAttribute{
						Optional: true,
						Validators: []validator.String{
							compressionValidator,
						},
					},

					"headers": schema.MapAttribute{
						Optional:    true,
						ElementType: types.StringType,
					},

					"max_retry": schema.Int64Attribute{
						Required: true,
						Validators: []validator.Int64{
							maxRetryValidator,
						},
					},

					"retry_interval_sec": schema.Int64Attribute{
						Required: true,
						Validators: []validator.Int64{
							retryIntervalSecValidator,
						},
					},

					"post_timeout_sec": schema.Int64Attribute{
						Optional: true,
						Validators: []validator.Int64{
							postTimeoutSecValidator,
						},
					},

					"security_token": schema.StringAttribute{
						Optional: true,
						Computed: true,
					},

					"version": schema.StringAttribute{
						Optional: true,
						Computed: true,
					},

					"access_key": schema.StringAttribute{
						Optional:  true,
						Sensitive: true,
					},

					"secret_key": schema.StringAttribute{
						Optional:  true,
						Sensitive: true,
					},

					"bucket": schema.StringAttribute{
						Optional: true,
					},

					"region": schema.StringAttribute{
						Optional: true,
					},

					"endpoint": schema.StringAttribute{
						Optional: true,
					},

					"object_prefix": schema.StringAttribute{
						Optional: true,
					},

					"use_ssl": schema.BoolAttribute{
						Optional: true,
					},

					"username": schema.StringAttribute{
						Optional: true,
					},

					"password": schema.StringAttribute{
						Optional:  true,
						Sensitive: true,
					},

					"host": schema.StringAttribute{
						Optional: true,
					},

					"port": schema.Int64Attribute{
						Optional: true,
						Validators: []validator.Int64{
							portValidator,
						},
					},

					"database": schema.StringAttribute{
						Optional: true,
					},

					"table_name": schema.StringAttribute{
						Optional: true,
					},

					"file_compression": schema.StringAttribute{
						Optional: true,
						Validators: []validator.String{
							fileCompressionValidator,
						},
					},

					"file_type": schema.StringAttribute{
						Optional: true,
						Validators: []validator.String{
							fileTypeValidator,
						},
					},

					"sslmode": schema.StringAttribute{
						Optional: true,
						Validators: []validator.String{
							sslmodeValidator,
						},
					},
				},
			},
		},
	}
}

// getWebhookAttributes extracts webhook attributes from the destination_attributes map
func getWebhookAttributes(destAttrs map[string]interface{}) (*streams.WebhookAttributes, error) {
	url, ok := destAttrs["url"].(string)
	if !ok {
		return nil, fmt.Errorf("url must be a string")
	}
	compression, ok := destAttrs["compression"].(string)
	if !ok {
		return nil, fmt.Errorf("compression must be a string")
	}
	headers, ok := destAttrs["headers"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("headers must be a map")
	}
	maxRetry, ok := destAttrs["max_retry"].(int64)
	if !ok {
		return nil, fmt.Errorf("max_retry must be an integer")
	}
	postTimeoutSec, ok := destAttrs["post_timeout_sec"].(int64)
	if !ok {
		return nil, fmt.Errorf("post_timeout_sec must be an integer")
	}
	retryIntervalSec, ok := destAttrs["retry_interval_sec"].(int64)
	if !ok {
		return nil, fmt.Errorf("retry_interval_sec must be an integer")
	}

	return &streams.WebhookAttributes{
		Url:              url,
		Compression:      compression,
		Headers:          headers,
		MaxRetry:         float32(maxRetry),
		PostTimeoutSec:   float32(postTimeoutSec),
		RetryIntervalSec: float32(retryIntervalSec),
		SecurityToken:    "",
	}, nil
}

// getS3Attributes extracts S3 attributes from the destination_attributes map
func getS3Attributes(destAttrs map[string]interface{}) (*streams.S3Attributes, error) {
	endpoint, ok := destAttrs["endpoint"].(string)
	if !ok {
		return nil, fmt.Errorf("endpoint must be a string")
	}
	accessKey, ok := destAttrs["access_key"].(string)
	if !ok {
		return nil, fmt.Errorf("access_key must be a string")
	}
	secretKey, ok := destAttrs["secret_key"].(string)
	if !ok {
		return nil, fmt.Errorf("secret_key must be a string")
	}
	bucket, ok := destAttrs["bucket"].(string)
	if !ok {
		return nil, fmt.Errorf("bucket must be a string")
	}
	objectPrefix, ok := destAttrs["object_prefix"].(string)
	if !ok {
		return nil, fmt.Errorf("object_prefix must be a string")
	}
	fileCompression, ok := destAttrs["file_compression"].(string)
	if !ok {
		return nil, fmt.Errorf("file_compression must be a string")
	}
	fileType, ok := destAttrs["file_type"].(string)
	if !ok {
		return nil, fmt.Errorf("file_type must be a string")
	}
	maxRetry, ok := destAttrs["max_retry"].(int64)
	if !ok {
		return nil, fmt.Errorf("max_retry must be an integer")
	}
	retryIntervalSec, ok := destAttrs["retry_interval_sec"].(int64)
	if !ok {
		return nil, fmt.Errorf("retry_interval_sec must be an integer")
	}
	useSsl, ok := destAttrs["use_ssl"].(bool)
	if !ok {
		return nil, fmt.Errorf("use_ssl must be a boolean")
	}

	return &streams.S3Attributes{
		Endpoint:         endpoint,
		AccessKey:        accessKey,
		SecretKey:        secretKey,
		Bucket:           bucket,
		ObjectPrefix:     objectPrefix,
		FileCompression:  fileCompression,
		FileType:         streams.S3AttributesFileType(fileType),
		MaxRetry:         float32(maxRetry),
		RetryIntervalSec: float32(retryIntervalSec),
		UseSsl:           useSsl,
	}, nil
}

// getPostgresAttributes extracts Postgres attributes from the destination_attributes map
func getPostgresAttributes(destAttrs map[string]interface{}) (*streams.PostgresAttributes, error) {
	username, ok := destAttrs["username"].(string)
	if !ok {
		return nil, fmt.Errorf("username must be a string")
	}
	password, ok := destAttrs["password"].(string)
	if !ok {
		return nil, fmt.Errorf("password must be a string")
	}
	host, ok := destAttrs["host"].(string)
	if !ok {
		return nil, fmt.Errorf("host must be a string")
	}
	port, ok := destAttrs["port"].(int64)
	if !ok {
		return nil, fmt.Errorf("port must be an integer")
	}
	database, ok := destAttrs["database"].(string)
	if !ok {
		return nil, fmt.Errorf("database must be a string")
	}
	accessKey, ok := destAttrs["access_key"].(string)
	if !ok {
		return nil, fmt.Errorf("access_key must be a string")
	}
	sslmode, ok := destAttrs["sslmode"].(string)
	if !ok {
		return nil, fmt.Errorf("sslmode must be a string")
	}
	tableName, ok := destAttrs["table_name"].(string)
	if !ok {
		return nil, fmt.Errorf("table_name must be a string")
	}
	maxRetry, ok := destAttrs["max_retry"].(int64)
	if !ok {
		return nil, fmt.Errorf("max_retry must be an integer")
	}
	retryIntervalSec, ok := destAttrs["retry_interval_sec"].(int64)
	if !ok {
		return nil, fmt.Errorf("retry_interval_sec must be an integer")
	}

	return &streams.PostgresAttributes{
		Username:         username,
		Password:         password,
		Host:             host,
		Port:             float32(port),
		Database:         database,
		AccessKey:        accessKey,
		Sslmode:          streams.PostgresAttributesSslmode(sslmode),
		TableName:        tableName,
		MaxRetry:         float32(maxRetry),
		RetryIntervalSec: float32(retryIntervalSec),
	}, nil
}

// readStreamFromAPI reads stream data from the API and updates the provided StreamResourceModel
func (r *StreamResource) readStreamFromAPI(ctx context.Context, streamID string) (*StreamResourceModel, error) {
	readResp, err := r.client.FindOneWithResponse(ctx, streamID)
	if err != nil {
		return nil, fmt.Errorf("error reading stream: %w", err)
	}

	if readResp.StatusCode() == 404 {
		return nil, fmt.Errorf("stream not found")
	}

	if readResp.StatusCode() != 200 {
		return nil, fmt.Errorf("API returned status code %d", readResp.StatusCode())
	}

	var result map[string]interface{}
	if err := json.Unmarshal(readResp.Body, &result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	// Create a new model and populate it with API data
	data := &StreamResourceModel{}

	// Update data based on API response
	if id, ok := result["id"].(string); ok {
		data.Id = types.StringValue(id)
	}
	if name, ok := result["name"].(string); ok {
		data.Name = types.StringValue(name)
	}
	if network, ok := result["network"].(string); ok {
		data.Network = types.StringValue(network)
	}
	if dataset, ok := result["dataset"].(string); ok {
		data.Dataset = types.StringValue(dataset)
	}
	if startRange, ok := result["start_range"].(float64); ok {
		data.StartRange = types.Int64Value(int64(startRange))
	}
	if endRange, ok := result["end_range"].(float64); ok {
		if endRange == -1 {
			data.EndRange = types.Int64Null()
		} else {
			data.EndRange = types.Int64Value(int64(endRange))
		}
	}
	if datasetBatchSize, ok := result["dataset_batch_size"].(float64); ok {
		data.DatasetBatchSize = types.Int64Value(int64(datasetBatchSize))
	}
	if includeStreamMetadata, ok := result["include_stream_metadata"].(string); ok {
		data.IncludeStreamMetadata = types.StringValue(includeStreamMetadata)
	}
	if destination, ok := result["destination"].(string); ok {
		data.Destination = types.StringValue(destination)
	}
	if status, ok := result["status"].(string); ok {
		data.Status = types.StringValue(status)
	}
	if elasticBatchEnabled, ok := result["elastic_batch_enabled"].(bool); ok {
		data.ElasticBatchEnabled = types.BoolValue(elasticBatchEnabled)
	}
	if region, ok := result["region"].(string); ok {
		data.Region = types.StringValue(region)
	}
	if filterFunction, ok := result["filter_function"].(string); ok {
		data.FilterFunction = types.StringValue(filterFunction)
	}
	if fixBlockReorgs, ok := result["fix_block_reorgs"].(float64); ok {
		data.FixBlockReorgs = types.Int64Value(int64(fixBlockReorgs))
	}
	if keepDistanceFromTip, ok := result["keep_distance_from_tip"].(float64); ok {
		data.KeepDistanceFromTip = types.Int64Value(int64(keepDistanceFromTip))
	}
	if notificationEmail, ok := result["notification_email"].(string); ok {
		data.NotificationEmail = types.StringValue(notificationEmail)
	}

	// Update destination_attributes
	if destAttrs, ok := result["destination_attributes"].(map[string]interface{}); ok {
		obj, err := r.updateDestinationAttributesFromAPI(destAttrs)
		if err != nil {
			return nil, fmt.Errorf("error updating destination_attributes: %w", err)
		}
		data.DestinationAttributes = obj
	}

	return data, nil
}

func (r *StreamResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data StreamResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Prepare data for API
	datasetBatchSize := float32(data.DatasetBatchSize.ValueInt64())
	startRange := float32(data.StartRange.ValueInt64())
	startRangePtr := &startRange

	var endRange *float32
	if !data.EndRange.IsNull() {
		val := float32(data.EndRange.ValueInt64())
		endRange = &val
	}

	var fixBlockReorgs *float32
	if !data.FixBlockReorgs.IsNull() {
		val := float32(data.FixBlockReorgs.ValueInt64())
		fixBlockReorgs = &val
	}

	var keepDistanceFromTip *float32
	if !data.KeepDistanceFromTip.IsNull() {
		val := float32(data.KeepDistanceFromTip.ValueInt64())
		keepDistanceFromTip = &val
	}

	var filterFunction string
	if !data.FilterFunction.IsNull() {
		filterFunction = data.FilterFunction.ValueString()
	} else {
		filterFunction = ""
	}

	var notificationEmail *string
	if !data.NotificationEmail.IsNull() {
		val := data.NotificationEmail.ValueString()
		notificationEmail = &val
	}

	// Convert destination_attributes to appropriate type based on destination
	destAttrs, err := r.convertDestinationAttributes(data.DestinationAttributes)
	if err != nil {
		resp.Diagnostics.AddError("Error converting destination_attributes", err.Error())
		return
	}

	// Create appropriate destination_attributes union type based on destination
	var destAttrsUnion streams.CreateStreamDto_DestinationAttributes

	switch data.Destination.ValueString() {
	case "webhook":
		webhookAttrs, err := getWebhookAttributes(destAttrs)
		if err != nil {
			resp.Diagnostics.AddError("Error converting destination_attributes", err.Error())
			return
		}
		if err := destAttrsUnion.FromWebhookAttributes(*webhookAttrs); err != nil {
			resp.Diagnostics.AddError("Error creating webhook destination_attributes", err.Error())
			return
		}

	case "s3":
		s3Attrs, err := getS3Attributes(destAttrs)
		if err != nil {
			resp.Diagnostics.AddError("Error converting destination_attributes", err.Error())
			return
		}
		if err := destAttrsUnion.FromS3Attributes(*s3Attrs); err != nil {
			resp.Diagnostics.AddError("Error creating S3 destination_attributes", err.Error())
			return
		}

	case "postgres":
		postgresAttrs, err := getPostgresAttributes(destAttrs)
		if err != nil {
			resp.Diagnostics.AddError("Error converting destination_attributes", err.Error())
			return
		}
		if err := destAttrsUnion.FromPostgresAttributes(*postgresAttrs); err != nil {
			resp.Diagnostics.AddError("Error creating Postgres destination_attributes", err.Error())
			return
		}

	default:
		resp.Diagnostics.AddError("Unsupported destination type", fmt.Sprintf("Destination type '%s' is not supported", data.Destination.ValueString()))
		return
	}

	createResp, err := r.client.CreateWithResponse(ctx, streams.CreateJSONRequestBody{
		Name:                  data.Name.ValueString(),
		Network:               streams.CreateStreamDtoNetwork(data.Network.ValueString()),
		Dataset:               streams.CreateStreamDtoDataset(data.Dataset.ValueString()),
		StartRange:            startRangePtr,
		DatasetBatchSize:      datasetBatchSize,
		IncludeStreamMetadata: streams.CreateStreamDtoIncludeStreamMetadata(data.IncludeStreamMetadata.ValueString()),
		Destination:           streams.CreateStreamDtoDestination(data.Destination.ValueString()),
		ElasticBatchEnabled:   data.ElasticBatchEnabled.ValueBool(),
		Status:                streams.CreateStreamDtoStatus(data.Status.ValueString()),
		FilterFunction:        filterFunction,
		DestinationAttributes: destAttrsUnion,
		Region:                streams.CreateStreamDtoRegion(data.Region.ValueString()),
		NotificationEmail:     notificationEmail,
		EndRange:              endRange,
		FixBlockReorgs:        fixBlockReorgs,
		KeepDistanceFromTip:   keepDistanceFromTip,
	})

	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("%s - Creating Stream", utils.ClientErrorSummary),
			utils.BuildClientErrorMessage(err),
		)
		return
	}

	if createResp.StatusCode() != 201 {
		tflog.Error(ctx, "Stream creation failed", map[string]interface{}{
			"status_code":   createResp.StatusCode(),
			"response_body": string(createResp.Body),
			"request_data": map[string]interface{}{
				"name":        data.Name.ValueString(),
				"network":     data.Network.ValueString(),
				"dataset":     data.Dataset.ValueString(),
				"destination": data.Destination.ValueString(),
				"region":      data.Region.ValueString(),
			},
		})

		m, err := utils.BuildRequestErrorMessage(createResp.Status(), createResp.Body)
		if err != nil {
			resp.Diagnostics.AddWarning(fmt.Sprintf("%s - Creating Stream", utils.InternalErrorSummary), utils.BuildInternalErrorMessage(err))
		}

		resp.Diagnostics.AddError(
			fmt.Sprintf("%s - Creating Stream", utils.RequestErrorSummary),
			m,
		)
		return
	}

	// Parse response and set ID
	var response map[string]interface{}
	if err := json.Unmarshal(createResp.Body, &response); err != nil {
		resp.Diagnostics.AddError("Error parsing response", fmt.Sprintf("Could not parse response from API: %v", err))
		return
	}

	if id, ok := response["id"].(string); ok {
		data.Id = types.StringValue(id)
	} else {
		resp.Diagnostics.AddError("Error reading ID", "Could not read ID from API response")
		return
	}

	// Read full stream data from API to get computed fields
	fullStreamData, err := r.readStreamFromAPI(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading stream", err.Error())
		return
	}

	// Update data with computed fields from API
	data.Name = fullStreamData.Name
	data.Network = fullStreamData.Network
	data.Dataset = fullStreamData.Dataset
	if data.StartRange.IsNull() {
		data.StartRange = fullStreamData.StartRange
	}
	data.EndRange = fullStreamData.EndRange
	data.DatasetBatchSize = fullStreamData.DatasetBatchSize
	data.IncludeStreamMetadata = fullStreamData.IncludeStreamMetadata
	data.Destination = fullStreamData.Destination
	data.Status = fullStreamData.Status
	data.ElasticBatchEnabled = fullStreamData.ElasticBatchEnabled
	data.Region = fullStreamData.Region
	data.FilterFunction = fullStreamData.FilterFunction
	data.DestinationAttributes = fullStreamData.DestinationAttributes

	tflog.Trace(ctx, "created a resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *StreamResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data StreamResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.RemoveWithResponse(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("%s - Deleting Stream", utils.ClientErrorSummary),
			utils.BuildClientErrorMessage(err),
		)
		return
	}

	if res.StatusCode() != 200 {
		m, err := utils.BuildRequestErrorMessage(res.Status(), res.Body)
		if err != nil {
			resp.Diagnostics.AddWarning(fmt.Sprintf("%s - Deleting Stream", utils.InternalErrorSummary), utils.BuildInternalErrorMessage(err))
		}

		resp.Diagnostics.AddError(
			fmt.Sprintf("%s - Deleting Stream", utils.RequestErrorSummary),
			m,
		)
		return
	}
}

func (r *StreamResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data StreamResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read stream data from API
	streamData, err := r.readStreamFromAPI(ctx, data.Id.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "stream not found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			fmt.Sprintf("%s - Reading Stream", utils.ClientErrorSummary),
			utils.BuildClientErrorMessage(err),
		)
		return
	}

	// Update state with data from API
	data.Name = streamData.Name
	data.Network = streamData.Network
	data.Dataset = streamData.Dataset
	data.StartRange = streamData.StartRange
	data.EndRange = streamData.EndRange
	data.DatasetBatchSize = streamData.DatasetBatchSize
	data.IncludeStreamMetadata = streamData.IncludeStreamMetadata
	data.Destination = streamData.Destination
	data.Status = streamData.Status
	data.ElasticBatchEnabled = streamData.ElasticBatchEnabled
	data.Region = streamData.Region
	data.FilterFunction = streamData.FilterFunction
	data.FixBlockReorgs = streamData.FixBlockReorgs
	data.KeepDistanceFromTip = streamData.KeepDistanceFromTip
	data.NotificationEmail = streamData.NotificationEmail
	data.DestinationAttributes = streamData.DestinationAttributes

	resp.State.Set(ctx, &data)
}

func (r *StreamResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan StreamResourceModel
	var state StreamResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Determine stream ID - prefer plan.Id if available, otherwise use state.Id
	var streamId string
	if !plan.Id.IsNull() && !plan.Id.IsUnknown() {
		streamId = plan.Id.ValueString()
	} else if !state.Id.IsNull() && !state.Id.IsUnknown() {
		streamId = state.Id.ValueString()
	} else {
		resp.Diagnostics.AddError("Invalid state", "Stream ID is null or unknown in both plan and state")
		return
	}

	tflog.Info(ctx, "Starting stream update", map[string]interface{}{
		"stream_id": streamId,
		"name":      plan.Name.ValueString(),
	})

	// Check current stream status
	streamData, err := r.readStreamFromAPI(ctx, streamId)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("%s - Reading Stream Status", utils.ClientErrorSummary),
			utils.BuildClientErrorMessage(err),
		)
		return
	}

	currentStatus := streamData.Status.ValueString()

	tflog.Info(ctx, "Current stream status", map[string]interface{}{
		"stream_id": streamId,
		"status":    currentStatus,
	})

	// If stream is active, pause it before update
	var wasActive bool
	if currentStatus == "active" {
		wasActive = true
		tflog.Info(ctx, "Pausing active stream before update", map[string]interface{}{
			"stream_id": streamId,
		})

		pauseResp, err := r.client.PauseStreamWithResponse(ctx, streamId)
		if err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("%s - Pausing Stream", utils.ClientErrorSummary),
				utils.BuildClientErrorMessage(err),
			)
			return
		}

		if pauseResp.StatusCode() != 200 && pauseResp.StatusCode() != 201 {
			m, err := utils.BuildRequestErrorMessage(pauseResp.Status(), pauseResp.Body)
			if err != nil {
				resp.Diagnostics.AddWarning(fmt.Sprintf("%s - Pausing Stream", utils.InternalErrorSummary), utils.BuildInternalErrorMessage(err))
			}

			resp.Diagnostics.AddError(
				fmt.Sprintf("%s - Pausing Stream", utils.RequestErrorSummary),
				m,
			)
			return
		}

		tflog.Info(ctx, "Stream paused successfully", map[string]interface{}{
			"stream_id": streamId,
		})
	}

	// Prepare required fields as pointers
	name := plan.Name.ValueString()
	startRange := float32(plan.StartRange.ValueInt64())
	datasetBatchSize := float32(plan.DatasetBatchSize.ValueInt64())
	elasticBatchEnabled := plan.ElasticBatchEnabled.ValueBool()
	includeStreamMetadata := streams.UpdateStreamDtoIncludeStreamMetadata(plan.IncludeStreamMetadata.ValueString())
	destination := streams.UpdateStreamDtoDestination(plan.Destination.ValueString())
	status := streams.UpdateStreamDtoStatus(plan.Status.ValueString())

	// Handle optional filter_function
	var filterFunction *string
	if !plan.FilterFunction.IsNull() {
		val := plan.FilterFunction.ValueString()
		filterFunction = &val
	}

	// Handle optional fields as pointers
	var endRange *float32
	if !plan.EndRange.IsNull() {
		val := float32(plan.EndRange.ValueInt64())
		endRange = &val
	}

	var fixBlockReorgs *float32
	if !plan.FixBlockReorgs.IsNull() {
		val := float32(plan.FixBlockReorgs.ValueInt64())
		fixBlockReorgs = &val
	}

	var keepDistanceFromTip *float32
	if !plan.KeepDistanceFromTip.IsNull() {
		val := float32(plan.KeepDistanceFromTip.ValueInt64())
		keepDistanceFromTip = &val
	}

	var notificationEmail *string
	if !plan.NotificationEmail.IsNull() {
		val := plan.NotificationEmail.ValueString()
		notificationEmail = &val
	}

	// Handle destination_attributes (optional)
	var destAttrsUnion *streams.UpdateStreamDto_DestinationAttributes
	if !plan.DestinationAttributes.IsNull() {
		destAttrs, err := r.convertDestinationAttributes(plan.DestinationAttributes)
		if err != nil {
			resp.Diagnostics.AddError("Error converting destination_attributes", err.Error())
			return
		}

		// Create appropriate destination_attributes union type based on destination
		var union streams.UpdateStreamDto_DestinationAttributes

		switch plan.Destination.ValueString() {
		case "webhook":
			webhookAttrs, err := getWebhookAttributes(destAttrs)
			if err != nil {
				resp.Diagnostics.AddError("Error converting destination_attributes", err.Error())
				return
			}
			if err := union.FromWebhookAttributes(*webhookAttrs); err != nil {
				resp.Diagnostics.AddError("Error creating webhook destination_attributes", err.Error())
				return
			}

		case "s3":
			s3Attrs, err := getS3Attributes(destAttrs)
			if err != nil {
				resp.Diagnostics.AddError("Error converting destination_attributes", err.Error())
				return
			}
			if err := union.FromS3Attributes(*s3Attrs); err != nil {
				resp.Diagnostics.AddError("Error creating S3 destination_attributes", err.Error())
				return
			}

		case "postgres":
			postgresAttrs, err := getPostgresAttributes(destAttrs)
			if err != nil {
				resp.Diagnostics.AddError("Error converting destination_attributes", err.Error())
				return
			}
			if err := union.FromPostgresAttributes(*postgresAttrs); err != nil {
				resp.Diagnostics.AddError("Error creating Postgres destination_attributes", err.Error())
				return
			}

		default:
			resp.Diagnostics.AddError("Unsupported destination type", fmt.Sprintf("Destination type '%s' is not supported", plan.Destination.ValueString()))
			return
		}

		destAttrsUnion = &union
	}

	// Execute stream update
	tflog.Info(ctx, "Updating stream configuration", map[string]interface{}{
		"stream_id": streamId,
		"name":      plan.Name.ValueString(),
	})

	updateResp, err := r.client.UpdateWithResponse(ctx, streamId, streams.UpdateJSONRequestBody{
		Name:                  &name,
		StartRange:            &startRange,
		EndRange:              endRange,
		DatasetBatchSize:      &datasetBatchSize,
		IncludeStreamMetadata: &includeStreamMetadata,
		Destination:           &destination,
		ElasticBatchEnabled:   &elasticBatchEnabled,
		Status:                &status,
		FilterFunction:        filterFunction,
		FixBlockReorgs:        fixBlockReorgs,
		KeepDistanceFromTip:   keepDistanceFromTip,
		NotificationEmail:     notificationEmail,
		DestinationAttributes: destAttrsUnion,
	})

	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("%s - Updating Stream", utils.ClientErrorSummary),
			utils.BuildClientErrorMessage(err),
		)
		return
	}

	if updateResp.StatusCode() != 200 {
		tflog.Error(ctx, "Stream update failed", map[string]interface{}{
			"status_code":   updateResp.StatusCode(),
			"response_body": string(updateResp.Body),
			"request_data": map[string]interface{}{
				"id":          streamId,
				"name":        plan.Name.ValueString(),
				"start_range": plan.StartRange.ValueInt64(),
				"destination": plan.Destination.ValueString(),
				"status":      plan.Status.ValueString(),
			},
		})

		m, err := utils.BuildRequestErrorMessage(updateResp.Status(), updateResp.Body)
		if err != nil {
			resp.Diagnostics.AddWarning(fmt.Sprintf("%s - Updating Stream", utils.InternalErrorSummary), utils.BuildInternalErrorMessage(err))
		}

		resp.Diagnostics.AddError(
			fmt.Sprintf("%s - Updating Stream", utils.RequestErrorSummary),
			m,
		)
		return
	}

	tflog.Info(ctx, "Stream updated successfully", map[string]interface{}{
		"stream_id": streamId,
	})

	// If stream was active before update, reactivate it
	if wasActive {
		tflog.Info(ctx, "Reactivating stream after update", map[string]interface{}{
			"stream_id": streamId,
		})

		activateResp, err := r.client.ActivateStreamWithResponse(ctx, streamId)
		if err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("%s - Activating Stream", utils.ClientErrorSummary),
				utils.BuildClientErrorMessage(err),
			)
			return
		}

		if activateResp.StatusCode() != 200 && activateResp.StatusCode() != 201 {
			m, err := utils.BuildRequestErrorMessage(activateResp.Status(), activateResp.Body)
			if err != nil {
				resp.Diagnostics.AddWarning(fmt.Sprintf("%s - Activating Stream", utils.InternalErrorSummary), utils.BuildInternalErrorMessage(err))
			}

			resp.Diagnostics.AddError(
				fmt.Sprintf("%s - Activating Stream", utils.RequestErrorSummary),
				m,
			)
			return
		}

		tflog.Info(ctx, "Stream reactivated successfully", map[string]interface{}{
			"stream_id": streamId,
		})
	}

	// Read full stream data from API to get computed fields
	fullStreamData, err := r.readStreamFromAPI(ctx, streamId)
	if err != nil {
		resp.Diagnostics.AddError("Error reading stream after update", err.Error())
		return
	}

	// Update plan with computed fields from API
	plan.Id = fullStreamData.Id
	plan.Name = fullStreamData.Name
	plan.Network = fullStreamData.Network
	plan.Dataset = fullStreamData.Dataset
	plan.StartRange = fullStreamData.StartRange
	plan.EndRange = fullStreamData.EndRange
	plan.DatasetBatchSize = fullStreamData.DatasetBatchSize
	plan.IncludeStreamMetadata = fullStreamData.IncludeStreamMetadata
	plan.Destination = fullStreamData.Destination
	plan.Status = fullStreamData.Status
	plan.ElasticBatchEnabled = fullStreamData.ElasticBatchEnabled
	plan.Region = fullStreamData.Region
	plan.FilterFunction = fullStreamData.FilterFunction
	plan.FixBlockReorgs = fullStreamData.FixBlockReorgs
	plan.KeepDistanceFromTip = fullStreamData.KeepDistanceFromTip
	plan.NotificationEmail = fullStreamData.NotificationEmail
	plan.DestinationAttributes = fullStreamData.DestinationAttributes

	// Save updated state
	resp.State.Set(ctx, &plan)
}

func (r *StreamResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// updateDestinationAttributesFromAPI converts destination_attributes from API to Terraform format.
func (r *StreamResource) updateDestinationAttributesFromAPI(destAttrs map[string]interface{}) (types.Object, error) {
	attrs := make(map[string]attr.Value)

	// Initialize all required fields with null values
	attrs["url"] = types.StringNull()
	attrs["compression"] = types.StringNull()
	attrs["headers"] = types.MapNull(types.StringType)
	attrs["max_retry"] = types.Int64Null()
	attrs["retry_interval_sec"] = types.Int64Null()
	attrs["post_timeout_sec"] = types.Int64Null()
	attrs["security_token"] = types.StringNull()
	attrs["version"] = types.StringNull()
	attrs["access_key"] = types.StringNull()
	attrs["secret_key"] = types.StringNull()
	attrs["bucket"] = types.StringNull()
	attrs["region"] = types.StringNull()
	attrs["endpoint"] = types.StringNull()
	attrs["object_prefix"] = types.StringNull()
	attrs["use_ssl"] = types.BoolNull()
	attrs["file_compression"] = types.StringNull()
	attrs["file_type"] = types.StringNull()
	attrs["username"] = types.StringNull()
	attrs["password"] = types.StringNull()
	attrs["host"] = types.StringNull()
	attrs["port"] = types.Int64Null()
	attrs["database"] = types.StringNull()
	attrs["table_name"] = types.StringNull()
	attrs["sslmode"] = types.StringNull()

	// Update with actual values from API
	for k, v := range destAttrs {
		switch val := v.(type) {
		case string:
			// Treat empty strings as null for optional fields that are not relevant for this destination type
			if val == "" && (k == "access_key" || k == "secret_key" || k == "bucket" || k == "region" || k == "file_compression" || k == "sslmode") {
				attrs[k] = types.StringNull()
			} else {
				attrs[k] = types.StringValue(val)
			}
		case float64:
			attrs[k] = types.Int64Value(int64(val))
		case bool:
			attrs[k] = types.BoolValue(val)
		case map[string]interface{}:
			// Handling headers as a map
			headerMap := make(map[string]attr.Value)
			for headerKey, headerVal := range val {
				if headerStr, ok := headerVal.(string); ok {
					headerMap[headerKey] = types.StringValue(headerStr)
				}
			}
			headersType := types.MapType{ElemType: types.StringType}
			headersMap, diags := types.MapValue(headersType.ElemType, headerMap)
			if diags.HasError() {
				return types.Object{}, fmt.Errorf("error creating headers map: %v", diags)
			}
			attrs[k] = headersMap
		}
	}

	objType := map[string]attr.Type{
		"url":                types.StringType,
		"compression":        types.StringType,
		"headers":            types.MapType{ElemType: types.StringType},
		"max_retry":          types.Int64Type,
		"retry_interval_sec": types.Int64Type,
		"post_timeout_sec":   types.Int64Type,
		"security_token":     types.StringType,
		"version":            types.StringType,
		"access_key":         types.StringType,
		"secret_key":         types.StringType,
		"bucket":             types.StringType,
		"region":             types.StringType,
		"endpoint":           types.StringType,
		"object_prefix":      types.StringType,
		"use_ssl":            types.BoolType,
		"file_compression":   types.StringType,
		"file_type":          types.StringType,
		"username":           types.StringType,
		"password":           types.StringType,
		"host":               types.StringType,
		"port":               types.Int64Type,
		"database":           types.StringType,
		"table_name":         types.StringType,
		"sslmode":            types.StringType,
	}

	obj, diags := types.ObjectValue(objType, attrs)
	if diags.HasError() {
		return types.Object{}, fmt.Errorf("error creating destination_attributes object: %v", diags)
	}

	return obj, nil
}

// convertDestinationAttributes converts destination_attributes from Terraform to API format.
func (r *StreamResource) convertDestinationAttributes(attrs types.Object) (map[string]interface{}, error) {
	destAttrs := make(map[string]interface{})
	attributes := attrs.Attributes()

	for k, v := range attributes {
		switch val := v.(type) {
		case types.String:
			destAttrs[k] = val.ValueString()
		case types.Int64:
			destAttrs[k] = val.ValueInt64()
		case types.Bool:
			destAttrs[k] = val.ValueBool()
		case types.Map:
			headers := make(map[string]interface{})
			elements := val.Elements()
			for key, value := range elements {
				if strVal, ok := value.(types.String); ok {
					headers[key] = strVal.ValueString()
				}
			}
			destAttrs[k] = headers
		default:
			return nil, fmt.Errorf("unsupported attribute type %s: %T", k, val)
		}
	}
	return destAttrs, nil
}
