package aws

import (
	"fmt"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/mediastore"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsMediaStoreContainer() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsMediaStoreContainerCreate,
		Read:   resourceAwsMediaStoreContainerRead,
		Update: resourceAwsMediaStoreContainerUpdate,
		Delete: resourceAwsMediaStoreContainerDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					if !regexp.MustCompile("^\\w+$").MatchString(value) {
						errors = append(errors, fmt.Errorf("%q must match \\w+", k))
					}
					return
				},
			},
			"policy": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validateIAMPolicyJson,
				DiffSuppressFunc: suppressEquivalentAwsPolicyDiffs,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsMediaStoreContainerCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).mediastoreconn

	input := &mediastore.CreateContainerInput{
		ContainerName: aws.String(d.Get("name").(string)),
	}

	_, err := conn.CreateContainer(input)
	if err != nil {
		return err
	}
	stateConf := &resource.StateChangeConf{
		Pending:    []string{mediastore.ContainerStatusCreating},
		Target:     []string{mediastore.ContainerStatusActive},
		Refresh:    mediaStoreContainerRefreshStatusFunc(conn, d.Get("name").(string)),
		Timeout:    10 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("[WARN] Error waiting for MediaStore Container status to be \"ACTIVE\": %s", err)
	}

	d.SetId(d.Get("name").(string))
	return resourceAwsMediaStoreContainerUpdate(d, meta)
}

func resourceAwsMediaStoreContainerRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).mediastoreconn

	input := &mediastore.DescribeContainerInput{
		ContainerName: aws.String(d.Id()),
	}
	resp, err := conn.DescribeContainer(input)
	if err != nil {
		return err
	}
	d.Set("arn", resp.Container.ARN)
	d.Set("endpoint", resp.Container.Endpoint)
	return nil
}

func resourceAwsMediaStoreContainerUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).mediastoreconn

	if !d.HasChange("policy") {
		return resourceAwsMediaStoreContainerRead(d, meta)
	}

	input := &mediastore.PutContainerPolicyInput{
		ContainerName: aws.String(d.Id()),
		Policy:        aws.String(d.Get("policy").(string)),
	}
	_, err := conn.PutContainerPolicy(input)
	if err != nil {
		return err
	}

	return resourceAwsMediaStoreContainerRead(d, meta)
}

func resourceAwsMediaStoreContainerDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).mediastoreconn

	input := &mediastore.DeleteContainerInput{
		ContainerName: aws.String(d.Id()),
	}
	_, err := conn.DeleteContainer(input)
	if err != nil {
		if ecrerr, ok := err.(awserr.Error); ok {
			switch ecrerr.Code() {
			case mediastore.ErrCodeContainerNotFoundException:
				d.SetId("")
				return nil
			}
		}
		return err
	}

	d.SetId("")
	return nil
}

func mediaStoreContainerRefreshStatusFunc(conn *mediastore.MediaStore, cn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &mediastore.DescribeContainerInput{
			ContainerName: aws.String(cn),
		}
		resp, err := conn.DescribeContainer(input)
		if err != nil {
			return nil, "failed", err
		}
		return resp, *resp.Container.Status, nil
	}
}
