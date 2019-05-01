package mgo

import (
	"fmt"
	"gobase/log"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type Trainer struct {
	Name string
	Age  int
	City string
}

func TestNew(t *testing.T) {
	logFile := "/Users/yeahyf/go/src/pubaws/conf/seelog.xml"
	log.SetLogConf(&logFile)
	address := "mongodb://yifan:123456@192.168.1.10:27017/yifants"
	timeout := time.Second * 3
	NewMongoClient(&address, &timeout)
	CloseClient()
}

func TestInsert(t *testing.T) {
	logFile := "/Users/yeahyf/go/src/pubaws/conf/seelog.xml"
	log.SetLogConf(&logFile)
	address := "mongodb://yifan:123456@192.168.1.10:27017/yifants"
	timeout := time.Second * 3
	NewMongoClient(&address, &timeout)
	dbName := "yifants"
	col := "trainers"
	document := Trainer{"杨语迟", 10, "杨林朱院村"}
	result, err := InsertOne(&dbName, &col, document)
	if err != nil {
		t.Fail()
	} else {
		fmt.Println(result)
	}
	CloseClient()
}

func TestInsertMany(t *testing.T) {
	logFile := "/Users/yeahyf/go/src/pubaws/conf/seelog.xml"
	log.SetLogConf(&logFile)
	address := "mongodb://yifan:123456@192.168.1.10:27017/yifants"
	timeout := time.Second * 3
	NewMongoClient(&address, &timeout)
	dbName := "yifants"
	col := "trainers"
	document1 := Trainer{"charlse", 10, "sz"}
	document2 := Trainer{"熊天豪", 10, "随州"}
	document3 := Trainer{"熊天豪2", 10, "随州1"}
	document4 := Trainer{"熊天豪3", 10, "随州3"}
	documents := []interface{}{document1, document2, document3, document4}

	result, err := InsertMany(&dbName, &col, documents)
	if err != nil {
		t.Fail()
	} else {
		fmt.Println(result)
	}
	CloseClient()
}

func TestDelete(t *testing.T) {
	logFile := "/Users/yeahyf/go/src/pubaws/conf/seelog.xml"
	log.SetLogConf(&logFile)
	address := "mongodb://yifan:123456@192.168.1.10:27017/yifants"
	timeout := time.Second * 3
	NewMongoClient(&address, &timeout)

	dbName := "yifants"
	col := "trainers"

	filer := bson.M{
		"name": "熊天豪",
		"age":  10,
	}
	result, err := Delete(&dbName, &col, filer)
	if err != nil {
		t.Fail()
	} else {
		fmt.Println(result)
	}
	CloseClient()
}

func TestSelect(t *testing.T) {
	logFile := "/Users/yeahyf/go/src/pubaws/conf/seelog.xml"
	log.SetLogConf(&logFile)
	address := "mongodb://yifan:123456@192.168.1.10:27017/yifants"
	timeout := time.Second * 3
	NewMongoClient(&address, &timeout)

	dbName := "yifants"
	col := "trainers"

	filer := bson.M{
		"name": "杨语迟",
		"age":  10,
	}
	result, err := Select(&dbName, &col, filer)
	if err != nil {
		t.Fail()
	} else {
		for _, v := range result {
			// k,v1 := range map[string]interface{}(v){
			fmt.Println(v)
			// }
		}
	}
	CloseClient()
}

func TestUpdate(t *testing.T) {
	logFile := "/Users/yeahyf/go/src/pubaws/conf/seelog.xml"
	log.SetLogConf(&logFile)
	address := "mongodb://yifan:123456@192.168.1.10:27017/yifants"
	timeout := time.Second * 3
	NewMongoClient(&address, &timeout)

	dbName := "yifants"
	col := "trainers"

	filer := bson.M{
		"name": "杨语迟",
		"age":  10,
	}

	update := bson.D{
		{"$set",
			bson.D{
				{"age", 12},
			}},
	}
	match, up, err := Update(&dbName, &col, &filer, &update)
	if err != nil {
		t.Fail()
	} else {
		fmt.Println(match, up)
	}
	CloseClient()
}
