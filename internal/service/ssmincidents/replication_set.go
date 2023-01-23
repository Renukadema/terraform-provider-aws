package ssmincidents

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssmincidents"
	"github.com/aws/aws-sdk-go-v2/service/ssmincidents/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ResNameReplicationSet = "Replication Set"
)

func ResourceReplicationSet() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceReplicationSetCreate,
		ReadContext:   resourceReplicationSetRead,
		UpdateContext: resourceReplicationSetUpdate,
		DeleteContext: resourceReplicationSetDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(120 * time.Minute),
			Update: schema.DefaultTimeout(120 * time.Minute),
			Delete: schema.DefaultTimeout(120 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"region": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"kms_key_arn": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "DefaultKey",
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status_update_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status_message": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			// all other computed fields in alphabetic order
			"created_by": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deletion_protected": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"last_modified_by": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_modified_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: resourceReplicationSetImport,
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceReplicationSetCreate(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*conns.AWSClient).SSMIncidentsClient()

	input := &ssmincidents.CreateReplicationSetInput{
		Regions: ExpandRegions(d.Get("region").(*schema.Set).List()),
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	if len(tags) > 0 {
		input.Tags = tags.IgnoreAWS().Map()
	}

	createReplicationSetOutput, err := client.CreateReplicationSet(context, input)
	if err != nil {
		return create.DiagError(names.SSMIncidents, create.ErrActionCreating, ResNameReplicationSet, "", err)
	}

	if createReplicationSetOutput == nil {
		return create.DiagError(names.SSMIncidents, create.ErrActionCreating, ResNameReplicationSet, "", errors.New("empty output"))
	}

	d.SetId(aws.ToString(createReplicationSetOutput.Arn))

	getReplicationSetInput := &ssmincidents.GetReplicationSetInput{
		Arn: aws.String(d.Id()),
	}

	if err := ssmincidents.NewWaitForReplicationSetActiveWaiter(client).Wait(context, getReplicationSetInput, d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.DiagError(names.SSMIncidents, create.ErrActionWaitingForCreation, ResNameReplicationSet, d.Id(), err)
	}

	return resourceReplicationSetRead(context, d, meta)
}

func resourceReplicationSetRead(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*conns.AWSClient).SSMIncidentsClient()

	replicationSet, err := FindReplicationSetByID(context, client, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SSMIncidents ReplicationSet (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.SSMIncidents, create.ErrActionReading, ResNameReplicationSet, d.Id(), err)
	}

	d.Set("arn", replicationSet.Arn)
	d.Set("created_by", replicationSet.CreatedBy)
	d.Set("created_time", replicationSet.CreatedTime.String())
	d.Set("deletion_protected", replicationSet.DeletionProtected)
	d.Set("last_modified_by", replicationSet.LastModifiedBy)
	d.Set("last_modified_time", replicationSet.LastModifiedTime.String())
	d.Set("status", replicationSet.Status)

	if err := d.Set("region", FlattenRegions(replicationSet.RegionMap)); err != nil {
		return create.DiagError(names.SSMIncidents, create.ErrActionSetting, ResNameReplicationSet, d.Id(), err)
	}

	if diagErr := SetResourceDataTags(context, d, meta, client, ResNameReplicationSet); diagErr != nil {
		return diagErr
	}

	return nil
}

func resourceReplicationSetUpdate(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*conns.AWSClient).SSMIncidentsClient()

	if d.HasChanges("region") {
		input := &ssmincidents.UpdateReplicationSetInput{
			Arn: aws.String(d.Id()),
		}

		if err := updateRegionsInput(d, input); err != nil {
			return create.DiagError(names.SSMIncidents, create.ErrActionUpdating, ResNameReplicationSet, d.Id(), err)
		}

		log.Printf("[DEBUG] Updating SSMIncidents ReplicationSet (%s): %#v", d.Id(), input)
		_, err := client.UpdateReplicationSet(context, input)
		if err != nil {
			return create.DiagError(names.SSMIncidents, create.ErrActionUpdating, ResNameReplicationSet, d.Id(), err)
		}

		getReplicationSetInput := &ssmincidents.GetReplicationSetInput{
			Arn: aws.String(d.Id()),
		}

		if err := ssmincidents.NewWaitForReplicationSetActiveWaiter(client).Wait(context, getReplicationSetInput, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return create.DiagError(names.SSMIncidents, create.ErrActionWaitingForUpdate, ResNameReplicationSet, d.Id(), err)
		}
	}

	// tags_all does not detect changes when tag value is "" while this change is detected by tags
	if d.HasChanges("tags_all", "tags") {
		log.Printf("[DEBUG] Updating SSMIncidents ReplicationSet tags")

		if err := UpdateResourceTags(context, client, d); err != nil {
			return create.DiagError(names.SSMIncidents, create.ErrActionUpdating, ResNameReplicationSet, d.Id(), err)
		}
	}

	return resourceReplicationSetRead(context, d, meta)
}

func resourceReplicationSetDelete(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*conns.AWSClient).SSMIncidentsClient()

	log.Printf("[INFO] Deleting SSMIncidents ReplicationSet %s", d.Id())

	_, err := client.DeleteReplicationSet(context, &ssmincidents.DeleteReplicationSetInput{
		Arn: aws.String(d.Id()),
	})

	if err != nil {
		var notFoundError *types.ResourceNotFoundException
		if errors.As(err, &notFoundError) {
			return nil
		}

		return create.DiagError(names.SSMIncidents, create.ErrActionDeleting, ResNameReplicationSet, d.Id(), err)
	}

	getReplicationSetInput := &ssmincidents.GetReplicationSetInput{
		Arn: aws.String(d.Id()),
	}

	if err := ssmincidents.NewWaitForReplicationSetDeletedWaiter(client).Wait(context, getReplicationSetInput, d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.DiagError(names.SSMIncidents, create.ErrActionWaitingForDeletion, ResNameReplicationSet, d.Id(), err)
	}

	return nil
}

// converts a list of regions to a map which maps region name to kms key arn
func regionListToRegionMap(list []interface{}) map[string]string {
	return MapListToMap(
		list,
		func(region interface{}) string {
			return region.(map[string]interface{})["name"].(string)
		},
		func(region interface{}) string {
			return region.(map[string]interface{})["kms_key_arn"].(string)
		},
	)

	// regionMap := make(map[string]string)
	// for _, region := range list {
	// 	regionData := region.(map[string]interface{})
	// 	regionName := regionData["name"].(string)
	// 	regionMap[regionName] = regionData["kms_key_arn"].(string)
	// }

	// return regionMap
}

// updates UpdateReplicationSetInput to include any required actions
// invalid updates return errors from AWS Api
func updateRegionsInput(d *schema.ResourceData, input *ssmincidents.UpdateReplicationSetInput) error {
	old, new := d.GetChange("region")
	oldRegions := regionListToRegionMap(old.(*schema.Set).List())
	newRegions := regionListToRegionMap(new.(*schema.Set).List())

	for region, oldcmk := range oldRegions {
		if newcmk, ok := newRegions[region]; !ok {
			// this region has been destroyed

			action := &types.UpdateReplicationSetActionMemberDeleteRegionAction{
				Value: types.DeleteRegionAction{
					RegionName: aws.String(region),
				},
			}

			input.Actions = append(input.Actions, action)
		} else {
			if oldcmk != newcmk {
				return fmt.Errorf("error: Incident Manager does not support updating Customer Managed Keys. To do this, remove the region, and then re-create it with the new key")
			}
		}
	}

	for region, newcmk := range newRegions {
		if _, ok := oldRegions[region]; !ok {
			// this region is newly created

			action := &types.UpdateReplicationSetActionMemberAddRegionAction{
				Value: types.AddRegionAction{
					RegionName: aws.String(region),
				},
			}

			if newcmk != "DefaultKey" {
				action.Value.SseKmsKeyId = aws.String(newcmk)
			}

			input.Actions = append(input.Actions, action)
		}
	}

	return nil
}

func resourceReplicationSetImport(context context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	client := meta.(*conns.AWSClient).SSMIncidentsClient()

	arn, err := GetReplicationSetARN(context, client)

	if err != nil {
		return nil, err
	}

	d.SetId(arn)

	return []*schema.ResourceData{d}, nil
}
