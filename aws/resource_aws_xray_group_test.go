package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/xray"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSXrayGroup_basic(t *testing.T) {
	var Group xray.Group
	resourceName := "aws_xray_group.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSXrayGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSXrayGroupBasicConfig(rName, "responsetime > 5"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckXrayGroupExists(resourceName, &Group),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "xray", regexp.MustCompile(`group/.+`)),
					resource.TestCheckResourceAttr(resourceName, "group_name", rName),
					resource.TestCheckResourceAttr(resourceName, "filter_expression", "responsetime > 5"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSXrayGroupBasicConfig(rName, "responsetime > 10"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckXrayGroupExists(resourceName, &Group),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "xray", regexp.MustCompile(`group/.+`)),
					resource.TestCheckResourceAttr(resourceName, "group_name", rName),
					resource.TestCheckResourceAttr(resourceName, "filter_expression", "responsetime > 10"),
				),
			},
		},
	})
}

func TestAccAWSXrayGroup_disappears(t *testing.T) {
	var Group xray.Group
	resourceName := "aws_xray_group.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSXrayGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSXrayGroupBasicConfig(rName, "responsetime > 5"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckXrayGroupExists(resourceName, &Group),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsXrayGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckXrayGroupExists(n string, Group *xray.Group) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No XRay Group ID is set")
		}
		conn := testAccProvider.Meta().(*AWSClient).xrayconn

		input := &xray.GetGroupInput{
			GroupARN: aws.String(rs.Primary.ID),
		}

		group, err := conn.GetGroup(input)

		if err != nil {
			return err
		}

		*Group = *group.Group

		return nil
	}
}

func testAccCheckAWSXrayGroupDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_xray_group" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).xrayconn

		input := &xray.GetGroupInput{
			GroupARN: aws.String(rs.Primary.ID),
		}

		group, err := conn.GetGroup(input)

		if isAWSErr(err, xray.ErrCodeInvalidRequestException, "Group not found") {
			continue
		}

		if err != nil {
			return err
		}

		if group != nil {
			return fmt.Errorf("Expected XRay Group to be destroyed, %s found", rs.Primary.ID)
		}
	}

	return nil
}

func testAccAWSXrayGroupBasicConfig(rName, expression string) string {
	return fmt.Sprintf(`
resource "aws_xray_group" "test" {
  group_name        = %[1]q
  filter_expression = %[2]q
}
`, rName, expression)
}
