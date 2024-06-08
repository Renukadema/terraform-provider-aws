// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	aws_sdkv2 "github.com/aws/aws-sdk-go-v2/aws"
	ec2_sdkv2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	InstanceReadyTimeout = 10 * time.Minute
	InstanceStartTimeout = 10 * time.Minute
	InstanceStopTimeout  = 10 * time.Minute

	// General timeout for IAM resource change to propagate.
	// See https://docs.aws.amazon.com/IAM/latest/UserGuide/troubleshoot_general.html#troubleshoot_general_eventual-consistency.
	// We have settled on 2 minutes as the best timeout value.
	iamPropagationTimeout = 2 * time.Minute

	// General timeout for EC2 resource changes to propagate.
	// See https://docs.aws.amazon.com/AWSEC2/latest/APIReference/query-api-troubleshooting.html#eventual-consistency.
	ec2PropagationTimeout = 5 * time.Minute // nosemgrep:ci.ec2-in-const-name, ci.ec2-in-var-name

	RouteNotFoundChecks                        = 1000 // Should exceed any reasonable custom timeout value.
	RouteTableNotFoundChecks                   = 1000 // Should exceed any reasonable custom timeout value.
	RouteTableAssociationCreatedNotFoundChecks = 1000 // Should exceed any reasonable custom timeout value.
	SecurityGroupNotFoundChecks                = 1000 // Should exceed any reasonable custom timeout value.
	InternetGatewayNotFoundChecks              = 1000 // Should exceed any reasonable custom timeout value.
)

const (
	AvailabilityZoneGroupOptInStatusTimeout = 10 * time.Minute
)

func WaitAvailabilityZoneGroupOptedIn(ctx context.Context, conn *ec2.EC2, name string) (*ec2.AvailabilityZone, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ec2.AvailabilityZoneOptInStatusNotOptedIn},
		Target:  []string{ec2.AvailabilityZoneOptInStatusOptedIn},
		Refresh: StatusAvailabilityZoneGroupOptInStatus(ctx, conn, name),
		Timeout: AvailabilityZoneGroupOptInStatusTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.AvailabilityZone); ok {
		return output, err
	}

	return nil, err
}

func WaitAvailabilityZoneGroupNotOptedIn(ctx context.Context, conn *ec2.EC2, name string) (*ec2.AvailabilityZone, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ec2.AvailabilityZoneOptInStatusOptedIn},
		Target:  []string{ec2.AvailabilityZoneOptInStatusNotOptedIn},
		Refresh: StatusAvailabilityZoneGroupOptInStatus(ctx, conn, name),
		Timeout: AvailabilityZoneGroupOptInStatusTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.AvailabilityZone); ok {
		return output, err
	}

	return nil, err
}

const (
	CapacityReservationActiveTimeout  = 2 * time.Minute
	CapacityReservationDeletedTimeout = 2 * time.Minute
)

func WaitCapacityReservationActive(ctx context.Context, conn *ec2.EC2, id string) (*ec2.CapacityReservation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ec2.CapacityReservationStatePending},
		Target:  []string{ec2.CapacityReservationStateActive},
		Refresh: StatusCapacityReservationState(ctx, conn, id),
		Timeout: CapacityReservationActiveTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.CapacityReservation); ok {
		return output, err
	}

	return nil, err
}

func WaitCapacityReservationDeleted(ctx context.Context, conn *ec2.EC2, id string) (*ec2.CapacityReservation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ec2.CapacityReservationStateActive},
		Target:  []string{},
		Refresh: StatusCapacityReservationState(ctx, conn, id),
		Timeout: CapacityReservationDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.CapacityReservation); ok {
		return output, err
	}

	return nil, err
}

const (
	// Maximum amount of time to wait for a LocalGatewayRouteTableVpcAssociation to return Associated
	LocalGatewayRouteTableVPCAssociationAssociatedTimeout = 5 * time.Minute

	// Maximum amount of time to wait for a LocalGatewayRouteTableVpcAssociation to return Disassociated
	LocalGatewayRouteTableVPCAssociationDisassociatedTimeout = 5 * time.Minute
)

// WaitLocalGatewayRouteTableVPCAssociationAssociated waits for a LocalGatewayRouteTableVpcAssociation to return Associated
func WaitLocalGatewayRouteTableVPCAssociationAssociated(ctx context.Context, conn *ec2.EC2, localGatewayRouteTableVpcAssociationID string) (*ec2.LocalGatewayRouteTableVpcAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ec2.RouteTableAssociationStateCodeAssociating},
		Target:  []string{ec2.RouteTableAssociationStateCodeAssociated},
		Refresh: StatusLocalGatewayRouteTableVPCAssociationState(ctx, conn, localGatewayRouteTableVpcAssociationID),
		Timeout: LocalGatewayRouteTableVPCAssociationAssociatedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.LocalGatewayRouteTableVpcAssociation); ok {
		return output, err
	}

	return nil, err
}

// WaitLocalGatewayRouteTableVPCAssociationDisassociated waits for a LocalGatewayRouteTableVpcAssociation to return Disassociated
func WaitLocalGatewayRouteTableVPCAssociationDisassociated(ctx context.Context, conn *ec2.EC2, localGatewayRouteTableVpcAssociationID string) (*ec2.LocalGatewayRouteTableVpcAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ec2.RouteTableAssociationStateCodeDisassociating},
		Target:  []string{ec2.RouteTableAssociationStateCodeDisassociated},
		Refresh: StatusLocalGatewayRouteTableVPCAssociationState(ctx, conn, localGatewayRouteTableVpcAssociationID),
		Timeout: LocalGatewayRouteTableVPCAssociationAssociatedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.LocalGatewayRouteTableVpcAssociation); ok {
		return output, err
	}

	return nil, err
}

func WaitFleet(ctx context.Context, conn *ec2.EC2, id string, pending, target []string, timeout, delay time.Duration) (*ec2.FleetData, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    pending,
		Target:     target,
		Refresh:    StatusFleetState(ctx, conn, id),
		Timeout:    timeout,
		Delay:      delay,
		MinTimeout: 1 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.FleetData); ok {
		return output, err
	}

	return nil, err
}

func WaitImageAvailable(ctx context.Context, conn *ec2.EC2, id string, timeout time.Duration) (*ec2.Image, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{ec2.ImageStatePending},
		Target:     []string{ec2.ImageStateAvailable},
		Refresh:    StatusImageState(ctx, conn, id),
		Timeout:    timeout,
		Delay:      amiRetryDelay,
		MinTimeout: amiRetryMinTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.Image); ok {
		if stateReason := output.StateReason; stateReason != nil {
			tfresource.SetLastError(err, errors.New(aws.StringValue(stateReason.Message)))
		}

		return output, err
	}

	return nil, err
}

func WaitImageDeleted(ctx context.Context, conn *ec2.EC2, id string, timeout time.Duration) (*ec2.Image, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{ec2.ImageStateAvailable, ec2.ImageStateFailed, ec2.ImageStatePending},
		Target:     []string{},
		Refresh:    StatusImageState(ctx, conn, id),
		Timeout:    timeout,
		Delay:      amiRetryDelay,
		MinTimeout: amiRetryMinTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.Image); ok {
		if stateReason := output.StateReason; stateReason != nil {
			tfresource.SetLastError(err, errors.New(aws.StringValue(stateReason.Message)))
		}

		return output, err
	}

	return nil, err
}

func WaitInstanceIAMInstanceProfileUpdated(ctx context.Context, conn *ec2.EC2, instanceID string, expectedValue string) (*ec2.Instance, error) {
	stateConf := &retry.StateChangeConf{
		Target:     []string{expectedValue},
		Refresh:    StatusInstanceIAMInstanceProfile(ctx, conn, instanceID),
		Timeout:    ec2PropagationTimeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.Instance); ok {
		return output, err
	}

	return nil, err
}

func WaitInstanceCapacityReservationSpecificationUpdated(ctx context.Context, conn *ec2.EC2, instanceID string, expectedValue *ec2.CapacityReservationSpecification) (*ec2.Instance, error) {
	stateConf := &retry.StateChangeConf{
		Target:     []string{strconv.FormatBool(true)},
		Refresh:    StatusInstanceCapacityReservationSpecificationEquals(ctx, conn, instanceID, expectedValue),
		Timeout:    ec2PropagationTimeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.Instance); ok {
		return output, err
	}

	return nil, err
}

func WaitInstanceMaintenanceOptionsAutoRecoveryUpdated(ctx context.Context, conn *ec2.EC2, id, expectedValue string, timeout time.Duration) (*ec2.InstanceMaintenanceOptions, error) {
	stateConf := &retry.StateChangeConf{
		Target:     []string{expectedValue},
		Refresh:    StatusInstanceMaintenanceOptionsAutoRecovery(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.InstanceMaintenanceOptions); ok {
		return output, err
	}

	return nil, err
}

func WaitInstanceMetadataOptionsApplied(ctx context.Context, conn *ec2.EC2, id string, timeout time.Duration) (*ec2.InstanceMetadataOptionsResponse, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{ec2.InstanceMetadataOptionsStatePending},
		Target:     []string{ec2.InstanceMetadataOptionsStateApplied},
		Refresh:    StatusInstanceMetadataOptionsState(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.InstanceMetadataOptionsResponse); ok {
		return output, err
	}

	return nil, err
}

func WaitInstanceRootBlockDeviceDeleteOnTerminationUpdated(ctx context.Context, conn *ec2.EC2, id string, expectedValue bool, timeout time.Duration) (*ec2.EbsInstanceBlockDevice, error) {
	stateConf := &retry.StateChangeConf{
		Target:     []string{strconv.FormatBool(expectedValue)},
		Refresh:    StatusInstanceRootBlockDeviceDeleteOnTermination(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.EbsInstanceBlockDevice); ok {
		return output, err
	}

	return nil, err
}

const ManagedPrefixListEntryCreateTimeout = 5 * time.Minute

func WaitSecurityGroupCreated(ctx context.Context, conn *ec2.EC2, id string, timeout time.Duration) (*ec2.SecurityGroup, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{SecurityGroupStatusCreated},
		Refresh:                   StatusSecurityGroup(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            SecurityGroupNotFoundChecks,
		ContinuousTargetOccurence: 3,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.SecurityGroup); ok {
		return output, err
	}

	return nil, err
}

const (
	SubnetIPv6CIDRBlockAssociationCreatedTimeout = 3 * time.Minute
	SubnetIPv6CIDRBlockAssociationDeletedTimeout = 3 * time.Minute
)

func WaitSubnetAvailable(ctx context.Context, conn *ec2.EC2, id string, timeout time.Duration) (*ec2.Subnet, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ec2.SubnetStatePending},
		Target:  []string{ec2.SubnetStateAvailable},
		Refresh: StatusSubnetState(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.Subnet); ok {
		return output, err
	}

	return nil, err
}

func WaitSubnetIPv6CIDRBlockAssociationCreated(ctx context.Context, conn *ec2.EC2, id string) (*ec2.SubnetCidrBlockState, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ec2.SubnetCidrBlockStateCodeAssociating, ec2.SubnetCidrBlockStateCodeDisassociated, ec2.SubnetCidrBlockStateCodeFailing},
		Target:  []string{ec2.SubnetCidrBlockStateCodeAssociated},
		Refresh: StatusSubnetIPv6CIDRBlockAssociationState(ctx, conn, id),
		Timeout: SubnetIPv6CIDRBlockAssociationCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.SubnetCidrBlockState); ok {
		if state := aws.StringValue(output.State); state == ec2.SubnetCidrBlockStateCodeFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

func WaitSubnetIPv6CIDRBlockAssociationDeleted(ctx context.Context, conn *ec2.EC2, id string) (*ec2.SubnetCidrBlockState, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ec2.SubnetCidrBlockStateCodeAssociated, ec2.SubnetCidrBlockStateCodeDisassociating, ec2.SubnetCidrBlockStateCodeFailing},
		Target:  []string{},
		Refresh: StatusSubnetIPv6CIDRBlockAssociationState(ctx, conn, id),
		Timeout: SubnetIPv6CIDRBlockAssociationDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.SubnetCidrBlockState); ok {
		if state := aws.StringValue(output.State); state == ec2.SubnetCidrBlockStateCodeFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

func waitSubnetAssignIPv6AddressOnCreationUpdated(ctx context.Context, conn *ec2.EC2, subnetID string, expectedValue bool) (*ec2.Subnet, error) {
	stateConf := &retry.StateChangeConf{
		Target:     []string{strconv.FormatBool(expectedValue)},
		Refresh:    StatusSubnetAssignIPv6AddressOnCreation(ctx, conn, subnetID),
		Timeout:    ec2PropagationTimeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.Subnet); ok {
		return output, err
	}

	return nil, err
}

func waitSubnetEnableLniAtDeviceIndexUpdated(ctx context.Context, conn *ec2.EC2, subnetID string, expectedValue int64) (*ec2.Subnet, error) {
	stateConf := &retry.StateChangeConf{
		Target:     []string{strconv.FormatInt(expectedValue, 10)},
		Refresh:    StatusSubnetEnableLniAtDeviceIndex(ctx, conn, subnetID),
		Timeout:    ec2PropagationTimeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.Subnet); ok {
		return output, err
	}

	return nil, err
}

func waitSubnetEnableDNS64Updated(ctx context.Context, conn *ec2.EC2, subnetID string, expectedValue bool) (*ec2.Subnet, error) {
	stateConf := &retry.StateChangeConf{
		Target:     []string{strconv.FormatBool(expectedValue)},
		Refresh:    StatusSubnetEnableDNS64(ctx, conn, subnetID),
		Timeout:    ec2PropagationTimeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.Subnet); ok {
		return output, err
	}

	return nil, err
}

func waitSubnetEnableResourceNameDNSAAAARecordOnLaunchUpdated(ctx context.Context, conn *ec2.EC2, subnetID string, expectedValue bool) (*ec2.Subnet, error) {
	stateConf := &retry.StateChangeConf{
		Target:     []string{strconv.FormatBool(expectedValue)},
		Refresh:    StatusSubnetEnableResourceNameDNSAAAARecordOnLaunch(ctx, conn, subnetID),
		Timeout:    ec2PropagationTimeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.Subnet); ok {
		return output, err
	}

	return nil, err
}

func waitSubnetEnableResourceNameDNSARecordOnLaunchUpdated(ctx context.Context, conn *ec2.EC2, subnetID string, expectedValue bool) (*ec2.Subnet, error) {
	stateConf := &retry.StateChangeConf{
		Target:     []string{strconv.FormatBool(expectedValue)},
		Refresh:    StatusSubnetEnableResourceNameDNSARecordOnLaunch(ctx, conn, subnetID),
		Timeout:    ec2PropagationTimeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.Subnet); ok {
		return output, err
	}

	return nil, err
}

func WaitSubnetMapCustomerOwnedIPOnLaunchUpdated(ctx context.Context, conn *ec2.EC2, subnetID string, expectedValue bool) (*ec2.Subnet, error) {
	stateConf := &retry.StateChangeConf{
		Target:     []string{strconv.FormatBool(expectedValue)},
		Refresh:    StatusSubnetMapCustomerOwnedIPOnLaunch(ctx, conn, subnetID),
		Timeout:    ec2PropagationTimeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.Subnet); ok {
		return output, err
	}

	return nil, err
}

func WaitSubnetMapPublicIPOnLaunchUpdated(ctx context.Context, conn *ec2.EC2, subnetID string, expectedValue bool) (*ec2.Subnet, error) {
	stateConf := &retry.StateChangeConf{
		Target:     []string{strconv.FormatBool(expectedValue)},
		Refresh:    StatusSubnetMapPublicIPOnLaunch(ctx, conn, subnetID),
		Timeout:    ec2PropagationTimeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.Subnet); ok {
		return output, err
	}

	return nil, err
}

func WaitSubnetPrivateDNSHostnameTypeOnLaunchUpdated(ctx context.Context, conn *ec2.EC2, subnetID string, expectedValue string) (*ec2.Subnet, error) {
	stateConf := &retry.StateChangeConf{
		Target:     []string{expectedValue},
		Refresh:    StatusSubnetPrivateDNSHostnameTypeOnLaunch(ctx, conn, subnetID),
		Timeout:    ec2PropagationTimeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.Subnet); ok {
		return output, err
	}

	return nil, err
}

func WaitVolumeCreated(ctx context.Context, conn *ec2_sdkv2.Client, id string, timeout time.Duration) (*awstypes.Volume, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.VolumeStateCreating),
		Target:     enum.Slice(awstypes.VolumeStateAvailable),
		Refresh:    StatusVolumeState(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Volume); ok {
		return output, err
	}

	return nil, err
}

func WaitVolumeDeleted(ctx context.Context, conn *ec2_sdkv2.Client, id string, timeout time.Duration) (*awstypes.Volume, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{ec2.VolumeStateDeleting},
		Target:     []string{},
		Refresh:    StatusVolumeState(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Volume); ok {
		return output, err
	}

	return nil, err
}

func WaitVolumeUpdated(ctx context.Context, conn *ec2_sdkv2.Client, id string, timeout time.Duration) (*awstypes.Volume, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.VolumeStateCreating, awstypes.VolumeState(awstypes.VolumeModificationStateModifying)),
		Target:     enum.Slice(awstypes.VolumeStateAvailable, awstypes.VolumeStateInUse),
		Refresh:    StatusVolumeState(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Volume); ok {
		return output, err
	}

	return nil, err
}

func WaitVolumeAttachmentCreated(ctx context.Context, conn *ec2.EC2, volumeID, instanceID, deviceName string, timeout time.Duration) (*ec2.VolumeAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{ec2.VolumeAttachmentStateAttaching},
		Target:     []string{ec2.VolumeAttachmentStateAttached},
		Refresh:    StatusVolumeAttachmentState(ctx, conn, volumeID, instanceID, deviceName),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.VolumeAttachment); ok {
		return output, err
	}

	return nil, err
}

func WaitVolumeModificationComplete(ctx context.Context, conn *ec2.EC2, id string, timeout time.Duration) (*ec2.VolumeModification, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ec2.VolumeModificationStateModifying},
		// The volume is useable once the state is "optimizing", but will not be at full performance.
		// Optimization can take hours. e.g. a full 1 TiB drive takes approximately 6 hours to optimize,
		// according to https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/monitoring-volume-modifications.html.
		Target:     []string{ec2.VolumeModificationStateCompleted, ec2.VolumeModificationStateOptimizing},
		Refresh:    StatusVolumeModificationState(ctx, conn, id),
		Timeout:    timeout,
		Delay:      30 * time.Second,
		MinTimeout: 30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.VolumeModification); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusMessage)))

		return output, err
	}

	return nil, err
}

const (
	vpcCreatedTimeout = 10 * time.Minute
	vpcDeletedTimeout = 5 * time.Minute
)

func WaitVPCCIDRBlockAssociationCreated(ctx context.Context, conn *ec2.EC2, id string, timeout time.Duration) (*ec2.VpcCidrBlockState, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{ec2.VpcCidrBlockStateCodeAssociating, ec2.VpcCidrBlockStateCodeDisassociated, ec2.VpcCidrBlockStateCodeFailing},
		Target:     []string{ec2.VpcCidrBlockStateCodeAssociated},
		Refresh:    StatusVPCCIDRBlockAssociationState(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.VpcCidrBlockState); ok {
		if state := aws.StringValue(output.State); state == ec2.VpcCidrBlockStateCodeFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

func WaitVPCCIDRBlockAssociationDeleted(ctx context.Context, conn *ec2.EC2, id string, timeout time.Duration) (*ec2.VpcCidrBlockState, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{ec2.VpcCidrBlockStateCodeAssociated, ec2.VpcCidrBlockStateCodeDisassociating, ec2.VpcCidrBlockStateCodeFailing},
		Target:     []string{},
		Refresh:    StatusVPCCIDRBlockAssociationState(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.VpcCidrBlockState); ok {
		if state := aws.StringValue(output.State); state == ec2.VpcCidrBlockStateCodeFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

const (
	vpcIPv6CIDRBlockAssociationCreatedTimeout = 10 * time.Minute
	vpcIPv6CIDRBlockAssociationDeletedTimeout = 5 * time.Minute
)

func WaitVPCIPv6CIDRBlockAssociationCreated(ctx context.Context, conn *ec2.EC2, id string, timeout time.Duration) (*ec2.VpcCidrBlockState, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{ec2.VpcCidrBlockStateCodeAssociating, ec2.VpcCidrBlockStateCodeDisassociated, ec2.VpcCidrBlockStateCodeFailing},
		Target:     []string{ec2.VpcCidrBlockStateCodeAssociated},
		Refresh:    StatusVPCIPv6CIDRBlockAssociationState(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.VpcCidrBlockState); ok {
		if state := aws.StringValue(output.State); state == ec2.VpcCidrBlockStateCodeFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

func WaitVPCIPv6CIDRBlockAssociationDeleted(ctx context.Context, conn *ec2.EC2, id string, timeout time.Duration) (*ec2.VpcCidrBlockState, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{ec2.VpcCidrBlockStateCodeAssociated, ec2.VpcCidrBlockStateCodeDisassociating, ec2.VpcCidrBlockStateCodeFailing},
		Target:     []string{},
		Refresh:    StatusVPCIPv6CIDRBlockAssociationState(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.VpcCidrBlockState); ok {
		if state := aws.StringValue(output.State); state == ec2.VpcCidrBlockStateCodeFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

func WaitVPCPeeringConnectionActive(ctx context.Context, conn *ec2.EC2, id string, timeout time.Duration) (*ec2.VpcPeeringConnection, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ec2.VpcPeeringConnectionStateReasonCodeInitiatingRequest, ec2.VpcPeeringConnectionStateReasonCodeProvisioning},
		Target:  []string{ec2.VpcPeeringConnectionStateReasonCodeActive, ec2.VpcPeeringConnectionStateReasonCodePendingAcceptance},
		Refresh: StatusVPCPeeringConnectionActive(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.VpcPeeringConnection); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.Status.Message)))

		return output, err
	}

	return nil, err
}

func WaitVPCPeeringConnectionDeleted(ctx context.Context, conn *ec2.EC2, id string, timeout time.Duration) (*ec2.VpcPeeringConnection, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			ec2.VpcPeeringConnectionStateReasonCodeActive,
			ec2.VpcPeeringConnectionStateReasonCodeDeleting,
			ec2.VpcPeeringConnectionStateReasonCodePendingAcceptance,
		},
		Target:  []string{},
		Refresh: StatusVPCPeeringConnectionDeleted(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.VpcPeeringConnection); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.Status.Message)))

		return output, err
	}

	return nil, err
}

func WaitNATGatewayCreated(ctx context.Context, conn *ec2.EC2, id string, timeout time.Duration) (*ec2.NatGateway, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ec2.NatGatewayStatePending},
		Target:  []string{ec2.NatGatewayStateAvailable},
		Refresh: StatusNATGatewayState(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.NatGateway); ok {
		if state := aws.StringValue(output.State); state == ec2.NatGatewayStateFailed {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(output.FailureCode), aws.StringValue(output.FailureMessage)))
		}

		return output, err
	}

	return nil, err
}

func WaitNATGatewayDeleted(ctx context.Context, conn *ec2.EC2, id string, timeout time.Duration) (*ec2.NatGateway, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{ec2.NatGatewayStateDeleting},
		Target:     []string{},
		Refresh:    StatusNATGatewayState(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.NatGateway); ok {
		if state := aws.StringValue(output.State); state == ec2.NatGatewayStateFailed {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(output.FailureCode), aws.StringValue(output.FailureMessage)))
		}

		return output, err
	}

	return nil, err
}

func WaitNATGatewayAddressAssigned(ctx context.Context, conn *ec2.EC2, natGatewayID, privateIP string, timeout time.Duration) (*ec2.NatGatewayAddress, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ec2.NatGatewayAddressStatusAssigning},
		Target:  []string{ec2.NatGatewayAddressStatusSucceeded},
		Refresh: StatusNATGatewayAddressByNATGatewayIDAndPrivateIP(ctx, conn, natGatewayID, privateIP),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.NatGatewayAddress); ok {
		if status := aws.StringValue(output.Status); status == ec2.NatGatewayAddressStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.FailureMessage)))
		}

		return output, err
	}

	return nil, err
}

func WaitNATGatewayAddressAssociated(ctx context.Context, conn *ec2.EC2, natGatewayID, allocationID string, timeout time.Duration) (*ec2.NatGatewayAddress, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ec2.NatGatewayAddressStatusAssociating},
		Target:  []string{ec2.NatGatewayAddressStatusSucceeded},
		Refresh: StatusNATGatewayAddressByNATGatewayIDAndAllocationID(ctx, conn, natGatewayID, allocationID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.NatGatewayAddress); ok {
		if status := aws.StringValue(output.Status); status == ec2.NatGatewayAddressStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.FailureMessage)))
		}

		return output, err
	}

	return nil, err
}

func WaitNATGatewayAddressDisassociated(ctx context.Context, conn *ec2.EC2, natGatewayID, allocationID string, timeout time.Duration) (*ec2.NatGatewayAddress, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ec2.NatGatewayAddressStatusSucceeded, ec2.NatGatewayAddressStatusDisassociating},
		Target:  []string{},
		Refresh: StatusNATGatewayAddressByNATGatewayIDAndAllocationID(ctx, conn, natGatewayID, allocationID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.NatGatewayAddress); ok {
		if status := aws.StringValue(output.Status); status == ec2.NatGatewayAddressStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.FailureMessage)))
		}

		return output, err
	}

	return nil, err
}

func WaitNATGatewayAddressUnassigned(ctx context.Context, conn *ec2.EC2, natGatewayID, privateIP string, timeout time.Duration) (*ec2.NatGatewayAddress, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ec2.NatGatewayAddressStatusUnassigning},
		Target:  []string{},
		Refresh: StatusNATGatewayAddressByNATGatewayIDAndPrivateIP(ctx, conn, natGatewayID, privateIP),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.NatGatewayAddress); ok {
		if status := aws.StringValue(output.Status); status == ec2.NatGatewayAddressStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.FailureMessage)))
		}

		return output, err
	}

	return nil, err
}

func waitEIPDomainNameAttributeUpdated(ctx context.Context, conn *ec2_sdkv2.Client, allocationID string, timeout time.Duration) (*awstypes.AddressAttribute, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{PTRUpdateStatusPending},
		Target:  []string{""},
		Timeout: timeout,
		Refresh: statusEIPDomainNameAttribute(ctx, conn, allocationID),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.AddressAttribute); ok {
		if v := output.PtrRecordUpdate; v != nil {
			tfresource.SetLastError(err, errors.New(aws_sdkv2.ToString(v.Reason)))
		}

		return output, err
	}

	return nil, err
}

func waitEIPDomainNameAttributeDeleted(ctx context.Context, conn *ec2_sdkv2.Client, allocationID string, timeout time.Duration) (*awstypes.AddressAttribute, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{PTRUpdateStatusPending},
		Target:  []string{},
		Timeout: timeout,
		Refresh: statusEIPDomainNameAttribute(ctx, conn, allocationID),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.AddressAttribute); ok {
		if v := output.PtrRecordUpdate; v != nil {
			tfresource.SetLastError(err, errors.New(aws_sdkv2.ToString(v.Reason)))
		}

		return output, err
	}

	return nil, err
}

func WaitHostCreated(ctx context.Context, conn *ec2.EC2, id string, timeout time.Duration) (*ec2.Host, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ec2.AllocationStatePending},
		Target:  []string{ec2.AllocationStateAvailable},
		Timeout: timeout,
		Refresh: StatusHostState(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.Host); ok {
		return output, err
	}

	return nil, err
}

func WaitHostUpdated(ctx context.Context, conn *ec2.EC2, id string, timeout time.Duration) (*ec2.Host, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ec2.AllocationStatePending},
		Target:  []string{ec2.AllocationStateAvailable},
		Timeout: timeout,
		Refresh: StatusHostState(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.Host); ok {
		return output, err
	}

	return nil, err
}

func WaitHostDeleted(ctx context.Context, conn *ec2.EC2, id string, timeout time.Duration) (*ec2.Host, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ec2.AllocationStateAvailable},
		Target:  []string{},
		Timeout: timeout,
		Refresh: StatusHostState(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.Host); ok {
		return output, err
	}

	return nil, err
}

const (
	dhcpOptionSetDeletedTimeout = 3 * time.Minute
)

func WaitInternetGatewayAttached(ctx context.Context, conn *ec2.EC2, internetGatewayID, vpcID string, timeout time.Duration) (*ec2.InternetGatewayAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending:        []string{ec2.AttachmentStatusAttaching},
		Target:         []string{InternetGatewayAttachmentStateAvailable},
		Timeout:        timeout,
		NotFoundChecks: InternetGatewayNotFoundChecks,
		Refresh:        StatusInternetGatewayAttachmentState(ctx, conn, internetGatewayID, vpcID),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.InternetGatewayAttachment); ok {
		return output, err
	}

	return nil, err
}

func WaitInternetGatewayDetached(ctx context.Context, conn *ec2.EC2, internetGatewayID, vpcID string, timeout time.Duration) (*ec2.InternetGatewayAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{InternetGatewayAttachmentStateAvailable, ec2.AttachmentStatusDetaching},
		Target:  []string{},
		Timeout: timeout,
		Refresh: StatusInternetGatewayAttachmentState(ctx, conn, internetGatewayID, vpcID),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.InternetGatewayAttachment); ok {
		return output, err
	}

	return nil, err
}

const (
	ManagedPrefixListTimeout = 15 * time.Minute
)

func WaitManagedPrefixListCreated(ctx context.Context, conn *ec2.EC2, id string) (*ec2.ManagedPrefixList, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ec2.PrefixListStateCreateInProgress},
		Target:  []string{ec2.PrefixListStateCreateComplete},
		Timeout: ManagedPrefixListTimeout,
		Refresh: StatusManagedPrefixListState(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.ManagedPrefixList); ok {
		if state := aws.StringValue(output.State); state == ec2.PrefixListStateCreateFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StateMessage)))
		}

		return output, err
	}

	return nil, err
}

func WaitManagedPrefixListModified(ctx context.Context, conn *ec2.EC2, id string) (*ec2.ManagedPrefixList, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ec2.PrefixListStateModifyInProgress},
		Target:  []string{ec2.PrefixListStateModifyComplete},
		Timeout: ManagedPrefixListTimeout,
		Refresh: StatusManagedPrefixListState(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.ManagedPrefixList); ok {
		if state := aws.StringValue(output.State); state == ec2.PrefixListStateModifyFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StateMessage)))
		}

		return output, err
	}

	return nil, err
}

func WaitManagedPrefixListDeleted(ctx context.Context, conn *ec2.EC2, id string) (*ec2.ManagedPrefixList, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ec2.PrefixListStateDeleteInProgress},
		Target:  []string{},
		Timeout: ManagedPrefixListTimeout,
		Refresh: StatusManagedPrefixListState(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.ManagedPrefixList); ok {
		if state := aws.StringValue(output.State); state == ec2.PrefixListStateDeleteFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StateMessage)))
		}

		return output, err
	}

	return nil, err
}

func WaitNetworkInsightsAnalysisCreated(ctx context.Context, conn *ec2.EC2, id string, timeout time.Duration) (*ec2.NetworkInsightsAnalysis, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{ec2.AnalysisStatusRunning},
		Target:     []string{ec2.AnalysisStatusSucceeded},
		Timeout:    timeout,
		Refresh:    StatusNetworkInsightsAnalysis(ctx, conn, id),
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.NetworkInsightsAnalysis); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusMessage)))

		return output, err
	}

	return nil, err
}

const (
	networkInterfaceAttachedTimeout = 5 * time.Minute
	NetworkInterfaceDetachedTimeout = 10 * time.Minute
)

func WaitNetworkInterfaceAttached(ctx context.Context, conn *ec2.EC2, id string, timeout time.Duration) (*ec2.NetworkInterfaceAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ec2.AttachmentStatusAttaching},
		Target:  []string{ec2.AttachmentStatusAttached},
		Timeout: timeout,
		Refresh: StatusNetworkInterfaceAttachmentStatus(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.NetworkInterfaceAttachment); ok {
		return output, err
	}

	return nil, err
}

func WaitNetworkInterfaceAvailableAfterUse(ctx context.Context, conn *ec2.EC2, id string, timeout time.Duration) (*ec2.NetworkInterface, error) {
	// Hyperplane attached ENI.
	// Wait for it to be moved into a removable state.
	stateConf := &retry.StateChangeConf{
		Pending:    []string{ec2.NetworkInterfaceStatusInUse},
		Target:     []string{ec2.NetworkInterfaceStatusAvailable},
		Timeout:    timeout,
		Refresh:    StatusNetworkInterfaceStatus(ctx, conn, id),
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
		// Handle EC2 ENI eventual consistency. It can take up to 3 minutes.
		ContinuousTargetOccurence: 18,
		NotFoundChecks:            1,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.NetworkInterface); ok {
		return output, err
	}

	return nil, err
}

func WaitNetworkInterfaceCreated(ctx context.Context, conn *ec2.EC2, id string, timeout time.Duration) (*ec2.NetworkInterface, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{NetworkInterfaceStatusPending},
		Target:  []string{ec2.NetworkInterfaceStatusAvailable},
		Timeout: timeout,
		Refresh: StatusNetworkInterfaceStatus(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.NetworkInterface); ok {
		return output, err
	}

	return nil, err
}

func WaitNetworkInterfaceDetached(ctx context.Context, conn *ec2.EC2, id string, timeout time.Duration) (*ec2.NetworkInterfaceAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ec2.AttachmentStatusAttached, ec2.AttachmentStatusDetaching},
		Target:  []string{ec2.AttachmentStatusDetached},
		Timeout: timeout,
		Refresh: StatusNetworkInterfaceAttachmentStatus(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.NetworkInterfaceAttachment); ok {
		return output, err
	}

	return nil, err
}

const (
	PlacementGroupCreatedTimeout = 5 * time.Minute
	PlacementGroupDeletedTimeout = 5 * time.Minute
)

func WaitPlacementGroupCreated(ctx context.Context, conn *ec2.EC2, name string) (*ec2.PlacementGroup, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ec2.PlacementGroupStatePending},
		Target:  []string{ec2.PlacementGroupStateAvailable},
		Timeout: PlacementGroupCreatedTimeout,
		Refresh: StatusPlacementGroupState(ctx, conn, name),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.PlacementGroup); ok {
		return output, err
	}

	return nil, err
}

func WaitPlacementGroupDeleted(ctx context.Context, conn *ec2.EC2, name string) (*ec2.PlacementGroup, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ec2.PlacementGroupStateDeleting},
		Target:  []string{},
		Timeout: PlacementGroupDeletedTimeout,
		Refresh: StatusPlacementGroupState(ctx, conn, name),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.PlacementGroup); ok {
		return output, err
	}

	return nil, err
}

func WaitSpotFleetRequestCreated(ctx context.Context, conn *ec2.EC2, id string, timeout time.Duration) (*ec2.SpotFleetRequestConfig, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{ec2.BatchStateSubmitted},
		Target:     []string{ec2.BatchStateActive},
		Refresh:    StatusSpotFleetRequestState(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.SpotFleetRequestConfig); ok {
		return output, err
	}

	return nil, err
}

func WaitSpotFleetRequestFulfilled(ctx context.Context, conn *ec2.EC2, id string, timeout time.Duration) (*ec2.SpotFleetRequestConfig, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{ec2.ActivityStatusPendingFulfillment},
		Target:     []string{ec2.ActivityStatusFulfilled},
		Refresh:    StatusSpotFleetActivityStatus(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.SpotFleetRequestConfig); ok {
		if activityStatus := aws.StringValue(output.ActivityStatus); activityStatus == ec2.ActivityStatusError {
			var errs []error

			input := &ec2.DescribeSpotFleetRequestHistoryInput{
				SpotFleetRequestId: aws.String(id),
				StartTime:          aws.Time(time.UnixMilli(0)),
			}

			if output, err := FindSpotFleetRequestHistoryRecords(ctx, conn, input); err == nil {
				for _, v := range output {
					if eventType := aws.StringValue(v.EventType); eventType == ec2.EventTypeError || eventType == ec2.EventTypeInformation {
						errs = append(errs, errors.New(v.String()))
					}
				}
			}

			tfresource.SetLastError(err, errors.Join(errs...))
		}

		return output, err
	}

	return nil, err
}

func WaitSpotFleetRequestUpdated(ctx context.Context, conn *ec2.EC2, id string, timeout time.Duration) (*ec2.SpotFleetRequestConfig, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{ec2.BatchStateModifying},
		Target:     []string{ec2.BatchStateActive},
		Refresh:    StatusSpotFleetRequestState(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.SpotFleetRequestConfig); ok {
		return output, err
	}

	return nil, err
}

func WaitSpotInstanceRequestFulfilled(ctx context.Context, conn *ec2.EC2, id string, timeout time.Duration) (*ec2.SpotInstanceRequest, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{spotInstanceRequestStatusCodePendingEvaluation, spotInstanceRequestStatusCodePendingFulfillment},
		Target:     []string{spotInstanceRequestStatusCodeFulfilled},
		Refresh:    StatusSpotInstanceRequest(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.SpotInstanceRequest); ok {
		if fault := output.Fault; fault != nil {
			errFault := fmt.Errorf("%s: %s", aws.StringValue(fault.Code), aws.StringValue(fault.Message))
			tfresource.SetLastError(err, fmt.Errorf("%s %w", aws.StringValue(output.Status.Message), errFault))
		} else {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.Status.Message)))
		}

		return output, err
	}

	return nil, err
}

func WaitVPCEndpointAccepted(ctx context.Context, conn *ec2.EC2, vpcEndpointID string, timeout time.Duration) (*ec2.VpcEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{vpcEndpointStatePendingAcceptance},
		Target:     []string{vpcEndpointStateAvailable},
		Timeout:    timeout,
		Refresh:    StatusVPCEndpointState(ctx, conn, vpcEndpointID),
		Delay:      5 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.VpcEndpoint); ok {
		if state, lastError := aws.StringValue(output.State), output.LastError; state == vpcEndpointStateFailed && lastError != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(lastError.Code), aws.StringValue(lastError.Message)))
		}

		return output, err
	}

	return nil, err
}

func WaitVPCEndpointAvailable(ctx context.Context, conn *ec2.EC2, vpcEndpointID string, timeout time.Duration) (*ec2.VpcEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{vpcEndpointStatePending},
		Target:     []string{vpcEndpointStateAvailable, vpcEndpointStatePendingAcceptance},
		Timeout:    timeout,
		Refresh:    StatusVPCEndpointState(ctx, conn, vpcEndpointID),
		Delay:      5 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.VpcEndpoint); ok {
		if state, lastError := aws.StringValue(output.State), output.LastError; state == vpcEndpointStateFailed && lastError != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(lastError.Code), aws.StringValue(lastError.Message)))
		}

		return output, err
	}

	return nil, err
}

func WaitVPCEndpointDeleted(ctx context.Context, conn *ec2.EC2, vpcEndpointID string, timeout time.Duration) (*ec2.VpcEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{vpcEndpointStateDeleting},
		Target:     []string{},
		Refresh:    StatusVPCEndpointState(ctx, conn, vpcEndpointID),
		Timeout:    timeout,
		Delay:      5 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.VpcEndpoint); ok {
		return output, err
	}

	return nil, err
}

func WaitVPCEndpointServiceAvailable(ctx context.Context, conn *ec2.EC2, id string, timeout time.Duration) (*ec2.ServiceConfiguration, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{ec2.ServiceStatePending},
		Target:     []string{ec2.ServiceStateAvailable},
		Refresh:    StatusVPCEndpointServiceStateAvailable(ctx, conn, id),
		Timeout:    timeout,
		Delay:      5 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.ServiceConfiguration); ok {
		return output, err
	}

	return nil, err
}

func WaitVPCEndpointServiceDeleted(ctx context.Context, conn *ec2.EC2, id string, timeout time.Duration) (*ec2.ServiceConfiguration, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{ec2.ServiceStateAvailable, ec2.ServiceStateDeleting},
		Target:     []string{},
		Timeout:    timeout,
		Refresh:    StatusVPCEndpointServiceStateDeleted(ctx, conn, id),
		Delay:      5 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.ServiceConfiguration); ok {
		return output, err
	}

	return nil, err
}

func WaitVPCEndpointRouteTableAssociationDeleted(ctx context.Context, conn *ec2.EC2, vpcEndpointID, routeTableID string) error {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{VPCEndpointRouteTableAssociationStatusReady},
		Target:                    []string{},
		Refresh:                   StatusVPCEndpointRouteTableAssociation(ctx, conn, vpcEndpointID, routeTableID),
		Timeout:                   ec2PropagationTimeout,
		ContinuousTargetOccurence: 2,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func WaitVPCEndpointRouteTableAssociationReady(ctx context.Context, conn *ec2.EC2, vpcEndpointID, routeTableID string) error {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{VPCEndpointRouteTableAssociationStatusReady},
		Refresh:                   StatusVPCEndpointRouteTableAssociation(ctx, conn, vpcEndpointID, routeTableID),
		Timeout:                   ec2PropagationTimeout,
		ContinuousTargetOccurence: 2,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func WaitEBSSnapshotImportComplete(ctx context.Context, conn *ec2_sdkv2.Client, importTaskID string, timeout time.Duration) (*awstypes.SnapshotTaskDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			EBSSnapshotImportStateActive,
			EBSSnapshotImportStateUpdating,
			EBSSnapshotImportStateValidating,
			EBSSnapshotImportStateValidated,
			EBSSnapshotImportStateConverting,
		},
		Target:  []string{EBSSnapshotImportStateCompleted},
		Refresh: StatusEBSSnapshotImport(ctx, conn, importTaskID),
		Timeout: timeout,
		Delay:   10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.SnapshotTaskDetail); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusMessage)))

		return output, err
	}

	return nil, err
}

const (
	ebsSnapshotArchivedTimeout = 60 * time.Minute
)

func waitEBSSnapshotTierArchive(ctx context.Context, conn *ec2_sdkv2.Client, id string, timeout time.Duration) (*awstypes.SnapshotTierStatus, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(TargetStorageTierStandard),
		Target:  enum.Slice(awstypes.TargetStorageTierArchive),
		Refresh: StatusSnapshotStorageTier(ctx, conn, id),
		Timeout: timeout,
		Delay:   10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.SnapshotTierStatus); ok {
		tfresource.SetLastError(err, fmt.Errorf("%s: %s", string(output.LastTieringOperationStatus), aws.StringValue(output.LastTieringOperationStatusDetail)))

		return output, err
	}

	return nil, err
}

func WaitInstanceConnectEndpointCreated(ctx context.Context, conn *ec2_sdkv2.Client, id string, timeout time.Duration) (*awstypes.Ec2InstanceConnectEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.Ec2InstanceConnectEndpointStateCreateInProgress),
		Target:  enum.Slice(awstypes.Ec2InstanceConnectEndpointStateCreateComplete),
		Refresh: statusInstanceConnectEndpoint(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Ec2InstanceConnectEndpoint); ok {
		tfresource.SetLastError(err, errors.New(aws_sdkv2.ToString(output.StateMessage)))

		return output, err
	}

	return nil, err
}

func WaitInstanceConnectEndpointDeleted(ctx context.Context, conn *ec2_sdkv2.Client, id string, timeout time.Duration) (*awstypes.Ec2InstanceConnectEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.Ec2InstanceConnectEndpointStateDeleteInProgress),
		Target:  []string{},
		Refresh: statusInstanceConnectEndpoint(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Ec2InstanceConnectEndpoint); ok {
		tfresource.SetLastError(err, errors.New(aws_sdkv2.ToString(output.StateMessage)))

		return output, err
	}

	return nil, err
}

func WaitImageBlockPublicAccessState(ctx context.Context, conn *ec2_sdkv2.Client, target string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Target:  []string{target},
		Refresh: StatusImageBlockPublicAccessState(ctx, conn),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func WaitVerifiedAccessEndpointCreated(ctx context.Context, conn *ec2_sdkv2.Client, id string, timeout time.Duration) (*awstypes.VerifiedAccessEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.VerifiedAccessEndpointStatusCodePending),
		Target:                    enum.Slice(awstypes.VerifiedAccessEndpointStatusCodeActive),
		Refresh:                   StatusVerifiedAccessEndpoint(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.VerifiedAccessEndpoint); ok {
		tfresource.SetLastError(err, errors.New(aws_sdkv2.ToString(output.Status.Message)))

		return output, err
	}

	return nil, err
}

func WaitVerifiedAccessEndpointUpdated(ctx context.Context, conn *ec2_sdkv2.Client, id string, timeout time.Duration) (*awstypes.VerifiedAccessEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.VerifiedAccessEndpointStatusCodeUpdating),
		Target:                    enum.Slice(awstypes.VerifiedAccessEndpointStatusCodeActive),
		Refresh:                   StatusVerifiedAccessEndpoint(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.VerifiedAccessEndpoint); ok {
		tfresource.SetLastError(err, errors.New(aws_sdkv2.ToString(output.Status.Message)))

		return output, err
	}

	return nil, err
}

func WaitVerifiedAccessEndpointDeleted(ctx context.Context, conn *ec2_sdkv2.Client, id string, timeout time.Duration) (*awstypes.VerifiedAccessEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.VerifiedAccessEndpointStatusCodeDeleting, awstypes.VerifiedAccessEndpointStatusCodeActive, awstypes.VerifiedAccessEndpointStatusCodeDeleted),
		Target:  []string{},
		Refresh: StatusVerifiedAccessEndpoint(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.VerifiedAccessEndpoint); ok {
		tfresource.SetLastError(err, errors.New(aws_sdkv2.ToString(output.Status.Message)))

		return output, err
	}

	return nil, err
}

func waitFastSnapshotRestoreCreated(ctx context.Context, conn *ec2_sdkv2.Client, availabilityZone, snapshotID string, timeout time.Duration) (*awstypes.DescribeFastSnapshotRestoreSuccessItem, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.FastSnapshotRestoreStateCodeEnabling, awstypes.FastSnapshotRestoreStateCodeOptimizing),
		Target:  enum.Slice(awstypes.FastSnapshotRestoreStateCodeEnabled),
		Refresh: statusFastSnapshotRestore(ctx, conn, availabilityZone, snapshotID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DescribeFastSnapshotRestoreSuccessItem); ok {
		return output, err
	}

	return nil, err
}

func waitFastSnapshotRestoreDeleted(ctx context.Context, conn *ec2_sdkv2.Client, availabilityZone, snapshotID string, timeout time.Duration) (*awstypes.DescribeFastSnapshotRestoreSuccessItem, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.FastSnapshotRestoreStateCodeDisabling, awstypes.FastSnapshotRestoreStateCodeOptimizing, awstypes.FastSnapshotRestoreStateCodeEnabled),
		Target:  []string{},
		Refresh: statusFastSnapshotRestore(ctx, conn, availabilityZone, snapshotID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DescribeFastSnapshotRestoreSuccessItem); ok {
		return output, err
	}

	return nil, err
}
