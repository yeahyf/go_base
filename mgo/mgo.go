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

var client *mongo.Client

//创建一个mongodb客户端
func NewMongoClient(address *string, timeout, maxsize, idletime int) error {

	var err error
	clientOptions := options.Client()
	client, err = mongo.Connect(context.TODO(), clientOptions.ApplyURI(*address),
		clientOptions.SetConnectTimeout(time.Duration(timeout)*time.Second),
		clientOptions.SetMaxPoolSize(uint16(maxsize)),
		clientOptions.SetMaxConnIdleTime(time.Duration(idletime)*time.Second))

	if err != nil {
		log.L.Error(err)
		return err
	}
	err = client.Ping(context.TODO(), readpref.Primary())
	if err != nil {
		log.L.Error(err)
		return err
	}
	log.L.Debug("Connected to MongoDB!")
	return nil
}

//======================

//插入一条记录
func InsertOne(dbName, colName *string, document interface{}) (interface{}, error) {
	collection := client.Database(*dbName).Collection(*colName)
	insertResult, err := collection.InsertOne(context.TODO(), document)
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
	collection := client.Database(*dbName).Collection(*colName)
	insertManyResult, err := collection.InsertMany(context.TODO(), documents)
	if err != nil {
		log.L.Error(err)
		return nil, err
	}
	log.L.Debug("Inserted multiple documents:", insertManyResult.InsertedIDs)
	return insertManyResult.InsertedIDs, err
}

//按照条件删除
func Delete(dbName, colName *string, key bson.M) (int64, error) {
	col := client.Database(*dbName).Collection(*colName)
	deleteResult, err := col.DeleteMany(context.TODO(), key)
	if err != nil {
		log.L.Error(err)
		return 0, err
	}
	log.L.Debugf("Deleted %v documents in the trainers collection\n", deleteResult.DeletedCount)
	return deleteResult.DeletedCount, nil
}

//======================

//更新数据
func Update(dbName, colName *string, filter bson.M, update bson.M) (int64, int64, error) {
	col := client.Database(*dbName).Collection(*colName)
	updateResult, err := col.UpdateMany(context.TODO(), filter, update)
	if err != nil {
		log.L.Error(err)
		return 0, 0, nil
	}
	log.L.Debugf("Matched %v documents and updated %v documents.\n", updateResult.MatchedCount, updateResult.ModifiedCount)
	return updateResult.MatchedCount, updateResult.ModifiedCount, nil
}

//======================

//根据条件查询
func Select(dbName, colName *string, key bson.M, max int64) ([]bson.M, error) {
	col := client.Database(*dbName).Collection(*colName)

	var findOptions *options.FindOptions
	var cursor *mongo.Cursor
	var err error
	if max > 0 {
		findOptions = options.Find()
		findOptions.SetLimit(max)
		cursor,err = col.Find(context.TODO(), key, findOptions)
	}else{
		cursor,err = col.Find(context.TODO(), key)
	}
	if err != nil {
		log.L.Error(err)
		return nil, err
	}
	defer cursor.Close(context.TODO())
	results := make([]bson.M, 0, 10)
	for cursor.Next(context.TODO()) {
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
	if client != nil {
		client.Disconnect(context.TODO())
	}
}
