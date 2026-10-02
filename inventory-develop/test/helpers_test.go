package test

import (
	"inventory/mocks"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/stretchr/testify/mock"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func fakeS3Request() *request.Request {
    svc := s3.New(session.Must(session.NewSession(&aws.Config{
        Region:      aws.String("us-east-1"),
        Credentials: credentials.NewStaticCredentials("test", "test", ""),
        Endpoint:    aws.String("http://127.0.0.1:0"),
        DisableSSL:  aws.Bool(true),
    })))

    req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
        Bucket: aws.String("test-bucket"),
        Key:    aws.String("test-key"),
    })

    return req
}

func mockPresignRequest(mockS3 *mocks.MockS3Client) {
	mockS3.
		// On("GetObjectRequest", mock.AnythingOfType("*s3.GetObjectInput")).
		On("GetObjectRequest", mock.Anything).
		Return(fakeS3Request(), &s3.GetObjectOutput{}).
		Maybe()
}