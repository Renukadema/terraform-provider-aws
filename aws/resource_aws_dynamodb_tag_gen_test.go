// Code generated by internal/tagresource/generator/main.go; DO NOT EDIT.

package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tagresource"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func testAccCheckDynamodbTagDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).dynamodbconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dynamodb_tag" {
			continue
		}

		identifier, key, err := tagresource.GetResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		_, err = keyvaluetags.DynamodbGetTag(conn, identifier, key)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("%s resource (%s) tag (%s) still exists", dynamodb.ServiceID, identifier, key)
	}

	return nil
}

func testAccCheckDynamodbTagExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("%s: missing resource ID", resourceName)
		}

		identifier, key, err := tagresource.GetResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*AWSClient).dynamodbconn

		_, err = keyvaluetags.DynamodbGetTag(conn, identifier, key)

		if err != nil {
			return err
		}

		return nil
	}
}
