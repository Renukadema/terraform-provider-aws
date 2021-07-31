package aws

import (
	"fmt"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsIAMRole() *schema.Resource {
	return &schema.Resource{
		Read: ClientInitCrudBaseFunc(dataSourceAwsIAMRoleRead),

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"assume_role_policy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"path": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"permissions_boundary": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"unique_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"create_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"max_session_duration": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"tags": tagsSchemaComputed(),
		},
	}
}

func dataSourceAwsIAMRoleRead(d *schema.ResourceData, meta interface{}) error {
	iamconn := meta.(*AWSClient).iamconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	name := d.Get("name").(string)

	input := &iam.GetRoleInput{
		RoleName: aws.String(name),
	}

	output, err := iamconn.GetRole(input)
	if err != nil {
		return fmt.Errorf("error reading IAM Role (%s): %w", name, err)
	}

	d.Set("arn", output.Role.Arn)
	if err := d.Set("create_date", output.Role.CreateDate.Format(time.RFC3339)); err != nil {
		return fmt.Errorf("error setting create_date: %w", err)
	}
	d.Set("description", output.Role.Description)
	d.Set("max_session_duration", output.Role.MaxSessionDuration)
	d.Set("name", output.Role.RoleName)
	d.Set("path", output.Role.Path)
	d.Set("permissions_boundary", "")
	if output.Role.PermissionsBoundary != nil {
		d.Set("permissions_boundary", output.Role.PermissionsBoundary.PermissionsBoundaryArn)
	}
	d.Set("unique_id", output.Role.RoleId)
	d.Set("tags", keyvaluetags.IamKeyValueTags(output.Role.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map())

	assumRolePolicy, err := url.QueryUnescape(aws.StringValue(output.Role.AssumeRolePolicyDocument))
	if err != nil {
		return fmt.Errorf("error parsing assume role policy document: %w", err)
	}
	if err := d.Set("assume_role_policy", assumRolePolicy); err != nil {
		return fmt.Errorf("error setting assume_role_policy: %w", err)
	}

	d.SetId(name)

	return nil
}
