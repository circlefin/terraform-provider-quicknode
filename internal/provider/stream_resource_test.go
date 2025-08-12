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
				if err != nil || resp.StatusCode() != 404 {
					return fmt.Errorf("Resource %s still exists", rs.Primary.ID)
				}
			}

			return nil
		},
	})
}

func TestAccQuicknodeStreamResourceS3(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing for S3 destination
			{
				Config: testAccQuickNodeStreamResourceS3(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("quicknode_stream.s3", "id"),
					resource.TestCheckResourceAttr("quicknode_stream.s3", "name", fmt.Sprintf("test-s3-stream-%s", rName)),
					resource.TestCheckResourceAttr("quicknode_stream.s3", "network", "ethereum-sepolia"),
					resource.TestCheckResourceAttr("quicknode_stream.s3", "dataset", "block_with_receipts_debug_trace"),
					resource.TestCheckResourceAttr("quicknode_stream.s3", "destination", "s3"),
					resource.TestCheckResourceAttr("quicknode_stream.s3", "status", "paused"),
					resource.TestCheckResourceAttr("quicknode_stream.s3", "region", "usa_east"),
					resource.TestCheckResourceAttr("quicknode_stream.s3", "elastic_batch_enabled", "true"),
					resource.TestCheckResourceAttr("quicknode_stream.s3", "dataset_batch_size", "1"),
					resource.TestCheckResourceAttr("quicknode_stream.s3", "include_stream_metadata", "body"),
					resource.TestCheckResourceAttr("quicknode_stream.s3", "start_range", "59274680"),
					resource.TestCheckResourceAttr("quicknode_stream.s3", "destination_attributes.max_retry", "3"),
					resource.TestCheckResourceAttr("quicknode_stream.s3", "destination_attributes.retry_interval_sec", "1"),
					resource.TestCheckResourceAttr("quicknode_stream.s3", "destination_attributes.file_compression", "gzip"),
					resource.TestCheckResourceAttr("quicknode_stream.s3", "destination_attributes.file_type", ".json"),
					resource.TestCheckResourceAttr("quicknode_stream.s3", "destination_attributes.use_ssl", "true"),
				),
			},
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
				if err != nil || resp.StatusCode() != 404 {
					return fmt.Errorf("Resource %s still exists", rs.Primary.ID)
				}
			}

			return nil
		},
	})
}

func TestAccQuicknodeStreamResourcePostgres(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing for Postgres destination
			{
				Config: testAccQuickNodeStreamResourcePostgres(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("quicknode_stream.postgres", "id"),
					resource.TestCheckResourceAttr("quicknode_stream.postgres", "name", fmt.Sprintf("test-postgres-stream-%s", rName)),
					resource.TestCheckResourceAttr("quicknode_stream.postgres", "network", "ethereum-sepolia"),
					resource.TestCheckResourceAttr("quicknode_stream.postgres", "dataset", "block_with_receipts_debug_trace"),
					resource.TestCheckResourceAttr("quicknode_stream.postgres", "destination", "postgres"),
					resource.TestCheckResourceAttr("quicknode_stream.postgres", "status", "paused"),
					resource.TestCheckResourceAttr("quicknode_stream.postgres", "region", "usa_east"),
					resource.TestCheckResourceAttr("quicknode_stream.postgres", "elastic_batch_enabled", "true"),
					resource.TestCheckResourceAttr("quicknode_stream.postgres", "dataset_batch_size", "1"),
					resource.TestCheckResourceAttr("quicknode_stream.postgres", "include_stream_metadata", "body"),
					resource.TestCheckResourceAttr("quicknode_stream.postgres", "start_range", "59274680"),
					resource.TestCheckResourceAttr("quicknode_stream.postgres", "destination_attributes.max_retry", "3"),
					resource.TestCheckResourceAttr("quicknode_stream.postgres", "destination_attributes.retry_interval_sec", "1"),
					resource.TestCheckResourceAttr("quicknode_stream.postgres", "destination_attributes.port", "5432"),
					resource.TestCheckResourceAttr("quicknode_stream.postgres", "destination_attributes.sslmode", "require"),
				),
			},
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
				if err != nil || resp.StatusCode() != 404 {
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

func testAccQuickNodeStreamResourceS3(name string) string {
	return providerConfig + fmt.Sprintf(`
resource "quicknode_stream" "s3" {
	name                    = "test-s3-stream-%s"
	network                 = "ethereum-sepolia"
	dataset                 = "block_with_receipts_debug_trace"
	start_range             = 59274680
	dataset_batch_size      = 1
	include_stream_metadata = "body"
	destination             = "s3"
	status                  = "paused"
	elastic_batch_enabled   = true
	region                  = "usa_east"

	destination_attributes = {
		endpoint         = "s3.amazonaws.com"
		access_key       = "YOUR_ACCESS_KEY"
		secret_key       = "YOUR_SECRET_KEY"
		bucket           = "my-quicknode-streams"
		object_prefix    = "ethereum-sepolia/"
		file_compression = "gzip"
		file_type        = ".json"
		max_retry        = 3
		retry_interval_sec = 1
		use_ssl          = true
	}
}`, name)
}

func testAccQuickNodeStreamResourcePostgres(name string) string {
	return providerConfig + fmt.Sprintf(`
resource "quicknode_stream" "postgres" {
	name                    = "test-postgres-stream-%s"
	network                 = "ethereum-sepolia"
	dataset                 = "block_with_receipts_debug_trace"
	start_range             = 59274680
	dataset_batch_size      = 1
	include_stream_metadata = "body"
	destination             = "postgres"
	status                  = "paused"
	elastic_batch_enabled   = true
	region                  = "usa_east"

	destination_attributes = {
		username         = "postgres"
		password         = "YOUR_PASSWORD"
		host             = "your-postgres-host.com"
		port             = 5432
		database         = "quicknode_streams"
		access_key       = "YOUR_ACCESS_KEY"
		sslmode          = "require"
		table_name       = "ethereum_sepolia_streams"
		max_retry        = 3
		retry_interval_sec = 1
	}
}`, name)
}
