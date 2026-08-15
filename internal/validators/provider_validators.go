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

package validators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type Int64AtLeastValidator struct {
	min int64
}

func (v Int64AtLeastValidator) Description(ctx context.Context) string {
	return fmt.Sprintf("value must be at least %d", v.min)
}

func (v Int64AtLeastValidator) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf("value must be at least %d", v.min)
}

func (v Int64AtLeastValidator) ValidateInt64(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueInt64()

	if value < v.min {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid value",
			fmt.Sprintf("Expected value to be at least %d, got: %d", v.min, value),
		)
	}
}

var RequestsPerSecondValidator = Int64AtLeastValidator{
	min: 1,
}
