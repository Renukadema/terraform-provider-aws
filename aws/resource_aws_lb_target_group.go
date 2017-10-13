package aws

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsLbTargetGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsLbTargetGroupCreate,
		Read:   resourceAwsLbTargetGroupRead,
		Update: resourceAwsLbTargetGroupUpdate,
		Delete: resourceAwsLbTargetGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"arn_suffix": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  validateAwsLbTargetGroupName,
			},
			"name_prefix": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateAwsLbTargetGroupNamePrefix,
			},

			"port": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateAwsLbTargetGroupPort,
			},

			"protocol": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateAwsLbTargetGroupProtocol,
			},

			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"deregistration_delay": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      300,
				ValidateFunc: validateAwsLbTargetGroupDeregistrationDelay,
			},

			"stickiness": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateAwsLbTargetGroupStickinessType,
						},
						"cookie_duration": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      86400,
							ValidateFunc: validateAwsLbTargetGroupStickinessCookieDuration,
						},
					},
				},
			},

			"health_check": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"interval": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  30,
						},

						"path": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateAwsLbTargetGroupHealthCheckPath,
						},

						"port": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "traffic-port",
							ValidateFunc: validateAwsLbTargetGroupHealthCheckPort,
						},

						"protocol": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "HTTP",
							StateFunc: func(v interface{}) string {
								return strings.ToUpper(v.(string))
							},
							ValidateFunc: validateAwsLbTargetGroupHealthCheckProtocol,
						},

						"timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      10,
							ValidateFunc: validateAwsLbTargetGroupHealthCheckTimeout,
						},

						"healthy_threshold": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      3,
							ValidateFunc: validateAwsLbTargetGroupHealthCheckHealthyThreshold,
						},

						"matcher": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"unhealthy_threshold": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      3,
							ValidateFunc: validateAwsLbTargetGroupHealthCheckHealthyThreshold,
						},
					},
				},
			},

			"tags": tagsSchema(),
		},
	}
}

func resourceAwsLbTargetGroupCreate(d *schema.ResourceData, meta interface{}) error {
	elbconn := meta.(*AWSClient).elbv2conn

	var groupName string
	if v, ok := d.GetOk("name"); ok {
		groupName = v.(string)
	} else if v, ok := d.GetOk("name_prefix"); ok {
		groupName = resource.PrefixedUniqueId(v.(string))
	} else {
		groupName = resource.PrefixedUniqueId("tf-")
	}

	params := &elbv2.CreateTargetGroupInput{
		Name:     aws.String(groupName),
		Port:     aws.Int64(int64(d.Get("port").(int))),
		Protocol: aws.String(d.Get("protocol").(string)),
		VpcId:    aws.String(d.Get("vpc_id").(string)),
	}

	if healthChecks := d.Get("health_check").([]interface{}); len(healthChecks) == 1 {
		healthCheck := healthChecks[0].(map[string]interface{})

		params.HealthCheckIntervalSeconds = aws.Int64(int64(healthCheck["interval"].(int)))
		params.HealthCheckPort = aws.String(healthCheck["port"].(string))
		params.HealthCheckProtocol = aws.String(healthCheck["protocol"].(string))
		params.HealthyThresholdCount = aws.Int64(int64(healthCheck["healthy_threshold"].(int)))

		if *params.Protocol != "TCP" {
			params.HealthCheckTimeoutSeconds = aws.Int64(int64(healthCheck["timeout"].(int)))
			params.HealthCheckPath = aws.String(healthCheck["path"].(string))
			params.Matcher = &elbv2.Matcher{
				HttpCode: aws.String(healthCheck["matcher"].(string)),
			}
			params.UnhealthyThresholdCount = aws.Int64(int64(healthCheck["unhealthy_threshold"].(int)))
		}
	}

	resp, err := elbconn.CreateTargetGroup(params)
	if err != nil {
		return errwrap.Wrapf("Error creating LB Target Group: {{err}}", err)
	}

	if len(resp.TargetGroups) == 0 {
		return errors.New("Error creating LB Target Group: no groups returned in response")
	}

	targetGroupArn := resp.TargetGroups[0].TargetGroupArn
	d.SetId(*targetGroupArn)

	return resourceAwsLbTargetGroupUpdate(d, meta)
}

func resourceAwsLbTargetGroupRead(d *schema.ResourceData, meta interface{}) error {
	elbconn := meta.(*AWSClient).elbv2conn

	resp, err := elbconn.DescribeTargetGroups(&elbv2.DescribeTargetGroupsInput{
		TargetGroupArns: []*string{aws.String(d.Id())},
	})
	if err != nil {
		if isTargetGroupNotFound(err) {
			log.Printf("[DEBUG] DescribeTargetGroups - removing %s from state", d.Id())
			d.SetId("")
			return nil
		}
		return errwrap.Wrapf("Error retrieving Target Group: {{err}}", err)
	}

	if len(resp.TargetGroups) != 1 {
		return fmt.Errorf("Error retrieving Target Group %q", d.Id())
	}

	return flattenAwsLbTargetGroupResource(d, meta, resp.TargetGroups[0])
}

func resourceAwsLbTargetGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	elbconn := meta.(*AWSClient).elbv2conn

	if err := setElbV2Tags(elbconn, d); err != nil {
		return errwrap.Wrapf("Error Modifying Tags on LB Target Group: {{err}}", err)
	}

	if d.HasChange("health_check") {
		healthChecks := d.Get("health_check").([]interface{})

		var params *elbv2.ModifyTargetGroupInput
		if len(healthChecks) == 1 {
			healthCheck := healthChecks[0].(map[string]interface{})

			params = &elbv2.ModifyTargetGroupInput{
				TargetGroupArn:        aws.String(d.Id()),
				HealthCheckPort:       aws.String(healthCheck["port"].(string)),
				HealthCheckProtocol:   aws.String(healthCheck["protocol"].(string)),
				HealthyThresholdCount: aws.Int64(int64(healthCheck["healthy_threshold"].(int))),
			}

			healthCheckProtocol := strings.ToLower(healthCheck["protocol"].(string))

			if healthCheckProtocol != "tcp" {
				params.Matcher = &elbv2.Matcher{
					HttpCode: aws.String(healthCheck["matcher"].(string)),
				}
				params.HealthCheckPath = aws.String(healthCheck["path"].(string))

				params.HealthCheckIntervalSeconds = aws.Int64(int64(healthCheck["interval"].(int)))
				params.UnhealthyThresholdCount = aws.Int64(int64(healthCheck["unhealthy_threshold"].(int)))
				params.HealthCheckIntervalSeconds = aws.Int64(int64(healthCheck["timeout"].(int)))
			}
		} else {
			params = &elbv2.ModifyTargetGroupInput{
				TargetGroupArn: aws.String(d.Id()),
			}
		}

		_, err := elbconn.ModifyTargetGroup(params)
		if err != nil {
			return errwrap.Wrapf("Error modifying Target Group: {{err}}", err)
		}
	}

	var attrs []*elbv2.TargetGroupAttribute

	if d.HasChange("deregistration_delay") {
		attrs = append(attrs, &elbv2.TargetGroupAttribute{
			Key:   aws.String("deregistration_delay.timeout_seconds"),
			Value: aws.String(fmt.Sprintf("%d", d.Get("deregistration_delay").(int))),
		})
	}

	if d.HasChange("stickiness") {
		stickinessBlocks := d.Get("stickiness").([]interface{})
		if len(stickinessBlocks) == 1 {
			stickiness := stickinessBlocks[0].(map[string]interface{})

			attrs = append(attrs,
				&elbv2.TargetGroupAttribute{
					Key:   aws.String("stickiness.enabled"),
					Value: aws.String(strconv.FormatBool(stickiness["enabled"].(bool))),
				},
				&elbv2.TargetGroupAttribute{
					Key:   aws.String("stickiness.type"),
					Value: aws.String(stickiness["type"].(string)),
				},
				&elbv2.TargetGroupAttribute{
					Key:   aws.String("stickiness.lb_cookie.duration_seconds"),
					Value: aws.String(fmt.Sprintf("%d", stickiness["cookie_duration"].(int))),
				})
		} else if len(stickinessBlocks) == 0 {
			attrs = append(attrs, &elbv2.TargetGroupAttribute{
				Key:   aws.String("stickiness.enabled"),
				Value: aws.String("false"),
			})
		}
	}

	if len(attrs) > 0 {
		params := &elbv2.ModifyTargetGroupAttributesInput{
			TargetGroupArn: aws.String(d.Id()),
			Attributes:     attrs,
		}

		_, err := elbconn.ModifyTargetGroupAttributes(params)
		if err != nil {
			return errwrap.Wrapf("Error modifying Target Group Attributes: {{err}}", err)
		}
	}

	return resourceAwsLbTargetGroupRead(d, meta)
}

func resourceAwsLbTargetGroupDelete(d *schema.ResourceData, meta interface{}) error {
	elbconn := meta.(*AWSClient).elbv2conn

	_, err := elbconn.DeleteTargetGroup(&elbv2.DeleteTargetGroupInput{
		TargetGroupArn: aws.String(d.Id()),
	})
	if err != nil {
		return errwrap.Wrapf("Error deleting Target Group: {{err}}", err)
	}

	return nil
}

func isTargetGroupNotFound(err error) bool {
	elberr, ok := err.(awserr.Error)
	return ok && elberr.Code() == "TargetGroupNotFound"
}

func validateAwsLbTargetGroupHealthCheckPath(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if len(value) > 1024 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be longer than 1024 characters: %q", k, value))
	}
	return
}

func validateAwsLbTargetGroupHealthCheckPort(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	if value == "traffic-port" {
		return
	}

	port, err := strconv.Atoi(value)
	if err != nil {
		errors = append(errors, fmt.Errorf("%q must be a valid port number (1-65536) or %q", k, "traffic-port"))
	}

	if port < 1 || port > 65536 {
		errors = append(errors, fmt.Errorf("%q must be a valid port number (1-65536) or %q", k, "traffic-port"))
	}

	return
}

func validateAwsLbTargetGroupHealthCheckHealthyThreshold(v interface{}, k string) (ws []string, errors []error) {
	value := v.(int)
	if value < 2 || value > 10 {
		errors = append(errors, fmt.Errorf("%q must be an integer between 2 and 10", k))
	}
	return
}

func validateAwsLbTargetGroupHealthCheckTimeout(v interface{}, k string) (ws []string, errors []error) {
	value := v.(int)
	if value < 2 || value > 60 {
		errors = append(errors, fmt.Errorf("%q must be an integer between 2 and 60", k))
	}
	return
}

func validateAwsLbTargetGroupHealthCheckProtocol(v interface{}, k string) (ws []string, errors []error) {
	value := strings.ToLower(v.(string))
	if value == "http" || value == "https" || value == "tcp" {
		return
	}

	errors = append(errors, fmt.Errorf("%q must be either %q, %q or %q", k, "HTTP", "HTTPS", "TCP"))
	return
}

func validateAwsLbTargetGroupPort(v interface{}, k string) (ws []string, errors []error) {
	port := v.(int)
	if port < 1 || port > 65536 {
		errors = append(errors, fmt.Errorf("%q must be a valid port number (1-65536)", k))
	}
	return
}

func validateAwsLbTargetGroupProtocol(v interface{}, k string) (ws []string, errors []error) {
	protocol := strings.ToLower(v.(string))
	if protocol == "http" || protocol == "https" || protocol == "tcp" {
		return
	}

	errors = append(errors, fmt.Errorf("%q must be either %q, %q or %q", k, "HTTP", "HTTPS", "TCP"))
	return
}

func validateAwsLbTargetGroupDeregistrationDelay(v interface{}, k string) (ws []string, errors []error) {
	delay := v.(int)
	if delay < 0 || delay > 3600 {
		errors = append(errors, fmt.Errorf("%q must be in the range 0-3600 seconds", k))
	}
	return
}

func validateAwsLbTargetGroupStickinessType(v interface{}, k string) (ws []string, errors []error) {
	stickinessType := v.(string)
	if stickinessType != "lb_cookie" {
		errors = append(errors, fmt.Errorf("%q must have the value %q", k, "lb_cookie"))
	}
	return
}

func validateAwsLbTargetGroupStickinessCookieDuration(v interface{}, k string) (ws []string, errors []error) {
	duration := v.(int)
	if duration < 1 || duration > 604800 {
		errors = append(errors, fmt.Errorf("%q must be a between 1 second and 1 week (1-604800 seconds))", k))
	}
	return
}

func lbTargetGroupSuffixFromARN(arn *string) string {
	if arn == nil {
		return ""
	}

	if arnComponents := regexp.MustCompile(`arn:.*:targetgroup/(.*)`).FindAllStringSubmatch(*arn, -1); len(arnComponents) == 1 {
		if len(arnComponents[0]) == 2 {
			return fmt.Sprintf("targetgroup/%s", arnComponents[0][1])
		}
	}

	return ""
}

// flattenAwsLbTargetGroupResource takes a *elbv2.TargetGroup and populates all respective resource fields.
func flattenAwsLbTargetGroupResource(d *schema.ResourceData, meta interface{}, targetGroup *elbv2.TargetGroup) error {
	elbconn := meta.(*AWSClient).elbv2conn

	d.Set("arn", targetGroup.TargetGroupArn)
	d.Set("arn_suffix", lbTargetGroupSuffixFromARN(targetGroup.TargetGroupArn))
	d.Set("name", targetGroup.TargetGroupName)
	d.Set("port", targetGroup.Port)
	d.Set("protocol", targetGroup.Protocol)
	d.Set("vpc_id", targetGroup.VpcId)

	healthCheck := make(map[string]interface{})
	healthCheck["interval"] = *targetGroup.HealthCheckIntervalSeconds
	healthCheck["port"] = *targetGroup.HealthCheckPort
	healthCheck["protocol"] = *targetGroup.HealthCheckProtocol
	healthCheck["timeout"] = *targetGroup.HealthCheckTimeoutSeconds
	healthCheck["healthy_threshold"] = *targetGroup.HealthyThresholdCount
	healthCheck["unhealthy_threshold"] = *targetGroup.UnhealthyThresholdCount
	if *targetGroup.Protocol != "TCP" {
		healthCheck["path"] = *targetGroup.HealthCheckPath
		healthCheck["matcher"] = *targetGroup.Matcher.HttpCode
	}

	d.Set("health_check", []interface{}{healthCheck})

	attrResp, err := elbconn.DescribeTargetGroupAttributes(&elbv2.DescribeTargetGroupAttributesInput{
		TargetGroupArn: aws.String(d.Id()),
	})
	if err != nil {
		return errwrap.Wrapf("Error retrieving Target Group Attributes: {{err}}", err)
	}

	stickinessMap := map[string]interface{}{}
	for _, attr := range attrResp.Attributes {
		switch *attr.Key {
		case "stickiness.enabled":
			enabled, err := strconv.ParseBool(*attr.Value)
			if err != nil {
				return fmt.Errorf("Error converting stickiness.enabled to bool: %s", *attr.Value)
			}
			stickinessMap["enabled"] = enabled
		case "stickiness.type":
			stickinessMap["type"] = *attr.Value
		case "stickiness.lb_cookie.duration_seconds":
			duration, err := strconv.Atoi(*attr.Value)
			if err != nil {
				return fmt.Errorf("Error converting stickiness.lb_cookie.duration_seconds to int: %s", *attr.Value)
			}
			stickinessMap["cookie_duration"] = duration
		case "deregistration_delay.timeout_seconds":
			timeout, err := strconv.Atoi(*attr.Value)
			if err != nil {
				return fmt.Errorf("Error converting deregistration_delay.timeout_seconds to int: %s", *attr.Value)
			}
			d.Set("deregistration_delay", timeout)
		}
	}

	if err := d.Set("stickiness", []interface{}{stickinessMap}); err != nil {
		return err
	}

	tagsResp, err := elbconn.DescribeTags(&elbv2.DescribeTagsInput{
		ResourceArns: []*string{aws.String(d.Id())},
	})
	if err != nil {
		return errwrap.Wrapf("Error retrieving Target Group Tags: {{err}}", err)
	}
	for _, t := range tagsResp.TagDescriptions {
		if *t.ResourceArn == d.Id() {
			if err := d.Set("tags", tagsToMapELBv2(t.Tags)); err != nil {
				return err
			}
		}
	}

	return nil
}
