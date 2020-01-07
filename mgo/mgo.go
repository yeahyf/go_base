///mongodb管理接口封装
package mgo

import (
	"context"
	"gobase/log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoDBClient struct {
	*mongo.Client
	dbName *string
}

//创建一个mongodb客户端
func NewMongoClient(address string, timeout, maxsize, idletime int) (*MongoDBClient, error) {
	clientOptions := options.Client()
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, clientOptions.ApplyURI(address),
		clientOptions.SetConnectTimeout(time.Duration(timeout)*time.Second),
		clientOptions.SetMaxPoolSize(uint64(maxsize)),
		clientOptions.SetMaxConnIdleTime(time.Duration(idletime)*time.Second))

	if err != nil {
		return nil, err
	}
	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}
	if log.IsDebug() {
		log.Debug("Connected to MongoDB!")
	}

	//提供一个默认的db名称
	dbName := "yifants"

	mongoClient := &MongoDBClient{
		client,
		&dbName,
	}
	return mongoClient, nil
}

//提供一个修改数据库名称的接口
func (c *MongoDBClient) SetDatabaseName(dbName *string) {
	c.dbName = dbName
}

//======================

//插入一条记录
func (c *MongoDBClient) InsertOne(colName *string, document interface{}) (interface{}, error) {
	collection := c.Database(*c.dbName).Collection(*colName)
	insertResult, err := collection.InsertOne(context.TODO(), document)
	if err != nil {
		return nil, err
	}
	if log.IsDebug() {
		log.Debug("Inserted a single document: ", insertResult.InsertedID)
	}
	return insertResult.InsertedID, nil
}

//插入多条记录
func (c *MongoDBClient) InsertMany(colName *string, documents []interface{}) ([]interface{}, error) {
	collection := c.Database(*c.dbName).Collection(*colName)
	insertManyResult, err := collection.InsertMany(context.TODO(), documents)
	if err != nil {
		return nil, err
	}
	if log.IsDebug() {
		log.Debug("Inserted multiple documents:", insertManyResult.InsertedIDs)
	}
	return insertManyResult.InsertedIDs, err
}

//按照条件删除
func (c *MongoDBClient) Delete(colName *string, key bson.M) (int64, error) {
	col := c.Database(*c.dbName).Collection(*colName)
	deleteResult, err := col.DeleteMany(context.TODO(), key)
	if err != nil {
		return 0, err
	}
	if log.IsDebug() {
		log.Debugf("Deleted %v documents in the trainers collection\n", deleteResult.DeletedCount)
	}
	return deleteResult.DeletedCount, nil
}

//======================

//更新数据
func (c *MongoDBClient) Update(colName *string, filter bson.M, update bson.M) (int64, int64, error) {
	col := c.Database(*c.dbName).Collection(*colName)
	updateResult, err := col.UpdateMany(context.TODO(), filter, update)
	if err != nil {
		return 0, 0, nil
	}
	if log.IsDebug() {
		log.Debugf("Matched %v documents and updated %v documents.\n", updateResult.MatchedCount, updateResult.ModifiedCount)
	}
	return updateResult.MatchedCount, updateResult.ModifiedCount, nil
}

//根据条件只查询一条
func (c *MongoDBClient) SelectOne(colName *string, filter bson.M, v interface{}) (interface{}, error) {
	col := c.Database(*c.dbName).Collection(*colName)
	selectResult := col.FindOne(context.TODO(), filter)

	if err := selectResult.Decode(v); err != nil {
		return nil, err
	}
	return v, nil
}

//根据条件查询
func (c *MongoDBClient) Select(colName *string, key bson.M, max int64) ([]bson.M, error) {
	col := c.Database(*c.dbName).Collection(*colName)

	var findOptions *options.FindOptions
	var cursor *mongo.Cursor
	var err error
	if max > 0 {
		findOptions = options.Find()
		findOptions.SetLimit(max)
		cursor, err = col.Find(context.TODO(), key, findOptions)
	} else {
		cursor, err = col.Find(context.TODO(), key)
	}
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer cursor.Close(context.TODO())
	results := make([]bson.M, 0, 10)
	for cursor.Next(context.TODO()) {
		var result bson.M
		err := cursor.Decode(&result)
		if err != nil {
			log.Error(err)
		}
		results = append(results, result)
	}
	if err := cursor.Err(); err != nil {
		log.Error(err)
		return nil, err
	}
	return results, nil
}

//======================

func (c *MongoDBClient) CloseClient() {
	c.Disconnect(context.TODO())
}
