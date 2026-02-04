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

	"github.com/circlefin/terraform-provider-quicknode/api/quicknode"
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
