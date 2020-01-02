package mgo

import (
	"fmt"
	log "gobase/zap"
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
	err := NewMongoClient(&address, 10, 2, 10)
	if err != nil {
		t.Fatal(err)
		t.Fail()
	}
	defer CloseClient()
}

func TestInsert(t *testing.T) {
	//logFile := "/Users/yeahyf/go/src/pubaws/conf/seelog.xml"
	log.SetLogConf(&logFile)
	address := "mongodb://yifan:123456@192.168.1.10:27017/yifants"
	err := NewMongoClient(&address, 10, 2, 10)
	if err != nil {
		t.Fatal(err)
		t.Fail()
	}
	defer CloseClient()

	dbName := "yifants"
	col := "trainers"
	document := Trainer{"杨", 10, "朱院村"}
	result, err := InsertOne(&dbName, &col, document)
	if err != nil {
		t.Error(err)
		t.Fail()
	} else {
		fmt.Println(result)
	}
}

func TestInsertMany(t *testing.T) {
	logFile := "/Users/yeahyf/go/src/pubaws/conf/seelog.xml"
	log.SetLogConf(&logFile)
	address := "mongodb://yifan:123456@192.168.1.10:27017/yifants"
	NewMongoClient(&address, 10, 2, 10)
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
	NewMongoClient(&address, 10, 2, 10)

	dbName := "yifants"
	col := "updapp"

	mytime := time.Now().Add(-time.Duration(3) * time.Hour * 24).Format("2006-01-02 15:04:05")
	filer := bson.M{"pubid": "a7fpmwda", "platform": "2", "status": "1", "utime": bson.M{"$lt": mytime}}

	result, err := Delete(&dbName, &col, filer)
	if err != nil {
		t.Error(err)
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
	NewMongoClient(&address, 10, 2, 10)

	dbName := "yifants"
	col := "trainers"

	filer := bson.M{
		"name": "杨",
		"age":  10,
	}
	result, err := Select(&dbName, &col, filer, 0)
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
	NewMongoClient(&address, 10, 2, 10)

	dbName := "yifants"
	col := "updapp"

	filer := bson.M{
		"pubid":    "bjluhixe",
		"platform": "1",
		"status":   "1",
	}

	update := bson.M{"$set": bson.M{"status": "0"}}
	match, up, err := Update(&dbName, &col, filer, update)
	if err != nil {
		t.Fail()
	} else {
		fmt.Println(match, up)
	}
	CloseClient()
}
