package awss3

import (
	"bytes"
	"testing"

	"github.com/yeahyf/go_base/strutil"
)

func TestUpdateObject(t *testing.T) {
	NewS3Client("us_w2_s3", "us-west-2", "storage-large-data")

	bigItems := make([]*BigStorageItem, 0, 1)

	var buf bytes.Buffer
	buf.WriteString("test")

	bigItem := &BigStorageItem{
		Key: strutil.Bytes2Str([]byte("fineboost/test/")),
		Buf: buf,
	}

	bigItems = append(bigItems, bigItem)

	err := Client.UpdateObject(bigItems)
	if err != nil {
		t.Failed()
	}
}
