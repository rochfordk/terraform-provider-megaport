package megaport

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/utilitywarehouse/terraform-provider-megaport/megaport/api"
)

func TestAccMegaportAwsVxc_basic(t *testing.T) {
	var vxcBefore api.ProductAssociatedVxc
	rName := "t" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	rId := acctest.RandStringFromCharSet(12, "012346789")
	rAsn := uint64(acctest.RandIntRange(1, 65535))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsVxcBasicConfig(rName, rId, rAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceExists("megaport_aws_vxc.test", &vxcBefore),
				),
			},
		},
	})
}

func testAccCheckResourceExists(n string, vxc *api.ProductAssociatedVxc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		cfg := testAccProvider.Meta().(*Config)
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("testAccCheckResourceExists: cannot find %q", n)
		}
		v, err := cfg.Client.GetCloudVxc(rs.Primary.ID)
		if err != nil {
			return err
		}
		*vxc = *v
		return nil
	}
}

func testAccCheckResourceDestroy(s *terraform.State) error {
	cfg := testAccProvider.Meta().(*Config)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "megaport_aws_vxc" { // TODO do we want to check all resources?
			continue
		}
		v, err := cfg.Client.GetCloudVxc(rs.Primary.ID)
		if err != nil {
			return err
		}
		if v != nil && v.ProvisioningStatus != api.ProductStatusDecommissioned {
			return fmt.Errorf("testAccCheckResourceDestroy: %s has not been destroyed", rs.Primary.ID)
		}
	}
	return nil
}

func testAccAwsVxcBasicConfig(name, accountId string, asn uint64) string {
	return fmt.Sprintf(`
data "megaport_location" "aws" {
  name_regex = "Equinix LD5"
}

data "megaport_partner_port" "aws" {
  name_regex   = "eu-west-1"
  connect_type = "AWS"
  location_id  = data.megaport_location.aws.id
}

data "megaport_location" "port" {
  name_regex = "Telehouse North"
}

resource "megaport_port" "port" {
  name        = "terraform_acctest_%s"
  location_id = data.megaport_location.port.id
  speed       = 1000
  term        = 1
}

resource "megaport_aws_vxc" "test" {
  name              = "terraform_acctest_%s"
  rate_limit        = 100
  invoice_reference = "terraform_acctest_ref_%s"

  a_end {
    product_uid = megaport_port.port.id
  }

  b_end {
    product_uid    = data.megaport_partner_port.aws.id
    aws_account_id = "%s"
    customer_asn   = %d
    type           = "private"
  }
}
`, name, name, name, accountId, asn)
}