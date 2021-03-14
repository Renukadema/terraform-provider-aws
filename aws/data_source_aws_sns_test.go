package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourceAwsSnsTopic_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsSnsTopicConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsSnsTopicCheck("data.aws_sns_topic.by_name"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsSnsTopic_fifo(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				PreConfig: testAccDataSourceAwsSnsTopicCreateFifo,
				Config:    testAccDataSourceAwsSnsTopicConfigFifo,
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsSnsTopicCheckFifo("data.aws_sns_topic.by_name_fifo"),
				),
			},
		},
	})
}

// TODO: Replace this function with terraform config once FIFO SNS is supported
func testAccDataSourceAwsSnsTopicCreateFifo() {
	conn := testAccProvider.Meta().(*AWSClient).snsconn
	params := &sns.CreateTopicInput{
		Name: aws.String("tf_test.fifo"),
		Attributes: map[string]*string{
			"FifoTopic": aws.String("true"),
		},
	}
	_, err := conn.CreateTopic(params)
	if err != nil {
		log.Printf("[INFO] Failed to create topic %s", err)
	} else {
		log.Printf("[INFO] Created FIFO SNS Topic")
	}
}

func testAccDataSourceAwsSnsTopicCheck(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", name)
		}

		snsTopicRs, ok := s.RootModule().Resources["aws_sns_topic.tf_test"]
		if !ok {
			return fmt.Errorf("can't find aws_sns_topic.tf_test in state")
		}

		attr := rs.Primary.Attributes

		if attr["name"] != snsTopicRs.Primary.Attributes["name"] {
			return fmt.Errorf(
				"name is %s; want %s",
				attr["name"],
				snsTopicRs.Primary.Attributes["name"],
			)
		}

		return nil
	}
}

func testAccDataSourceAwsSnsTopicCheckFifo(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", name)
		}
		//snsTopicRs, ok := s.RootModule().Resources["aws_sns_topic.tf_test_fifo"]
		//if !ok {
		//	return fmt.Errorf("can't find aws_sns_topic.tf_test_fifo in state")
		//}
		//attr := rs.Primary.Attributes

		//if attr["name"] != snsTopicRs.Primary.Attributes["name"] {
		//	return fmt.Errorf(
		//		"name is %s; want %s",
		//		attr["name"],
		//		snsTopicRs.Primary.Attributes["name"],
		//	)
		//}
		if rs.Primary.Attributes["name"] != "tf_test.fifo" {
			return fmt.Errorf(
				"name is %s; want %s",
				rs.Primary.Attributes["name"],
				"tf_test.fifo")
		}

		return nil
	}
}

const testAccDataSourceAwsSnsTopicConfig = `
resource "aws_sns_topic" "tf_wrong1" {
  name = "wrong1"
}

resource "aws_sns_topic" "tf_test" {
  name = "tf_test"
}

resource "aws_sns_topic" "tf_wrong2" {
  name = "wrong2"
}

data "aws_sns_topic" "by_name" {
  name = aws_sns_topic.tf_test.name
  depends_on = [aws_sns_topic.tf_test]
}
`

const testAccDataSourceAwsSnsTopicConfigFifo = `
# Can't work until fifo support is added
#resource "aws_sns_topic" "tf_test_fifo" {
#  name = "tf_test.fifo"
#}

data "aws_sns_topic" "by_name_fifo" {
  name = "tf_test.fifo"
}
`
