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
	"fmt"
	"os"
	"testing"

	"github.com/circlefin/terraform-provider-quicknode/api/streams"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/securityprovider"
)

func TestAccMinimalQuicknodeStreamResource(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccQuickNodeStreamResource(rName, "webhook"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("quicknode_stream.main", "id"),
					resource.TestCheckResourceAttr("quicknode_stream.main", "name", fmt.Sprintf("test-stream-%s", rName)),
					resource.TestCheckResourceAttr("quicknode_stream.main", "network", "ethereum-sepolia"),
					resource.TestCheckResourceAttr("quicknode_stream.main", "dataset", "block_with_receipts_debug_trace"),
					resource.TestCheckResourceAttr("quicknode_stream.main", "destination", "webhook"),
					resource.TestCheckResourceAttr("quicknode_stream.main", "status", "paused"),
					resource.TestCheckResourceAttr("quicknode_stream.main", "region", "usa_east"),
					resource.TestCheckResourceAttr("quicknode_stream.main", "elastic_batch_enabled", "true"),
					resource.TestCheckResourceAttr("quicknode_stream.main", "dataset_batch_size", "1"),
					resource.TestCheckResourceAttr("quicknode_stream.main", "include_stream_metadata", "body"),
					resource.TestCheckResourceAttr("quicknode_stream.main", "start_range", "59274680"),
					resource.TestCheckResourceAttr("quicknode_stream.main", "destination_attributes.max_retry", "3"),
					resource.TestCheckResourceAttr("quicknode_stream.main", "destination_attributes.retry_interval_sec", "1"),
					resource.TestCheckResourceAttr("quicknode_stream.main", "destination_attributes.compression", "none"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "quicknode_stream.main",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccQuickNodeStreamResourceUpdated(rName, "webhook"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("quicknode_stream.main", "name", fmt.Sprintf("test-stream-updated-%s", rName)),
					resource.TestCheckResourceAttr("quicknode_stream.main", "status", "active"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
		CheckDestroy: func(s *terraform.State) error {
			apiKey := os.Getenv("QUICKNODE_APIKEY")
			bearerTokenProvider, _ := securityprovider.NewSecurityProviderBearerToken(apiKey)
			client, _ := streams.NewClientWithResponses("https://api.quicknode.com", streams.WithRequestEditorFn(bearerTokenProvider.Intercept))

			for _, rs := range s.RootModule().Resources {
				if rs.Type != "quicknode_stream" {
					continue
				}

				resp, err := client.FindOneWithResponse(t.Context(), rs.Primary.ID)
				if err != nil {
					if resp != nil && resp.StatusCode() == 404 {
						// Stream not found - this is what we want
						continue
					}
					// Other errors might indicate the stream still exists
					return fmt.Errorf("Error checking if resource %s exists: %v", rs.Primary.ID, err)
				}

				if resp.StatusCode() == 200 {
					return fmt.Errorf("Resource %s still exists", rs.Primary.ID)
				}
			}

			return nil
		},
	})
}





func testAccQuickNodeStreamResource(name string, destination string) string {
	return providerConfig + fmt.Sprintf(`
resource "quicknode_stream" "main" {
	name                    = "test-stream-%s"
	network                 = "ethereum-sepolia"
	dataset                 = "block_with_receipts_debug_trace"
	start_range             = 59274680
	dataset_batch_size      = 1
	include_stream_metadata = "body"
	destination             = "%s"
	status                  = "paused"
	elastic_batch_enabled   = true
	region                  = "usa_east"

	destination_attributes = {
		url                = "https://webhook.site/your-unique-url"
		compression        = "none"
		headers            = {
			"Content-Type" = "application/json"
		}
		max_retry          = 3
		retry_interval_sec = 1
		post_timeout_sec   = 30
	}
}`, name, destination)
}

func testAccQuickNodeStreamResourceUpdated(name string, destination string) string {
	return providerConfig + fmt.Sprintf(`
resource "quicknode_stream" "main" {
	name                    = "test-stream-updated-%s"
	network                 = "ethereum-sepolia"
	dataset                 = "block_with_receipts_debug_trace"
	start_range             = 59274680
	dataset_batch_size      = 1
	include_stream_metadata = "body"
	destination             = "%s"
	status                  = "active"
	elastic_batch_enabled   = true
	region                  = "usa_east"

	destination_attributes = {
		url                = "https://webhook.site/your-unique-url"
		compression        = "none"
		headers            = {
			"Content-Type" = "application/json"
		}
		max_retry          = 3
		retry_interval_sec = 1
		post_timeout_sec   = 30
	}
}`, name, destination)
}
