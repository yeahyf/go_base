package ddb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/yeahyf/go_base/log"
)

// DynamoDBInfo 封装DynamoDB配置信息
type DynamoDBInfo struct {
	Profile   string
	Region    string
	Mode      int
	LocalAddr string
}

// NewDynamoDBClient 构建DynamoDB的操作对象
func NewDynamoDBClient(info *DynamoDBInfo) *dynamodb.Client {
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
	return client
}
