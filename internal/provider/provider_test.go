package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"gitea": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccPreCheck validates that the required environment is set up for acceptance tests.
func testAccPreCheck(t *testing.T) {
	// Check if Gitea is configured
	if v := os.Getenv("GITEA_HOSTNAME"); v == "" {
		t.Fatal("GITEA_HOSTNAME must be set for acceptance tests. Expected: http://localhost:3000")
	}
	if v := os.Getenv("GITEA_USERNAME"); v == "" {
		t.Fatal("GITEA_USERNAME must be set for acceptance tests. Expected: root")
	}
	if v := os.Getenv("GITEA_PASSWORD"); v == "" {
		t.Fatal("GITEA_PASSWORD must be set for acceptance tests. Expected: admin1234")
	}
}

// providerConfig returns a basic provider configuration for testing.
func providerConfig() string {
	return `
provider "gitea" {
  gitea_username = "root"
  gitea_password = "admin1234"
  gitea_hostname = "http://localhost:3000"
}
`
}
