package s3

import (
	"gobase/log"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func Delete(destKey string, s3Service *s3.S3, bucket string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(destKey),
	}

	_, err := s3Service.DeleteObject(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				log.L.Error(s3.ErrCodeNoSuchKey, aerr.Error())
			}
		} else {
			log.L.Error(err.Error())
		}
		return err
	}
	//log.L.Info(result)
	return nil
}

func Upload(srcFile *string, destKey *string, s3Service *s3.S3, bucket string) error {
	input := &s3.PutObjectInput{
		Body:   aws.ReadSeekCloser(strings.NewReader(*srcFile)),
		Bucket: aws.String(bucket),
		Key:    aws.String(*destKey),
	}

	_, err := s3Service.PutObject(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				log.L.Error(s3.ErrCodeNoSuchKey, aerr.Error())
			}
		} else {
			log.L.Error(err.Error())
		}
		return err
	}
	//log.L.Info(result)
	return nil
}

func Download(srcFile *string, destPath *string, s3Service *s3.S3, bucket string) error {
	getObjectInput := &s3.GetObjectInput{
		Bucket: aws.String(bucket), //目标存储桶
		Key:    aws.String(*srcFile),
	}

	getObjectOutput, err := s3Service.GetObject(getObjectInput)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchKey:
				log.L.Error(s3.ErrCodeNoSuchKey, aerr.Error())
			default:
				log.L.Error(aerr.Error())
			}
		} else {
			log.L.Error(err.Error())
		}
		return err
	}
	fileContent, err := ioutil.ReadAll(getObjectOutput.Body)
	if err == nil {
		ioutil.WriteFile(path.Join(*destPath, *srcFile), fileContent, os.ModePerm)
		log.L.Info("File: " + *srcFile + " download success!")
	} else {
		log.L.Error(err.Error())
	}
	return err
}

func GetS3Service(region string) (*s3.S3, error) {
	var s3Service *s3.S3
	//创建会话，默认采用配置方式，区域直接硬编码
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewSharedCredentials("", "default"),
	})
	if err != nil {
		log.L.Errorf("Get AWS Session Error!")
		return s3Service, err
	}

	_, err = sess.Config.Credentials.Get()

	if err != nil {
		log.L.Errorf("AWS Config Credentials Error!")
		return s3Service, err
	}
	return s3.New(sess), nil
}
