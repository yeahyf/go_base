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
	logFile := "/Users/yeahyf/go/src/pubaws/conf/zap.json"
	log.SetLogConf(&logFile)
	address := "mongodb://yifan:123456@192.168.1.10:27017/yifants"
	c, err := NewMongoClient(address, 10, 2, 10)
	if err != nil {
		t.Fatal(err)
		t.Fail()
	}
	defer c.CloseClient()
}

func TestInsert(t *testing.T) {
	logFile := "/Users/yeahyf/go/src/pubaws/conf/zap.json"
	log.SetLogConf(&logFile)
	address := "mongodb://yifan:123456@192.168.1.10:27017/yifants"
	c, err := NewMongoClient(address, 10, 2, 10)
	if err != nil {
		t.Fatal(err)
		t.Fail()
	}
	defer c.CloseClient()

	//dbName := "yifants"
	col := "trainers"
	document := Trainer{"杨", 10, "朱院村"}
	result, err := c.InsertOne(&col, document)
	if err != nil {
		t.Error(err)
		t.Fail()
	} else {
		fmt.Println(result)
	}
}

func TestInsertMany(t *testing.T) {
	logFile := "/Users/yeahyf/go/src/pubaws/conf/zap.json"
	log.SetLogConf(&logFile)
	address := "mongodb://yifan:123456@192.168.1.10:27017/yifants"
	c, err := NewMongoClient(address, 10, 2, 10)
	defer c.CloseClient()
	//dbName := "yifants"
	col := "trainers"
	document1 := Trainer{"charlse", 10, "sz"}
	document2 := Trainer{"熊天豪", 10, "随州"}
	document3 := Trainer{"熊天豪2", 10, "随州1"}
	document4 := Trainer{"熊天豪3", 10, "随州3"}
	documents := []interface{}{document1, document2, document3, document4}

	result, err := c.InsertMany(&col, documents)
	if err != nil {
		t.Fail()
	} else {
		fmt.Println(result)
	}

}

func TestDelete(t *testing.T) {
	logFile := "/Users/yeahyf/go/src/pubaws/conf/zap.json"
	log.SetLogConf(&logFile)
	address := "mongodb://yifan:123456@192.168.1.10:27017/yifants"
	c, err := NewMongoClient(address, 10, 2, 10)
	defer c.CloseClient()

	//dbName := "yifants"
	col := "updapp"

	mytime := time.Now().Add(-time.Duration(3) * time.Hour * 24).Format("2006-01-02 15:04:05")
	filer := bson.M{"pubid": "a7fpmwda", "platform": "2", "status": "1", "utime": bson.M{"$lt": mytime}}

	result, err := c.Delete(&col, filer)
	if err != nil {
		t.Error(err)
		t.Fail()
	} else {
		fmt.Println(result)
	}
}

type Geo struct {
	Code string `bson:"code"`
	Name string `bson:"name"`
}

// func TestSelectOne(t *testing.T) {
// 	logFile := "/Users/yeahyf/go/src/pubaws/conf/zap.json"
// 	log.SetLogConf(&logFile)
// 	address := "mongodb://yifan:123456@192.168.1.10:27017/yifants"
// 	c, err := NewMongoClient(address, 10, 2, 10)
// 	defer c.CloseClient()

// 	//dbName := "yifants"
// 	col := "geo"

// 	filer := bson.M{
// 		"code": "us",
// 	}

// 	geo := &Geo{}
// 	//result, err := c.Select(&col, filer, 0)
// 	err = c.SelectOne(&col, filer, geo) //查询全部
// 	if err != nil {
// 		t.Fail()
// 	} else {
// 		fmt.Println(geo.Name)
// 	}

// }

type GeoExt struct {
	Code string   `bson:"code"`
	Geo  []string `bson:"geo"`
	Name string   `bson:"name"`
}

func TestSelectAll(t *testing.T) {
	logFile := "/Users/yeahyf/go/src/pubaws/conf/zap.json"
	log.SetLogConf(&logFile)
	address := "mongodb://yifan:123456@192.168.1.10:27017/yifants"
	c, err := NewMongoClient(address, 10, 2, 10)
	defer c.CloseClient()

	//dbName := "yifants"
	col := "geoext"

	// filer := bson.M{
	// 	"name": "杨",
	// 	"age":  10,
	// }
	//result, err := c.Select(&col, filer, 0)
	result, err := c.Select(&col, nil, 0) //查询全部
	if err != nil {
		t.Fail()
	} else {
		for _, v := range result {
			g := &GeoExt{}

			bsonBytes, _ := bson.Marshal(v)
			bson.Unmarshal(bsonBytes, g)

			fmt.Println(g.Code, g.Name, g.Geo)
		}
	}

}

// func TestUpdate(t *testing.T) {
// 	logFile := "/Users/yeahyf/go/src/pubaws/conf/zap.json"
// 	log.SetLogConf(&logFile)
// 	address := "mongodb://yifan:123456@192.168.1.10:27017/yifants"
// 	c, err := NewMongoClient(address, 10, 2, 10)
// 	defer c.CloseClient()

// 	//dbName := "yifants"
// 	col := "updapp"

// 	filer := bson.M{
// 		"pubid":    "bjluhixe",
// 		"platform": "1",
// 		"status":   "1",
// 	}

// 	update := bson.M{"$set": bson.M{"status": "0"}}
// 	match, up, err := c.Update(&col, filer, update)
// 	if err != nil {
// 		t.Fail()
// 	} else {
// 		fmt.Println(match, up)
// 	}
// }
