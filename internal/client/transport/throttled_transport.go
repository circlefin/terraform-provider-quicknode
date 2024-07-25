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

package transport

import (
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
	"golang.org/x/time/rate"
)

var _ http.RoundTripper = &ThrottledTransport{}

type ThrottledTransport struct {
	roundTripper http.RoundTripper
	ratelimiter  *rate.Limiter
}

func (c *ThrottledTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	err := c.ratelimiter.Wait(r.Context())
	if err != nil {
		return nil, err
	}
	return c.roundTripper.RoundTrip(r)
}

func NewThrottledTransport(rt http.RoundTripper, rl *rate.Limiter) http.RoundTripper {
	return &ThrottledTransport{
		roundTripper: rt,
		ratelimiter:  rl,
	}
}

func NewRetryableThrottledClient(tokens int) *http.Client {
	limiter := rate.NewLimiter(rate.Limit(tokens), tokens)
	retryableclient := retryablehttp.NewClient()

	// Ensure that retries also respect the rate limit.
	retryableclient.PrepareRetry = func(req *http.Request) error {
		return limiter.Wait(req.Context())
	}

	client := retryableclient.StandardClient()

	transport := NewThrottledTransport(client.Transport, limiter)
	client.Transport = transport

	return client
}
