package awss3

import (
	"testing"
	"time"
)

var bak = map[string]string{
	"z7qzbwn63zeo": "2019-01-01",
	"cb9utai4sq9s": "2021-05-01",
	"nzygzy9nszy8": "2021-11-01",
	"zasesb89qj28": "2020-01-01",
	"bs7y4uppkyyo": "2022-01-01",
	"ocloxiquo5xc": "2022-01-01",
	"w6sonmxovnr4": "2022-01-01",
	"3uu9v4dktngg": "2022-01-01",
	"b5k8fig1dqf4": "2022-01-01",
}

var Time = "2006-01-02"

var bakTime map[string]time.Time
var bucketName = "adjust-magic-seven"

func Test_List(t *testing.T) {
	bakTime = make(map[string]time.Time, 9)
	for k, v := range bak {
		t, _ := time.Parse(Time, v)
		bakTime[k] = t
	}
	s3Client := NewS3Client("default", "eu-central-1", bucketName)
	marker := ""
	for i := 0; i < 1000; i++ {
		list, err := s3Client.ListObjects(&marker)
		if err != nil {
			t.Fail()
		}
		temp := make([]*BigStorageItem, 0, len(list))
		for _, v := range list {
			key := *v.Key
			prefix := key[0:12]
			if giveT, ok := bakTime[prefix]; ok {
				ts := key[13:23]
				t, _ := time.Parse(Time, ts)
				if giveT.After(t) { //删除
					temp = append(temp, v)
				}
			} else {
				temp = append(temp, v)
			}
		}
		if len(temp) == 0 {
			continue
		}

		//_ = s3Client.DeleteObjects(temp)

		for _, v := range temp {
			t.Log(*v.Key)
			err = s3Client.DeleteObject(v)
			ver := "null"
			v.Version = &ver
			err = s3Client.DeleteObject(v)
			marker = *v.Key
		}
		if err != nil {
			t.Log(err)
		}
	}
}

func Test_List2(t *testing.T) {
	s3Client := NewS3Client("default", "eu-central-1", bucketName)
	marker := ""
	result, err := s3Client.ListObjectVersion(&marker)
	if err != nil {
		t.Fail()
	}
	t.Log(result)
}

func Test_Delete(t *testing.T) {
	s3Client := NewS3Client("default", "eu-central-1", bucketName)
	key := "13vgxnxiii00_2019-08-12T170000_d6ecdd3fbcfaff677deb68ecda4e7379.csv.gz"
	ver := "xRvJ1jRynRsNBBVuV3eicU2FB6CRmlSf"
	item := &BigStorageItem{
		Key:     &key,
		Version: &ver,
	}

	err := s3Client.DeleteObject(item)
	if err != nil {
		t.Fail()
	}

}
