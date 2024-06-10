// Code generated by internal/generate/servicepackages/main.go; DO NOT EDIT.

package backup

import (
	"context"

	aws_sdkv1 "github.com/aws/aws-sdk-go/aws"
	endpoints_sdkv1 "github.com/aws/aws-sdk-go/aws/endpoints"
	session_sdkv1 "github.com/aws/aws-sdk-go/aws/session"
	backup_sdkv1 "github.com/aws/aws-sdk-go/service/backup"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type servicePackage struct{}

func (p *servicePackage) FrameworkDataSources(ctx context.Context) []*types.ServicePackageFrameworkDataSource {
	return []*types.ServicePackageFrameworkDataSource{}
}

func (p *servicePackage) FrameworkResources(ctx context.Context) []*types.ServicePackageFrameworkResource {
	return []*types.ServicePackageFrameworkResource{}
}

func (p *servicePackage) SDKDataSources(ctx context.Context) []*types.ServicePackageSDKDataSource {
	return []*types.ServicePackageSDKDataSource{
		{
			Factory:  DataSourceFramework,
			TypeName: "aws_backup_framework",
		},
		{
			Factory:  DataSourcePlan,
			TypeName: "aws_backup_plan",
		},
		{
			Factory:  DataSourceReportPlan,
			TypeName: "aws_backup_report_plan",
		},
		{
			Factory:  DataSourceSelection,
			TypeName: "aws_backup_selection",
		},
		{
			Factory:  DataSourceVault,
			TypeName: "aws_backup_vault",
		},
	}
}

func (p *servicePackage) SDKResources(ctx context.Context) []*types.ServicePackageSDKResource {
	return []*types.ServicePackageSDKResource{
		{
			Factory:  ResourceFramework,
			TypeName: "aws_backup_framework",
			Name:     "Framework",
			Tags: &types.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
		{
			Factory:  ResourceGlobalSettings,
			TypeName: "aws_backup_global_settings",
		},
		{
			Factory:  ResourcePlan,
			TypeName: "aws_backup_plan",
			Name:     "Plan",
			Tags: &types.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
		{
			Factory:  ResourceRegionSettings,
			TypeName: "aws_backup_region_settings",
		},
		{
			Factory:  ResourceReportPlan,
			TypeName: "aws_backup_report_plan",
			Name:     "Report Plan",
			Tags: &types.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
		{
			Factory:  ResourceSelection,
			TypeName: "aws_backup_selection",
		},
		{
			Factory:  ResourceVault,
			TypeName: "aws_backup_vault",
			Name:     "Vault",
			Tags: &types.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
			},
		},
		{
			Factory:  ResourceVaultLockConfiguration,
			TypeName: "aws_backup_vault_lock_configuration",
		},
		{
			Factory:  ResourceVaultNotifications,
			TypeName: "aws_backup_vault_notifications",
		},
		{
			Factory:  ResourceVaultPolicy,
			TypeName: "aws_backup_vault_policy",
		},
	}
}

func (p *servicePackage) ServicePackageName() string {
	return names.Backup
}

// NewConn returns a new AWS SDK for Go v1 client for this service package's AWS API.
func (p *servicePackage) NewConn(ctx context.Context, config map[string]any) (*backup_sdkv1.Backup, error) {
	sess := config[names.AttrSession].(*session_sdkv1.Session)

	cfg := aws_sdkv1.Config{}

	if endpoint := config[names.AttrEndpoint].(string); endpoint != "" {
		tflog.Debug(ctx, "setting endpoint", map[string]any{
			"tf_aws.endpoint": endpoint,
		})
		cfg.Endpoint = aws_sdkv1.String(endpoint)

		if sess.Config.UseFIPSEndpoint == endpoints_sdkv1.FIPSEndpointStateEnabled {
			tflog.Debug(ctx, "endpoint set, ignoring UseFIPSEndpoint setting")
			cfg.UseFIPSEndpoint = endpoints_sdkv1.FIPSEndpointStateDisabled
		}
	}

	return backup_sdkv1.New(sess.Copy(&cfg)), nil
}

func ServicePackage(ctx context.Context) conns.ServicePackage {
	return &servicePackage{}
}
