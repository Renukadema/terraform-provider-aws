package aws

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSDynamoDbTableItem_basic(t *testing.T) {
	var conf dynamodb.GetItemOutput

	tableName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(8))
	hashKey := "hashKey"
	itemContent := `{
	"hashKey": {"S": "something"},
	"one": {"N": "11111"},
	"two": {"N": "22222"},
	"three": {"N": "33333"},
	"four": {"N": "44444"}
}`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbItemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbItemConfigBasic(tableName, hashKey, itemContent),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDynamoDbTableItemExists("aws_dynamodb_table_item.test", &conf),
					testAccCheckAWSDynamoDbTableItemCount(tableName, 1),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test", "hash_key", hashKey),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test", "table_name", tableName),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test", "item", itemContent+"\n"),
				),
			},
		},
	})
}

func TestAccAWSDynamoDbTableItem_rangeKey(t *testing.T) {
	var conf dynamodb.GetItemOutput

	tableName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(8))
	hashKey := "hashKey"
	rangeKey := "rangeKey"
	itemContent := `{
	"hashKey": {"S": "something"},
	"rangeKey": {"S": "something-else"},
	"one": {"N": "11111"},
	"two": {"N": "22222"},
	"three": {"N": "33333"},
	"four": {"N": "44444"}
}`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbItemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbItemConfigWithRangeKey(tableName, hashKey, rangeKey, itemContent),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDynamoDbTableItemExists("aws_dynamodb_table_item.test", &conf),
					testAccCheckAWSDynamoDbTableItemCount(tableName, 1),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test", "hash_key", hashKey),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test", "range_key", rangeKey),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test", "table_name", tableName),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test", "item", itemContent+"\n"),
				),
			},
		},
	})
}

func TestAccAWSDynamoDbTableItem_withMultipleItems(t *testing.T) {
	var conf1 dynamodb.GetItemOutput
	var conf2 dynamodb.GetItemOutput

	tableName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(8))
	hashKey := "hashKey"
	rangeKey := "rangeKey"
	firstItem := `{
	"hashKey": {"S": "something"},
	"rangeKey": {"S": "first"},
	"one": {"N": "11111"},
	"two": {"N": "22222"},
	"three": {"N": "33333"}
}`
	secondItem := `{
	"hashKey": {"S": "something"},
	"rangeKey": {"S": "second"},
	"one": {"S": "one"},
	"two": {"S": "two"},
	"three": {"S": "three"},
	"four": {"S": "four"}
}`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbItemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbItemConfigWithMultipleItems(tableName, hashKey, rangeKey, firstItem, secondItem),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDynamoDbTableItemExists("aws_dynamodb_table_item.test1", &conf1),
					testAccCheckAWSDynamoDbTableItemExists("aws_dynamodb_table_item.test2", &conf2),
					testAccCheckAWSDynamoDbTableItemCount(tableName, 2),

					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test1", "hash_key", hashKey),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test1", "range_key", rangeKey),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test1", "table_name", tableName),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test1", "item", firstItem+"\n"),

					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test2", "hash_key", hashKey),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test2", "range_key", rangeKey),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test2", "table_name", tableName),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test2", "item", secondItem+"\n"),
				),
			},
		},
	})
}

func TestAccAWSDynamoDbTableItem_update(t *testing.T) {
	var conf dynamodb.GetItemOutput

	tableName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(8))
	hashKey := "hashKey"

	itemBefore := `{
	"hashKey": {"S": "before"},
	"one": {"N": "11111"},
	"two": {"N": "22222"},
	"three": {"N": "33333"},
	"four": {"N": "44444"}
}`
	itemAfter := `{
	"hashKey": {"S": "before"},
	"one": {"N": "11111"},
	"two": {"N": "22222"},
	"new": {"S": "shiny new one"}
}`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbItemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbItemConfigBasic(tableName, hashKey, itemBefore),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDynamoDbTableItemExists("aws_dynamodb_table_item.test", &conf),
					testAccCheckAWSDynamoDbTableItemCount(tableName, 1),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test", "hash_key", hashKey),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test", "table_name", tableName),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test", "item", itemBefore+"\n"),
				),
			},
			{
				Config: testAccAWSDynamoDbItemConfigBasic(tableName, hashKey, itemAfter),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDynamoDbTableItemExists("aws_dynamodb_table_item.test", &conf),
					testAccCheckAWSDynamoDbTableItemCount(tableName, 1),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test", "hash_key", hashKey),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test", "table_name", tableName),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test", "item", itemAfter+"\n"),
				),
			},
		},
	})
}

func TestAccAWSDynamoDbTableItem_updateWithRangeKey(t *testing.T) {
	var conf dynamodb.GetItemOutput

	tableName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(8))
	hashKey := "hashKey"
	rangeKey := "rangeKey"

	itemBefore := `{
	"hashKey": {"S": "before"},
	"rangeKey": {"S": "rangeBefore"},
	"value": {"S": "valueBefore"}
}`
	itemAfter := `{
	"hashKey": {"S": "before"},
	"rangeKey": {"S": "rangeAfter"},
	"value": {"S": "valueAfter"}
}`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbItemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbItemConfigWithRangeKey(tableName, hashKey, rangeKey, itemBefore),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDynamoDbTableItemExists("aws_dynamodb_table_item.test", &conf),
					testAccCheckAWSDynamoDbTableItemCount(tableName, 1),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test", "hash_key", hashKey),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test", "range_key", rangeKey),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test", "table_name", tableName),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test", "item", itemBefore+"\n"),
				),
			},
			{
				Config: testAccAWSDynamoDbItemConfigWithRangeKey(tableName, hashKey, rangeKey, itemAfter),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDynamoDbTableItemExists("aws_dynamodb_table_item.test", &conf),
					testAccCheckAWSDynamoDbTableItemCount(tableName, 1),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test", "hash_key", hashKey),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test", "range_key", rangeKey),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test", "table_name", tableName),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test", "item", itemAfter+"\n"),
				),
			},
		},
	})
}

func TestAccCheckAWSDynamoDbItem_importWithHashKey(t *testing.T) {
	var conf dynamodb.GetItemOutput

	tableName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(8))
	hashKey := "hashKey"
	hashKeyValue := "importHashKey"

	item := fmt.Sprintf(`{
	"%s": {"S": "%s"},
	"otherAttrS": {"S": "otherStringValue"},
	"otherAttrN": {"N": "123"},
	"otherAttrSS": {"SS": ["a", "b", "c"]},
	"otherAttrNS": {"NS": ["0", "1.1", "-2.22"]},
	"otherAttrBool": {"BOOL": false},
	"otherAttrNull": {"NULL": true}
}`, hashKey, hashKeyValue)

	checkFn := func(s []*terraform.InstanceState) error {
		return testAccAWSDynamoDbItemCompareItemAttribute(item, s)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbItemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbItemImportWithHashKey(tableName, hashKey, item),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDynamoDbTableItemExists("aws_dynamodb_table_item.test", &conf),
					testAccCheckAWSDynamoDbTableItemCount(tableName, 1),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test", "hash_key", hashKey),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test", "table_name", tableName),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test", "item", item+"\n"),
				),
			},
			{
				ResourceName:            "aws_dynamodb_table_item.test",
				ImportState:             true,
				ImportStateId:           fmt.Sprintf("%s|%s", tableName, hashKeyValue),
				ImportStateCheck:        checkFn,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"item"},
			},
			{
				ResourceName:            "aws_dynamodb_table_item.test",
				ImportState:             true,
				ImportStateId:           fmt.Sprintf("[ \"%s\", \"%s\" ]", tableName, hashKeyValue),
				ImportStateCheck:        checkFn,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"item"},
			},
		},
	})
}

func TestAccCheckAWSDynamoDbItem_importWithRangeKey(t *testing.T) {
	var conf dynamodb.GetItemOutput

	tableName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(8))
	hashKey := "hashKey"
	hashKeyValue := "dGhpcyB0ZXh0IGlzIGJhc2U2NC1lbmNvZGVk"
	rangeKey := "rangeKey"
	rangeKeyValue := "100"

	item := fmt.Sprintf(`{
	"%s": {"B": "%s"},
	"%s": {"N": "%s"},
	"otherAttrS": {"S": "otherStringValue"},
	"otherAttrN": {"N": "123"},
	"otherAttrSS": {"SS": ["a", "b", "c"]},
	"otherAttrNS": {"NS": ["0", "1.1", "-2.22"]},
	"otherAttrBool": {"BOOL": false},
	"otherAttrNull": {"NULL": true}
}`, hashKey, hashKeyValue, rangeKey, rangeKeyValue)

	checkFn := func(s []*terraform.InstanceState) error {
		return testAccAWSDynamoDbItemCompareItemAttribute(item, s)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbItemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbItemImportWithRangeKey(tableName, hashKey, rangeKey, item),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDynamoDbTableItemExists("aws_dynamodb_table_item.test", &conf),
					testAccCheckAWSDynamoDbTableItemCount(tableName, 1),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test", "hash_key", hashKey),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test", "range_key", rangeKey),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test", "table_name", tableName),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test", "item", item+"\n"),
				),
			},
			{
				ResourceName:            "aws_dynamodb_table_item.test",
				ImportState:             true,
				ImportStateId:           fmt.Sprintf("%s|%s|%s", tableName, hashKeyValue, rangeKeyValue),
				ImportStateCheck:        checkFn,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"item"},
			},
			{
				ResourceName:            "aws_dynamodb_table_item.test",
				ImportState:             true,
				ImportStateId:           fmt.Sprintf("[ \"%s\", \"%s\", \"%s\" ]", tableName, hashKeyValue, rangeKeyValue),
				ImportStateCheck:        checkFn,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"item"},
			},
		},
	})
}

func testAccCheckAWSDynamoDbItemDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).dynamodbconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dynamodb_table_item" {
			continue
		}

		attrs := rs.Primary.Attributes
		attributes, err := expandDynamoDbTableItemAttributes(attrs["item"])
		if err != nil {
			return err
		}

		result, err := conn.GetItem(&dynamodb.GetItemInput{
			TableName:                aws.String(attrs["table_name"]),
			ConsistentRead:           aws.Bool(true),
			Key:                      buildDynamoDbTableItemQueryKey(attributes, attrs["hash_key"], attrs["range_key"]),
			ProjectionExpression:     buildDynamoDbProjectionExpression(attributes),
			ExpressionAttributeNames: buildDynamoDbExpressionAttributeNames(attributes),
		})
		if err != nil {
			if isAWSErr(err, dynamodb.ErrCodeResourceNotFoundException, "") {
				return nil
			}
			return fmt.Errorf("Error retrieving DynamoDB table item: %s", err)
		}
		if result.Item == nil {
			return nil
		}

		return fmt.Errorf("DynamoDB table item %s still exists.", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAWSDynamoDbTableItemExists(n string, item *dynamodb.GetItemOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DynamoDB table item ID specified!")
		}

		conn := testAccProvider.Meta().(*AWSClient).dynamodbconn

		attrs := rs.Primary.Attributes
		attributes, err := expandDynamoDbTableItemAttributes(attrs["item"])
		if err != nil {
			return err
		}

		result, err := conn.GetItem(&dynamodb.GetItemInput{
			TableName:                aws.String(attrs["table_name"]),
			ConsistentRead:           aws.Bool(true),
			Key:                      buildDynamoDbTableItemQueryKey(attributes, attrs["hash_key"], attrs["range_key"]),
			ProjectionExpression:     buildDynamoDbProjectionExpression(attributes),
			ExpressionAttributeNames: buildDynamoDbExpressionAttributeNames(attributes),
		})
		if err != nil {
			return fmt.Errorf("Problem getting table item '%s': %s", rs.Primary.ID, err)
		}

		*item = *result

		return nil
	}
}

func testAccCheckAWSDynamoDbTableItemCount(tableName string, count int64) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).dynamodbconn
		out, err := conn.Scan(&dynamodb.ScanInput{
			ConsistentRead: aws.Bool(true),
			TableName:      aws.String(tableName),
			Select:         aws.String(dynamodb.SelectCount),
		})
		if err != nil {
			return err
		}
		expectedCount := count
		if *out.Count != expectedCount {
			return fmt.Errorf("Expected %d items, got %d", expectedCount, *out.Count)
		}
		return nil
	}
}

func testAccAWSDynamoDbItemConfigBasic(tableName, hashKey, item string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = "%s"
  read_capacity  = 10
  write_capacity = 10
  hash_key       = "%s"

  attribute {
    name = "%s"
    type = "S"
  }
}

resource "aws_dynamodb_table_item" "test" {
  table_name = "${aws_dynamodb_table.test.name}"
  hash_key   = "${aws_dynamodb_table.test.hash_key}"

  item = <<ITEM
%s
ITEM
}
`, tableName, hashKey, hashKey, item)
}

func testAccAWSDynamoDbItemConfigWithRangeKey(tableName, hashKey, rangeKey, item string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = "%s"
  read_capacity  = 10
  write_capacity = 10
  hash_key       = "%s"
  range_key      = "%s"

  attribute {
    name = "%s"
    type = "S"
  }

  attribute {
    name = "%s"
    type = "S"
  }
}

resource "aws_dynamodb_table_item" "test" {
  table_name = "${aws_dynamodb_table.test.name}"
  hash_key   = "${aws_dynamodb_table.test.hash_key}"
  range_key  = "${aws_dynamodb_table.test.range_key}"

  item = <<ITEM
%s
ITEM
}
`, tableName, hashKey, rangeKey, hashKey, rangeKey, item)
}

func testAccAWSDynamoDbItemConfigWithMultipleItems(tableName, hashKey, rangeKey, firstItem, secondItem string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = "%s"
  read_capacity  = 10
  write_capacity = 10
  hash_key       = "%s"
  range_key      = "%s"

  attribute {
    name = "%s"
    type = "S"
  }

  attribute {
    name = "%s"
    type = "S"
  }
}

resource "aws_dynamodb_table_item" "test1" {
  table_name = "${aws_dynamodb_table.test.name}"
  hash_key   = "${aws_dynamodb_table.test.hash_key}"
  range_key  = "${aws_dynamodb_table.test.range_key}"

  item = <<ITEM
%s
ITEM
}

resource "aws_dynamodb_table_item" "test2" {
  table_name = "${aws_dynamodb_table.test.name}"
  hash_key   = "${aws_dynamodb_table.test.hash_key}"
  range_key  = "${aws_dynamodb_table.test.range_key}"

  item = <<ITEM
%s
ITEM
}
`, tableName, hashKey, rangeKey, hashKey, rangeKey, firstItem, secondItem)
}

func testAccAWSDynamoDbItemCompareItemAttribute(item string, s []*terraform.InstanceState) error {
	if len(s) != 1 {
		return fmt.Errorf("expected 1 state: %#v", s)
	}
	var a, b map[string]interface{}

	err := json.Unmarshal([]byte(item), &a)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(s[0].Attributes["item"]), &b)
	if err != nil {
		return err
	}

	if !reflect.DeepEqual(a, b) {
		return fmt.Errorf("item attributes not equal:\n\texpected: %#v\n\tactual: %#v", a, b)
	}

	return nil
}

func testAccAWSDynamoDbItemImportWithHashKey(tableName, hashKey, item string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = "%s"
  read_capacity  = 5
  write_capacity = 5
	hash_key       = "%s"

  attribute {
    name = "%s"
    type = "S"
  }
}

resource "aws_dynamodb_table_item" "test" {
  table_name = "${aws_dynamodb_table.test.name}"
	hash_key   = "${aws_dynamodb_table.test.hash_key}"

  item = <<ITEM
%s
ITEM
}
`, tableName, hashKey, hashKey, item)
}

func testAccAWSDynamoDbItemImportWithRangeKey(tableName, hashKey, rangeKey, item string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = "%s"
  read_capacity  = 5
  write_capacity = 5
	hash_key       = "%s"
	range_key      = "%s"

  attribute {
    name = "%s"
    type = "B"
  }

  attribute {
    name = "%s"
    type = "N"
  }
}

resource "aws_dynamodb_table_item" "test" {
  table_name = "${aws_dynamodb_table.test.name}"
	hash_key   = "${aws_dynamodb_table.test.hash_key}"
	range_key  = "${aws_dynamodb_table.test.range_key}"

  item = <<ITEM
%s
ITEM
}
`, tableName, hashKey, rangeKey, hashKey, rangeKey, item)
}
