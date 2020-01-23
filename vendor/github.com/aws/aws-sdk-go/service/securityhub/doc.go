// Code generated by private/model/cli/gen-api/main.go. DO NOT EDIT.

// Package securityhub provides the client and types for making API
// requests to AWS SecurityHub.
//
// Security Hub provides you with a comprehensive view of the security state
// of your AWS environment and resources. It also provides you with the compliance
// status of your environment based on CIS AWS Foundations compliance checks.
// Security Hub collects security data from AWS accounts, services, and integrated
// third-party products and helps you analyze security trends in your environment
// to identify the highest priority security issues. For more information about
// Security Hub, see the AWS Security Hub User Guide (https://docs.aws.amazon.com/securityhub/latest/userguide/what-is-securityhub.html) .
//
// When you use operations in the Security Hub API, the requests are executed
// only in the AWS Region that is currently active or in the specific AWS Region
// that you specify in your request. Any configuration or settings change that
// results from the operation is applied only to that Region. To make the same
// change in other Regions, execute the same command for each Region to apply
// the change to. For example, if your Region is set to us-west-2, when you
// use CreateMembers to add a member account to Security Hub, the association
// of the member account with the master account is created only in the us-west-2
// Region. Security Hub must be enabled for the member account in the same Region
// that the invite was sent from.
//
// The following throttling limits apply to using Security Hub API operations:
//
//    * GetFindings - RateLimit of 3 requests per second, and a BurstLimit of
//    6 requests per second.
//
//    * UpdateFindings - RateLimit of 1 request per second, and a BurstLimit
//    of 5 requests per second.
//
//    * All other operations - RateLimit of 10 request per second, and a BurstLimit
//    of 30 requests per second.
//
// See https://docs.aws.amazon.com/goto/WebAPI/securityhub-2018-10-26 for more information on this service.
//
// See securityhub package documentation for more information.
// https://docs.aws.amazon.com/sdk-for-go/api/service/securityhub/
//
// Using the Client
//
// To contact AWS SecurityHub with the SDK use the New function to create
// a new service client. With that client you can make API requests to the service.
// These clients are safe to use concurrently.
//
// See the SDK's documentation for more information on how to use the SDK.
// https://docs.aws.amazon.com/sdk-for-go/api/
//
// See aws.Config documentation for more information on configuring SDK clients.
// https://docs.aws.amazon.com/sdk-for-go/api/aws/#Config
//
// See the AWS SecurityHub client SecurityHub for more
// information on creating client for this service.
// https://docs.aws.amazon.com/sdk-for-go/api/service/securityhub/#New
package securityhub
