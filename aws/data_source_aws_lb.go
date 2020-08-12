package aws

import (
	"bytes"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsLb() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsLbRead,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"arn_suffix": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"internal": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"load_balancer_type": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"security_groups": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
				Set:      schema.HashString,
			},

			"subnets": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
				Set:      schema.HashString,
			},

			"subnet_mapping": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subnet_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"allocation_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"private_ipv4_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
				Set: func(v interface{}) int {
					var buf bytes.Buffer
					m := v.(map[string]interface{})
					buf.WriteString(fmt.Sprintf("%s-", m["subnet_id"].(string)))
					if m["allocation_id"] != "" {
						buf.WriteString(fmt.Sprintf("%s-", m["allocation_id"].(string)))
					}
					if m["private_ipv4_address"] != "" {
						buf.WriteString(fmt.Sprintf("%s-", m["private_ipv4_address"].(string)))
					}
					return hashcode.String(buf.String())
				},
			},

			"access_logs": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bucket": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"prefix": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},

			"enable_deletion_protection": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"enable_http2": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"idle_timeout": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"drop_invalid_header_fields": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"ip_address_type": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tagsSchemaComputed(),
		},
	}
}

func dataSourceAwsLbRead(d *schema.ResourceData, meta interface{}) error {
	elbconn := meta.(*AWSClient).elbv2conn
	lbArn := d.Get("arn").(string)
	lbName := d.Get("name").(string)

	describeLbOpts := &elbv2.DescribeLoadBalancersInput{}
	switch {
	case lbArn != "":
		describeLbOpts.LoadBalancerArns = []*string{aws.String(lbArn)}
	case lbName != "":
		describeLbOpts.Names = []*string{aws.String(lbName)}
	}

	log.Printf("[DEBUG] Reading Load Balancer: %s", describeLbOpts)
	describeResp, err := elbconn.DescribeLoadBalancers(describeLbOpts)
	if err != nil {
		return fmt.Errorf("Error retrieving LB: %s", err)
	}
	if len(describeResp.LoadBalancers) != 1 {
		return fmt.Errorf("Search returned %d results, please revise so only one is returned", len(describeResp.LoadBalancers))
	}
	d.SetId(aws.StringValue(describeResp.LoadBalancers[0].LoadBalancerArn))

	return flattenAwsLbResource(d, meta, describeResp.LoadBalancers[0])
}
