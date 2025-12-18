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

package validators

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type StringOneOfValidator struct {
	values []string
}

func (v StringOneOfValidator) Description(ctx context.Context) string {
	return fmt.Sprintf("value must be one of: %v", v.values)
}

func (v StringOneOfValidator) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf("value must be one of: %v", v.values)
}

func (v StringOneOfValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()

	for _, validValue := range v.values {
		if value == validValue {
			return
		}
	}

	resp.Diagnostics.AddAttributeError(
		req.Path,
		"Invalid value",
		fmt.Sprintf("Expected value to be one of: %v, got: %s", v.values, value),
	)
}

type StringRegexpValidator struct {
	regexp  *regexp.Regexp
	message string
}

func (v StringRegexpValidator) Description(ctx context.Context) string {
	return v.message
}

func (v StringRegexpValidator) MarkdownDescription(ctx context.Context) string {
	return v.message
}

func (v StringRegexpValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()

	if !v.regexp.MatchString(value) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid value",
			v.message,
		)
	}
}

type Int64RangeValidator struct {
	min int64
	max int64
}

func (v Int64RangeValidator) Description(ctx context.Context) string {
	return fmt.Sprintf("value must be between %d and %d", v.min, v.max)
}

func (v Int64RangeValidator) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf("value must be between %d and %d", v.min, v.max)
}

func (v Int64RangeValidator) ValidateInt64(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueInt64()

	if value < v.min || value > v.max {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid value",
			fmt.Sprintf("Expected value to be between %d and %d, got: %d", v.min, v.max, value),
		)
	}
}

var (
	NetworkValidator = StringOneOfValidator{
		values: []string{
			"abstract-mainnet", "abstract-testnet", "arbitrum-mainnet", "arbitrum-sepolia", "arc-testnet",
			"avalanche-fuji", "avalanche-mainnet", "b3-mainnet", "b3-sepolia",
			"base-mainnet", "base-sepolia", "bera-mainnet", "bera-bepolia",
			"bitcoin-mainnet", "blast-mainnet", "blast-sepolia", "bnbchain-mainnet",
			"bnbchain-testnet", "celo-mainnet", "cyber-mainnet", "cyber-sepolia",
			"ethereum-holesky", "ethereum-hoodi", "ethereum-mainnet", "ethereum-sepolia",
			"fantom-mainnet", "flow-mainnet", "flow-testnet", "gnosis-mainnet",
			"gravity-alpham", "hedera-mainnet", "hedera-testnet", "hemi-mainnet",
			"hemi-testnet", "hyperliquid-mainnet", "imx-mainnet", "imx-testnet",
			"injective-mainnet", "injective-testnet", "ink-mainnet", "ink-sepolia",
			"joc-mainnet", "kaia-mainnet", "kaia-testnet", "lens-mainnet",
			"lens-testnet", "linea-mainnet", "mantle-mainnet", "mantle-sepolia",
			"monad-testnet", "morph-holesky", "morph-mainnet", "nova-mainnet",
			"omni-mainnet", "omni-omega", "optimism-mainnet", "optimism-sepolia",
			"peaq-mainnet", "plasma-testnet", "polygon-amoy", "polygon-mainnet",
			"race-mainnet", "race-testnet", "sahara-testnet", "scroll-mainnet",
			"scroll-testnet", "sei-mainnet", "sei-testnet", "solana-devnet",
			"solana-mainnet", "solana-testnet", "soneium-mainnet", "sonic-mainnet",
			"sophon-mainnet", "sophon-testnet", "story-aeneid", "story-mainnet",
			"tron-mainnet", "unichain-mainnet", "unichain-sepolia", "vana-mainnet",
			"vana-moksha", "worldchain-mainnet", "worldchain-sepolia", "xai-mainnet",
			"xai-sepolia", "xrp-mainnet", "xrp-testnet", "zerog-galileo",
			"zkevm-cardona", "zkevm-mainnet", "zksync-mainnet", "zksync-sepolia",
			"zora-mainnet",
		},
	}

	DatasetValidator = StringOneOfValidator{
		values: []string{
			"block", "block_with_receipts", "receipts", "logs", "transactions",
			"trace_blocks", "debug_traces", "block_with_receipts_debug_trace",
			"block_with_receipts_trace_block", "programs_with_logs", "ledger",
		},
	}

	MetadataValidator = StringOneOfValidator{
		values: []string{"body", "header", "none"},
	}

	DestinationValidator = StringOneOfValidator{
		values: []string{"webhook", "s3", "function", "postgres"},
	}

	StatusValidator = StringOneOfValidator{
		values: []string{"active", "paused", "terminated", "completed"},
	}

	RegionValidator = StringOneOfValidator{
		values: []string{"usa_east", "europe_central", "asia_east"},
	}

	CompressionValidator = StringOneOfValidator{
		values: []string{"none", "gzip"},
	}

	FileCompressionValidator = StringOneOfValidator{
		values: []string{"none", "gzip"},
	}

	FileTypeValidator = StringOneOfValidator{
		values: []string{".json", ".parquet"},
	}

	SslmodeValidator = StringOneOfValidator{
		values: []string{"disable", "require"},
	}

	SecurityTokenValidator = StringRegexpValidator{
		regexp:  regexp.MustCompile(`^.{32,64}$`),
		message: "security token must be between 32-64 characters",
	}

	EmailValidator = StringRegexpValidator{
		regexp:  regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`),
		message: "Invalid email format",
	}

	StartRangeValidator = Int64RangeValidator{
		min: 0,
		max: 999999999999,
	}

	EndRangeValidator = Int64RangeValidator{
		min: 0,
		max: 999999999999,
	}

	DatasetBatchSizeValidator = Int64RangeValidator{
		min: 1,
		max: 1000,
	}

	FixBlockReorgsValidator = Int64RangeValidator{
		min: 0,
		max: 1,
	}

	KeepDistanceFromTipValidator = Int64RangeValidator{
		min: 0,
		max: 10000,
	}

	MaxRetryValidator = Int64RangeValidator{
		min: 0,
		max: 100,
	}

	RetryIntervalSecValidator = Int64RangeValidator{
		min: 1,
		max: 3600,
	}

	PostTimeoutSecValidator = Int64RangeValidator{
		min: 1,
		max: 3600,
	}

	PortValidator = Int64RangeValidator{
		min: 1,
		max: 65535,
	}
)
