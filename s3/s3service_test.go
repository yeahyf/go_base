package s3

import (
	"fmt"
	"testing"
	//"github.com/aws/aws-sdk-go/service/s3"
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

func TestGetObjectsList(t *testing.T) {
	region := "eu-central-1"
	s3Service, err := GetS3Service(&region)
	if err != nil {
		fmt.Println("=====")
		t.Fail()
		return
	}

	bucket := "adjust-magic-seven"
	prefix := "11r8xw5s6yps_2018-07111"

	result, err1 := GetObjectsList(&bucket, &prefix, s3Service)
	if err1 != nil {
		fmt.Println("====1")
		t.Fail()
	} else {
		t.Log("ok")
		t.Log(len(result))
		for _, key := range result {
			t.Log(*key)
		}
	}

}
