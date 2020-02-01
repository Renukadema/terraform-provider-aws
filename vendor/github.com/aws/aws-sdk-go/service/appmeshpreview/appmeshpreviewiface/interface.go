// Code generated by private/model/cli/gen-api/main.go. DO NOT EDIT.

// Package appmeshpreviewiface provides an interface to enable mocking the AWS App Mesh Preview service client
// for testing your code.
//
// It is important to note that this interface will have breaking changes
// when the service model is updated and adds new API operations, paginators,
// and waiters.
package appmeshpreviewiface

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/appmeshpreview"
)

// AppMeshPreviewAPI provides an interface to enable mocking the
// appmeshpreview.AppMeshPreview service client's API operation,
// paginators, and waiters. This make unit testing your code that calls out
// to the SDK's service client's calls easier.
//
// The best way to use this interface is so the SDK's service client's calls
// can be stubbed out for unit testing your code with the SDK without needing
// to inject custom request handlers into the SDK's request pipeline.
//
//    // myFunc uses an SDK service client to make a request to
//    // AWS App Mesh Preview.
//    func myFunc(svc appmeshpreviewiface.AppMeshPreviewAPI) bool {
//        // Make svc.CreateMesh request
//    }
//
//    func main() {
//        sess := session.New()
//        svc := appmeshpreview.New(sess)
//
//        myFunc(svc)
//    }
//
// In your _test.go file:
//
//    // Define a mock struct to be used in your unit tests of myFunc.
//    type mockAppMeshPreviewClient struct {
//        appmeshpreviewiface.AppMeshPreviewAPI
//    }
//    func (m *mockAppMeshPreviewClient) CreateMesh(input *appmeshpreview.CreateMeshInput) (*appmeshpreview.CreateMeshOutput, error) {
//        // mock response/functionality
//    }
//
//    func TestMyFunc(t *testing.T) {
//        // Setup Test
//        mockSvc := &mockAppMeshPreviewClient{}
//
//        myfunc(mockSvc)
//
//        // Verify myFunc's functionality
//    }
//
// It is important to note that this interface will have breaking changes
// when the service model is updated and adds new API operations, paginators,
// and waiters. Its suggested to use the pattern above for testing, or using
// tooling to generate mocks to satisfy the interfaces.
type AppMeshPreviewAPI interface {
	CreateMesh(*appmeshpreview.CreateMeshInput) (*appmeshpreview.CreateMeshOutput, error)
	CreateMeshWithContext(aws.Context, *appmeshpreview.CreateMeshInput, ...request.Option) (*appmeshpreview.CreateMeshOutput, error)
	CreateMeshRequest(*appmeshpreview.CreateMeshInput) (*request.Request, *appmeshpreview.CreateMeshOutput)

	CreateRoute(*appmeshpreview.CreateRouteInput) (*appmeshpreview.CreateRouteOutput, error)
	CreateRouteWithContext(aws.Context, *appmeshpreview.CreateRouteInput, ...request.Option) (*appmeshpreview.CreateRouteOutput, error)
	CreateRouteRequest(*appmeshpreview.CreateRouteInput) (*request.Request, *appmeshpreview.CreateRouteOutput)

	CreateVirtualNode(*appmeshpreview.CreateVirtualNodeInput) (*appmeshpreview.CreateVirtualNodeOutput, error)
	CreateVirtualNodeWithContext(aws.Context, *appmeshpreview.CreateVirtualNodeInput, ...request.Option) (*appmeshpreview.CreateVirtualNodeOutput, error)
	CreateVirtualNodeRequest(*appmeshpreview.CreateVirtualNodeInput) (*request.Request, *appmeshpreview.CreateVirtualNodeOutput)

	CreateVirtualRouter(*appmeshpreview.CreateVirtualRouterInput) (*appmeshpreview.CreateVirtualRouterOutput, error)
	CreateVirtualRouterWithContext(aws.Context, *appmeshpreview.CreateVirtualRouterInput, ...request.Option) (*appmeshpreview.CreateVirtualRouterOutput, error)
	CreateVirtualRouterRequest(*appmeshpreview.CreateVirtualRouterInput) (*request.Request, *appmeshpreview.CreateVirtualRouterOutput)

	CreateVirtualService(*appmeshpreview.CreateVirtualServiceInput) (*appmeshpreview.CreateVirtualServiceOutput, error)
	CreateVirtualServiceWithContext(aws.Context, *appmeshpreview.CreateVirtualServiceInput, ...request.Option) (*appmeshpreview.CreateVirtualServiceOutput, error)
	CreateVirtualServiceRequest(*appmeshpreview.CreateVirtualServiceInput) (*request.Request, *appmeshpreview.CreateVirtualServiceOutput)

	DeleteMesh(*appmeshpreview.DeleteMeshInput) (*appmeshpreview.DeleteMeshOutput, error)
	DeleteMeshWithContext(aws.Context, *appmeshpreview.DeleteMeshInput, ...request.Option) (*appmeshpreview.DeleteMeshOutput, error)
	DeleteMeshRequest(*appmeshpreview.DeleteMeshInput) (*request.Request, *appmeshpreview.DeleteMeshOutput)

	DeleteRoute(*appmeshpreview.DeleteRouteInput) (*appmeshpreview.DeleteRouteOutput, error)
	DeleteRouteWithContext(aws.Context, *appmeshpreview.DeleteRouteInput, ...request.Option) (*appmeshpreview.DeleteRouteOutput, error)
	DeleteRouteRequest(*appmeshpreview.DeleteRouteInput) (*request.Request, *appmeshpreview.DeleteRouteOutput)

	DeleteVirtualNode(*appmeshpreview.DeleteVirtualNodeInput) (*appmeshpreview.DeleteVirtualNodeOutput, error)
	DeleteVirtualNodeWithContext(aws.Context, *appmeshpreview.DeleteVirtualNodeInput, ...request.Option) (*appmeshpreview.DeleteVirtualNodeOutput, error)
	DeleteVirtualNodeRequest(*appmeshpreview.DeleteVirtualNodeInput) (*request.Request, *appmeshpreview.DeleteVirtualNodeOutput)

	DeleteVirtualRouter(*appmeshpreview.DeleteVirtualRouterInput) (*appmeshpreview.DeleteVirtualRouterOutput, error)
	DeleteVirtualRouterWithContext(aws.Context, *appmeshpreview.DeleteVirtualRouterInput, ...request.Option) (*appmeshpreview.DeleteVirtualRouterOutput, error)
	DeleteVirtualRouterRequest(*appmeshpreview.DeleteVirtualRouterInput) (*request.Request, *appmeshpreview.DeleteVirtualRouterOutput)

	DeleteVirtualService(*appmeshpreview.DeleteVirtualServiceInput) (*appmeshpreview.DeleteVirtualServiceOutput, error)
	DeleteVirtualServiceWithContext(aws.Context, *appmeshpreview.DeleteVirtualServiceInput, ...request.Option) (*appmeshpreview.DeleteVirtualServiceOutput, error)
	DeleteVirtualServiceRequest(*appmeshpreview.DeleteVirtualServiceInput) (*request.Request, *appmeshpreview.DeleteVirtualServiceOutput)

	DescribeMesh(*appmeshpreview.DescribeMeshInput) (*appmeshpreview.DescribeMeshOutput, error)
	DescribeMeshWithContext(aws.Context, *appmeshpreview.DescribeMeshInput, ...request.Option) (*appmeshpreview.DescribeMeshOutput, error)
	DescribeMeshRequest(*appmeshpreview.DescribeMeshInput) (*request.Request, *appmeshpreview.DescribeMeshOutput)

	DescribeRoute(*appmeshpreview.DescribeRouteInput) (*appmeshpreview.DescribeRouteOutput, error)
	DescribeRouteWithContext(aws.Context, *appmeshpreview.DescribeRouteInput, ...request.Option) (*appmeshpreview.DescribeRouteOutput, error)
	DescribeRouteRequest(*appmeshpreview.DescribeRouteInput) (*request.Request, *appmeshpreview.DescribeRouteOutput)

	DescribeVirtualNode(*appmeshpreview.DescribeVirtualNodeInput) (*appmeshpreview.DescribeVirtualNodeOutput, error)
	DescribeVirtualNodeWithContext(aws.Context, *appmeshpreview.DescribeVirtualNodeInput, ...request.Option) (*appmeshpreview.DescribeVirtualNodeOutput, error)
	DescribeVirtualNodeRequest(*appmeshpreview.DescribeVirtualNodeInput) (*request.Request, *appmeshpreview.DescribeVirtualNodeOutput)

	DescribeVirtualRouter(*appmeshpreview.DescribeVirtualRouterInput) (*appmeshpreview.DescribeVirtualRouterOutput, error)
	DescribeVirtualRouterWithContext(aws.Context, *appmeshpreview.DescribeVirtualRouterInput, ...request.Option) (*appmeshpreview.DescribeVirtualRouterOutput, error)
	DescribeVirtualRouterRequest(*appmeshpreview.DescribeVirtualRouterInput) (*request.Request, *appmeshpreview.DescribeVirtualRouterOutput)

	DescribeVirtualService(*appmeshpreview.DescribeVirtualServiceInput) (*appmeshpreview.DescribeVirtualServiceOutput, error)
	DescribeVirtualServiceWithContext(aws.Context, *appmeshpreview.DescribeVirtualServiceInput, ...request.Option) (*appmeshpreview.DescribeVirtualServiceOutput, error)
	DescribeVirtualServiceRequest(*appmeshpreview.DescribeVirtualServiceInput) (*request.Request, *appmeshpreview.DescribeVirtualServiceOutput)

	ListMeshes(*appmeshpreview.ListMeshesInput) (*appmeshpreview.ListMeshesOutput, error)
	ListMeshesWithContext(aws.Context, *appmeshpreview.ListMeshesInput, ...request.Option) (*appmeshpreview.ListMeshesOutput, error)
	ListMeshesRequest(*appmeshpreview.ListMeshesInput) (*request.Request, *appmeshpreview.ListMeshesOutput)

	ListRoutes(*appmeshpreview.ListRoutesInput) (*appmeshpreview.ListRoutesOutput, error)
	ListRoutesWithContext(aws.Context, *appmeshpreview.ListRoutesInput, ...request.Option) (*appmeshpreview.ListRoutesOutput, error)
	ListRoutesRequest(*appmeshpreview.ListRoutesInput) (*request.Request, *appmeshpreview.ListRoutesOutput)

	ListVirtualNodes(*appmeshpreview.ListVirtualNodesInput) (*appmeshpreview.ListVirtualNodesOutput, error)
	ListVirtualNodesWithContext(aws.Context, *appmeshpreview.ListVirtualNodesInput, ...request.Option) (*appmeshpreview.ListVirtualNodesOutput, error)
	ListVirtualNodesRequest(*appmeshpreview.ListVirtualNodesInput) (*request.Request, *appmeshpreview.ListVirtualNodesOutput)

	ListVirtualRouters(*appmeshpreview.ListVirtualRoutersInput) (*appmeshpreview.ListVirtualRoutersOutput, error)
	ListVirtualRoutersWithContext(aws.Context, *appmeshpreview.ListVirtualRoutersInput, ...request.Option) (*appmeshpreview.ListVirtualRoutersOutput, error)
	ListVirtualRoutersRequest(*appmeshpreview.ListVirtualRoutersInput) (*request.Request, *appmeshpreview.ListVirtualRoutersOutput)

	ListVirtualServices(*appmeshpreview.ListVirtualServicesInput) (*appmeshpreview.ListVirtualServicesOutput, error)
	ListVirtualServicesWithContext(aws.Context, *appmeshpreview.ListVirtualServicesInput, ...request.Option) (*appmeshpreview.ListVirtualServicesOutput, error)
	ListVirtualServicesRequest(*appmeshpreview.ListVirtualServicesInput) (*request.Request, *appmeshpreview.ListVirtualServicesOutput)

	UpdateMesh(*appmeshpreview.UpdateMeshInput) (*appmeshpreview.UpdateMeshOutput, error)
	UpdateMeshWithContext(aws.Context, *appmeshpreview.UpdateMeshInput, ...request.Option) (*appmeshpreview.UpdateMeshOutput, error)
	UpdateMeshRequest(*appmeshpreview.UpdateMeshInput) (*request.Request, *appmeshpreview.UpdateMeshOutput)

	UpdateRoute(*appmeshpreview.UpdateRouteInput) (*appmeshpreview.UpdateRouteOutput, error)
	UpdateRouteWithContext(aws.Context, *appmeshpreview.UpdateRouteInput, ...request.Option) (*appmeshpreview.UpdateRouteOutput, error)
	UpdateRouteRequest(*appmeshpreview.UpdateRouteInput) (*request.Request, *appmeshpreview.UpdateRouteOutput)

	UpdateVirtualNode(*appmeshpreview.UpdateVirtualNodeInput) (*appmeshpreview.UpdateVirtualNodeOutput, error)
	UpdateVirtualNodeWithContext(aws.Context, *appmeshpreview.UpdateVirtualNodeInput, ...request.Option) (*appmeshpreview.UpdateVirtualNodeOutput, error)
	UpdateVirtualNodeRequest(*appmeshpreview.UpdateVirtualNodeInput) (*request.Request, *appmeshpreview.UpdateVirtualNodeOutput)

	UpdateVirtualRouter(*appmeshpreview.UpdateVirtualRouterInput) (*appmeshpreview.UpdateVirtualRouterOutput, error)
	UpdateVirtualRouterWithContext(aws.Context, *appmeshpreview.UpdateVirtualRouterInput, ...request.Option) (*appmeshpreview.UpdateVirtualRouterOutput, error)
	UpdateVirtualRouterRequest(*appmeshpreview.UpdateVirtualRouterInput) (*request.Request, *appmeshpreview.UpdateVirtualRouterOutput)

	UpdateVirtualService(*appmeshpreview.UpdateVirtualServiceInput) (*appmeshpreview.UpdateVirtualServiceOutput, error)
	UpdateVirtualServiceWithContext(aws.Context, *appmeshpreview.UpdateVirtualServiceInput, ...request.Option) (*appmeshpreview.UpdateVirtualServiceOutput, error)
	UpdateVirtualServiceRequest(*appmeshpreview.UpdateVirtualServiceInput) (*request.Request, *appmeshpreview.UpdateVirtualServiceOutput)
}

var _ AppMeshPreviewAPI = (*appmeshpreview.AppMeshPreview)(nil)
