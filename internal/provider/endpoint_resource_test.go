package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/circlefin/terraform-provider-quicknode/api/quicknode"
	"github.com/deepmap/oapi-codegen/pkg/securityprovider"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccMinimalQuicknodeEndpointResource(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccQuickNodeResource(rName, "created-by-terraform"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(fmt.Sprintf("quicknode_endpoint.%s", rName), "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      fmt.Sprintf("quicknode_endpoint.%s", rName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccQuickNodeResource(rName, "updated-by-terraform"),
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

				resp, err := client.GetV0EndpointsId(context.Background(), rs.Primary.ID)
				if err != nil || resp.StatusCode == 200 {
					return fmt.Errorf("Resource %s still exists", rs.Primary.ID)
				}
			}

			return nil
		},
	})
}

func testAccQuickNodeResource(name string, label string) string {
	return providerConfig + fmt.Sprintf(`
resource "quicknode_endpoint" "%s" {
	network = "mainnet"
	chain   = "eth"
	label   = "%s"
}`, name, fmt.Sprintf("%s-%s", name, label))
}
