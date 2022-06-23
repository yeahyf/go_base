package awss3

import (
	"bytes"
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/smallnest/rpcx/log"
)

//BigStorageItem 数据对象封装
type BigStorageItem struct {
	Key *string
	Buf bytes.Buffer
}

//S3Client S3操作代理对象
type S3Client struct {
	Client     *s3.Client
	BucketName string
}

var Client *S3Client

//WriteAt 大数据对象写入方法
func (v *BigStorageItem) WriteAt(p []byte, _ int64) (int, error) {
	size, err := v.Buf.Write(p)
	if err != nil {
		return 0, err
	}
	return size, nil
}

//NewS3Client 构建新的S3操作代理对象
func NewS3Client(profile string, region string, bucketName string) {
	if Client != nil {
		return
	}
	s3Cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithSharedConfigProfile(profile),
		config.WithRegion(region))
	if err != nil {
		log.Errorf("couldn't create aws s3 cfg! %v", err)
		panic(err)
	}
	var client *s3.Client
	client = s3.NewFromConfig(s3Cfg)

	Client = &S3Client{
		Client:     client,
		BucketName: bucketName,
	}
}

//UpdateObject 批量上传对象到S3
func (client S3Client) UpdateObject(items []*BigStorageItem) error {
	uploader := manager.NewUploader(client.Client)
	for _, v := range items {
		_, err := uploader.Upload(context.Background(), &s3.PutObjectInput{
			Bucket: aws.String(client.BucketName),
			Key:    v.Key,
			Body:   &v.Buf,
			ACL:    "private",
		})
		if err != nil {
			log.Errorf("couldn't upload data to s3 %v", err)
			return err
		}
	}
	return nil
}

//DownloadObject 下载单个S3对象
func (client S3Client) DownloadObject(item *BigStorageItem) error {
	downloader := manager.NewDownloader(client.Client)
	_, err := downloader.Download(context.Background(), item, &s3.GetObjectInput{
		Bucket: aws.String(client.BucketName),
		Key:    item.Key,
	})

	if err != nil {
		log.Errorf("couldn't download data %v,key=%s", err, *item.Key)
	}
	return err
}

//DownloadObjects 批量下载对象
func (client S3Client) DownloadObjects(items []*BigStorageItem) error {
	for _, item := range items {
		err := client.DownloadObject(item)
		if err != nil {
			return err
		}
	}
	return nil
}

//DeleteObjects 删除s3对象
func (client S3Client) DeleteObjects(items []*BigStorageItem) error {
	keys := make([]types.ObjectIdentifier, 0, len(items))
	for _, v := range items {
		keys = append(keys, types.ObjectIdentifier{Key: v.Key})
	}
	deleteKeys := &types.Delete{
		Objects: keys,
	}
	_, err := client.Client.DeleteObjects(context.Background(),
		&s3.DeleteObjectsInput{
			Bucket: aws.String(client.BucketName),
			Delete: deleteKeys,
		})
	return err
}
