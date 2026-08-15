// Copyright 2026 Circle Internet Group, Inc.  All rights reserved.
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
	"testing"

	frameworkprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestsPerSecondValidation(t *testing.T) {
	ctx := context.Background()
	var schemaResponse frameworkprovider.SchemaResponse
	(&QuickNodeProvider{}).Schema(ctx, frameworkprovider.SchemaRequest{}, &schemaResponse)

	attribute, ok := schemaResponse.Schema.Attributes["requests_per_second"].(schema.Int64Attribute)
	require.True(t, ok)
	require.Len(t, attribute.Validators, 1)

	tests := []struct {
		name      string
		value     types.Int64
		wantError bool
	}{
		{name: "null", value: types.Int64Null()},
		{name: "unknown", value: types.Int64Unknown()},
		{name: "negative", value: types.Int64Value(-1), wantError: true},
		{name: "zero", value: types.Int64Value(0), wantError: true},
		{name: "one", value: types.Int64Value(1)},
		{name: "positive", value: types.Int64Value(5)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := validator.Int64Request{ConfigValue: test.value}
			var response validator.Int64Response

			attribute.Validators[0].ValidateInt64(ctx, request, &response)

			if test.wantError {
				require.Len(t, response.Diagnostics, 1)
				assert.Equal(t, "Invalid value", response.Diagnostics[0].Summary())
				assert.Contains(t, response.Diagnostics[0].Detail(), "Expected value to be at least 1")
				return
			}

			assert.Empty(t, response.Diagnostics)
		})
	}
}
