package s3

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
)

func TestUpload(t *testing.T) {
	s3Service, err := GetS3Service("us-east-2")
	if err != nil {
		return
	}
	srcFile := "/Users/yeahyf/go/src/gobase/s3/s3service.go"
	destKey := "s3service.go"
	//公共读
	err = Upload(&srcFile, &destKey, s3Service, "campaign-resource-lib", s3.ObjectCannedACLPublicRead)
	if err != nil {
		t.Fail()
	}
}

func TestDelete(t *testing.T) {
	s3Service, err := GetS3Service("us-east-2")
	if err != nil {
		return
	}
	destKey := "s3service.go"
	err = Delete(destKey, s3Service, "campaign-resource-lib")
	if err != nil {
		t.Fail()
	}
}
