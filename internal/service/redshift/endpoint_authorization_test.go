package redshift_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/redshift"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfredshift "github.com/hashicorp/terraform-provider-aws/internal/service/redshift"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccRedshiftEndpointAuthorization_basic(t *testing.T) {
	var v redshift.EndpointAuthorization
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(18))
	resourceName := "aws_redshift_endpoint_authorization.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, redshift.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(t),
		CheckDestroy:             testAccCheckEndpointAuthorizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointAuthorizationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointAuthorizationExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "cluster_identifier", "aws_redshift_cluster.test", "cluster_identifier"),
					resource.TestCheckResourceAttrPair(resourceName, "account", "data.aws_caller_identity.test", "account_id"),
					resource.TestCheckResourceAttr(resourceName, "allowed_all_vpcs", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "grantee", "data.aws_caller_identity.test", "account_id"),
					acctest.CheckResourceAttrAccountID(resourceName, "grantor"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_delete"},
			},
		},
	})
}

func TestAccRedshiftEndpointAuthorization_vpcs(t *testing.T) {
	var v redshift.EndpointAuthorization
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(18))
	resourceName := "aws_redshift_endpoint_authorization.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, redshift.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(t),
		CheckDestroy:             testAccCheckEndpointAuthorizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointAuthorizationConfig_vpcs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointAuthorizationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "vpc_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "allowed_all_vpcs", "false"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_delete"},
			},
			{
				Config: testAccEndpointAuthorizationConfig_vpcsUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointAuthorizationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "vpc_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "allowed_all_vpcs", "false"),
				),
			},
			{
				Config: testAccEndpointAuthorizationConfig_vpcs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointAuthorizationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "vpc_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "allowed_all_vpcs", "false"),
				),
			},
		},
	})
}

func TestAccRedshiftEndpointAuthorization_disappears(t *testing.T) {
	var v redshift.EndpointAuthorization
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(18))
	resourceName := "aws_redshift_endpoint_authorization.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, redshift.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(t),
		CheckDestroy:             testAccCheckEndpointAuthorizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointAuthorizationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointAuthorizationExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfredshift.ResourceEndpointAuthorization(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRedshiftEndpointAuthorization_disappears_cluster(t *testing.T) {
	var v redshift.EndpointAuthorization
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(18))
	resourceName := "aws_redshift_endpoint_authorization.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, redshift.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(t),
		CheckDestroy:             testAccCheckEndpointAuthorizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointAuthorizationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointAuthorizationExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfredshift.ResourceCluster(), "aws_redshift_cluster.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckEndpointAuthorizationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_redshift_endpoint_authorization" {
			continue
		}

		_, err := tfredshift.FindEndpointAuthorizationById(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Redshift Endpoint Authorization %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckEndpointAuthorizationExists(n string, v *redshift.EndpointAuthorization) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Redshift Endpoint Authorization ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn

		output, err := tfredshift.FindEndpointAuthorizationById(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccEndpointAuthorizationConfigBase(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2),
		acctest.ConfigAlternateAccountProvider(),
		fmt.Sprintf(`
resource "aws_redshift_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_redshift_cluster" "test" {
  cluster_identifier                   = %[1]q
  availability_zone                    = data.aws_availability_zones.available.names[0]
  database_name                        = "mydb"
  master_username                      = "foo_test"
  master_password                      = "Mustbe8characters"
  node_type                            = "ra3.xlplus"
  automated_snapshot_retention_period  = 1
  allow_version_upgrade                = false
  skip_final_snapshot                  = true
  availability_zone_relocation_enabled = true
  publicly_accessible                  = false
}

data "aws_caller_identity" "test" {
  provider = "awsalternate"
}
`, rName))
}

func testAccEndpointAuthorizationConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccEndpointAuthorizationConfigBase(rName), `
resource "aws_redshift_endpoint_authorization" "test" {
  account            = data.aws_caller_identity.test.account_id
  cluster_identifier = aws_redshift_cluster.test.cluster_identifier
}
`)
}

func testAccEndpointAuthorizationConfig_vpcs(rName string) string {
	return acctest.ConfigCompose(testAccEndpointAuthorizationConfigBase(rName), fmt.Sprintf(`
resource "aws_vpc" "test2" {
  provider = "awsalternate"

  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_redshift_endpoint_authorization" "test" {
  account            = data.aws_caller_identity.test.account_id
  cluster_identifier = aws_redshift_cluster.test.cluster_identifier
  vpc_ids            = [aws_vpc.test2.id]
}
`, rName))
}

func testAccEndpointAuthorizationConfig_vpcsUpdated(rName string) string {
	return acctest.ConfigCompose(testAccEndpointAuthorizationConfigBase(rName), fmt.Sprintf(`
resource "aws_vpc" "test2" {
  provider = "awsalternate"

  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "test3" {
  provider = "awsalternate"

  cidr_block = "11.0.0.0/16"

  tags = {
    Name = "%[1]s-2"
  }
}

resource "aws_redshift_endpoint_authorization" "test" {
  account            = data.aws_caller_identity.test.account_id
  cluster_identifier = aws_redshift_cluster.test.cluster_identifier
  vpc_ids            = [aws_vpc.test2.id, aws_vpc.test3.id]
}
`, rName))
}
