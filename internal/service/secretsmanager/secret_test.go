// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secretsmanager_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsecretsmanager "github.com/hashicorp/terraform-provider-aws/internal/service/secretsmanager"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSecretsManagerSecret_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var secret secretsmanager.DescribeSecretOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecretDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretConfig_name(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecretExists(ctx, resourceName, &secret),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "secretsmanager", regexache.MustCompile(fmt.Sprintf("secret:%s-[[:alnum:]]+$", rName))),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "force_overwrite_replica_secret", "false"),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "recovery_window_in_days", "30"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"recovery_window_in_days", "force_overwrite_replica_secret"},
			},
		},
	})
}

func TestAccSecretsManagerSecret_withNamePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var secret secretsmanager.DescribeSecretOutput
	resourceName := "aws_secretsmanager_secret.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecretDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretConfig_namePrefix("tf-acc-test-prefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(ctx, resourceName, &secret),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, "name", "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "tf-acc-test-prefix-"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"recovery_window_in_days", "force_overwrite_replica_secret"},
			},
		},
	})
}

func TestAccSecretsManagerSecret_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var secret secretsmanager.DescribeSecretOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecretDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(ctx, resourceName, &secret),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsecretsmanager.ResourceSecret(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSecretsManagerSecret_description(t *testing.T) {
	ctx := acctest.Context(t)
	var secret secretsmanager.DescribeSecretOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecretDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(ctx, resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				Config: testAccSecretConfig_description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(ctx, resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"recovery_window_in_days", "force_overwrite_replica_secret"},
			},
		},
	})
}

func TestAccSecretsManagerSecret_secretValues(t *testing.T) {
	ctx := acctest.Context(t)
	var secret secretsmanager.DescribeSecretOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecretDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretConfig_secretString(rName, "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(ctx, resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "secret_string", "value1"),
				),
			},
			{
				Config: testAccSecretConfig_secretString(rName, "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(ctx, resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "secret_string", "value2"),
				),
			},
			{
				Config: testAccSecretConfig_secretBinary(rName, itypes.Base64EncodeOnce([]byte("value3"))),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(ctx, resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "secret_string", itypes.Base64EncodeOnce([]byte("value3"))),
				),
			},
			{
				Config: testAccSecretConfig_secretBinary(rName, itypes.Base64EncodeOnce([]byte("value4"))),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(ctx, resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "secret_string", itypes.Base64EncodeOnce([]byte("value4"))),
				),
			},
			{
				Config: testAccSecretConfig_secretString(rName, "value6"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(ctx, resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "secret_string", "value6"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"recovery_window_in_days", "force_overwrite_replica_secret"},
			},
		},
	})
}

func TestAccSecretsManagerSecret_basicReplica(t *testing.T) {
	ctx := acctest.Context(t)
	var secret secretsmanager.DescribeSecretOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t); acctest.PreCheckMultipleRegion(t, 2) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 2),
		CheckDestroy:             testAccCheckSecretDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretConfig_basicReplica(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(ctx, resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "force_overwrite_replica_secret", "false"),
					resource.TestCheckResourceAttr(resourceName, "replica.#", "1"),
				),
			},
		},
	})
}

func TestAccSecretsManagerSecret_overwriteReplica(t *testing.T) {
	ctx := acctest.Context(t)
	var secret secretsmanager.DescribeSecretOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t); acctest.PreCheckMultipleRegion(t, 3) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 3),
		CheckDestroy:             testAccCheckSecretDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretConfig_overwriteReplica(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(ctx, resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "force_overwrite_replica_secret", "true"),
				),
			},
			{
				Config: testAccSecretConfig_overwriteReplicaUpdate(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(ctx, resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "force_overwrite_replica_secret", "true"),
				),
			},
			{
				Config: testAccSecretConfig_overwriteReplica(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(ctx, resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "force_overwrite_replica_secret", "false"),
				),
			},
		},
	})
}

func TestAccSecretsManagerSecret_kmsKeyID(t *testing.T) {
	ctx := acctest.Context(t)
	var secret secretsmanager.DescribeSecretOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecretDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretConfig_kmsKeyID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(ctx, resourceName, &secret),
					resource.TestCheckResourceAttrSet(resourceName, "kms_key_id"),
				),
			},
			{
				Config: testAccSecretConfig_kmsKeyIDUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(ctx, resourceName, &secret),
					resource.TestCheckResourceAttrSet(resourceName, "kms_key_id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"recovery_window_in_days", "force_overwrite_replica_secret"},
			},
		},
	})
}

func TestAccSecretsManagerSecret_RecoveryWindowInDays_recreate(t *testing.T) {
	ctx := acctest.Context(t)
	var secret secretsmanager.DescribeSecretOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecretDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretConfig_recoveryWindowInDays(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(ctx, resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "recovery_window_in_days", "0"),
				),
			},
			{
				Config: testAccSecretConfig_recoveryWindowInDays(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(ctx, resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "recovery_window_in_days", "0"),
				),
				Taint: []string{resourceName},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"recovery_window_in_days", "force_overwrite_replica_secret"},
			},
		},
	})
}

func TestAccSecretsManagerSecret_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var secret secretsmanager.DescribeSecretOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecretDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(ctx, resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"recovery_window_in_days", "force_overwrite_replica_secret"},
			},
			{
				Config: testAccSecretConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(ctx, resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccSecretConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(ctx, resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccSecretsManagerSecret_policy(t *testing.T) {
	ctx := acctest.Context(t)
	var secret secretsmanager.DescribeSecretOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_secretsmanager_secret.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecretsManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecretDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecretConfig_policy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(ctx, resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "description", "San Holo feat. Duskus"),
					resource.TestMatchResourceAttr(resourceName, "policy",
						regexache.MustCompile(`{"Action":"secretsmanager:GetSecretValue".+`)),
				),
			},
			{
				Config: testAccSecretConfig_policyEmpty(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(ctx, resourceName, &secret),
					resource.TestCheckResourceAttr(resourceName, "description", "Poliça"),
					resource.TestCheckResourceAttr(resourceName, "policy", ""),
				),
			},
			{
				Config: testAccSecretConfig_policy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretExists(ctx, resourceName, &secret),
					resource.TestMatchResourceAttr(resourceName, "policy",
						regexache.MustCompile(`{"Action":"secretsmanager:GetSecretValue".+`)),
				),
			},
		},
	})
}

func testAccCheckSecretDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SecretsManagerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_secretsmanager_secret" {
				continue
			}

			_, err := tfsecretsmanager.FindSecretByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Secrets Manager Secret %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckSecretExists(ctx context.Context, n string, v *secretsmanager.DescribeSecretOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SecretsManagerClient(ctx)

		output, err := tfsecretsmanager.FindSecretByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SecretsManagerClient(ctx)

	input := &secretsmanager.ListSecretsInput{}

	_, err := conn.ListSecrets(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccSecretConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  description = %[1]q
  name        = %[2]q
}
`, description, rName)
}

func testAccSecretConfig_secretString(rName, secretString string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  secret_string = %[1]q
  name          = %[2]q
}
`, secretString, rName)
}

func testAccSecretConfig_secretBinary(rName, secretBinary string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  secret_binary = %[1]q
  name          = %[2]q
}
`, secretBinary, rName)
}

func testAccSecretConfig_basicReplica(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigMultipleRegionProvider(2), fmt.Sprintf(`
data "aws_region" "alternate" {
  provider = awsalternate
}

resource "aws_secretsmanager_secret" "test" {
  name = %[1]q

  replica {
    region = data.aws_region.alternate.name
  }
}
`, rName))
}

func testAccSecretConfig_overwriteReplica(rName string, force_overwrite_replica_secret bool) string {
	return acctest.ConfigCompose(acctest.ConfigMultipleRegionProvider(3), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  provider                = awsalternate
  deletion_window_in_days = 7
}

resource "aws_kms_key" "test2" {
  provider                = awsthird
  deletion_window_in_days = 7
}

data "aws_region" "alternate" {
  provider = awsalternate
}

resource "aws_secretsmanager_secret" "test" {
  name                           = %[1]q
  force_overwrite_replica_secret = %[2]t

  replica {
    kms_key_id = aws_kms_key.test.key_id
    region     = data.aws_region.alternate.name
  }
}
`, rName, force_overwrite_replica_secret))
}

func testAccSecretConfig_overwriteReplicaUpdate(rName string, force_overwrite_replica_secret bool) string {
	return acctest.ConfigCompose(acctest.ConfigMultipleRegionProvider(3), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  provider                = awsalternate
  deletion_window_in_days = 7
}

resource "aws_kms_key" "test2" {
  provider                = awsthird
  deletion_window_in_days = 7
}

data "aws_region" "third" {
  provider = awsthird
}

resource "aws_secretsmanager_secret" "test" {
  name                           = %[1]q
  force_overwrite_replica_secret = %[2]t

  replica {
    kms_key_id = aws_kms_key.test2.key_id
    region     = data.aws_region.third.name
  }
}
`, rName, force_overwrite_replica_secret))
}

func testAccSecretConfig_name(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}
`, rName)
}

func testAccSecretConfig_namePrefix(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name_prefix = %[1]q
}
`, rName)
}

func testAccSecretConfig_kmsKeyID(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test1" {
  deletion_window_in_days = 7
}

resource "aws_kms_key" "test2" {
  deletion_window_in_days = 7
}

resource "aws_secretsmanager_secret" "test" {
  kms_key_id = aws_kms_key.test1.id
  name       = %[1]q
}
`, rName)
}

func testAccSecretConfig_kmsKeyIDUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test1" {
  deletion_window_in_days = 7
}

resource "aws_kms_key" "test2" {
  deletion_window_in_days = 7
}

resource "aws_secretsmanager_secret" "test" {
  kms_key_id = aws_kms_key.test2.id
  name       = %[1]q
}
`, rName)
}

func testAccSecretConfig_recoveryWindowInDays(rName string, recoveryWindowInDays int) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name                    = %[1]q
  recovery_window_in_days = %[2]d
}
`, rName, recoveryWindowInDays)
}

func testAccSecretConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccSecretConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccSecretConfig_policy(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Principal = {
        Service = "ec2.amazonaws.com"
      },
      Effect = "Allow"
      Sid    = ""
    }]
  })
}

resource "aws_secretsmanager_secret" "test" {
  name        = %[1]q
  description = "San Holo feat. Duskus"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Sid    = "EnableAllPermissions"
      Effect = "Allow"
      Principal = {
        AWS = aws_iam_role.test.arn
      }
      Action   = "secretsmanager:GetSecretValue"
      Resource = "*"
    }]
  })
}
`, rName)
}

func testAccSecretConfig_policyEmpty(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Principal = {
        Service = "ec2.amazonaws.com"
      },
      Effect = "Allow"
      Sid    = ""
    }]
  })
}

resource "aws_secretsmanager_secret" "test" {
  name        = %[1]q
  description = "Poliça"

  policy = "{}"
}
`, rName)
}
