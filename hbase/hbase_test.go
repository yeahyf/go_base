package hbase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yeahyf/go_base/utils"

	"github.com/yeahyf/go_base/log"
)

const (
	URL       = "http://ld-8vb1yr869xuausw73-proxy-lindorm-pub.lindorm.rds.aliyuncs.com:9190"
	USER      = "root"
	PASSWORD  = "root"
	SpaceName = "yifan_test"
)

var (
	archiveFamilyName  = "a"
	extendFamilyName   = "e"
	timeStampQualifier = "t"
)

var conf = &PoolConf{
	SpaceName,
	URL,
	USER,
	PASSWORD,

	1,
	2,
	5,
	300 * time.Second,
	3600 * time.Second,
}

func TestCreateNameSpace(t *testing.T) {
	file := "/Users/yeahyf/workproject/archive_service_system/conf/zap.json"
	//cfg.Load(&file)
	log.SetLogConf(&file)
	pool := NewConnPool(thriftHBaseConnFactory, conf)
	defer pool.Close()
	conn, err := pool.Get(context.Background())
	if err != nil {
		t.Fatal(err)
		return
	}
	defer pool.Put(conn)

	err = conn.CreateNameSpace()
	if err != nil {
		t.Fatal(err)
		return
	}
}
func TestDeleteNameSpace(t *testing.T) {
	file := "/Users/yeahyf/workproject/archive_service_system/conf/zap.json"
	//cfg.Load(&file)
	log.SetLogConf(&file)
	pool := NewConnPool(thriftHBaseConnFactory, conf)
	defer pool.Close()
	conn, err := pool.Get(context.Background())
	if err != nil {
		t.Fatal(err)
		return
	}
	defer pool.Put(conn)

	err = conn.DeleteNameSpace()
	if err != nil {
		t.Fatal(err)
	}
}
func TestListAllTable(t *testing.T) {
	file := "/Users/yeahyf/workproject/archive_service_system/conf/zap.json"
	//cfg.Load(&file)
	//Init()
	log.SetLogConf(&file)
	pool := NewConnPool(thriftHBaseConnFactory, conf)
	defer pool.Close()
	conn, err := pool.Get(context.Background())
	if err != nil {
		t.Fatal(err)
		return
	}
	defer pool.Put(conn)

	list, err := conn.ListAllTable()
	if err != nil {
		t.Fatal(err)
		return
	}
	t.Log(list)
}
func TestCreateTable(t *testing.T) {
	file := "/Users/yeahyf/workproject/archive_service_system/conf/zap.json"
	//cfg.Load(&file)
	log.SetLogConf(&file)
	pool := NewConnPool(thriftHBaseConnFactory, conf)
	defer pool.Close()
	conn, err := pool.Get(context.Background())
	if err != nil {
		t.Fatal(err)
		return
	}
	defer pool.Put(conn)

	tableName := "playcity"
	familys := []string{"a", "e"}
	err = conn.CreateTable(tableName, familys)
	if err != nil {
		t.Fatal(err)
		return
	}
}

func TestCreateTableWithVersion(t *testing.T) {
	file := "/Users/yeahyf/workproject/archive_service_system/conf/zap.json"
	//cfg.Load(&file)
	log.SetLogConf(&file)
	pool := NewConnPool(thriftHBaseConnFactory, conf)
	defer pool.Close()
	conn, err := pool.Get(context.Background())
	if err != nil {
		t.Fatal(err)
		return
	}
	defer pool.Put(conn)

	tableName := "new_2"
	familys := []string{"a", "e"}
	err = conn.CreateTableWithVer(tableName, familys, 10)
	if err != nil {
		t.Fatal(err)
		return
	}
}

func TestDeleteTable(t *testing.T) {
	file := "/Users/yeahyf/workproject/archive_service_system/conf/zap.json"
	//cfg.Load(&file)
	log.SetLogConf(&file)
	pool := NewConnPool(thriftHBaseConnFactory, conf)
	defer pool.Close()
	conn, err := pool.Get(context.Background())
	if err != nil {
		t.Fatal(err)
		return
	}
	defer pool.Put(conn)

	tableName := "playcity"
	err = conn.DisableTable(tableName)
	err = conn.DeleteTable(tableName)
	if err != nil {
		t.Fatal(err)
		return
	}
}
func TestFetchRow(t *testing.T) {
	file := "/Users/yeahyf/workproject/archive_service_system/conf/zap.json"
	//cfg.Load(&file)
	log.SetLogConf(&file)
	pool := NewConnPool(thriftHBaseConnFactory, conf)
	defer pool.Close()
	conn, err := pool.Get(context.Background())
	if err != nil {
		t.Fatal(err)
		return
	}
	defer pool.Put(conn)

	tableName := "playcity"
	rowKey := "2gknb1qvkfu:0"
	m, err := conn.FetchRow(tableName, rowKey, nil)
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range m {
		if k == "t" {
			t.Log(k, utils.GetInt64FromBytes(v))
		} else {
			t.Log(k, string(v))
		}
	}

}
func TestDeleteRow(t *testing.T) {
	file := "/Users/yeahyf/workproject/archive_service_system/conf/zap.json"
	//cfg.Load(&file)
	log.SetLogConf(&file)
	pool := NewConnPool(thriftHBaseConnFactory, conf)
	defer pool.Close()
	conn, err := pool.Get(context.Background())
	if err != nil {
		t.Fatal(err)
		return
	}
	defer pool.Put(conn)

	tableName := "new_2"
	rowKey := "2gknb1qvkfu:2"

	err = conn.DeleteRow(tableName, rowKey)
	if err != nil {
		if !errors.Is(err, RowNotFoundErr) {
			t.Fatal(err)
		}
	} else {
		t.Log("delete successfully")
	}

}
func TestUpdateRow(t *testing.T) {
	file := "/Users/yeahyf/workproject/archive_service_system/conf/zap.json"
	//cfg.Load(&file)
	log.SetLogConf(&file)
	pool := NewConnPool(thriftHBaseConnFactory, conf)
	defer pool.Close()
	conn, err := pool.Get(context.Background())
	if err != nil {
		t.Fatal(err)
		return
	}
	defer pool.Put(conn)

	tableName := "new_2"
	rowKey := "2gknb1qvkfu:2"

	data := make(map[string][]byte, 3)
	a1 := []byte("1_0")
	a2 := []byte("2_0")
	a3 := []byte("3_0")
	a4 := []byte("4_0")

	data["a3"] = a1
	data["a4"] = a2
	data["a1"] = a3
	data["a2"] = a4

	extends := make(map[string][]byte, 1)
	ts := time.Now().Unix()
	tdata := utils.GetBytesForInt64(uint64(ts))
	extends[timeStampQualifier] = tdata

	m := make(map[string]map[string][]byte)
	m[archiveFamilyName] = data
	m[extendFamilyName] = extends

	err = conn.UpdateRow(tableName, rowKey, m)
	if err != nil {
		t.Fatal(err)
	}
}
func TestExistRow(t *testing.T) {
	file := "/Users/yeahyf/workproject/archive_service_system/conf/zap.json"
	//cfg.Load(&file)
	log.SetLogConf(&file)
	pool := NewConnPool(thriftHBaseConnFactory, conf)
	defer pool.Close()
	conn, err := pool.Get(context.Background())
	if err != nil {
		t.Fatal(err)
		return
	}
	defer pool.Put(conn)

	tableName := "playcity"
	rowKey := "2gknb1qvkfu:0"

	exist, err := conn.ExistRow(tableName, rowKey)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(exist)
	}

}
func TestDeleteColumns(t *testing.T) {
	file := "/Users/yeahyf/workproject/archive_service_system/conf/zap.json"
	//cfg.Load(&file)
	log.SetLogConf(&file)
	pool := NewConnPool(thriftHBaseConnFactory, conf)
	defer pool.Close()
	conn, err := pool.Get(context.Background())
	if err != nil {
		t.Fatal(err)
		return
	}
	defer pool.Put(conn)

	tableName := "playcity"
	rowKey := "2gknb1qvkfu:0"

	m := make(map[string][]string)
	s := []string{"a4", "a3", "a1"}
	t1 := []string{"t"}
	m["a"] = s
	m["e"] = t1

	err = conn.DeleteColumns(tableName, rowKey, m)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log("delete successful")
	}

}

func TestFetchRowWithVersion(t *testing.T) {
	file := "/Users/yeahyf/workproject/archive_service_system/conf/zap.json"
	//cfg.Load(&file)
	log.SetLogConf(&file)
	pool := NewConnPool(thriftHBaseConnFactory, conf)
	defer pool.Close()
	conn, err := pool.Get(context.Background())
	if err != nil {
		t.Fatal(err)
		return
	}
	defer pool.Put(conn)

	tableName := "new_2"
	rowKey := "2gknb1qvkfu:2"
	m, err := conn.FetchRowByVer(tableName, rowKey, nil, 11)
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range m {
		if k == "t" {
			t.Log(k, utils.GetInt64FromBytes(v))
		} else {
			t.Log(k, string(v))
		}
	}

}
