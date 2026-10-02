package mocks

import (
	"database/sql"

	"inventory/db"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockDatabase implements db.Database for tests.
type MockDatabase struct {
	mock.Mock
}

// helper to call m.Called with variadic expansion
func call(m *mock.Mock, args ...interface{}) mock.Arguments {
	return m.Called(args...)
}

// Utility to expand a fixed arg + variadic slice into one []interface{}
func buildArgs(first interface{}, rest []interface{}) []interface{} {
	res := make([]interface{}, 0, 1+len(rest))
	res = append(res, first)
	res = append(res, rest...)
	return res
}

// Utility to expand only variadic slice (when no fixed first arg)
func buildArgsFromSlice(rest []interface{}) []interface{} {
	res := make([]interface{}, 0, len(rest))
	res = append(res, rest...)
	return res
}

// ---- Methods ----
// Note: methods that accept variadic args must call m.Called with those args expanded.

type MockS3Client struct {
	mock.Mock
}

// AbortMultipartUpload implements [s3iface.S3API].
func (m *MockS3Client) AbortMultipartUpload(*s3.AbortMultipartUploadInput) (*s3.AbortMultipartUploadOutput, error) {
	panic("unimplemented")
}

// AbortMultipartUploadRequest implements [s3iface.S3API].
func (m *MockS3Client) AbortMultipartUploadRequest(*s3.AbortMultipartUploadInput) (*request.Request, *s3.AbortMultipartUploadOutput) {
	panic("unimplemented")
}

// AbortMultipartUploadWithContext implements [s3iface.S3API].
func (m *MockS3Client) AbortMultipartUploadWithContext(aws.Context, *s3.AbortMultipartUploadInput, ...request.Option) (*s3.AbortMultipartUploadOutput, error) {
	panic("unimplemented")
}

// CompleteMultipartUpload implements [s3iface.S3API].
func (m *MockS3Client) CompleteMultipartUpload(*s3.CompleteMultipartUploadInput) (*s3.CompleteMultipartUploadOutput, error) {
	panic("unimplemented")
}

// CompleteMultipartUploadRequest implements [s3iface.S3API].
func (m *MockS3Client) CompleteMultipartUploadRequest(*s3.CompleteMultipartUploadInput) (*request.Request, *s3.CompleteMultipartUploadOutput) {
	panic("unimplemented")
}

// CompleteMultipartUploadWithContext implements [s3iface.S3API].
func (m *MockS3Client) CompleteMultipartUploadWithContext(aws.Context, *s3.CompleteMultipartUploadInput, ...request.Option) (*s3.CompleteMultipartUploadOutput, error) {
	panic("unimplemented")
}

// CopyObject implements [s3iface.S3API].
func (m *MockS3Client) CopyObject(*s3.CopyObjectInput) (*s3.CopyObjectOutput, error) {
	panic("unimplemented")
}

// CopyObjectRequest implements [s3iface.S3API].
func (m *MockS3Client) CopyObjectRequest(*s3.CopyObjectInput) (*request.Request, *s3.CopyObjectOutput) {
	panic("unimplemented")
}

// CopyObjectWithContext implements [s3iface.S3API].
func (m *MockS3Client) CopyObjectWithContext(aws.Context, *s3.CopyObjectInput, ...request.Option) (*s3.CopyObjectOutput, error) {
	panic("unimplemented")
}

// CreateBucket implements [s3iface.S3API].
func (m *MockS3Client) CreateBucket(*s3.CreateBucketInput) (*s3.CreateBucketOutput, error) {
	panic("unimplemented")
}

// CreateBucketRequest implements [s3iface.S3API].
func (m *MockS3Client) CreateBucketRequest(*s3.CreateBucketInput) (*request.Request, *s3.CreateBucketOutput) {
	panic("unimplemented")
}

// CreateBucketWithContext implements [s3iface.S3API].
func (m *MockS3Client) CreateBucketWithContext(aws.Context, *s3.CreateBucketInput, ...request.Option) (*s3.CreateBucketOutput, error) {
	panic("unimplemented")
}

// CreateMultipartUpload implements [s3iface.S3API].
func (m *MockS3Client) CreateMultipartUpload(*s3.CreateMultipartUploadInput) (*s3.CreateMultipartUploadOutput, error) {
	panic("unimplemented")
}

// CreateMultipartUploadRequest implements [s3iface.S3API].
func (m *MockS3Client) CreateMultipartUploadRequest(*s3.CreateMultipartUploadInput) (*request.Request, *s3.CreateMultipartUploadOutput) {
	panic("unimplemented")
}

// CreateMultipartUploadWithContext implements [s3iface.S3API].
func (m *MockS3Client) CreateMultipartUploadWithContext(aws.Context, *s3.CreateMultipartUploadInput, ...request.Option) (*s3.CreateMultipartUploadOutput, error) {
	panic("unimplemented")
}

// CreateSession implements [s3iface.S3API].
func (m *MockS3Client) CreateSession(*s3.CreateSessionInput) (*s3.CreateSessionOutput, error) {
	panic("unimplemented")
}

// CreateSessionRequest implements [s3iface.S3API].
func (m *MockS3Client) CreateSessionRequest(*s3.CreateSessionInput) (*request.Request, *s3.CreateSessionOutput) {
	panic("unimplemented")
}

// CreateSessionWithContext implements [s3iface.S3API].
func (m *MockS3Client) CreateSessionWithContext(aws.Context, *s3.CreateSessionInput, ...request.Option) (*s3.CreateSessionOutput, error) {
	panic("unimplemented")
}

// DeleteBucket implements [s3iface.S3API].
func (m *MockS3Client) DeleteBucket(*s3.DeleteBucketInput) (*s3.DeleteBucketOutput, error) {
	panic("unimplemented")
}

// DeleteBucketAnalyticsConfiguration implements [s3iface.S3API].
func (m *MockS3Client) DeleteBucketAnalyticsConfiguration(*s3.DeleteBucketAnalyticsConfigurationInput) (*s3.DeleteBucketAnalyticsConfigurationOutput, error) {
	panic("unimplemented")
}

// DeleteBucketAnalyticsConfigurationRequest implements [s3iface.S3API].
func (m *MockS3Client) DeleteBucketAnalyticsConfigurationRequest(*s3.DeleteBucketAnalyticsConfigurationInput) (*request.Request, *s3.DeleteBucketAnalyticsConfigurationOutput) {
	panic("unimplemented")
}

// DeleteBucketAnalyticsConfigurationWithContext implements [s3iface.S3API].
func (m *MockS3Client) DeleteBucketAnalyticsConfigurationWithContext(aws.Context, *s3.DeleteBucketAnalyticsConfigurationInput, ...request.Option) (*s3.DeleteBucketAnalyticsConfigurationOutput, error) {
	panic("unimplemented")
}

// DeleteBucketCors implements [s3iface.S3API].
func (m *MockS3Client) DeleteBucketCors(*s3.DeleteBucketCorsInput) (*s3.DeleteBucketCorsOutput, error) {
	panic("unimplemented")
}

// DeleteBucketCorsRequest implements [s3iface.S3API].
func (m *MockS3Client) DeleteBucketCorsRequest(*s3.DeleteBucketCorsInput) (*request.Request, *s3.DeleteBucketCorsOutput) {
	panic("unimplemented")
}

// DeleteBucketCorsWithContext implements [s3iface.S3API].
func (m *MockS3Client) DeleteBucketCorsWithContext(aws.Context, *s3.DeleteBucketCorsInput, ...request.Option) (*s3.DeleteBucketCorsOutput, error) {
	panic("unimplemented")
}

// DeleteBucketEncryption implements [s3iface.S3API].
func (m *MockS3Client) DeleteBucketEncryption(*s3.DeleteBucketEncryptionInput) (*s3.DeleteBucketEncryptionOutput, error) {
	panic("unimplemented")
}

// DeleteBucketEncryptionRequest implements [s3iface.S3API].
func (m *MockS3Client) DeleteBucketEncryptionRequest(*s3.DeleteBucketEncryptionInput) (*request.Request, *s3.DeleteBucketEncryptionOutput) {
	panic("unimplemented")
}

// DeleteBucketEncryptionWithContext implements [s3iface.S3API].
func (m *MockS3Client) DeleteBucketEncryptionWithContext(aws.Context, *s3.DeleteBucketEncryptionInput, ...request.Option) (*s3.DeleteBucketEncryptionOutput, error) {
	panic("unimplemented")
}

// DeleteBucketIntelligentTieringConfiguration implements [s3iface.S3API].
func (m *MockS3Client) DeleteBucketIntelligentTieringConfiguration(*s3.DeleteBucketIntelligentTieringConfigurationInput) (*s3.DeleteBucketIntelligentTieringConfigurationOutput, error) {
	panic("unimplemented")
}

// DeleteBucketIntelligentTieringConfigurationRequest implements [s3iface.S3API].
func (m *MockS3Client) DeleteBucketIntelligentTieringConfigurationRequest(*s3.DeleteBucketIntelligentTieringConfigurationInput) (*request.Request, *s3.DeleteBucketIntelligentTieringConfigurationOutput) {
	panic("unimplemented")
}

// DeleteBucketIntelligentTieringConfigurationWithContext implements [s3iface.S3API].
func (m *MockS3Client) DeleteBucketIntelligentTieringConfigurationWithContext(aws.Context, *s3.DeleteBucketIntelligentTieringConfigurationInput, ...request.Option) (*s3.DeleteBucketIntelligentTieringConfigurationOutput, error) {
	panic("unimplemented")
}

// DeleteBucketInventoryConfiguration implements [s3iface.S3API].
func (m *MockS3Client) DeleteBucketInventoryConfiguration(*s3.DeleteBucketInventoryConfigurationInput) (*s3.DeleteBucketInventoryConfigurationOutput, error) {
	panic("unimplemented")
}

// DeleteBucketInventoryConfigurationRequest implements [s3iface.S3API].
func (m *MockS3Client) DeleteBucketInventoryConfigurationRequest(*s3.DeleteBucketInventoryConfigurationInput) (*request.Request, *s3.DeleteBucketInventoryConfigurationOutput) {
	panic("unimplemented")
}

// DeleteBucketInventoryConfigurationWithContext implements [s3iface.S3API].
func (m *MockS3Client) DeleteBucketInventoryConfigurationWithContext(aws.Context, *s3.DeleteBucketInventoryConfigurationInput, ...request.Option) (*s3.DeleteBucketInventoryConfigurationOutput, error) {
	panic("unimplemented")
}

// DeleteBucketLifecycle implements [s3iface.S3API].
func (m *MockS3Client) DeleteBucketLifecycle(*s3.DeleteBucketLifecycleInput) (*s3.DeleteBucketLifecycleOutput, error) {
	panic("unimplemented")
}

// DeleteBucketLifecycleRequest implements [s3iface.S3API].
func (m *MockS3Client) DeleteBucketLifecycleRequest(*s3.DeleteBucketLifecycleInput) (*request.Request, *s3.DeleteBucketLifecycleOutput) {
	panic("unimplemented")
}

// DeleteBucketLifecycleWithContext implements [s3iface.S3API].
func (m *MockS3Client) DeleteBucketLifecycleWithContext(aws.Context, *s3.DeleteBucketLifecycleInput, ...request.Option) (*s3.DeleteBucketLifecycleOutput, error) {
	panic("unimplemented")
}

// DeleteBucketMetricsConfiguration implements [s3iface.S3API].
func (m *MockS3Client) DeleteBucketMetricsConfiguration(*s3.DeleteBucketMetricsConfigurationInput) (*s3.DeleteBucketMetricsConfigurationOutput, error) {
	panic("unimplemented")
}

// DeleteBucketMetricsConfigurationRequest implements [s3iface.S3API].
func (m *MockS3Client) DeleteBucketMetricsConfigurationRequest(*s3.DeleteBucketMetricsConfigurationInput) (*request.Request, *s3.DeleteBucketMetricsConfigurationOutput) {
	panic("unimplemented")
}

// DeleteBucketMetricsConfigurationWithContext implements [s3iface.S3API].
func (m *MockS3Client) DeleteBucketMetricsConfigurationWithContext(aws.Context, *s3.DeleteBucketMetricsConfigurationInput, ...request.Option) (*s3.DeleteBucketMetricsConfigurationOutput, error) {
	panic("unimplemented")
}

// DeleteBucketOwnershipControls implements [s3iface.S3API].
func (m *MockS3Client) DeleteBucketOwnershipControls(*s3.DeleteBucketOwnershipControlsInput) (*s3.DeleteBucketOwnershipControlsOutput, error) {
	panic("unimplemented")
}

// DeleteBucketOwnershipControlsRequest implements [s3iface.S3API].
func (m *MockS3Client) DeleteBucketOwnershipControlsRequest(*s3.DeleteBucketOwnershipControlsInput) (*request.Request, *s3.DeleteBucketOwnershipControlsOutput) {
	panic("unimplemented")
}

// DeleteBucketOwnershipControlsWithContext implements [s3iface.S3API].
func (m *MockS3Client) DeleteBucketOwnershipControlsWithContext(aws.Context, *s3.DeleteBucketOwnershipControlsInput, ...request.Option) (*s3.DeleteBucketOwnershipControlsOutput, error) {
	panic("unimplemented")
}

// DeleteBucketPolicy implements [s3iface.S3API].
func (m *MockS3Client) DeleteBucketPolicy(*s3.DeleteBucketPolicyInput) (*s3.DeleteBucketPolicyOutput, error) {
	panic("unimplemented")
}

// DeleteBucketPolicyRequest implements [s3iface.S3API].
func (m *MockS3Client) DeleteBucketPolicyRequest(*s3.DeleteBucketPolicyInput) (*request.Request, *s3.DeleteBucketPolicyOutput) {
	panic("unimplemented")
}

// DeleteBucketPolicyWithContext implements [s3iface.S3API].
func (m *MockS3Client) DeleteBucketPolicyWithContext(aws.Context, *s3.DeleteBucketPolicyInput, ...request.Option) (*s3.DeleteBucketPolicyOutput, error) {
	panic("unimplemented")
}

// DeleteBucketReplication implements [s3iface.S3API].
func (m *MockS3Client) DeleteBucketReplication(*s3.DeleteBucketReplicationInput) (*s3.DeleteBucketReplicationOutput, error) {
	panic("unimplemented")
}

// DeleteBucketReplicationRequest implements [s3iface.S3API].
func (m *MockS3Client) DeleteBucketReplicationRequest(*s3.DeleteBucketReplicationInput) (*request.Request, *s3.DeleteBucketReplicationOutput) {
	panic("unimplemented")
}

// DeleteBucketReplicationWithContext implements [s3iface.S3API].
func (m *MockS3Client) DeleteBucketReplicationWithContext(aws.Context, *s3.DeleteBucketReplicationInput, ...request.Option) (*s3.DeleteBucketReplicationOutput, error) {
	panic("unimplemented")
}

// DeleteBucketRequest implements [s3iface.S3API].
func (m *MockS3Client) DeleteBucketRequest(*s3.DeleteBucketInput) (*request.Request, *s3.DeleteBucketOutput) {
	panic("unimplemented")
}

// DeleteBucketTagging implements [s3iface.S3API].
func (m *MockS3Client) DeleteBucketTagging(*s3.DeleteBucketTaggingInput) (*s3.DeleteBucketTaggingOutput, error) {
	panic("unimplemented")
}

// DeleteBucketTaggingRequest implements [s3iface.S3API].
func (m *MockS3Client) DeleteBucketTaggingRequest(*s3.DeleteBucketTaggingInput) (*request.Request, *s3.DeleteBucketTaggingOutput) {
	panic("unimplemented")
}

// DeleteBucketTaggingWithContext implements [s3iface.S3API].
func (m *MockS3Client) DeleteBucketTaggingWithContext(aws.Context, *s3.DeleteBucketTaggingInput, ...request.Option) (*s3.DeleteBucketTaggingOutput, error) {
	panic("unimplemented")
}

// DeleteBucketWebsite implements [s3iface.S3API].
func (m *MockS3Client) DeleteBucketWebsite(*s3.DeleteBucketWebsiteInput) (*s3.DeleteBucketWebsiteOutput, error) {
	panic("unimplemented")
}

// DeleteBucketWebsiteRequest implements [s3iface.S3API].
func (m *MockS3Client) DeleteBucketWebsiteRequest(*s3.DeleteBucketWebsiteInput) (*request.Request, *s3.DeleteBucketWebsiteOutput) {
	panic("unimplemented")
}

// DeleteBucketWebsiteWithContext implements [s3iface.S3API].
func (m *MockS3Client) DeleteBucketWebsiteWithContext(aws.Context, *s3.DeleteBucketWebsiteInput, ...request.Option) (*s3.DeleteBucketWebsiteOutput, error) {
	panic("unimplemented")
}

// DeleteBucketWithContext implements [s3iface.S3API].
func (m *MockS3Client) DeleteBucketWithContext(aws.Context, *s3.DeleteBucketInput, ...request.Option) (*s3.DeleteBucketOutput, error) {
	panic("unimplemented")
}

// DeleteObject implements [s3iface.S3API].
func (m *MockS3Client) DeleteObject(*s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error) {
	panic("unimplemented")
}

// DeleteObjectRequest implements [s3iface.S3API].
func (m *MockS3Client) DeleteObjectRequest(*s3.DeleteObjectInput) (*request.Request, *s3.DeleteObjectOutput) {
	panic("unimplemented")
}

// DeleteObjectTagging implements [s3iface.S3API].
func (m *MockS3Client) DeleteObjectTagging(*s3.DeleteObjectTaggingInput) (*s3.DeleteObjectTaggingOutput, error) {
	panic("unimplemented")
}

// DeleteObjectTaggingRequest implements [s3iface.S3API].
func (m *MockS3Client) DeleteObjectTaggingRequest(*s3.DeleteObjectTaggingInput) (*request.Request, *s3.DeleteObjectTaggingOutput) {
	panic("unimplemented")
}

// DeleteObjectTaggingWithContext implements [s3iface.S3API].
func (m *MockS3Client) DeleteObjectTaggingWithContext(aws.Context, *s3.DeleteObjectTaggingInput, ...request.Option) (*s3.DeleteObjectTaggingOutput, error) {
	panic("unimplemented")
}

// DeleteObjectWithContext implements [s3iface.S3API].
func (m *MockS3Client) DeleteObjectWithContext(aws.Context, *s3.DeleteObjectInput, ...request.Option) (*s3.DeleteObjectOutput, error) {
	panic("unimplemented")
}

// DeleteObjects implements [s3iface.S3API].
func (m *MockS3Client) DeleteObjects(*s3.DeleteObjectsInput) (*s3.DeleteObjectsOutput, error) {
	panic("unimplemented")
}

// DeleteObjectsRequest implements [s3iface.S3API].
func (m *MockS3Client) DeleteObjectsRequest(*s3.DeleteObjectsInput) (*request.Request, *s3.DeleteObjectsOutput) {
	panic("unimplemented")
}

// DeleteObjectsWithContext implements [s3iface.S3API].
func (m *MockS3Client) DeleteObjectsWithContext(aws.Context, *s3.DeleteObjectsInput, ...request.Option) (*s3.DeleteObjectsOutput, error) {
	panic("unimplemented")
}

// DeletePublicAccessBlock implements [s3iface.S3API].
func (m *MockS3Client) DeletePublicAccessBlock(*s3.DeletePublicAccessBlockInput) (*s3.DeletePublicAccessBlockOutput, error) {
	panic("unimplemented")
}

// DeletePublicAccessBlockRequest implements [s3iface.S3API].
func (m *MockS3Client) DeletePublicAccessBlockRequest(*s3.DeletePublicAccessBlockInput) (*request.Request, *s3.DeletePublicAccessBlockOutput) {
	panic("unimplemented")
}

// DeletePublicAccessBlockWithContext implements [s3iface.S3API].
func (m *MockS3Client) DeletePublicAccessBlockWithContext(aws.Context, *s3.DeletePublicAccessBlockInput, ...request.Option) (*s3.DeletePublicAccessBlockOutput, error) {
	panic("unimplemented")
}

// GetBucketAccelerateConfiguration implements [s3iface.S3API].
func (m *MockS3Client) GetBucketAccelerateConfiguration(*s3.GetBucketAccelerateConfigurationInput) (*s3.GetBucketAccelerateConfigurationOutput, error) {
	panic("unimplemented")
}

// GetBucketAccelerateConfigurationRequest implements [s3iface.S3API].
func (m *MockS3Client) GetBucketAccelerateConfigurationRequest(*s3.GetBucketAccelerateConfigurationInput) (*request.Request, *s3.GetBucketAccelerateConfigurationOutput) {
	panic("unimplemented")
}

// GetBucketAccelerateConfigurationWithContext implements [s3iface.S3API].
func (m *MockS3Client) GetBucketAccelerateConfigurationWithContext(aws.Context, *s3.GetBucketAccelerateConfigurationInput, ...request.Option) (*s3.GetBucketAccelerateConfigurationOutput, error) {
	panic("unimplemented")
}

// GetBucketAcl implements [s3iface.S3API].
func (m *MockS3Client) GetBucketAcl(*s3.GetBucketAclInput) (*s3.GetBucketAclOutput, error) {
	panic("unimplemented")
}

// GetBucketAclRequest implements [s3iface.S3API].
func (m *MockS3Client) GetBucketAclRequest(*s3.GetBucketAclInput) (*request.Request, *s3.GetBucketAclOutput) {
	panic("unimplemented")
}

// GetBucketAclWithContext implements [s3iface.S3API].
func (m *MockS3Client) GetBucketAclWithContext(aws.Context, *s3.GetBucketAclInput, ...request.Option) (*s3.GetBucketAclOutput, error) {
	panic("unimplemented")
}

// GetBucketAnalyticsConfiguration implements [s3iface.S3API].
func (m *MockS3Client) GetBucketAnalyticsConfiguration(*s3.GetBucketAnalyticsConfigurationInput) (*s3.GetBucketAnalyticsConfigurationOutput, error) {
	panic("unimplemented")
}

// GetBucketAnalyticsConfigurationRequest implements [s3iface.S3API].
func (m *MockS3Client) GetBucketAnalyticsConfigurationRequest(*s3.GetBucketAnalyticsConfigurationInput) (*request.Request, *s3.GetBucketAnalyticsConfigurationOutput) {
	panic("unimplemented")
}

// GetBucketAnalyticsConfigurationWithContext implements [s3iface.S3API].
func (m *MockS3Client) GetBucketAnalyticsConfigurationWithContext(aws.Context, *s3.GetBucketAnalyticsConfigurationInput, ...request.Option) (*s3.GetBucketAnalyticsConfigurationOutput, error) {
	panic("unimplemented")
}

// GetBucketCors implements [s3iface.S3API].
func (m *MockS3Client) GetBucketCors(*s3.GetBucketCorsInput) (*s3.GetBucketCorsOutput, error) {
	panic("unimplemented")
}

// GetBucketCorsRequest implements [s3iface.S3API].
func (m *MockS3Client) GetBucketCorsRequest(*s3.GetBucketCorsInput) (*request.Request, *s3.GetBucketCorsOutput) {
	panic("unimplemented")
}

// GetBucketCorsWithContext implements [s3iface.S3API].
func (m *MockS3Client) GetBucketCorsWithContext(aws.Context, *s3.GetBucketCorsInput, ...request.Option) (*s3.GetBucketCorsOutput, error) {
	panic("unimplemented")
}

// GetBucketEncryption implements [s3iface.S3API].
func (m *MockS3Client) GetBucketEncryption(*s3.GetBucketEncryptionInput) (*s3.GetBucketEncryptionOutput, error) {
	panic("unimplemented")
}

// GetBucketEncryptionRequest implements [s3iface.S3API].
func (m *MockS3Client) GetBucketEncryptionRequest(*s3.GetBucketEncryptionInput) (*request.Request, *s3.GetBucketEncryptionOutput) {
	panic("unimplemented")
}

// GetBucketEncryptionWithContext implements [s3iface.S3API].
func (m *MockS3Client) GetBucketEncryptionWithContext(aws.Context, *s3.GetBucketEncryptionInput, ...request.Option) (*s3.GetBucketEncryptionOutput, error) {
	panic("unimplemented")
}

// GetBucketIntelligentTieringConfiguration implements [s3iface.S3API].
func (m *MockS3Client) GetBucketIntelligentTieringConfiguration(*s3.GetBucketIntelligentTieringConfigurationInput) (*s3.GetBucketIntelligentTieringConfigurationOutput, error) {
	panic("unimplemented")
}

// GetBucketIntelligentTieringConfigurationRequest implements [s3iface.S3API].
func (m *MockS3Client) GetBucketIntelligentTieringConfigurationRequest(*s3.GetBucketIntelligentTieringConfigurationInput) (*request.Request, *s3.GetBucketIntelligentTieringConfigurationOutput) {
	panic("unimplemented")
}

// GetBucketIntelligentTieringConfigurationWithContext implements [s3iface.S3API].
func (m *MockS3Client) GetBucketIntelligentTieringConfigurationWithContext(aws.Context, *s3.GetBucketIntelligentTieringConfigurationInput, ...request.Option) (*s3.GetBucketIntelligentTieringConfigurationOutput, error) {
	panic("unimplemented")
}

// GetBucketInventoryConfiguration implements [s3iface.S3API].
func (m *MockS3Client) GetBucketInventoryConfiguration(*s3.GetBucketInventoryConfigurationInput) (*s3.GetBucketInventoryConfigurationOutput, error) {
	panic("unimplemented")
}

// GetBucketInventoryConfigurationRequest implements [s3iface.S3API].
func (m *MockS3Client) GetBucketInventoryConfigurationRequest(*s3.GetBucketInventoryConfigurationInput) (*request.Request, *s3.GetBucketInventoryConfigurationOutput) {
	panic("unimplemented")
}

// GetBucketInventoryConfigurationWithContext implements [s3iface.S3API].
func (m *MockS3Client) GetBucketInventoryConfigurationWithContext(aws.Context, *s3.GetBucketInventoryConfigurationInput, ...request.Option) (*s3.GetBucketInventoryConfigurationOutput, error) {
	panic("unimplemented")
}

// GetBucketLifecycle implements [s3iface.S3API].
func (m *MockS3Client) GetBucketLifecycle(*s3.GetBucketLifecycleInput) (*s3.GetBucketLifecycleOutput, error) {
	panic("unimplemented")
}

// GetBucketLifecycleConfiguration implements [s3iface.S3API].
func (m *MockS3Client) GetBucketLifecycleConfiguration(*s3.GetBucketLifecycleConfigurationInput) (*s3.GetBucketLifecycleConfigurationOutput, error) {
	panic("unimplemented")
}

// GetBucketLifecycleConfigurationRequest implements [s3iface.S3API].
func (m *MockS3Client) GetBucketLifecycleConfigurationRequest(*s3.GetBucketLifecycleConfigurationInput) (*request.Request, *s3.GetBucketLifecycleConfigurationOutput) {
	panic("unimplemented")
}

// GetBucketLifecycleConfigurationWithContext implements [s3iface.S3API].
func (m *MockS3Client) GetBucketLifecycleConfigurationWithContext(aws.Context, *s3.GetBucketLifecycleConfigurationInput, ...request.Option) (*s3.GetBucketLifecycleConfigurationOutput, error) {
	panic("unimplemented")
}

// GetBucketLifecycleRequest implements [s3iface.S3API].
func (m *MockS3Client) GetBucketLifecycleRequest(*s3.GetBucketLifecycleInput) (*request.Request, *s3.GetBucketLifecycleOutput) {
	panic("unimplemented")
}

// GetBucketLifecycleWithContext implements [s3iface.S3API].
func (m *MockS3Client) GetBucketLifecycleWithContext(aws.Context, *s3.GetBucketLifecycleInput, ...request.Option) (*s3.GetBucketLifecycleOutput, error) {
	panic("unimplemented")
}

// GetBucketLocation implements [s3iface.S3API].
func (m *MockS3Client) GetBucketLocation(*s3.GetBucketLocationInput) (*s3.GetBucketLocationOutput, error) {
	panic("unimplemented")
}

// GetBucketLocationRequest implements [s3iface.S3API].
func (m *MockS3Client) GetBucketLocationRequest(*s3.GetBucketLocationInput) (*request.Request, *s3.GetBucketLocationOutput) {
	panic("unimplemented")
}

// GetBucketLocationWithContext implements [s3iface.S3API].
func (m *MockS3Client) GetBucketLocationWithContext(aws.Context, *s3.GetBucketLocationInput, ...request.Option) (*s3.GetBucketLocationOutput, error) {
	panic("unimplemented")
}

// GetBucketLogging implements [s3iface.S3API].
func (m *MockS3Client) GetBucketLogging(*s3.GetBucketLoggingInput) (*s3.GetBucketLoggingOutput, error) {
	panic("unimplemented")
}

// GetBucketLoggingRequest implements [s3iface.S3API].
func (m *MockS3Client) GetBucketLoggingRequest(*s3.GetBucketLoggingInput) (*request.Request, *s3.GetBucketLoggingOutput) {
	panic("unimplemented")
}

// GetBucketLoggingWithContext implements [s3iface.S3API].
func (m *MockS3Client) GetBucketLoggingWithContext(aws.Context, *s3.GetBucketLoggingInput, ...request.Option) (*s3.GetBucketLoggingOutput, error) {
	panic("unimplemented")
}

// GetBucketMetricsConfiguration implements [s3iface.S3API].
func (m *MockS3Client) GetBucketMetricsConfiguration(*s3.GetBucketMetricsConfigurationInput) (*s3.GetBucketMetricsConfigurationOutput, error) {
	panic("unimplemented")
}

// GetBucketMetricsConfigurationRequest implements [s3iface.S3API].
func (m *MockS3Client) GetBucketMetricsConfigurationRequest(*s3.GetBucketMetricsConfigurationInput) (*request.Request, *s3.GetBucketMetricsConfigurationOutput) {
	panic("unimplemented")
}

// GetBucketMetricsConfigurationWithContext implements [s3iface.S3API].
func (m *MockS3Client) GetBucketMetricsConfigurationWithContext(aws.Context, *s3.GetBucketMetricsConfigurationInput, ...request.Option) (*s3.GetBucketMetricsConfigurationOutput, error) {
	panic("unimplemented")
}

// GetBucketNotification implements [s3iface.S3API].
func (m *MockS3Client) GetBucketNotification(*s3.GetBucketNotificationConfigurationRequest) (*s3.NotificationConfigurationDeprecated, error) {
	panic("unimplemented")
}

// GetBucketNotificationConfiguration implements [s3iface.S3API].
func (m *MockS3Client) GetBucketNotificationConfiguration(*s3.GetBucketNotificationConfigurationRequest) (*s3.NotificationConfiguration, error) {
	panic("unimplemented")
}

// GetBucketNotificationConfigurationRequest implements [s3iface.S3API].
func (m *MockS3Client) GetBucketNotificationConfigurationRequest(*s3.GetBucketNotificationConfigurationRequest) (*request.Request, *s3.NotificationConfiguration) {
	panic("unimplemented")
}

// GetBucketNotificationConfigurationWithContext implements [s3iface.S3API].
func (m *MockS3Client) GetBucketNotificationConfigurationWithContext(aws.Context, *s3.GetBucketNotificationConfigurationRequest, ...request.Option) (*s3.NotificationConfiguration, error) {
	panic("unimplemented")
}

// GetBucketNotificationRequest implements [s3iface.S3API].
func (m *MockS3Client) GetBucketNotificationRequest(*s3.GetBucketNotificationConfigurationRequest) (*request.Request, *s3.NotificationConfigurationDeprecated) {
	panic("unimplemented")
}

// GetBucketNotificationWithContext implements [s3iface.S3API].
func (m *MockS3Client) GetBucketNotificationWithContext(aws.Context, *s3.GetBucketNotificationConfigurationRequest, ...request.Option) (*s3.NotificationConfigurationDeprecated, error) {
	panic("unimplemented")
}

// GetBucketOwnershipControls implements [s3iface.S3API].
func (m *MockS3Client) GetBucketOwnershipControls(*s3.GetBucketOwnershipControlsInput) (*s3.GetBucketOwnershipControlsOutput, error) {
	panic("unimplemented")
}

// GetBucketOwnershipControlsRequest implements [s3iface.S3API].
func (m *MockS3Client) GetBucketOwnershipControlsRequest(*s3.GetBucketOwnershipControlsInput) (*request.Request, *s3.GetBucketOwnershipControlsOutput) {
	panic("unimplemented")
}

// GetBucketOwnershipControlsWithContext implements [s3iface.S3API].
func (m *MockS3Client) GetBucketOwnershipControlsWithContext(aws.Context, *s3.GetBucketOwnershipControlsInput, ...request.Option) (*s3.GetBucketOwnershipControlsOutput, error) {
	panic("unimplemented")
}

// GetBucketPolicy implements [s3iface.S3API].
func (m *MockS3Client) GetBucketPolicy(*s3.GetBucketPolicyInput) (*s3.GetBucketPolicyOutput, error) {
	panic("unimplemented")
}

// GetBucketPolicyRequest implements [s3iface.S3API].
func (m *MockS3Client) GetBucketPolicyRequest(*s3.GetBucketPolicyInput) (*request.Request, *s3.GetBucketPolicyOutput) {
	panic("unimplemented")
}

// GetBucketPolicyStatus implements [s3iface.S3API].
func (m *MockS3Client) GetBucketPolicyStatus(*s3.GetBucketPolicyStatusInput) (*s3.GetBucketPolicyStatusOutput, error) {
	panic("unimplemented")
}

// GetBucketPolicyStatusRequest implements [s3iface.S3API].
func (m *MockS3Client) GetBucketPolicyStatusRequest(*s3.GetBucketPolicyStatusInput) (*request.Request, *s3.GetBucketPolicyStatusOutput) {
	panic("unimplemented")
}

// GetBucketPolicyStatusWithContext implements [s3iface.S3API].
func (m *MockS3Client) GetBucketPolicyStatusWithContext(aws.Context, *s3.GetBucketPolicyStatusInput, ...request.Option) (*s3.GetBucketPolicyStatusOutput, error) {
	panic("unimplemented")
}

// GetBucketPolicyWithContext implements [s3iface.S3API].
func (m *MockS3Client) GetBucketPolicyWithContext(aws.Context, *s3.GetBucketPolicyInput, ...request.Option) (*s3.GetBucketPolicyOutput, error) {
	panic("unimplemented")
}

// GetBucketReplication implements [s3iface.S3API].
func (m *MockS3Client) GetBucketReplication(*s3.GetBucketReplicationInput) (*s3.GetBucketReplicationOutput, error) {
	panic("unimplemented")
}

// GetBucketReplicationRequest implements [s3iface.S3API].
func (m *MockS3Client) GetBucketReplicationRequest(*s3.GetBucketReplicationInput) (*request.Request, *s3.GetBucketReplicationOutput) {
	panic("unimplemented")
}

// GetBucketReplicationWithContext implements [s3iface.S3API].
func (m *MockS3Client) GetBucketReplicationWithContext(aws.Context, *s3.GetBucketReplicationInput, ...request.Option) (*s3.GetBucketReplicationOutput, error) {
	panic("unimplemented")
}

// GetBucketRequestPayment implements [s3iface.S3API].
func (m *MockS3Client) GetBucketRequestPayment(*s3.GetBucketRequestPaymentInput) (*s3.GetBucketRequestPaymentOutput, error) {
	panic("unimplemented")
}

// GetBucketRequestPaymentRequest implements [s3iface.S3API].
func (m *MockS3Client) GetBucketRequestPaymentRequest(*s3.GetBucketRequestPaymentInput) (*request.Request, *s3.GetBucketRequestPaymentOutput) {
	panic("unimplemented")
}

// GetBucketRequestPaymentWithContext implements [s3iface.S3API].
func (m *MockS3Client) GetBucketRequestPaymentWithContext(aws.Context, *s3.GetBucketRequestPaymentInput, ...request.Option) (*s3.GetBucketRequestPaymentOutput, error) {
	panic("unimplemented")
}

// GetBucketTagging implements [s3iface.S3API].
func (m *MockS3Client) GetBucketTagging(*s3.GetBucketTaggingInput) (*s3.GetBucketTaggingOutput, error) {
	panic("unimplemented")
}

// GetBucketTaggingRequest implements [s3iface.S3API].
func (m *MockS3Client) GetBucketTaggingRequest(*s3.GetBucketTaggingInput) (*request.Request, *s3.GetBucketTaggingOutput) {
	panic("unimplemented")
}

// GetBucketTaggingWithContext implements [s3iface.S3API].
func (m *MockS3Client) GetBucketTaggingWithContext(aws.Context, *s3.GetBucketTaggingInput, ...request.Option) (*s3.GetBucketTaggingOutput, error) {
	panic("unimplemented")
}

// GetBucketVersioning implements [s3iface.S3API].
func (m *MockS3Client) GetBucketVersioning(*s3.GetBucketVersioningInput) (*s3.GetBucketVersioningOutput, error) {
	panic("unimplemented")
}

// GetBucketVersioningRequest implements [s3iface.S3API].
func (m *MockS3Client) GetBucketVersioningRequest(*s3.GetBucketVersioningInput) (*request.Request, *s3.GetBucketVersioningOutput) {
	panic("unimplemented")
}

// GetBucketVersioningWithContext implements [s3iface.S3API].
func (m *MockS3Client) GetBucketVersioningWithContext(aws.Context, *s3.GetBucketVersioningInput, ...request.Option) (*s3.GetBucketVersioningOutput, error) {
	panic("unimplemented")
}

// GetBucketWebsite implements [s3iface.S3API].
func (m *MockS3Client) GetBucketWebsite(*s3.GetBucketWebsiteInput) (*s3.GetBucketWebsiteOutput, error) {
	panic("unimplemented")
}

// GetBucketWebsiteRequest implements [s3iface.S3API].
func (m *MockS3Client) GetBucketWebsiteRequest(*s3.GetBucketWebsiteInput) (*request.Request, *s3.GetBucketWebsiteOutput) {
	panic("unimplemented")
}

// GetBucketWebsiteWithContext implements [s3iface.S3API].
func (m *MockS3Client) GetBucketWebsiteWithContext(aws.Context, *s3.GetBucketWebsiteInput, ...request.Option) (*s3.GetBucketWebsiteOutput, error) {
	panic("unimplemented")
}

// GetObject implements [s3iface.S3API].
func (m *MockS3Client) GetObject(*s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	panic("unimplemented")
}

// GetObjectAcl implements [s3iface.S3API].
func (m *MockS3Client) GetObjectAcl(*s3.GetObjectAclInput) (*s3.GetObjectAclOutput, error) {
	panic("unimplemented")
}

// GetObjectAclRequest implements [s3iface.S3API].
func (m *MockS3Client) GetObjectAclRequest(*s3.GetObjectAclInput) (*request.Request, *s3.GetObjectAclOutput) {
	panic("unimplemented")
}

// GetObjectAclWithContext implements [s3iface.S3API].
func (m *MockS3Client) GetObjectAclWithContext(aws.Context, *s3.GetObjectAclInput, ...request.Option) (*s3.GetObjectAclOutput, error) {
	panic("unimplemented")
}

// GetObjectAttributes implements [s3iface.S3API].
func (m *MockS3Client) GetObjectAttributes(*s3.GetObjectAttributesInput) (*s3.GetObjectAttributesOutput, error) {
	panic("unimplemented")
}

// GetObjectAttributesRequest implements [s3iface.S3API].
func (m *MockS3Client) GetObjectAttributesRequest(*s3.GetObjectAttributesInput) (*request.Request, *s3.GetObjectAttributesOutput) {
	panic("unimplemented")
}

// GetObjectAttributesWithContext implements [s3iface.S3API].
func (m *MockS3Client) GetObjectAttributesWithContext(aws.Context, *s3.GetObjectAttributesInput, ...request.Option) (*s3.GetObjectAttributesOutput, error) {
	panic("unimplemented")
}

// GetObjectLegalHold implements [s3iface.S3API].
func (m *MockS3Client) GetObjectLegalHold(*s3.GetObjectLegalHoldInput) (*s3.GetObjectLegalHoldOutput, error) {
	panic("unimplemented")
}

// GetObjectLegalHoldRequest implements [s3iface.S3API].
func (m *MockS3Client) GetObjectLegalHoldRequest(*s3.GetObjectLegalHoldInput) (*request.Request, *s3.GetObjectLegalHoldOutput) {
	panic("unimplemented")
}

// GetObjectLegalHoldWithContext implements [s3iface.S3API].
func (m *MockS3Client) GetObjectLegalHoldWithContext(aws.Context, *s3.GetObjectLegalHoldInput, ...request.Option) (*s3.GetObjectLegalHoldOutput, error) {
	panic("unimplemented")
}

// GetObjectLockConfiguration implements [s3iface.S3API].
func (m *MockS3Client) GetObjectLockConfiguration(*s3.GetObjectLockConfigurationInput) (*s3.GetObjectLockConfigurationOutput, error) {
	panic("unimplemented")
}

// GetObjectLockConfigurationRequest implements [s3iface.S3API].
func (m *MockS3Client) GetObjectLockConfigurationRequest(*s3.GetObjectLockConfigurationInput) (*request.Request, *s3.GetObjectLockConfigurationOutput) {
	panic("unimplemented")
}

// GetObjectLockConfigurationWithContext implements [s3iface.S3API].
func (m *MockS3Client) GetObjectLockConfigurationWithContext(aws.Context, *s3.GetObjectLockConfigurationInput, ...request.Option) (*s3.GetObjectLockConfigurationOutput, error) {
	panic("unimplemented")
}

// GetObjectRequest implements [s3iface.S3API].
func (m *MockS3Client) GetObjectRequest(input *s3.GetObjectInput) (*request.Request, *s3.GetObjectOutput) {
	args := m.Called(input)
	var req *request.Request
	var out *s3.GetObjectOutput
	if args.Get(0) != nil {
		req = args.Get(0).(*request.Request)
	}
	if args.Get(1) != nil {
		out = args.Get(1).(*s3.GetObjectOutput)
	}
	return req, out
}

// GetObjectRetention implements [s3iface.S3API].
func (m *MockS3Client) GetObjectRetention(*s3.GetObjectRetentionInput) (*s3.GetObjectRetentionOutput, error) {
	panic("unimplemented")
}

// GetObjectRetentionRequest implements [s3iface.S3API].
func (m *MockS3Client) GetObjectRetentionRequest(*s3.GetObjectRetentionInput) (*request.Request, *s3.GetObjectRetentionOutput) {
	panic("unimplemented")
}

// GetObjectRetentionWithContext implements [s3iface.S3API].
func (m *MockS3Client) GetObjectRetentionWithContext(aws.Context, *s3.GetObjectRetentionInput, ...request.Option) (*s3.GetObjectRetentionOutput, error) {
	panic("unimplemented")
}

// GetObjectTagging implements [s3iface.S3API].
func (m *MockS3Client) GetObjectTagging(*s3.GetObjectTaggingInput) (*s3.GetObjectTaggingOutput, error) {
	panic("unimplemented")
}

// GetObjectTaggingRequest implements [s3iface.S3API].
func (m *MockS3Client) GetObjectTaggingRequest(*s3.GetObjectTaggingInput) (*request.Request, *s3.GetObjectTaggingOutput) {
	panic("unimplemented")
}

// GetObjectTaggingWithContext implements [s3iface.S3API].
func (m *MockS3Client) GetObjectTaggingWithContext(aws.Context, *s3.GetObjectTaggingInput, ...request.Option) (*s3.GetObjectTaggingOutput, error) {
	panic("unimplemented")
}

// GetObjectTorrent implements [s3iface.S3API].
func (m *MockS3Client) GetObjectTorrent(*s3.GetObjectTorrentInput) (*s3.GetObjectTorrentOutput, error) {
	panic("unimplemented")
}

// GetObjectTorrentRequest implements [s3iface.S3API].
func (m *MockS3Client) GetObjectTorrentRequest(*s3.GetObjectTorrentInput) (*request.Request, *s3.GetObjectTorrentOutput) {
	panic("unimplemented")
}

// GetObjectTorrentWithContext implements [s3iface.S3API].
func (m *MockS3Client) GetObjectTorrentWithContext(aws.Context, *s3.GetObjectTorrentInput, ...request.Option) (*s3.GetObjectTorrentOutput, error) {
	panic("unimplemented")
}

// GetObjectWithContext implements [s3iface.S3API].
func (m *MockS3Client) GetObjectWithContext(aws.Context, *s3.GetObjectInput, ...request.Option) (*s3.GetObjectOutput, error) {
	panic("unimplemented")
}

// GetPublicAccessBlock implements [s3iface.S3API].
func (m *MockS3Client) GetPublicAccessBlock(*s3.GetPublicAccessBlockInput) (*s3.GetPublicAccessBlockOutput, error) {
	panic("unimplemented")
}

// GetPublicAccessBlockRequest implements [s3iface.S3API].
func (m *MockS3Client) GetPublicAccessBlockRequest(*s3.GetPublicAccessBlockInput) (*request.Request, *s3.GetPublicAccessBlockOutput) {
	panic("unimplemented")
}

// GetPublicAccessBlockWithContext implements [s3iface.S3API].
func (m *MockS3Client) GetPublicAccessBlockWithContext(aws.Context, *s3.GetPublicAccessBlockInput, ...request.Option) (*s3.GetPublicAccessBlockOutput, error) {
	panic("unimplemented")
}

// HeadBucket implements [s3iface.S3API].
func (m *MockS3Client) HeadBucket(*s3.HeadBucketInput) (*s3.HeadBucketOutput, error) {
	panic("unimplemented")
}

// HeadBucketRequest implements [s3iface.S3API].
func (m *MockS3Client) HeadBucketRequest(*s3.HeadBucketInput) (*request.Request, *s3.HeadBucketOutput) {
	panic("unimplemented")
}

// HeadBucketWithContext implements [s3iface.S3API].
func (m *MockS3Client) HeadBucketWithContext(aws.Context, *s3.HeadBucketInput, ...request.Option) (*s3.HeadBucketOutput, error) {
	panic("unimplemented")
}

// HeadObject implements [s3iface.S3API].
func (m *MockS3Client) HeadObject(*s3.HeadObjectInput) (*s3.HeadObjectOutput, error) {
	panic("unimplemented")
}

// HeadObjectRequest implements [s3iface.S3API].
func (m *MockS3Client) HeadObjectRequest(*s3.HeadObjectInput) (*request.Request, *s3.HeadObjectOutput) {
	panic("unimplemented")
}

// HeadObjectWithContext implements [s3iface.S3API].
func (m *MockS3Client) HeadObjectWithContext(aws.Context, *s3.HeadObjectInput, ...request.Option) (*s3.HeadObjectOutput, error) {
	panic("unimplemented")
}

// ListBucketAnalyticsConfigurations implements [s3iface.S3API].
func (m *MockS3Client) ListBucketAnalyticsConfigurations(*s3.ListBucketAnalyticsConfigurationsInput) (*s3.ListBucketAnalyticsConfigurationsOutput, error) {
	panic("unimplemented")
}

// ListBucketAnalyticsConfigurationsRequest implements [s3iface.S3API].
func (m *MockS3Client) ListBucketAnalyticsConfigurationsRequest(*s3.ListBucketAnalyticsConfigurationsInput) (*request.Request, *s3.ListBucketAnalyticsConfigurationsOutput) {
	panic("unimplemented")
}

// ListBucketAnalyticsConfigurationsWithContext implements [s3iface.S3API].
func (m *MockS3Client) ListBucketAnalyticsConfigurationsWithContext(aws.Context, *s3.ListBucketAnalyticsConfigurationsInput, ...request.Option) (*s3.ListBucketAnalyticsConfigurationsOutput, error) {
	panic("unimplemented")
}

// ListBucketIntelligentTieringConfigurations implements [s3iface.S3API].
func (m *MockS3Client) ListBucketIntelligentTieringConfigurations(*s3.ListBucketIntelligentTieringConfigurationsInput) (*s3.ListBucketIntelligentTieringConfigurationsOutput, error) {
	panic("unimplemented")
}

// ListBucketIntelligentTieringConfigurationsRequest implements [s3iface.S3API].
func (m *MockS3Client) ListBucketIntelligentTieringConfigurationsRequest(*s3.ListBucketIntelligentTieringConfigurationsInput) (*request.Request, *s3.ListBucketIntelligentTieringConfigurationsOutput) {
	panic("unimplemented")
}

// ListBucketIntelligentTieringConfigurationsWithContext implements [s3iface.S3API].
func (m *MockS3Client) ListBucketIntelligentTieringConfigurationsWithContext(aws.Context, *s3.ListBucketIntelligentTieringConfigurationsInput, ...request.Option) (*s3.ListBucketIntelligentTieringConfigurationsOutput, error) {
	panic("unimplemented")
}

// ListBucketInventoryConfigurations implements [s3iface.S3API].
func (m *MockS3Client) ListBucketInventoryConfigurations(*s3.ListBucketInventoryConfigurationsInput) (*s3.ListBucketInventoryConfigurationsOutput, error) {
	panic("unimplemented")
}

// ListBucketInventoryConfigurationsRequest implements [s3iface.S3API].
func (m *MockS3Client) ListBucketInventoryConfigurationsRequest(*s3.ListBucketInventoryConfigurationsInput) (*request.Request, *s3.ListBucketInventoryConfigurationsOutput) {
	panic("unimplemented")
}

// ListBucketInventoryConfigurationsWithContext implements [s3iface.S3API].
func (m *MockS3Client) ListBucketInventoryConfigurationsWithContext(aws.Context, *s3.ListBucketInventoryConfigurationsInput, ...request.Option) (*s3.ListBucketInventoryConfigurationsOutput, error) {
	panic("unimplemented")
}

// ListBucketMetricsConfigurations implements [s3iface.S3API].
func (m *MockS3Client) ListBucketMetricsConfigurations(*s3.ListBucketMetricsConfigurationsInput) (*s3.ListBucketMetricsConfigurationsOutput, error) {
	panic("unimplemented")
}

// ListBucketMetricsConfigurationsRequest implements [s3iface.S3API].
func (m *MockS3Client) ListBucketMetricsConfigurationsRequest(*s3.ListBucketMetricsConfigurationsInput) (*request.Request, *s3.ListBucketMetricsConfigurationsOutput) {
	panic("unimplemented")
}

// ListBucketMetricsConfigurationsWithContext implements [s3iface.S3API].
func (m *MockS3Client) ListBucketMetricsConfigurationsWithContext(aws.Context, *s3.ListBucketMetricsConfigurationsInput, ...request.Option) (*s3.ListBucketMetricsConfigurationsOutput, error) {
	panic("unimplemented")
}

// ListBuckets implements [s3iface.S3API].
func (m *MockS3Client) ListBuckets(*s3.ListBucketsInput) (*s3.ListBucketsOutput, error) {
	panic("unimplemented")
}

// ListBucketsRequest implements [s3iface.S3API].
func (m *MockS3Client) ListBucketsRequest(*s3.ListBucketsInput) (*request.Request, *s3.ListBucketsOutput) {
	panic("unimplemented")
}

// ListBucketsWithContext implements [s3iface.S3API].
func (m *MockS3Client) ListBucketsWithContext(aws.Context, *s3.ListBucketsInput, ...request.Option) (*s3.ListBucketsOutput, error) {
	panic("unimplemented")
}

// ListDirectoryBuckets implements [s3iface.S3API].
func (m *MockS3Client) ListDirectoryBuckets(*s3.ListDirectoryBucketsInput) (*s3.ListDirectoryBucketsOutput, error) {
	panic("unimplemented")
}

// ListDirectoryBucketsPages implements [s3iface.S3API].
func (m *MockS3Client) ListDirectoryBucketsPages(*s3.ListDirectoryBucketsInput, func(*s3.ListDirectoryBucketsOutput, bool) bool) error {
	panic("unimplemented")
}

// ListDirectoryBucketsPagesWithContext implements [s3iface.S3API].
func (m *MockS3Client) ListDirectoryBucketsPagesWithContext(aws.Context, *s3.ListDirectoryBucketsInput, func(*s3.ListDirectoryBucketsOutput, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// ListDirectoryBucketsRequest implements [s3iface.S3API].
func (m *MockS3Client) ListDirectoryBucketsRequest(*s3.ListDirectoryBucketsInput) (*request.Request, *s3.ListDirectoryBucketsOutput) {
	panic("unimplemented")
}

// ListDirectoryBucketsWithContext implements [s3iface.S3API].
func (m *MockS3Client) ListDirectoryBucketsWithContext(aws.Context, *s3.ListDirectoryBucketsInput, ...request.Option) (*s3.ListDirectoryBucketsOutput, error) {
	panic("unimplemented")
}

// ListMultipartUploads implements [s3iface.S3API].
func (m *MockS3Client) ListMultipartUploads(*s3.ListMultipartUploadsInput) (*s3.ListMultipartUploadsOutput, error) {
	panic("unimplemented")
}

// ListMultipartUploadsPages implements [s3iface.S3API].
func (m *MockS3Client) ListMultipartUploadsPages(*s3.ListMultipartUploadsInput, func(*s3.ListMultipartUploadsOutput, bool) bool) error {
	panic("unimplemented")
}

// ListMultipartUploadsPagesWithContext implements [s3iface.S3API].
func (m *MockS3Client) ListMultipartUploadsPagesWithContext(aws.Context, *s3.ListMultipartUploadsInput, func(*s3.ListMultipartUploadsOutput, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// ListMultipartUploadsRequest implements [s3iface.S3API].
func (m *MockS3Client) ListMultipartUploadsRequest(*s3.ListMultipartUploadsInput) (*request.Request, *s3.ListMultipartUploadsOutput) {
	panic("unimplemented")
}

// ListMultipartUploadsWithContext implements [s3iface.S3API].
func (m *MockS3Client) ListMultipartUploadsWithContext(aws.Context, *s3.ListMultipartUploadsInput, ...request.Option) (*s3.ListMultipartUploadsOutput, error) {
	panic("unimplemented")
}

// ListObjectVersions implements [s3iface.S3API].
func (m *MockS3Client) ListObjectVersions(*s3.ListObjectVersionsInput) (*s3.ListObjectVersionsOutput, error) {
	panic("unimplemented")
}

// ListObjectVersionsPages implements [s3iface.S3API].
func (m *MockS3Client) ListObjectVersionsPages(*s3.ListObjectVersionsInput, func(*s3.ListObjectVersionsOutput, bool) bool) error {
	panic("unimplemented")
}

// ListObjectVersionsPagesWithContext implements [s3iface.S3API].
func (m *MockS3Client) ListObjectVersionsPagesWithContext(aws.Context, *s3.ListObjectVersionsInput, func(*s3.ListObjectVersionsOutput, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// ListObjectVersionsRequest implements [s3iface.S3API].
func (m *MockS3Client) ListObjectVersionsRequest(*s3.ListObjectVersionsInput) (*request.Request, *s3.ListObjectVersionsOutput) {
	panic("unimplemented")
}

// ListObjectVersionsWithContext implements [s3iface.S3API].
func (m *MockS3Client) ListObjectVersionsWithContext(aws.Context, *s3.ListObjectVersionsInput, ...request.Option) (*s3.ListObjectVersionsOutput, error) {
	panic("unimplemented")
}

// ListObjects implements [s3iface.S3API].
func (m *MockS3Client) ListObjects(*s3.ListObjectsInput) (*s3.ListObjectsOutput, error) {
	panic("unimplemented")
}

// ListObjectsPages implements [s3iface.S3API].
func (m *MockS3Client) ListObjectsPages(*s3.ListObjectsInput, func(*s3.ListObjectsOutput, bool) bool) error {
	panic("unimplemented")
}

// ListObjectsPagesWithContext implements [s3iface.S3API].
func (m *MockS3Client) ListObjectsPagesWithContext(aws.Context, *s3.ListObjectsInput, func(*s3.ListObjectsOutput, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// ListObjectsRequest implements [s3iface.S3API].
func (m *MockS3Client) ListObjectsRequest(*s3.ListObjectsInput) (*request.Request, *s3.ListObjectsOutput) {
	panic("unimplemented")
}

// ListObjectsV2 implements [s3iface.S3API].
func (m *MockS3Client) ListObjectsV2(*s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error) {
	panic("unimplemented")
}

// ListObjectsV2Pages implements [s3iface.S3API].
func (m *MockS3Client) ListObjectsV2Pages(*s3.ListObjectsV2Input, func(*s3.ListObjectsV2Output, bool) bool) error {
	panic("unimplemented")
}

// ListObjectsV2PagesWithContext implements [s3iface.S3API].
func (m *MockS3Client) ListObjectsV2PagesWithContext(aws.Context, *s3.ListObjectsV2Input, func(*s3.ListObjectsV2Output, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// ListObjectsV2Request implements [s3iface.S3API].
func (m *MockS3Client) ListObjectsV2Request(*s3.ListObjectsV2Input) (*request.Request, *s3.ListObjectsV2Output) {
	panic("unimplemented")
}

// ListObjectsV2WithContext implements [s3iface.S3API].
func (m *MockS3Client) ListObjectsV2WithContext(aws.Context, *s3.ListObjectsV2Input, ...request.Option) (*s3.ListObjectsV2Output, error) {
	panic("unimplemented")
}

// ListObjectsWithContext implements [s3iface.S3API].
func (m *MockS3Client) ListObjectsWithContext(aws.Context, *s3.ListObjectsInput, ...request.Option) (*s3.ListObjectsOutput, error) {
	panic("unimplemented")
}

// ListParts implements [s3iface.S3API].
func (m *MockS3Client) ListParts(*s3.ListPartsInput) (*s3.ListPartsOutput, error) {
	panic("unimplemented")
}

// ListPartsPages implements [s3iface.S3API].
func (m *MockS3Client) ListPartsPages(*s3.ListPartsInput, func(*s3.ListPartsOutput, bool) bool) error {
	panic("unimplemented")
}

// ListPartsPagesWithContext implements [s3iface.S3API].
func (m *MockS3Client) ListPartsPagesWithContext(aws.Context, *s3.ListPartsInput, func(*s3.ListPartsOutput, bool) bool, ...request.Option) error {
	panic("unimplemented")
}

// ListPartsRequest implements [s3iface.S3API].
func (m *MockS3Client) ListPartsRequest(*s3.ListPartsInput) (*request.Request, *s3.ListPartsOutput) {
	panic("unimplemented")
}

// ListPartsWithContext implements [s3iface.S3API].
func (m *MockS3Client) ListPartsWithContext(aws.Context, *s3.ListPartsInput, ...request.Option) (*s3.ListPartsOutput, error) {
	panic("unimplemented")
}

// PutBucketAccelerateConfiguration implements [s3iface.S3API].
func (m *MockS3Client) PutBucketAccelerateConfiguration(*s3.PutBucketAccelerateConfigurationInput) (*s3.PutBucketAccelerateConfigurationOutput, error) {
	panic("unimplemented")
}

// PutBucketAccelerateConfigurationRequest implements [s3iface.S3API].
func (m *MockS3Client) PutBucketAccelerateConfigurationRequest(*s3.PutBucketAccelerateConfigurationInput) (*request.Request, *s3.PutBucketAccelerateConfigurationOutput) {
	panic("unimplemented")
}

// PutBucketAccelerateConfigurationWithContext implements [s3iface.S3API].
func (m *MockS3Client) PutBucketAccelerateConfigurationWithContext(aws.Context, *s3.PutBucketAccelerateConfigurationInput, ...request.Option) (*s3.PutBucketAccelerateConfigurationOutput, error) {
	panic("unimplemented")
}

// PutBucketAcl implements [s3iface.S3API].
func (m *MockS3Client) PutBucketAcl(*s3.PutBucketAclInput) (*s3.PutBucketAclOutput, error) {
	panic("unimplemented")
}

// PutBucketAclRequest implements [s3iface.S3API].
func (m *MockS3Client) PutBucketAclRequest(*s3.PutBucketAclInput) (*request.Request, *s3.PutBucketAclOutput) {
	panic("unimplemented")
}

// PutBucketAclWithContext implements [s3iface.S3API].
func (m *MockS3Client) PutBucketAclWithContext(aws.Context, *s3.PutBucketAclInput, ...request.Option) (*s3.PutBucketAclOutput, error) {
	panic("unimplemented")
}

// PutBucketAnalyticsConfiguration implements [s3iface.S3API].
func (m *MockS3Client) PutBucketAnalyticsConfiguration(*s3.PutBucketAnalyticsConfigurationInput) (*s3.PutBucketAnalyticsConfigurationOutput, error) {
	panic("unimplemented")
}

// PutBucketAnalyticsConfigurationRequest implements [s3iface.S3API].
func (m *MockS3Client) PutBucketAnalyticsConfigurationRequest(*s3.PutBucketAnalyticsConfigurationInput) (*request.Request, *s3.PutBucketAnalyticsConfigurationOutput) {
	panic("unimplemented")
}

// PutBucketAnalyticsConfigurationWithContext implements [s3iface.S3API].
func (m *MockS3Client) PutBucketAnalyticsConfigurationWithContext(aws.Context, *s3.PutBucketAnalyticsConfigurationInput, ...request.Option) (*s3.PutBucketAnalyticsConfigurationOutput, error) {
	panic("unimplemented")
}

// PutBucketCors implements [s3iface.S3API].
func (m *MockS3Client) PutBucketCors(*s3.PutBucketCorsInput) (*s3.PutBucketCorsOutput, error) {
	panic("unimplemented")
}

// PutBucketCorsRequest implements [s3iface.S3API].
func (m *MockS3Client) PutBucketCorsRequest(*s3.PutBucketCorsInput) (*request.Request, *s3.PutBucketCorsOutput) {
	panic("unimplemented")
}

// PutBucketCorsWithContext implements [s3iface.S3API].
func (m *MockS3Client) PutBucketCorsWithContext(aws.Context, *s3.PutBucketCorsInput, ...request.Option) (*s3.PutBucketCorsOutput, error) {
	panic("unimplemented")
}

// PutBucketEncryption implements [s3iface.S3API].
func (m *MockS3Client) PutBucketEncryption(*s3.PutBucketEncryptionInput) (*s3.PutBucketEncryptionOutput, error) {
	panic("unimplemented")
}

// PutBucketEncryptionRequest implements [s3iface.S3API].
func (m *MockS3Client) PutBucketEncryptionRequest(*s3.PutBucketEncryptionInput) (*request.Request, *s3.PutBucketEncryptionOutput) {
	panic("unimplemented")
}

// PutBucketEncryptionWithContext implements [s3iface.S3API].
func (m *MockS3Client) PutBucketEncryptionWithContext(aws.Context, *s3.PutBucketEncryptionInput, ...request.Option) (*s3.PutBucketEncryptionOutput, error) {
	panic("unimplemented")
}

// PutBucketIntelligentTieringConfiguration implements [s3iface.S3API].
func (m *MockS3Client) PutBucketIntelligentTieringConfiguration(*s3.PutBucketIntelligentTieringConfigurationInput) (*s3.PutBucketIntelligentTieringConfigurationOutput, error) {
	panic("unimplemented")
}

// PutBucketIntelligentTieringConfigurationRequest implements [s3iface.S3API].
func (m *MockS3Client) PutBucketIntelligentTieringConfigurationRequest(*s3.PutBucketIntelligentTieringConfigurationInput) (*request.Request, *s3.PutBucketIntelligentTieringConfigurationOutput) {
	panic("unimplemented")
}

// PutBucketIntelligentTieringConfigurationWithContext implements [s3iface.S3API].
func (m *MockS3Client) PutBucketIntelligentTieringConfigurationWithContext(aws.Context, *s3.PutBucketIntelligentTieringConfigurationInput, ...request.Option) (*s3.PutBucketIntelligentTieringConfigurationOutput, error) {
	panic("unimplemented")
}

// PutBucketInventoryConfiguration implements [s3iface.S3API].
func (m *MockS3Client) PutBucketInventoryConfiguration(*s3.PutBucketInventoryConfigurationInput) (*s3.PutBucketInventoryConfigurationOutput, error) {
	panic("unimplemented")
}

// PutBucketInventoryConfigurationRequest implements [s3iface.S3API].
func (m *MockS3Client) PutBucketInventoryConfigurationRequest(*s3.PutBucketInventoryConfigurationInput) (*request.Request, *s3.PutBucketInventoryConfigurationOutput) {
	panic("unimplemented")
}

// PutBucketInventoryConfigurationWithContext implements [s3iface.S3API].
func (m *MockS3Client) PutBucketInventoryConfigurationWithContext(aws.Context, *s3.PutBucketInventoryConfigurationInput, ...request.Option) (*s3.PutBucketInventoryConfigurationOutput, error) {
	panic("unimplemented")
}

// PutBucketLifecycle implements [s3iface.S3API].
func (m *MockS3Client) PutBucketLifecycle(*s3.PutBucketLifecycleInput) (*s3.PutBucketLifecycleOutput, error) {
	panic("unimplemented")
}

// PutBucketLifecycleConfiguration implements [s3iface.S3API].
func (m *MockS3Client) PutBucketLifecycleConfiguration(*s3.PutBucketLifecycleConfigurationInput) (*s3.PutBucketLifecycleConfigurationOutput, error) {
	panic("unimplemented")
}

// PutBucketLifecycleConfigurationRequest implements [s3iface.S3API].
func (m *MockS3Client) PutBucketLifecycleConfigurationRequest(*s3.PutBucketLifecycleConfigurationInput) (*request.Request, *s3.PutBucketLifecycleConfigurationOutput) {
	panic("unimplemented")
}

// PutBucketLifecycleConfigurationWithContext implements [s3iface.S3API].
func (m *MockS3Client) PutBucketLifecycleConfigurationWithContext(aws.Context, *s3.PutBucketLifecycleConfigurationInput, ...request.Option) (*s3.PutBucketLifecycleConfigurationOutput, error) {
	panic("unimplemented")
}

// PutBucketLifecycleRequest implements [s3iface.S3API].
func (m *MockS3Client) PutBucketLifecycleRequest(*s3.PutBucketLifecycleInput) (*request.Request, *s3.PutBucketLifecycleOutput) {
	panic("unimplemented")
}

// PutBucketLifecycleWithContext implements [s3iface.S3API].
func (m *MockS3Client) PutBucketLifecycleWithContext(aws.Context, *s3.PutBucketLifecycleInput, ...request.Option) (*s3.PutBucketLifecycleOutput, error) {
	panic("unimplemented")
}

// PutBucketLogging implements [s3iface.S3API].
func (m *MockS3Client) PutBucketLogging(*s3.PutBucketLoggingInput) (*s3.PutBucketLoggingOutput, error) {
	panic("unimplemented")
}

// PutBucketLoggingRequest implements [s3iface.S3API].
func (m *MockS3Client) PutBucketLoggingRequest(*s3.PutBucketLoggingInput) (*request.Request, *s3.PutBucketLoggingOutput) {
	panic("unimplemented")
}

// PutBucketLoggingWithContext implements [s3iface.S3API].
func (m *MockS3Client) PutBucketLoggingWithContext(aws.Context, *s3.PutBucketLoggingInput, ...request.Option) (*s3.PutBucketLoggingOutput, error) {
	panic("unimplemented")
}

// PutBucketMetricsConfiguration implements [s3iface.S3API].
func (m *MockS3Client) PutBucketMetricsConfiguration(*s3.PutBucketMetricsConfigurationInput) (*s3.PutBucketMetricsConfigurationOutput, error) {
	panic("unimplemented")
}

// PutBucketMetricsConfigurationRequest implements [s3iface.S3API].
func (m *MockS3Client) PutBucketMetricsConfigurationRequest(*s3.PutBucketMetricsConfigurationInput) (*request.Request, *s3.PutBucketMetricsConfigurationOutput) {
	panic("unimplemented")
}

// PutBucketMetricsConfigurationWithContext implements [s3iface.S3API].
func (m *MockS3Client) PutBucketMetricsConfigurationWithContext(aws.Context, *s3.PutBucketMetricsConfigurationInput, ...request.Option) (*s3.PutBucketMetricsConfigurationOutput, error) {
	panic("unimplemented")
}

// PutBucketNotification implements [s3iface.S3API].
func (m *MockS3Client) PutBucketNotification(*s3.PutBucketNotificationInput) (*s3.PutBucketNotificationOutput, error) {
	panic("unimplemented")
}

// PutBucketNotificationConfiguration implements [s3iface.S3API].
func (m *MockS3Client) PutBucketNotificationConfiguration(*s3.PutBucketNotificationConfigurationInput) (*s3.PutBucketNotificationConfigurationOutput, error) {
	panic("unimplemented")
}

// PutBucketNotificationConfigurationRequest implements [s3iface.S3API].
func (m *MockS3Client) PutBucketNotificationConfigurationRequest(*s3.PutBucketNotificationConfigurationInput) (*request.Request, *s3.PutBucketNotificationConfigurationOutput) {
	panic("unimplemented")
}

// PutBucketNotificationConfigurationWithContext implements [s3iface.S3API].
func (m *MockS3Client) PutBucketNotificationConfigurationWithContext(aws.Context, *s3.PutBucketNotificationConfigurationInput, ...request.Option) (*s3.PutBucketNotificationConfigurationOutput, error) {
	panic("unimplemented")
}

// PutBucketNotificationRequest implements [s3iface.S3API].
func (m *MockS3Client) PutBucketNotificationRequest(*s3.PutBucketNotificationInput) (*request.Request, *s3.PutBucketNotificationOutput) {
	panic("unimplemented")
}

// PutBucketNotificationWithContext implements [s3iface.S3API].
func (m *MockS3Client) PutBucketNotificationWithContext(aws.Context, *s3.PutBucketNotificationInput, ...request.Option) (*s3.PutBucketNotificationOutput, error) {
	panic("unimplemented")
}

// PutBucketOwnershipControls implements [s3iface.S3API].
func (m *MockS3Client) PutBucketOwnershipControls(*s3.PutBucketOwnershipControlsInput) (*s3.PutBucketOwnershipControlsOutput, error) {
	panic("unimplemented")
}

// PutBucketOwnershipControlsRequest implements [s3iface.S3API].
func (m *MockS3Client) PutBucketOwnershipControlsRequest(*s3.PutBucketOwnershipControlsInput) (*request.Request, *s3.PutBucketOwnershipControlsOutput) {
	panic("unimplemented")
}

// PutBucketOwnershipControlsWithContext implements [s3iface.S3API].
func (m *MockS3Client) PutBucketOwnershipControlsWithContext(aws.Context, *s3.PutBucketOwnershipControlsInput, ...request.Option) (*s3.PutBucketOwnershipControlsOutput, error) {
	panic("unimplemented")
}

// PutBucketPolicy implements [s3iface.S3API].
func (m *MockS3Client) PutBucketPolicy(*s3.PutBucketPolicyInput) (*s3.PutBucketPolicyOutput, error) {
	panic("unimplemented")
}

// PutBucketPolicyRequest implements [s3iface.S3API].
func (m *MockS3Client) PutBucketPolicyRequest(*s3.PutBucketPolicyInput) (*request.Request, *s3.PutBucketPolicyOutput) {
	panic("unimplemented")
}

// PutBucketPolicyWithContext implements [s3iface.S3API].
func (m *MockS3Client) PutBucketPolicyWithContext(aws.Context, *s3.PutBucketPolicyInput, ...request.Option) (*s3.PutBucketPolicyOutput, error) {
	panic("unimplemented")
}

// PutBucketReplication implements [s3iface.S3API].
func (m *MockS3Client) PutBucketReplication(*s3.PutBucketReplicationInput) (*s3.PutBucketReplicationOutput, error) {
	panic("unimplemented")
}

// PutBucketReplicationRequest implements [s3iface.S3API].
func (m *MockS3Client) PutBucketReplicationRequest(*s3.PutBucketReplicationInput) (*request.Request, *s3.PutBucketReplicationOutput) {
	panic("unimplemented")
}

// PutBucketReplicationWithContext implements [s3iface.S3API].
func (m *MockS3Client) PutBucketReplicationWithContext(aws.Context, *s3.PutBucketReplicationInput, ...request.Option) (*s3.PutBucketReplicationOutput, error) {
	panic("unimplemented")
}

// PutBucketRequestPayment implements [s3iface.S3API].
func (m *MockS3Client) PutBucketRequestPayment(*s3.PutBucketRequestPaymentInput) (*s3.PutBucketRequestPaymentOutput, error) {
	panic("unimplemented")
}

// PutBucketRequestPaymentRequest implements [s3iface.S3API].
func (m *MockS3Client) PutBucketRequestPaymentRequest(*s3.PutBucketRequestPaymentInput) (*request.Request, *s3.PutBucketRequestPaymentOutput) {
	panic("unimplemented")
}

// PutBucketRequestPaymentWithContext implements [s3iface.S3API].
func (m *MockS3Client) PutBucketRequestPaymentWithContext(aws.Context, *s3.PutBucketRequestPaymentInput, ...request.Option) (*s3.PutBucketRequestPaymentOutput, error) {
	panic("unimplemented")
}

// PutBucketTagging implements [s3iface.S3API].
func (m *MockS3Client) PutBucketTagging(*s3.PutBucketTaggingInput) (*s3.PutBucketTaggingOutput, error) {
	panic("unimplemented")
}

// PutBucketTaggingRequest implements [s3iface.S3API].
func (m *MockS3Client) PutBucketTaggingRequest(*s3.PutBucketTaggingInput) (*request.Request, *s3.PutBucketTaggingOutput) {
	panic("unimplemented")
}

// PutBucketTaggingWithContext implements [s3iface.S3API].
func (m *MockS3Client) PutBucketTaggingWithContext(aws.Context, *s3.PutBucketTaggingInput, ...request.Option) (*s3.PutBucketTaggingOutput, error) {
	panic("unimplemented")
}

// PutBucketVersioning implements [s3iface.S3API].
func (m *MockS3Client) PutBucketVersioning(*s3.PutBucketVersioningInput) (*s3.PutBucketVersioningOutput, error) {
	panic("unimplemented")
}

// PutBucketVersioningRequest implements [s3iface.S3API].
func (m *MockS3Client) PutBucketVersioningRequest(*s3.PutBucketVersioningInput) (*request.Request, *s3.PutBucketVersioningOutput) {
	panic("unimplemented")
}

// PutBucketVersioningWithContext implements [s3iface.S3API].
func (m *MockS3Client) PutBucketVersioningWithContext(aws.Context, *s3.PutBucketVersioningInput, ...request.Option) (*s3.PutBucketVersioningOutput, error) {
	panic("unimplemented")
}

// PutBucketWebsite implements [s3iface.S3API].
func (m *MockS3Client) PutBucketWebsite(*s3.PutBucketWebsiteInput) (*s3.PutBucketWebsiteOutput, error) {
	panic("unimplemented")
}

// PutBucketWebsiteRequest implements [s3iface.S3API].
func (m *MockS3Client) PutBucketWebsiteRequest(*s3.PutBucketWebsiteInput) (*request.Request, *s3.PutBucketWebsiteOutput) {
	panic("unimplemented")
}

// PutBucketWebsiteWithContext implements [s3iface.S3API].
func (m *MockS3Client) PutBucketWebsiteWithContext(aws.Context, *s3.PutBucketWebsiteInput, ...request.Option) (*s3.PutBucketWebsiteOutput, error) {
	panic("unimplemented")
}

// PutObjectAcl implements [s3iface.S3API].
func (m *MockS3Client) PutObjectAcl(*s3.PutObjectAclInput) (*s3.PutObjectAclOutput, error) {
	panic("unimplemented")
}

// PutObjectAclRequest implements [s3iface.S3API].
func (m *MockS3Client) PutObjectAclRequest(*s3.PutObjectAclInput) (*request.Request, *s3.PutObjectAclOutput) {
	panic("unimplemented")
}

// PutObjectAclWithContext implements [s3iface.S3API].
func (m *MockS3Client) PutObjectAclWithContext(aws.Context, *s3.PutObjectAclInput, ...request.Option) (*s3.PutObjectAclOutput, error) {
	panic("unimplemented")
}

// PutObjectLegalHold implements [s3iface.S3API].
func (m *MockS3Client) PutObjectLegalHold(*s3.PutObjectLegalHoldInput) (*s3.PutObjectLegalHoldOutput, error) {
	panic("unimplemented")
}

// PutObjectLegalHoldRequest implements [s3iface.S3API].
func (m *MockS3Client) PutObjectLegalHoldRequest(*s3.PutObjectLegalHoldInput) (*request.Request, *s3.PutObjectLegalHoldOutput) {
	panic("unimplemented")
}

// PutObjectLegalHoldWithContext implements [s3iface.S3API].
func (m *MockS3Client) PutObjectLegalHoldWithContext(aws.Context, *s3.PutObjectLegalHoldInput, ...request.Option) (*s3.PutObjectLegalHoldOutput, error) {
	panic("unimplemented")
}

// PutObjectLockConfiguration implements [s3iface.S3API].
func (m *MockS3Client) PutObjectLockConfiguration(*s3.PutObjectLockConfigurationInput) (*s3.PutObjectLockConfigurationOutput, error) {
	panic("unimplemented")
}

// PutObjectLockConfigurationRequest implements [s3iface.S3API].
func (m *MockS3Client) PutObjectLockConfigurationRequest(*s3.PutObjectLockConfigurationInput) (*request.Request, *s3.PutObjectLockConfigurationOutput) {
	panic("unimplemented")
}

// PutObjectLockConfigurationWithContext implements [s3iface.S3API].
func (m *MockS3Client) PutObjectLockConfigurationWithContext(aws.Context, *s3.PutObjectLockConfigurationInput, ...request.Option) (*s3.PutObjectLockConfigurationOutput, error) {
	panic("unimplemented")
}

// PutObjectRequest implements [s3iface.S3API].
func (m *MockS3Client) PutObjectRequest(*s3.PutObjectInput) (*request.Request, *s3.PutObjectOutput) {
	panic("unimplemented")
}

// PutObjectRetention implements [s3iface.S3API].
func (m *MockS3Client) PutObjectRetention(*s3.PutObjectRetentionInput) (*s3.PutObjectRetentionOutput, error) {
	panic("unimplemented")
}

// PutObjectRetentionRequest implements [s3iface.S3API].
func (m *MockS3Client) PutObjectRetentionRequest(*s3.PutObjectRetentionInput) (*request.Request, *s3.PutObjectRetentionOutput) {
	panic("unimplemented")
}

// PutObjectRetentionWithContext implements [s3iface.S3API].
func (m *MockS3Client) PutObjectRetentionWithContext(aws.Context, *s3.PutObjectRetentionInput, ...request.Option) (*s3.PutObjectRetentionOutput, error) {
	panic("unimplemented")
}

// PutObjectTagging implements [s3iface.S3API].
func (m *MockS3Client) PutObjectTagging(*s3.PutObjectTaggingInput) (*s3.PutObjectTaggingOutput, error) {
	panic("unimplemented")
}

// PutObjectTaggingRequest implements [s3iface.S3API].
func (m *MockS3Client) PutObjectTaggingRequest(*s3.PutObjectTaggingInput) (*request.Request, *s3.PutObjectTaggingOutput) {
	panic("unimplemented")
}

// PutObjectTaggingWithContext implements [s3iface.S3API].
func (m *MockS3Client) PutObjectTaggingWithContext(aws.Context, *s3.PutObjectTaggingInput, ...request.Option) (*s3.PutObjectTaggingOutput, error) {
	panic("unimplemented")
}

// PutObjectWithContext implements [s3iface.S3API].
func (m *MockS3Client) PutObjectWithContext(aws.Context, *s3.PutObjectInput, ...request.Option) (*s3.PutObjectOutput, error) {
	panic("unimplemented")
}

// PutPublicAccessBlock implements [s3iface.S3API].
func (m *MockS3Client) PutPublicAccessBlock(*s3.PutPublicAccessBlockInput) (*s3.PutPublicAccessBlockOutput, error) {
	panic("unimplemented")
}

// PutPublicAccessBlockRequest implements [s3iface.S3API].
func (m *MockS3Client) PutPublicAccessBlockRequest(*s3.PutPublicAccessBlockInput) (*request.Request, *s3.PutPublicAccessBlockOutput) {
	panic("unimplemented")
}

// PutPublicAccessBlockWithContext implements [s3iface.S3API].
func (m *MockS3Client) PutPublicAccessBlockWithContext(aws.Context, *s3.PutPublicAccessBlockInput, ...request.Option) (*s3.PutPublicAccessBlockOutput, error) {
	panic("unimplemented")
}

// RestoreObject implements [s3iface.S3API].
func (m *MockS3Client) RestoreObject(*s3.RestoreObjectInput) (*s3.RestoreObjectOutput, error) {
	panic("unimplemented")
}

// RestoreObjectRequest implements [s3iface.S3API].
func (m *MockS3Client) RestoreObjectRequest(*s3.RestoreObjectInput) (*request.Request, *s3.RestoreObjectOutput) {
	panic("unimplemented")
}

// RestoreObjectWithContext implements [s3iface.S3API].
func (m *MockS3Client) RestoreObjectWithContext(aws.Context, *s3.RestoreObjectInput, ...request.Option) (*s3.RestoreObjectOutput, error) {
	panic("unimplemented")
}

// SelectObjectContent implements [s3iface.S3API].
func (m *MockS3Client) SelectObjectContent(*s3.SelectObjectContentInput) (*s3.SelectObjectContentOutput, error) {
	panic("unimplemented")
}

// SelectObjectContentRequest implements [s3iface.S3API].
func (m *MockS3Client) SelectObjectContentRequest(*s3.SelectObjectContentInput) (*request.Request, *s3.SelectObjectContentOutput) {
	panic("unimplemented")
}

// SelectObjectContentWithContext implements [s3iface.S3API].
func (m *MockS3Client) SelectObjectContentWithContext(aws.Context, *s3.SelectObjectContentInput, ...request.Option) (*s3.SelectObjectContentOutput, error) {
	panic("unimplemented")
}

// UploadPart implements [s3iface.S3API].
func (m *MockS3Client) UploadPart(*s3.UploadPartInput) (*s3.UploadPartOutput, error) {
	panic("unimplemented")
}

// UploadPartCopy implements [s3iface.S3API].
func (m *MockS3Client) UploadPartCopy(*s3.UploadPartCopyInput) (*s3.UploadPartCopyOutput, error) {
	panic("unimplemented")
}

// UploadPartCopyRequest implements [s3iface.S3API].
func (m *MockS3Client) UploadPartCopyRequest(*s3.UploadPartCopyInput) (*request.Request, *s3.UploadPartCopyOutput) {
	panic("unimplemented")
}

// UploadPartCopyWithContext implements [s3iface.S3API].
func (m *MockS3Client) UploadPartCopyWithContext(aws.Context, *s3.UploadPartCopyInput, ...request.Option) (*s3.UploadPartCopyOutput, error) {
	panic("unimplemented")
}

// UploadPartRequest implements [s3iface.S3API].
func (m *MockS3Client) UploadPartRequest(*s3.UploadPartInput) (*request.Request, *s3.UploadPartOutput) {
	panic("unimplemented")
}

// UploadPartWithContext implements [s3iface.S3API].
func (m *MockS3Client) UploadPartWithContext(aws.Context, *s3.UploadPartInput, ...request.Option) (*s3.UploadPartOutput, error) {
	panic("unimplemented")
}

// WaitUntilBucketExists implements [s3iface.S3API].
func (m *MockS3Client) WaitUntilBucketExists(*s3.HeadBucketInput) error {
	panic("unimplemented")
}

// WaitUntilBucketExistsWithContext implements [s3iface.S3API].
func (m *MockS3Client) WaitUntilBucketExistsWithContext(aws.Context, *s3.HeadBucketInput, ...request.WaiterOption) error {
	panic("unimplemented")
}

// WaitUntilBucketNotExists implements [s3iface.S3API].
func (m *MockS3Client) WaitUntilBucketNotExists(*s3.HeadBucketInput) error {
	panic("unimplemented")
}

// WaitUntilBucketNotExistsWithContext implements [s3iface.S3API].
func (m *MockS3Client) WaitUntilBucketNotExistsWithContext(aws.Context, *s3.HeadBucketInput, ...request.WaiterOption) error {
	panic("unimplemented")
}

// WaitUntilObjectExists implements [s3iface.S3API].
func (m *MockS3Client) WaitUntilObjectExists(*s3.HeadObjectInput) error {
	panic("unimplemented")
}

// WaitUntilObjectExistsWithContext implements [s3iface.S3API].
func (m *MockS3Client) WaitUntilObjectExistsWithContext(aws.Context, *s3.HeadObjectInput, ...request.WaiterOption) error {
	panic("unimplemented")
}

// WaitUntilObjectNotExists implements [s3iface.S3API].
func (m *MockS3Client) WaitUntilObjectNotExists(*s3.HeadObjectInput) error {
	panic("unimplemented")
}

// WaitUntilObjectNotExistsWithContext implements [s3iface.S3API].
func (m *MockS3Client) WaitUntilObjectNotExistsWithContext(aws.Context, *s3.HeadObjectInput, ...request.WaiterOption) error {
	panic("unimplemented")
}

// WriteGetObjectResponse implements [s3iface.S3API].
func (m *MockS3Client) WriteGetObjectResponse(*s3.WriteGetObjectResponseInput) (*s3.WriteGetObjectResponseOutput, error) {
	panic("unimplemented")
}

// WriteGetObjectResponseRequest implements [s3iface.S3API].
func (m *MockS3Client) WriteGetObjectResponseRequest(*s3.WriteGetObjectResponseInput) (*request.Request, *s3.WriteGetObjectResponseOutput) {
	panic("unimplemented")
}

// WriteGetObjectResponseWithContext implements [s3iface.S3API].
func (m *MockS3Client) WriteGetObjectResponseWithContext(aws.Context, *s3.WriteGetObjectResponseInput, ...request.Option) (*s3.WriteGetObjectResponseOutput, error) {
	panic("unimplemented")
}

func (m *MockS3Client) PutObject(input *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	args := m.Called(input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*s3.PutObjectOutput), args.Error(1)
}

func (m *MockDatabase) Where(query interface{}, args ...interface{}) db.Database {
	callArgs := buildArgs(query, args)
	ret := m.Called(callArgs...)
	if ret.Get(0) == nil {
		return nil
	}
	// ensure the stored value implements db.Database
	return ret.Get(0).(db.Database)
}

func (m *MockDatabase) First(dest interface{}, conditions ...interface{}) error {
	callArgs := buildArgs(dest, conditions)
	ret := m.Called(callArgs...)
	return ret.Error(0)
}

func (m *MockDatabase) Create(value interface{}) error {
	ret := m.Called(value)
	return ret.Error(0)
}

func (m *MockDatabase) Model(value interface{}) db.Database {
	ret := m.Called(value)
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(db.Database)
}

func (m *MockDatabase) Count(count *int64) db.Database {
	ret := m.Called(count)
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(db.Database)
}

func (m *MockDatabase) Error() error {
	ret := m.Called()
	return ret.Error(0)
}

func (m *MockDatabase) Save(value interface{}) error {
	ret := m.Called(value)
	return ret.Error(0)
}

func (m *MockDatabase) Preload(column string, conditions ...interface{}) db.Database {
	callArgs := buildArgsFromSlice(append([]interface{}{column}, conditions...))
	ret := m.Called(callArgs...)
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(db.Database)
}

func (m *MockDatabase) Delete(value interface{}, conditions ...interface{}) error {
	callArgs := buildArgs(value, conditions)
	ret := m.Called(callArgs...)
	return ret.Error(0)
}

func (m *MockDatabase) Find(dest interface{}, conds ...interface{}) error {
	callArgs := buildArgs(dest, conds)
	ret := m.Called(callArgs...)
	return ret.Error(0)
}

func (m *MockDatabase) AutoMigrate(dst ...interface{}) error {
	ret := m.Called(dst)
	return ret.Error(0)
}

func (m *MockDatabase) Offset(offset int) db.Database {
	ret := m.Called(offset)
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(db.Database)
}

func (m *MockDatabase) Limit(limit int) db.Database {
	ret := m.Called(limit)
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(db.Database)
}

func (m *MockDatabase) Scan(dest interface{}) error {
	ret := m.Called(dest)
	return ret.Error(0)
}

func (m *MockDatabase) Updates(values ...interface{}) error {
	ret := m.Called(values...)
	return ret.Error(0)
}

func (m *MockDatabase) Update(column string, value interface{}) error {
	ret := m.Called(column, value)
	return ret.Error(0)
}

func (m *MockDatabase) Association(column string) *gorm.Association {
	ret := m.Called(column)
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(*gorm.Association)
}

func (m *MockDatabase) Select(query interface{}) db.Database {
	ret := m.Called(query)
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(db.Database)
}

func (m *MockDatabase) Table(name string) db.Database {
	ret := m.Called(name)
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(db.Database)
}

func (m *MockDatabase) Joins(query string, args ...interface{}) db.Database {
	callArgs := buildArgs(query, args)
	ret := m.Called(callArgs...)
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(db.Database)
}

func (m *MockDatabase) Group(query string) db.Database {
	ret := m.Called(query)
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(db.Database)
}

func (m *MockDatabase) Having(query interface{}, args ...interface{}) db.Database {
	callArgs := buildArgs(query, args)
	ret := m.Called(callArgs...)
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(db.Database)
}

func (m *MockDatabase) Order(value interface{}) db.Database {
	ret := m.Called(value)
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(db.Database)
}

func (m *MockDatabase) Scopes(funcs ...func(db.Database) db.Database) db.Database {
	ret := m.Called(funcs)
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(db.Database)
}

func (m *MockDatabase) Take(dest interface{}, conditions ...interface{}) error {
	callArgs := buildArgs(dest, conditions)
	ret := m.Called(callArgs...)
	return ret.Error(0)
}

func (m *MockDatabase) Last(dest interface{}, conditions ...interface{}) error {
	callArgs := buildArgs(dest, conditions)
	ret := m.Called(callArgs...)
	return ret.Error(0)
}

func (m *MockDatabase) UpdateColumns(values interface{}) error {
	ret := m.Called(values)
	return ret.Error(0)
}

func (m *MockDatabase) CreateInBatches(value interface{}, batchSize int) error {
	ret := m.Called(value, batchSize)
	return ret.Error(0)
}

func (m *MockDatabase) FindInBatches(dest interface{}, batchSize int, fc func(db.Database, int) error) error {
	ret := m.Called(dest, batchSize, fc)
	return ret.Error(0)
}

func (m *MockDatabase) UpdateInBatches(value interface{}, batchSize int) error {
	ret := m.Called(value, batchSize)
	return ret.Error(0)
}

func (m *MockDatabase) Transaction(fc func(db.Database) error, opts ...*sql.TxOptions) error {
	callArgs := buildArgsFromSlice(append([]interface{}{fc}, interfaceSliceFromTxOptions(opts)...))
	ret := m.Called(callArgs...)
	return ret.Error(0)
}

// helper to convert []*sql.TxOptions to []interface{}
func interfaceSliceFromTxOptions(opts []*sql.TxOptions) []interface{} {
	res := make([]interface{}, 0, len(opts))
	for _, o := range opts {
		res = append(res, o)
	}
	return res
}

func (m *MockDatabase) Begin(opts ...*sql.TxOptions) db.Database {
	callArgs := buildArgsFromSlice(interfaceSliceFromTxOptions(opts))
	ret := m.Called(callArgs...)
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(db.Database)
}

func (m *MockDatabase) Commit() error {
	ret := m.Called()
	return ret.Error(0)
}

func (m *MockDatabase) Rollback() error {
	ret := m.Called()
	return ret.Error(0)
}

func (m *MockDatabase) Raw(sql string, values ...interface{}) db.Database {
	callArgs := buildArgs(sql, values)
	ret := m.Called(callArgs...)
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(db.Database)
}

func (m *MockDatabase) Exec(sql string, values ...interface{}) error {
	callArgs := buildArgs(sql, values)
	ret := m.Called(callArgs...)
	return ret.Error(0)
}

func (m *MockDatabase) Pluck(column string, dest interface{}) error {
	ret := m.Called(column, dest)
	return ret.Error(0)
}

func (m *MockDatabase) Migrator() gorm.Migrator {
	ret := m.Called()
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(gorm.Migrator)
}

func (m *MockDatabase) Distinct(args ...interface{}) db.Database {
	ret := m.Called(args)
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(db.Database)
}

func (m *MockDatabase) Debug() db.Database {
	ret := m.Called()
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(db.Database)
}

func (m *MockDatabase) Session(config *gorm.Session) db.Database {
	ret := m.Called(config)
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(db.Database)
}

func (m *MockDatabase) GetUnderlying() *gorm.DB {
	ret := m.Called()
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(*gorm.DB)
}

func (m *MockDatabase) Unscoped() db.Database {
	ret := m.Called()
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(db.Database)
}
