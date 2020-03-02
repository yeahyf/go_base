package s3

import (
	"bytes"

	"github.com/yeahyf/go_base/strutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	awss3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/smallnest/rpcx/log"
)

type BigStorageItem struct {
	UID      string
	BundleID string
	Key      string
	Value    []byte
}

func (v *BigStorageItem) WriteAt(p []byte, off int64) (n int, err error) {
	v.Value = p[:]
	return len(p), nil
}

//统一构建存档的Key
func (v *BigStorageItem) GetStorageKey() *string {
	var buf bytes.Buffer
	buf.WriteString(v.UID)
	buf.WriteByte('/')
	buf.WriteString(v.BundleID)
	buf.WriteByte('_')
	buf.WriteString(v.Key)
	return strutil.Bytes2Str(buf.Bytes())
}

//根据profile获取s3操作对象
func NewSession(region, profile *string) (*session.Session, error) {
	//创建会话，默认采用配置方式，区域直接硬编码
	sess, err := session.NewSession(&aws.Config{
		Region:      region,
		Credentials: credentials.NewSharedCredentials("", *profile),
	})
	if err != nil {
		log.Errorf("Get AWS Session Error!")
		return nil, err
	}
	return sess, nil
}

func UploadData(sess *session.Session, bucket *string, items []*BigStorageItem) error {
	svc := s3manager.NewUploader(sess)

	objects := make([]s3manager.BatchUploadObject, 0, len(items))
	//批量构建对象
	for _, v := range items {
		buf := bytes.Buffer{}
		buf.Grow(len(v.Value))
		buf.Write(v.Value)
		batchObject := s3manager.BatchUploadObject{
			Object: &s3manager.UploadInput{
				Key:    v.GetStorageKey(),
				Bucket: aws.String(*bucket),
				Body:   &buf,
				ACL:    aws.String("private"),
			},
		}
		objects = append(objects, batchObject)
	}
	//保存
	iter := &s3manager.UploadObjectsIterator{Objects: objects}
	return svc.UploadWithIterator(aws.BackgroundContext(), iter)
}

func DownloadData(sess *session.Session, bucket *string, items []*BigStorageItem) error {
	svc := s3manager.NewDownloader(sess)

	objects := make([]s3manager.BatchDownloadObject, 0, len(items))
	//批量构建对象
	for _, v := range items {
		batchObject := s3manager.BatchDownloadObject{
			Object: &awss3.GetObjectInput{
				Key:    v.GetStorageKey(),
				Bucket: aws.String(*bucket),
			},
			Writer: v,
		}
		objects = append(objects, batchObject)
	}

	iter := &s3manager.DownloadObjectsIterator{
		Objects: objects,
	}
	return svc.DownloadWithIterator(aws.BackgroundContext(), iter)
}
