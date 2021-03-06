package megaport

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/utilitywarehouse/terraform-provider-megaport/megaport/api"
)

var (
	testAccProviders map[string]terraform.ResourceProvider
	testAccProvider  *schema.Provider
)

func init() {
	testAccProvider = Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"megaport": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("Provider.InternalValidate(): %s", err)
	}
	testAccPreCheck(t)
	c := testAccProvider.Meta()
	if c == nil {
		t.Fatalf("Provider metadata is nil")
	}
	cfg, ok := c.(*Config)
	if !ok {
		t.Fatalf("Could not extract Config from Provider metadata")
	}
	if cfg.Client == nil {
		t.Fatalf("Config does not include a valid Client")
	}
	if cfg.Client.BaseURL != api.EndpointStaging {
		t.Fatalf("Unexpected Provider endpoint: %s", cfg.Client.BaseURL)
	}
	if cfg.Client.Token != os.Getenv("MEGAPORT_TOKEN") {
		t.Fatalf("Provider token does not match the environment variable MEGAPORT_TOKEN")
	}
}

func testAccPreCheck(t *testing.T) {
	if os.Getenv("MEGAPORT_TOKEN") == "" {
		t.Fatal("MEGAPORT_TOKEN must be set for acceptance tests")
	}
	if err := os.Setenv("MEGAPORT_API_ENDPOINT", api.EndpointStaging); err != nil {
		t.Fatal(err)
	}
	if err := testAccProvider.Configure(terraform.NewResourceConfigRaw(nil)); err != nil {
		t.Fatal(err)
	}
}
