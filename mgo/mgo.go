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

var MongoClient *mongo.Client
var timeOut time.Duration
var ctx context.Context

//创建一个mongodb客户端
func NewMongoClient(address *string, timeout *time.Duration) {
	timeOut := *timeout
	ctx, _ = context.WithTimeout(context.Background(), timeOut*time.Second)
	var err error
	MongoClient, err = mongo.Connect(ctx, options.Client().ApplyURI(*address))
	if err != nil {
		log.L.Error(err)
		panic(err)
	}
	err = MongoClient.Ping(ctx, readpref.Primary())
	if err != nil {
		log.L.Error(err)
		panic(err)
	}
	log.L.Debug("Connected to MongoDB!")
}

//======================

//插入一条记录
func InsertOne(dbName, colName *string, document interface{}) (interface{}, error) {
	col := MongoClient.Database(*dbName).Collection(*colName)
	insertResult, err := col.InsertOne(ctx, document)
	if err != nil {
		log.L.Error(err)
		return nil, err
	} else {
		log.L.Debug("Inserted a single document: ", insertResult.InsertedID)
	}
	return insertResult.InsertedID, nil
}

//插入多条记录
func InsertMany(dbName, colName *string, documents []interface{}) ([]interface{}, error) {
	col := MongoClient.Database(*dbName).Collection(*colName)
	insertManyResult, err := col.InsertMany(ctx, documents)
	if err != nil {
		log.L.Error(err)
		return nil, err
	}
	log.L.Debug("Inserted multiple documents:", insertManyResult.InsertedIDs)
	return insertManyResult.InsertedIDs, err
}

//======================

//按照条件删除
func Delete(dbName, colName *string, key bson.M) (int64, error) {
	col := MongoClient.Database(*dbName).Collection(*colName)
	deleteResult, err := col.DeleteMany(ctx, key)
	if err != nil {
		log.L.Error(err)
		return 0, err
	}
	log.L.Debugf("Deleted %v documents in the trainers collection\n", deleteResult.DeletedCount)
	return deleteResult.DeletedCount, nil
}

//======================

//更新数据
func Update(dbName, colName *string, filter *bson.M, update *bson.D) (int64, int64, error) {
	col := MongoClient.Database(*dbName).Collection(*colName)
	updateResult, err := col.UpdateMany(ctx, filter, update)
	if err != nil {
		log.L.Error(err)
		return 0, 0, nil
	}
	log.L.Debugf("Matched %v documents and updated %v documents.\n", updateResult.MatchedCount, updateResult.ModifiedCount)
	return updateResult.MatchedCount, updateResult.ModifiedCount, nil
}

//======================

//根据条件查询
func Select(dbName, colName *string, key bson.M) ([]bson.M, error) {
	col := MongoClient.Database(*dbName).Collection(*colName)
	cursor, err := col.Find(ctx, key)
	if err != nil {
		log.L.Error(err)
		return nil, err
	}
	defer cursor.Close(ctx)

	results := make([]bson.M, 0, 10)

	for cursor.Next(ctx) {
		var result bson.M
		err := cursor.Decode(&result)
		if err != nil {
			log.L.Error(err)
		}
		results = append(results, result)
	}
	if err := cursor.Err(); err != nil {
		log.L.Error(err)
		return nil, err
	}
	return results, nil
}

//======================

func CloseClient() {
	if MongoClient != nil {
		MongoClient.Disconnect(context.TODO())
	}
}
