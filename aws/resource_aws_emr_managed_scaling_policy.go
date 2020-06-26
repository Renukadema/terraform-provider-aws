package aws

import (
	"log"
	// "time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/emr"
	// "github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	// "github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAwsEMRManagedScalingPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEMRManagedScalingPolicyCreate,
		Read:   resourceAwsEMRManagedScalingPolicyRead,
		Delete: resourceAwsEMRManagedScalingPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"compute_limits": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem:     computeLimitsSchema(),
			},
		},
	}
}
func computeLimitsSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"unit_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validateAwsEmrComputeLimitsUnitType(),
			},
			"minimum_capacity_units": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
				Default:  1,
			},
			"maximum_capacity_units": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
				Default:  1,
			},
			"maximum_core_capacity_units": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
				Default:  1,
			},
			"maximum_ondemand_capacity_units": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
				Default:  1,
			},
		},
	}
}

func resourceAwsEMRManagedScalingPolicyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).emrconn

	if l := d.Get("compute_limits").(*schema.Set).List(); len(l) > 0 && l[0] != nil {
		cl := l[0].(map[string]interface{})
		computeLimits := &emr.ComputeLimits{
			UnitType:                     aws.String(cl["unit_type"].(string)),
			MinimumCapacityUnits:         aws.Int64(int64(cl["minimum_capacity_units"].(int))),
			MaximumCapacityUnits:         aws.Int64(int64(cl["maximum_capacity_units"].(int))),
			MaximumCoreCapacityUnits:     aws.Int64(int64(cl["maximum_core_capacity_units"].(int))),
			MaximumOnDemandCapacityUnits: aws.Int64(int64(cl["maximum_ondemand_capacity_units"].(int))),
		}
		managedScalingPolicy := &emr.ManagedScalingPolicy{
			ComputeLimits: computeLimits,
		}

		_, err := conn.PutManagedScalingPolicy(&emr.PutManagedScalingPolicyInput{
			ClusterId:            aws.String(d.Get("cluster_id").(string)),
			ManagedScalingPolicy: managedScalingPolicy,
		})

		if err != nil {
			log.Printf("[ERROR] EMR.PutManagedScalingPolicy %s", err)
			return err
		}
	}

	d.SetId(d.Get("cluster_id").(string))

	return nil
	// return resourceAwsEMRManagedScalingPolicyRead(d, meta)
}

func resourceAwsEMRManagedScalingPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).emrconn
	resp, err := conn.GetManagedScalingPolicy(&emr.GetManagedScalingPolicyInput{
		ClusterId: aws.String(d.Get("cluster_id").(string)),
	})
	if err != nil {
		if isAWSErr(err, "InvalidRequestException", "does not exist") {
			log.Printf("[WARN] EMR Managed Scaling Policy (%s) not found, removing from state", d.Get("cluster_id").(string))
			d.SetId("")
			return nil
		}
		return err
	}

	if resp.ManagedScalingPolicy != nil {
		attrs := make(map[string]interface{})
		attrs["unit_type"] = *resp.ManagedScalingPolicy.ComputeLimits.UnitType
		attrs["minimum_capacity_units"] = *resp.ManagedScalingPolicy.ComputeLimits.MinimumCapacityUnits
		attrs["maximum_capacity_units"] = *resp.ManagedScalingPolicy.ComputeLimits.MaximumCapacityUnits
		attrs["maximum_core_capacity_units"] = *resp.ManagedScalingPolicy.ComputeLimits.MaximumCoreCapacityUnits
		attrs["maximum_ondemand_capacity_units"] = *resp.ManagedScalingPolicy.ComputeLimits.MaximumOnDemandCapacityUnits

		computeLimits := make([]interface{}, 0)
		computeLimits = append(computeLimits, attrs)
		d.Set("compute_limits", computeLimits)
	}

	return nil
}

func resourceAwsEMRManagedScalingPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).emrconn
	_, err := conn.RemoveManagedScalingPolicy(&emr.RemoveManagedScalingPolicyInput{
		ClusterId: aws.String(d.Get("cluster_id").(string)),
	})
	if err != nil {
		if isAWSErr(err, "InvalidRequestException", "does not exist") {
			return nil
		}
		return err
	}
	return nil
}
