// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53profiles_test

import (
	"context"
	"fmt"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53profiles/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfroute53profiles "github.com/hashicorp/terraform-provider-aws/internal/service/route53profiles"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRoute53ProfilesProfile_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53profiles_profile.test"
	var v awstypes.Profile

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ProfilesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProfileConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProfileExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckProfileExists(ctx context.Context, name string, r *awstypes.Profile) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53ProfilesClient(ctx)

		output, err := tfroute53profiles.FindProfileByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*r = *output

		return nil
	}
}

func testAccCheckProfileDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53ProfilesClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_route53profiles_profile" {
				continue
			}

			_, err := tfroute53profiles.FindProfileByID(ctx, conn, rs.Primary.ID)
			if err == nil {
				return fmt.Errorf("route 53 Profile %s still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccProfileConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_route53profiles_profile" "test" {
  name = %[1]q
}
`, rName)
}
