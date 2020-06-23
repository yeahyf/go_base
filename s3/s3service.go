///aws s3接口封装
package s3

import (
	"bytes"
	"net/http"
	"os"
	"path"

	"github.com/yeahyf/go_base/log"

	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func Delete(destKey *string, s3Service *s3.S3, bucket *string) error {
	input := &s3.DeleteObjectInput{
		Bucket: bucket,
		Key:    destKey,
	}

	_, err := s3Service.DeleteObject(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				log.Error(s3.ErrCodeNoSuchKey, aerr.Error())
			}
		} else {
			log.Error(err.Error())
		}
		return err
	}
	//log.Info(result)
	return nil
}

func UploadBuffer(data []byte, destKey *string, s3Service *s3.S3, bucket *string, acl *string) error {
	contentType := "application/octet-stream"

	input := &s3.PutObjectInput{
		Bucket:        bucket,
		Key:           destKey,
		ACL:           acl,
		Body:          bytes.NewReader(data),
		ContentLength: aws.Int64(int64(len(data))),
		ContentType:   &contentType,
		StorageClass:  aws.String(s3.ObjectStorageClassIntelligentTiering),
	}

	_, err := s3Service.PutObject(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				log.Error(s3.ErrCodeNoSuchKey, aerr.Error())
			}
		} else {
			log.Error(err.Error())
		}
		return err
	}
	//log.Info(result)
	return nil
}

func Upload(srcFile *string, destKey *string, s3Service *s3.S3, bucket *string, acl *string) error {
	file, err := os.Open(*srcFile)
	if err != nil {
		log.Error(err)
		return err
	}
	defer file.Close()

	fileInfo, _ := file.Stat()
	var size = fileInfo.Size()
	buffer := make([]byte, size)
	file.Read(buffer)
	contentType := http.DetectContentType(buffer)

	input := &s3.PutObjectInput{
		Bucket:        bucket,
		Key:           destKey,
		ACL:           acl,
		Body:          bytes.NewReader(buffer),
		ContentLength: aws.Int64(size),
		ContentType:   &contentType,
		StorageClass:  aws.String(s3.ObjectStorageClassIntelligentTiering),
	}

	_, err = s3Service.PutObject(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				log.Error(s3.ErrCodeNoSuchKey, aerr.Error())
			}
		} else {
			log.Error(err.Error())
		}
		return err
	}
	//log.Info(result)
	return nil
}

//使用新的接口处理下载
func Download(srcFile *string, destPath *string, region, bucket string) error {
	sess, err := GetNewSession("", region)
	if err != nil {
		log.Error("Create S3 Session Error!!!")
		return err
	}
	downloader := s3manager.NewDownloader(sess)
	dataFile := path.Join(*destPath, *srcFile)
	f, err := os.Create(dataFile)
	if err != nil {
		log.Error("os Create error: ", err)
		return err
	}
	defer f.Close()

	numBytes, err := downloader.Download(f,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    srcFile,
		})

	if err != nil {
		log.Errorf("Unable to download item %s, %v", dataFile, err)
		return err
	}
	log.Info("File download finished! ", numBytes, *srcFile)
	return nil
}

func GetS3ServiceAccessID(region, accessid, accesskey *string) (*s3.S3, error) {
	var s3Service *s3.S3
	//创建会话，默认采用配置方式，区域直接硬编码
	sess, err := session.NewSession(&aws.Config{
		Region:      region,
		Credentials: credentials.NewStaticCredentials(*accessid, *accesskey, ""),
	})
	if err != nil {
		log.Errorf("Get AWS Session Error!")
		return s3Service, err
	}

	_, err = sess.Config.Credentials.Get()

	if err != nil {
		log.Errorf("AWS Config Credentials Error!")
		return s3Service, err
	}
	return s3.New(sess), nil
}

//根据profile获取s3操作对象
func GetS3ServiceProfile(region, profile *string) (*s3.S3, error) {
	var s3Service *s3.S3
	//创建会话，默认采用配置方式，区域直接硬编码
	sess, err := session.NewSession(&aws.Config{
		Region:      region,
		Credentials: credentials.NewSharedCredentials("", *profile),
	})
	if err != nil {
		log.Errorf("Get AWS Session Error!")
		return s3Service, err
	}

	_, err = sess.Config.Credentials.Get()

	if err != nil {
		log.Errorf("AWS Config Credentials Error!")
		return s3Service, err
	}
	return s3.New(sess), nil
}

//使用默认的profile进行处理，profile为.aws目录下面配置的内容
func GetS3Service(region *string) (*s3.S3, error) {
	profile := "default"
	return GetS3ServiceProfile(region, &profile)
}

//根据业务要求获取符合前缀要求的存放在aws s3中文件的列表
func GetObjectsList(bucket *string, prefix *string, s3Service *s3.S3) ([]*string, error) {
	input := &s3.ListObjectsInput{
		Bucket: bucket,
		Prefix: prefix,
	}

	output, err := s3Service.ListObjects(input)
	if err != nil {
		log.Errorf("Get List Object Error! prefix = %s", *prefix)
		return nil, err
	} else {
		//只需要关心Key即可
		size := len(output.Contents)
		//定义一个字符串指针切片
		result := make([]*string, 0, size)
		for _, object := range output.Contents {
			result = append(result, object.Key)
		}
		return result, nil
	}
}

func GetNewSession(profile, region string) (*session.Session, error) {
	if profile == "" {
		profile = "default"
	}
	sess, err := session.NewSession(&aws.Config{
		Credentials:      credentials.NewSharedCredentials("", profile),
		Region:           aws.String(region),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(false), //virtual-host style方式，不要修改
	})
	return sess, err
}
