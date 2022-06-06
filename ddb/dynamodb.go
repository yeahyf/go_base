package ddb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/yeahyf/go_base/log"
)

var storageClient *StorageClient

//StorageClient 封装DynamoDB操作对象
type StorageClient struct {
	DynamoDbClient *dynamodb.Client
	TableName      string
}

//DynamoDBInfo 封装DynamoDB配置信息
type DynamoDBInfo struct {
	Profile   string
	Region    string
	Mode      int
	LocalAddr string
	TableName string
}

//NewDynamoDBClient 构建DynamoDB的操作对象
func NewDynamoDBClient(info DynamoDBInfo) *StorageClient {
	if storageClient != nil {
		return storageClient
	}
	dynamodbCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithSharedConfigProfile(info.Profile),
		config.WithRegion(info.Region))
	if err != nil {
		log.Errorf("couldn't create aws dynamodb cfg! %v", err)
		panic(err)
	}
	var client *dynamodb.Client
	if info.Mode == 0 {
		client = dynamodb.NewFromConfig(dynamodbCfg, func(options *dynamodb.Options) {
			options.EndpointResolver = dynamodb.EndpointResolverFromURL(info.LocalAddr)
		})
	} else {
		client = dynamodb.NewFromConfig(dynamodbCfg)
	}
	storageClient = &StorageClient{
		DynamoDbClient: client,
		TableName:      info.TableName,
	}
	return storageClient
}
