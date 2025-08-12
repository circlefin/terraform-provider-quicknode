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
	"encoding/base64"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// FilterDataSource provides an optional way to read JavaScript filter files and encode them to base64.
// This data source is optional - you can also use Terraform's built-in function directly:
//   filter_function = base64encode(file("${path.module}/filter.js"))
//
// The data source provides additional benefits like:
// - Better error handling for file operations
// - Validation of file content
// - Debugging capabilities (shows raw code in output)
// - Future extensibility for advanced features

// FilterDataSourceModel describes the data structure.
type FilterDataSourceModel struct {
	FilePath      types.String `tfsdk:"file_path"`
	FilterCode    types.String `tfsdk:"filter_code"`
	Base64Encoded types.String `tfsdk:"base64_encoded"`
}

// FilterDataSource implements datasource.DataSource.
type FilterDataSource struct{}

// Metadata returns the data source type name.
func (d *FilterDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_filter"
}

// Schema defines the schema for the data source.
func (d *FilterDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Data source for QuickNode Stream filters",
		Attributes: map[string]schema.Attribute{
			"file_path": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Path to JavaScript filter file",
			},
			"filter_code": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Raw JavaScript filter code",
			},
			"base64_encoded": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Base64 encoded filter code for QuickNode API",
			},
		},
	}
}

// Read reads the data source.
func (d *FilterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data FilterDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read file content
	fileContent, err := os.ReadFile(data.FilePath.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading filter file", fmt.Sprintf("Could not read file %s: %v", data.FilePath.ValueString(), err))
		return
	}

	// Set filter code
	data.FilterCode = types.StringValue(string(fileContent))

	// Encode to base64
	base64Encoded := base64.StdEncoding.EncodeToString(fileContent)
	data.Base64Encoded = types.StringValue(base64Encoded)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// NewFilterDataSource returns a new instance of the data source.
func NewFilterDataSource() datasource.DataSource {
	return &FilterDataSource{}
}
