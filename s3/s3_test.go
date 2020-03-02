package s3

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
)

func TestUploadData(t *testing.T) {
	sess, err := NewSession(aws.String(endpoints.UsWest2RegionID),
		aws.String("us_w2_s3"))
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	bucket := "storage-large-data"

	items := []*BigStorageItem{
		&BigStorageItem{
			UID:      "123",
			BundleID: "1231",
			Key:      "_ActionFeature1",
			Value:    []byte("123123alskdjfalskjflaskd;fjaslk;dfjasl;dkfj1"),
		},
	}

	err = UploadData(sess, &bucket, items)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
}

func TestDownloadData(t *testing.T) {
	sess, err := NewSession(aws.String(endpoints.UsWest2RegionID),
		aws.String("us_w2_s3"))
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	bucket := "storage-large-data"

	items := []*BigStorageItem{
		&BigStorageItem{
			UID:      "123",
			BundleID: "1231",
			Key:      "_ActionFeature1",
			Value:    []byte("123123alskdjfalskjflaskd;fjaslk;dfjasl;dkfj1"),
		},
	}

	err = UploadData(sess, &bucket, items)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	err = DownloadData(sess, &bucket, items)

	if err != nil {
		t.Error(err)
		t.Fail()
	} else {
		for i := 0; i < len(items); i++ {
			t.Log(items[i].UID + items[i].Key + items[i].BundleID)
			t.Log(string(items[i].Value))
		}
	}
}

func TestUploadDownload(t *testing.T) {
	sess, err := NewSession(aws.String(endpoints.UsWest2RegionID),
		aws.String("us_w2_s3"))
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	bucket := "storage-large-data"

	items := []*BigStorageItem{
		&BigStorageItem{
			UID:      "123",
			BundleID: "1231",
			Key:      "_ActionFeature1",
		},
	}

	err = DownloadData(sess, &bucket, items)

	if err != nil {
		t.Error(err)
		t.Fail()
	} else {
		for i := 0; i < len(items); i++ {
			t.Log(items[i].UID + items[i].Key + items[i].BundleID)
			t.Log(string(items[i].Value))
		}
	}
}
