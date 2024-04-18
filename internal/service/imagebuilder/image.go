// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package imagebuilder

import (
	"context"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/imagebuilder"
	awstypes "github.com/aws/aws-sdk-go-v2/service/imagebuilder/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_imagebuilder_image", name="Image")
// @Tags(identifierAttribute="id")
func ResourceImage() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceImageCreate,
		ReadWithoutTimeout:   resourceImageRead,
		UpdateWithoutTimeout: resourceImageUpdate,
		DeleteWithoutTimeout: resourceImageDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"container_recipe_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^arn:aws[^:]*:imagebuilder:[^:]+:(?:\d{12}|aws):container-recipe/[0-9a-z_-]+/\d+\.\d+\.\d+$`), "valid container recipe ARN must be provided"),
				ExactlyOneOf: []string{"container_recipe_arn", "image_recipe_arn"},
			},
			"date_created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"distribution_configuration_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^arn:aws[^:]*:imagebuilder:[^:]+:(?:\d{12}|aws):distribution-configuration/[0-9a-z_-]+$`), "valid distribution configuration ARN must be provided"),
			},
			"enhanced_image_metadata_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  true,
			},
			"image_recipe_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^arn:aws[^:]*:imagebuilder:[^:]+:(?:\d{12}|aws):image-recipe/[0-9a-z_-]+/\d+\.\d+\.\d+$`), "valid image recipe ARN must be provided"),
				ExactlyOneOf: []string{"container_recipe_arn", "image_recipe_arn"},
			},
			"image_scanning_configuration": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ecr_configuration": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"container_tags": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"repository_name": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"image_scanning_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			"image_tests_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"image_tests_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							ForceNew: true,
							Default:  true,
						},
						"timeout_minutes": {
							Type:         schema.TypeInt,
							Optional:     true,
							ForceNew:     true,
							Default:      720,
							ValidateFunc: validation.IntBetween(60, 1440),
						},
					},
				},
			},
			"infrastructure_configuration_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^arn:aws[^:]*:imagebuilder:[^:]+:(?:\d{12}|aws):infrastructure-configuration/[0-9a-z_-]+$`), "valid infrastructure configuration ARN must be provided"),
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"os_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"output_resources": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"amis": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"account_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"description": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"image": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"region": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"containers": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"image_uris": {
										Type:     schema.TypeSet,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"region": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"platform": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceImageCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderClient(ctx)

	input := &imagebuilder.CreateImageInput{
		ClientToken:                  aws.String(id.UniqueId()),
		EnhancedImageMetadataEnabled: aws.Bool(d.Get("enhanced_image_metadata_enabled").(bool)),
		Tags:                         getTagsIn(ctx),
	}

	if v, ok := d.GetOk("container_recipe_arn"); ok {
		input.ContainerRecipeArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("distribution_configuration_arn"); ok {
		input.DistributionConfigurationArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("image_recipe_arn"); ok {
		input.ImageRecipeArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("image_scanning_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.ImageScanningConfiguration = expandImageScanningConfiguration(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("image_tests_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.ImageTestsConfiguration = expandImageTestConfiguration(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("infrastructure_configuration_arn"); ok {
		input.InfrastructureConfigurationArn = aws.String(v.(string))
	}

	output, err := conn.CreateImage(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Image Builder Image: %s", err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "creating Image Builder Image: empty response")
	}

	d.SetId(aws.ToString(output.ImageBuildVersionArn))

	if _, err := waitImageStatusAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Image Builder Image (%s) to become available: %s", d.Id(), err)
	}

	return append(diags, resourceImageRead(ctx, d, meta)...)
}

func resourceImageRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderClient(ctx)

	input := &imagebuilder.GetImageInput{
		ImageBuildVersionArn: aws.String(d.Id()),
	}

	output, err := conn.GetImage(ctx, input)

	if !d.IsNewResource() && errs.MessageContains(err, ResourceNotFoundException, "cannot be found") {
		log.Printf("[WARN] Image Builder Image (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Image Builder Image (%s): %s", d.Id(), err)
	}

	if output == nil || output.Image == nil {
		return sdkdiag.AppendErrorf(diags, "getting Image Builder Image (%s): empty response", d.Id())
	}

	image := output.Image

	d.Set("arn", image.Arn)
	d.Set("date_created", image.DateCreated)

	if image.ContainerRecipe != nil {
		d.Set("container_recipe_arn", image.ContainerRecipe.Arn)
	}

	if image.DistributionConfiguration != nil {
		d.Set("distribution_configuration_arn", image.DistributionConfiguration.Arn)
	}

	d.Set("enhanced_image_metadata_enabled", image.EnhancedImageMetadataEnabled)

	if image.ImageRecipe != nil {
		d.Set("image_recipe_arn", image.ImageRecipe.Arn)
	}

	if image.ImageScanningConfiguration != nil {
		d.Set("image_scanning_configuration", []interface{}{flattenImageScanningConfiguration(image.ImageScanningConfiguration)})
	} else {
		d.Set("image_scanning_configuration", nil)
	}

	if image.ImageTestsConfiguration != nil {
		d.Set("image_tests_configuration", []interface{}{flattenImageTestsConfiguration(image.ImageTestsConfiguration)})
	} else {
		d.Set("image_tests_configuration", nil)
	}

	if image.InfrastructureConfiguration != nil {
		d.Set("infrastructure_configuration_arn", image.InfrastructureConfiguration.Arn)
	}

	d.Set("name", image.Name)
	d.Set("platform", image.Platform)
	d.Set("os_version", image.OsVersion)

	if image.OutputResources != nil {
		d.Set("output_resources", []interface{}{flattenOutputResources(image.OutputResources)})
	} else {
		d.Set("output_resources", nil)
	}

	setTagsOut(ctx, image.Tags)

	d.Set("version", image.Version)

	return diags
}

func resourceImageUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceImageRead(ctx, d, meta)...)
}

func resourceImageDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderClient(ctx)

	input := &imagebuilder.DeleteImageInput{
		ImageBuildVersionArn: aws.String(d.Id()),
	}

	_, err := conn.DeleteImage(ctx, input)

	if errs.MessageContains(err, ResourceNotFoundException, "cannot be found") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Image Builder Image (%s): %s", d.Id(), err)
	}

	return diags
}

func flattenOutputResources(apiObject *awstypes.OutputResources) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Amis; v != nil {
		tfMap["amis"] = flattenAMIs(v)
	}

	if v := apiObject.Containers; v != nil {
		tfMap["containers"] = flattenContainers(v)
	}

	return tfMap
}

func flattenAMI(apiObject awstypes.Ami) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.AccountId; v != nil {
		tfMap["account_id"] = aws.ToString(v)
	}

	if v := apiObject.Description; v != nil {
		tfMap["description"] = aws.ToString(v)
	}

	if v := apiObject.Image; v != nil {
		tfMap["image"] = aws.ToString(v)
	}

	if v := apiObject.Name; v != nil {
		tfMap["name"] = aws.ToString(v)
	}

	if v := apiObject.Region; v != nil {
		tfMap["region"] = aws.ToString(v)
	}

	return tfMap
}

func flattenAMIs(apiObjects []awstypes.Ami) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenAMI(apiObject))
	}

	return tfList
}

func flattenContainer(apiObject awstypes.Container) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.ImageUris; v != nil {
		tfMap["image_uris"] = aws.StringSlice(v)
	}

	if v := apiObject.Region; v != nil {
		tfMap["region"] = aws.ToString(v)
	}

	return tfMap
}

func flattenContainers(apiObjects []awstypes.Container) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenContainer(apiObject))
	}

	return tfList
}
