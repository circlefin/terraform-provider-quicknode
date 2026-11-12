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
	"net/http"
	"testing"

	"github.com/circlefin/terraform-provider-quicknode/api/streams"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// findOneStubClient embeds the full ClientWithResponsesInterface so it
// satisfies the type without having to hand-roll every method. Only the
// FindOne call is exercised by these tests; any other call panics via the nil
// embedded interface.
type findOneStubClient struct {
	streams.ClientWithResponsesInterface

	body   string
	status int
}

func (s *findOneStubClient) FindOneWithResponse(_ context.Context, _ string, _ ...streams.RequestEditorFn) (*streams.FindOneResponse, error) {
	status := s.status
	if status == 0 {
		status = http.StatusOK
	}
	return &streams.FindOneResponse{
		HTTPResponse: &http.Response{StatusCode: status, Status: http.StatusText(status)},
		Body:         []byte(s.body),
	}, nil
}

func TestReadStreamFromAPI_EmptyFilterFunction(t *testing.T) {
	cases := []struct {
		name      string
		fallback  types.String
		wantNull  bool
		wantValue string
	}{
		{
			name:      "unset attribute stays null when the API reports an empty filter",
			fallback:  types.StringNull(),
			wantNull:  true,
			wantValue: "",
		},
		{
			name:      "explicitly configured empty filter is preserved",
			fallback:  types.StringValue(""),
			wantNull:  false,
			wantValue: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stub := &findOneStubClient{body: `{"id":"stream-1","filter_function":""}`}
			r := &StreamResource{client: stub}

			data, err := r.readStreamFromAPI(context.Background(), "stream-1", &StreamResourceModel{
				FilterFunction: tc.fallback,
			})
			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			if data.FilterFunction.IsNull() != tc.wantNull {
				t.Fatalf("expected null=%v, got %v", tc.wantNull, data.FilterFunction)
			}
			if !tc.wantNull && data.FilterFunction.ValueString() != tc.wantValue {
				t.Errorf("expected filter_function %q, got %q", tc.wantValue, data.FilterFunction.ValueString())
			}
		})
	}
}

func TestReadStreamFromAPI_ZeroIntegerAttributes(t *testing.T) {
	cases := []struct {
		name     string
		fallback types.Int64
		wantNull bool
	}{
		{
			name:     "unset attribute stays null when the API reports zero",
			fallback: types.Int64Null(),
			wantNull: true,
		},
		{
			name:     "explicitly configured zero is preserved",
			fallback: types.Int64Value(0),
			wantNull: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stub := &findOneStubClient{body: `{"id":"stream-1","fix_block_reorgs":0,"keep_distance_from_tip":0}`}
			r := &StreamResource{client: stub}

			data, err := r.readStreamFromAPI(context.Background(), "stream-1", &StreamResourceModel{
				FixBlockReorgs:      tc.fallback,
				KeepDistanceFromTip: tc.fallback,
			})
			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			if data.FixBlockReorgs.IsNull() != tc.wantNull {
				t.Errorf("expected fix_block_reorgs null=%v, got %v", tc.wantNull, data.FixBlockReorgs)
			}
			if data.KeepDistanceFromTip.IsNull() != tc.wantNull {
				t.Errorf("expected keep_distance_from_tip null=%v, got %v", tc.wantNull, data.KeepDistanceFromTip)
			}
		})
	}
}

func TestReadStreamFromAPI_OmittedFieldKeepsConfiguredValue(t *testing.T) {
	stub := &findOneStubClient{body: `{"id":"stream-1"}`}
	r := &StreamResource{client: stub}

	data, err := r.readStreamFromAPI(context.Background(), "stream-1", &StreamResourceModel{
		FilterFunction:      types.StringValue("block.number > 10"),
		FixBlockReorgs:      types.Int64Value(0),
		KeepDistanceFromTip: types.Int64Value(0),
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if data.FilterFunction.ValueString() != "block.number > 10" {
		t.Errorf("expected the configured filter_function to be preserved, got %q", data.FilterFunction.ValueString())
	}
	if data.FixBlockReorgs.ValueInt64() != 0 || data.FixBlockReorgs.IsNull() {
		t.Errorf("expected the configured fix_block_reorgs to be preserved, got %v", data.FixBlockReorgs)
	}
	if data.KeepDistanceFromTip.ValueInt64() != 0 || data.KeepDistanceFromTip.IsNull() {
		t.Errorf("expected the configured keep_distance_from_tip to be preserved, got %v", data.KeepDistanceFromTip)
	}
}

func TestEndpointLabel(t *testing.T) {
	empty := ""
	nonEmpty := "my endpoint"

	cases := []struct {
		name  string
		prior types.String
		api   *string
		want  types.String
	}{
		{
			name:  "unset label stays null when the API reports an empty label",
			prior: types.StringNull(),
			api:   &empty,
			want:  types.StringNull(),
		},
		{
			name:  "unset label stays null when the API reports no label",
			prior: types.StringNull(),
			api:   nil,
			want:  types.StringNull(),
		},
		{
			name:  "explicitly configured empty label is preserved",
			prior: types.StringValue(""),
			api:   &empty,
			want:  types.StringValue(""),
		},
		{
			name:  "explicitly configured empty label is preserved when the API reports no label",
			prior: types.StringValue(""),
			api:   nil,
			want:  types.StringValue(""),
		},
		{
			name:  "the API label wins when it is not empty",
			prior: types.StringValue(""),
			api:   &nonEmpty,
			want:  types.StringValue(nonEmpty),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := endpointLabel(tc.prior, tc.api)
			if !got.Equal(tc.want) {
				t.Errorf("expected %v, got %v", tc.want, got)
			}
		})
	}
}
