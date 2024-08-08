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

package transport_test

import (
	"net/http"
	"testing"

	"github.com/circlefin/terraform-provider-quicknode/internal/client/transport"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

type MockRoundTripper struct{}

func (rt *MockRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	return &http.Response{}, nil
}

func TestThrottledTransport(t *testing.T) {
	for _, tc := range []struct {
		name        string
		limit       int
		expectError bool
	}{
		{
			"if no requests available, expect error",
			0,
			true,
		},
		{
			"if requests available, expect no error",
			1,
			false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			transport := transport.NewThrottledTransport(&MockRoundTripper{}, rate.NewLimiter(rate.Limit(tc.limit), tc.limit))
			resp, err := transport.RoundTrip(&http.Request{})
			if tc.expectError && assert.Error(t, err) {
				assert.EqualError(t, err, "rate: Wait(n=1) exceeds limiter's burst 0")
			} else {
				assert.NotNil(t, resp)
			}

		})
	}
}
