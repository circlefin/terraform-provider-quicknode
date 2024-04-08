package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const (
	// providerConfig is a shared configuration to combine with the actual
	// test configuration so the HashiCups client is properly configured.
	// It is also possible to use the QUICKNODE_ environment variables instead,
	// such as updating the Makefile and running the testing through that tool.
	providerConfig = `
provider "quicknode" {
}
`
)

var (
	// testAccProtoV6ProviderFactories are used to instantiate a provider during
	// acceptance testing. The factory function will be invoked for every Terraform
	// CLI command executed to create a provider server to which the CLI can
	// reattach.
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"quicknode": providerserver.NewProtocol6WithError(New("test")()),
	}
)

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("QUICKNODE_APIKEY"); v == "" {
		t.Fatal("QUICKNODE_APIKEY must be set for acceptance tests")
	}
}
