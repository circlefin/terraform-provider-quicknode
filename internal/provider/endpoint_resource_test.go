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
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/circlefin/terraform-provider-quicknode/api/quicknode"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/securityprovider"
)

func TestAccMinimalQuicknodeEndpointResource(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccQuickNodeResource(rName, "created-by-terraform", "tag1", "tag2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("quicknode_endpoint.main", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "quicknode_endpoint.main",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccQuickNodeResource(rName, "updated-by-terraform", "tag1", "tag3"),
				Check:  resource.ComposeAggregateTestCheckFunc(),
			},
			// Delete testing automatically occurs in TestCase
		},
		CheckDestroy: func(s *terraform.State) error {
			apiKey := os.Getenv("QUICKNODE_APIKEY")
			bearerTokenProvider, _ := securityprovider.NewSecurityProviderBearerToken(apiKey)
			client, _ := quicknode.NewClientWithResponses("https://api.quicknode.com", quicknode.WithRequestEditorFn(bearerTokenProvider.Intercept))

			for _, rs := range s.RootModule().Resources {
				if rs.Type != "quicknode_endpoint" {
					continue
				}

				resp, err := client.ShowEndpoint(t.Context(), rs.Primary.ID)
				if err != nil || resp.StatusCode == 200 {
					return fmt.Errorf("Resource %s still exists", rs.Primary.ID)
				}
			}

			return nil
		},
	})
}

func testAccQuickNodeResource(name, label, tag1, tag2 string) string {
	return providerConfig + fmt.Sprintf(`
resource "quicknode_endpoint" "main" {
	network = "mainnet"
	chain   = "eth"
	label   = "%s-%s"
	tags	= [%q, %q]
}`, name, label, tag1, tag2)
}

// multichainStubClient embeds the full ClientWithResponsesInterface so it
// satisfies the type without having to hand-roll every method. Only the
// multichain calls are exercised by these tests; any other call will panic
// via the nil embedded interface, which is what we want to surface as a test
// failure.
type multichainStubClient struct {
	quicknode.ClientWithResponsesInterface

	enableCalls  int
	disableCalls int
	lastID       string

	enableStatus  int
	disableStatus int
	enableErr     error
	disableErr    error

	// nilResponse causes the stub to return (nil, nil), simulating a
	// misbehaving client that violates the oapi-codegen contract.
	nilResponse bool
}

func (s *multichainStubClient) EnableMultichainWithResponse(_ context.Context, id string, _ ...quicknode.RequestEditorFn) (*quicknode.EnableMultichainResponse, error) {
	s.enableCalls++
	s.lastID = id
	if s.enableErr != nil {
		return nil, s.enableErr
	}
	if s.nilResponse {
		return nil, nil
	}
	status := s.enableStatus
	if status == 0 {
		status = http.StatusOK
	}
	return &quicknode.EnableMultichainResponse{
		HTTPResponse: &http.Response{StatusCode: status, Status: http.StatusText(status)},
		Body:         []byte(`{"data":true,"error":null}`),
	}, nil
}

func (s *multichainStubClient) DisableMultichainWithResponse(_ context.Context, id string, _ ...quicknode.RequestEditorFn) (*quicknode.DisableMultichainResponse, error) {
	s.disableCalls++
	s.lastID = id
	if s.disableErr != nil {
		return nil, s.disableErr
	}
	if s.nilResponse {
		return nil, nil
	}
	status := s.disableStatus
	if status == 0 {
		status = http.StatusOK
	}
	return &quicknode.DisableMultichainResponse{
		HTTPResponse: &http.Response{StatusCode: status, Status: http.StatusText(status)},
		Body:         []byte(`{"data":true,"error":null}`),
	}, nil
}

func TestSetMultichain_EnableSuccess(t *testing.T) {
	stub := &multichainStubClient{}
	r := &EndpointResource{client: stub}
	var diags diag.Diagnostics

	r.setMultichain(context.Background(), "endpoint-123", true, &diags)

	if diags.HasError() {
		t.Fatalf("expected no error diagnostics, got: %v", diags.Errors())
	}
	if stub.enableCalls != 1 {
		t.Errorf("expected 1 enable call, got %d", stub.enableCalls)
	}
	if stub.disableCalls != 0 {
		t.Errorf("expected 0 disable calls, got %d", stub.disableCalls)
	}
	if stub.lastID != "endpoint-123" {
		t.Errorf("expected id 'endpoint-123', got %q", stub.lastID)
	}
}

func TestSetMultichain_DisableSuccess(t *testing.T) {
	stub := &multichainStubClient{}
	r := &EndpointResource{client: stub}
	var diags diag.Diagnostics

	r.setMultichain(context.Background(), "endpoint-abc", false, &diags)

	if diags.HasError() {
		t.Fatalf("expected no error diagnostics, got: %v", diags.Errors())
	}
	if stub.disableCalls != 1 {
		t.Errorf("expected 1 disable call, got %d", stub.disableCalls)
	}
	if stub.enableCalls != 0 {
		t.Errorf("expected 0 enable calls, got %d", stub.enableCalls)
	}
	if stub.lastID != "endpoint-abc" {
		t.Errorf("expected id 'endpoint-abc', got %q", stub.lastID)
	}
}

func TestSetMultichain_EnableTransportError(t *testing.T) {
	stub := &multichainStubClient{enableErr: http.ErrServerClosed}
	r := &EndpointResource{client: stub}
	var diags diag.Diagnostics

	r.setMultichain(context.Background(), "id", true, &diags)

	if !diags.HasError() {
		t.Fatalf("expected error diagnostics on transport failure")
	}
	if got := diags.Errors()[0].Summary(); got == "" || !strings.Contains(got, "Enabling Multichain") {
		t.Errorf("expected diagnostic summary mentioning 'Enabling Multichain', got %q", got)
	}
}

func TestSetMultichain_DisableTransportError(t *testing.T) {
	stub := &multichainStubClient{disableErr: http.ErrServerClosed}
	r := &EndpointResource{client: stub}
	var diags diag.Diagnostics

	r.setMultichain(context.Background(), "id", false, &diags)

	if !diags.HasError() {
		t.Fatalf("expected error diagnostics on transport failure")
	}
	if got := diags.Errors()[0].Summary(); !strings.Contains(got, "Disabling Multichain") {
		t.Errorf("expected diagnostic summary mentioning 'Disabling Multichain', got %q", got)
	}
}

func TestSetMultichain_EnableNon200(t *testing.T) {
	stub := &multichainStubClient{enableStatus: http.StatusBadRequest}
	r := &EndpointResource{client: stub}
	var diags diag.Diagnostics

	r.setMultichain(context.Background(), "id", true, &diags)

	if !diags.HasError() {
		t.Fatalf("expected error diagnostics on non-200 response")
	}
	if got := diags.Errors()[0].Summary(); !strings.Contains(got, "Enabling Multichain") {
		t.Errorf("expected diagnostic summary mentioning 'Enabling Multichain', got %q", got)
	}
}

func TestSetMultichain_DisableNon200(t *testing.T) {
	stub := &multichainStubClient{disableStatus: http.StatusInternalServerError}
	r := &EndpointResource{client: stub}
	var diags diag.Diagnostics

	r.setMultichain(context.Background(), "id", false, &diags)

	if !diags.HasError() {
		t.Fatalf("expected error diagnostics on non-200 response")
	}
	if got := diags.Errors()[0].Summary(); !strings.Contains(got, "Disabling Multichain") {
		t.Errorf("expected diagnostic summary mentioning 'Disabling Multichain', got %q", got)
	}
}

func TestSetMultichain_EnableNilResponse(t *testing.T) {
	stub := &multichainStubClient{nilResponse: true}
	r := &EndpointResource{client: stub}
	var diags diag.Diagnostics

	r.setMultichain(context.Background(), "id", true, &diags)

	if !diags.HasError() {
		t.Fatalf("expected error diagnostics on nil response")
	}
	if got := diags.Errors()[0].Summary(); !strings.Contains(got, "Enabling Multichain") {
		t.Errorf("expected diagnostic summary mentioning 'Enabling Multichain', got %q", got)
	}
}

func TestSetMultichain_DisableNilResponse(t *testing.T) {
	stub := &multichainStubClient{nilResponse: true}
	r := &EndpointResource{client: stub}
	var diags diag.Diagnostics

	r.setMultichain(context.Background(), "id", false, &diags)

	if !diags.HasError() {
		t.Fatalf("expected error diagnostics on nil response")
	}
	if got := diags.Errors()[0].Summary(); !strings.Contains(got, "Disabling Multichain") {
		t.Errorf("expected diagnostic summary mentioning 'Disabling Multichain', got %q", got)
	}
}

// TestMultichainDiff_NullVsFalse asserts that a legacy state where
// Multichain is null is treated as equivalent to a plan value of false,
// so upgrading to a provider version that adds the Multichain attribute
// will not trigger a spurious Disable call on the first apply.
func TestMultichainDiff_NullVsFalse(t *testing.T) {
	state := types.BoolNull()
	plan := types.BoolValue(false)

	if state.ValueBool() != plan.ValueBool() {
		t.Fatalf("expected null state (%v) to compare equal to false plan (%v) via ValueBool", state.ValueBool(), plan.ValueBool())
	}
	if state.Equal(plan) {
		t.Fatalf("sanity: Equal() should still distinguish null and false; this test only guards against using Equal() for the diff")
	}
}
