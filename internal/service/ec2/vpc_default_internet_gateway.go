// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_default_internet_gateway", name="Internet Gateway")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func ResourceDefaultInternetGateway() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDefaultInternetGatewayCreate,
		ReadWithoutTimeout:   resourceInternetGatewayRead,
		UpdateWithoutTimeout: resourceInternetGatewayUpdate,
		DeleteWithoutTimeout: resourceDefaultInternetGatewayDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"existing_default_internet_gateway": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrForceDestroy: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceDefaultInternetGatewayCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Check if there is a default VPC
	input := &ec2.DescribeVpcsInput{
		Filters: newAttributeFilterList(
			map[string]string{
				"isDefault": "true",
			},
		),
	}
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)
	vpc, err := FindVPC(ctx, conn, input)
	if err == nil {
		log.Print("[INFO] Found existing attached EC2 Internet Gateway")
		input := &ec2.DescribeInternetGatewaysInput{}
		input.Filters = newAttributeFilterList(map[string]string{
			"attachment.vpc-id": *vpc.VpcId,
		})

		igw, err := FindInternetGateway(ctx, conn, input)
		log.Printf("found igw with ID: %s", igw)

		if err == nil {
			d.SetId(aws.StringValue(igw.InternetGatewayId))
			d.Set("existing_default_internet_gateway", true)
		} else if tfresource.NotFound(err) {
			log.Print("[INFO] Found default VPC without attached EC2 Internet Gateway. Creating and attaching one")
			input := &ec2.CreateInternetGatewayInput{}
			output, err := conn.CreateInternetGatewayWithContext(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "Creating EC2 Internet Gateway: %s", err)
			}

			igw = output.InternetGateway

			d.SetId(aws.StringValue(igw.InternetGatewayId))
			d.Set("existing_default_internet_gateway", false)

			if err := attachInternetGateway(ctx, conn, d.Id(), *vpc.VpcId, d.Timeout(schema.TimeoutDelete)); err != nil {
				return sdkdiag.AppendErrorf(diags, "Attaching EC2 Internet Gateway (%s): %s", d.Id(), err)
			}
		} else {
			return sdkdiag.AppendErrorf(diags, "Creating EC2 Internet Gateway (%s): %s", d.Id(), err)
		}
	}
	return append(diags, resourceInternetGatewayRead(ctx, d, meta)...)
}

func resourceDefaultInternetGatewayDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	if d.Get(names.AttrForceDestroy).(bool) {
		// See if the VPC assigned to the IGW has the isDefault property
		conn := meta.(*conns.AWSClient).EC2Conn(ctx)
		input := &ec2.DescribeVpcsInput{
			Filters: newAttributeFilterList(
				map[string]string{
					"isDefault": "true",
					"vpc-id":    d.Get(names.AttrVPCID).(string),
				},
			),
		}
		_, err := FindVPC(ctx, conn, input)

		if err == nil {
			// Detach if it is attached.
			if v, ok := d.GetOk(names.AttrVPCID); ok {
				if err := detachInternetGateway(ctx, conn, d.Id(), v.(string), d.Timeout(schema.TimeoutDelete)); err != nil {
					return sdkdiag.AppendErrorf(diags, "deleting EC2 Internet Gateway (%s): %s", d.Id(), err)
				}

				input := &ec2.DeleteInternetGatewayInput{
					InternetGatewayId: aws.String(d.Id()),
				}

				log.Printf("[INFO] Deleting Internet Gateway: %s", d.Id())
				_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutDelete), func() (interface{}, error) {
					return conn.DeleteInternetGatewayWithContext(ctx, input)
				}, errCodeDependencyViolation)

				if tfawserr.ErrCodeEquals(err, errCodeInvalidInternetGatewayIDNotFound) {
					return diags
				}

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "deleting EC2 Internet Gateway (%s): %s", d.Id(), err)
				}
			}
		}
	}
	log.Printf("[INFO] Skipping Internet Gateway: %s", d.Id())
	return diags
}
