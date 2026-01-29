package hbase

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/yeahyf/go_base/file"
	"github.com/yeahyf/go_base/strutil"
	"github.com/yeahyf/go_base/utils"

	"github.com/yeahyf/go_base/log"
)

var (
	// URL       = getEnvOrFatal("HBASE_URL")
	// USER      = getEnvOrFatal("HBASE_USER")
	// PASSWORD  = getEnvOrFatal("HBASE_PASSWORD")
	// SpaceName = getEnvOrFatal("HBASE_NAMESPACE")
	URL       = "http://ld-wz9nuxthmu1a8701f-proxy-lindorm-pub.lindorm.rds.aliyuncs.com:9190"
	USER      = "root"
	PASSWORD  = "pfneaaU1"
	SpaceName = "ass_test"
)

func getEnvOrFatal(envKey string) string {
	value := os.Getenv(envKey)
	if value == "" {
		panic(fmt.Sprintf("Environment variable %s is required but not set", envKey))
	}
	return value
}

var (
	archiveFamilyName  = "a"
	extendFamilyName   = "e"
	timeStampQualifier = "t"
)

var conf = &PoolConf{
	URL,
	USER,
	PASSWORD,

	1,
	2,
	10,
	300 * time.Second,
	3600 * time.Second,
}

func TestCreateNameSpace(t *testing.T) {
	file := "/Users/yeahyf/workproject/archive_service_system/conf/zap.json"
	// cfg.Load(&file)
	log.SetLogConf(&file)
	pool := NewPoolByCfg(conf)
	defer pool.Close()
	conn, err := pool.Get(context.Background())
	if err != nil {
		t.Fatal(err)
		return
	}
	defer pool.Put(conn)

	err = conn.CreateNameSpace("game_asset")
	if err != nil {
		t.Fatal(err)
		return
	}
}

func TestDeleteNameSpace(t *testing.T) {
	file := "/Users/yeahyf/workproject/archive_service_system/conf/zap.json"
	// cfg.Load(&file)
	log.SetLogConf(&file)
	pool := NewPoolByCfg(conf)
	defer pool.Close()
	conn, err := pool.Get(context.Background())
	if err != nil {
		t.Fatal(err)
		return
	}
	defer pool.Put(conn)

	err = conn.DeleteNameSpace(SpaceName)
	if err != nil {
		t.Fatal(err)
	}
}

func TestListAllTable(t *testing.T) {
	file := "/Users/yeahyf/workproject/archive_service_system/conf/zap.json"
	// cfg.Load(&file)
	// Init()
	log.SetLogConf(&file)
	pool := NewPoolByCfg(conf)
	defer pool.Close()
	conn, err := pool.Get(context.Background())
	if err != nil {
		t.Fatal(err)
		return
	}
	defer pool.Put(conn)

	list, err := conn.ListAllTable(SpaceName)
	if err != nil {
		t.Fatal(err)
		return
	}
	// println(list)
	t.Log(len(list))
	for i := range len(list) {
		t.Log(list[i])
		println(list[i])
	}
}

func TestCreateTable(t *testing.T) {
	file := "/Users/yeahyf/workproject/archive_service_system/conf/zap.json"
	// cfg.Load(&file)
	log.SetLogConf(&file)
	pool := NewPoolByCfg(conf)
	defer pool.Close()
	conn, err := pool.Get(context.Background())
	if err != nil {
		t.Fatal(err)
		return
	}
	defer pool.Put(conn)

	tableName := "wars0"
	familys := []string{"a", "e"}
	err = conn.CreateTable(SpaceName, tableName, familys)
	if err != nil {
		t.Fatal(err)
		return
	}
}

func TestExistsTable(t *testing.T) {
	file := "/Users/yeahyf/workproject/archive_service_system/conf/zap.json"
	// cfg.Load(&file)
	log.SetLogConf(&file)
	pool := NewPoolByCfg(conf)
	defer pool.Close()
	conn, err := pool.Get(context.Background())
	if err != nil {
		t.Fatal(err)
		return
	}
	defer pool.Put(conn)

	tableName := "playcity1"

	result, err := conn.ExistTable(SpaceName, tableName)
	if err != nil {
		t.Fatal(err)
		return
	} else {
		t.Log(result)
	}
}

func TestCreateTableWithVersion(t *testing.T) {
	file := "/Users/yeahyf/workproject/archive_service_system/conf/zap.json"
	// cfg.Load(&file)
	log.SetLogConf(&file)
	pool := NewPoolByCfg(conf)
	defer pool.Close()
	conn, err := pool.Get(context.Background())
	if err != nil {
		t.Fatal(err)
		return
	}
	defer pool.Put(conn)

	tableName := "new_with_version"
	familys := []string{"a", "e"}
	err = conn.CreateTableWithVer(SpaceName, tableName, familys, 10)
	if err != nil {
		t.Fatal(err)
		return
	}
}

func TestDeleteTable(t *testing.T) {
	file := "/Users/yeahyf/workproject/archive_service_system/conf/zap.json"
	// cfg.Load(&file)
	log.SetLogConf(&file)
	pool := newConnPool(thriftHBaseConnFactory, conf)
	defer pool.Close()
	conn, err := pool.Get(context.Background())
	if err != nil {
		t.Fatal(err)
		return
	}
	defer pool.Put(conn)

	tableName := "playcity"
	err = conn.DisableTable(SpaceName, tableName)
	err = conn.DeleteTable(SpaceName, tableName)
	if err != nil {
		t.Fatal(err)
		return
	}
}

func TestFetchRow(t *testing.T) {
	file1 := "/Users/yeahyf/workproject/archive_service_system/conf/zap.json"
	// cfg.Load(&file)
	log.SetLogConf(&file1)
	pool := NewPoolByCfg(conf)
	defer pool.Close()

	// 读取文件

	conn, err := pool.Get(context.Background())
	if err != nil {
		t.Fatal(err)
		return
	}
	defer pool.Put(conn)

	tableName := "wars0"
	rowKey := "4gszshxwmx6:0"

	columnKeys := make(map[string][]string, 4)
	columnKeys["a"] = []string{"abnormal", "abnormal_time", "abnormal_reason", "abnormal_remark"}

	m, err := conn.FetchRow("ass_test", tableName, rowKey, columnKeys)
	if err != nil {
		t.Fatal(err)
	}
	result := make(map[string]string)
	for k, v := range m {
		value, e := strutil.Gunzip(v)
		if e != nil {
			result[k] = string(v)
		} else {
			result[k] = string(value)
		}
	}
	for k, v := range result {
		fmt.Printf("%s:%s\n", k, v)
	}

}

func CompressString(s string) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	_, err := gz.Write([]byte(s))
	if err != nil {
		return nil, err
	}

	if err := gz.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func TestZip(t *testing.T) {
	zip, _ := CompressString("Hello World")
	r, _ := strutil.Gunzip(zip)
	println(string(r))
}

func TestDeleteRow(t *testing.T) {
	file := "/Users/yeahyf/workproject/archive_service_system/conf/zap.json"
	// cfg.Load(&file)
	log.SetLogConf(&file)
	pool := NewPoolByCfg(conf)
	defer pool.Close()
	conn, err := pool.Get(context.Background())
	if err != nil {
		t.Fatal(err)
		return
	}
	defer pool.Put(conn)

	tableName := "wars0"
	rowKey := "4gszshxwmx6:0"

	columnKeys := make(map[string][]string, 4)
	columnKeys["a"] = []string{"abnormal", "abnormal_time", "abnormal_reason", "abnormal_remark"}
	err = conn.DeleteColumns(SpaceName, tableName, rowKey, columnKeys)
	if err != nil {
		if !errors.Is(err, RowNotFoundErr) {
			t.Fatal(err)
		}
	} else {
		t.Log("delete successfully")
	}
}

// 更新数据
func TestUpdateRow(t *testing.T) {
	file1 := "/Users/yeahyf/workproject/archive_service_system/conf/zap.json"
	// cfg.Load(&file)
	log.SetLogConf(&file1)
	pool := NewPoolByCfg(conf)
	defer pool.Close()

	source := make(map[string]struct{}, 273)

	file.ReadLine("/Users/yeahyf/修复玩家ID02.csv", func(line *string) {
		source[*line] = struct{}{}
	})

	// 从文件中读取数据

	file.ReadLine("/Users/yeahyf/datadelete.txt", func(line *string) {
		aLine := *line

		id := aLine[0:11]

		uid := id + "0"
		if _, ok := source[uid]; !ok {
			return
		}

		d := aLine[13:]
		aMap := make(map[string]string, 4)
		err := json.Unmarshal([]byte(d), &aMap)
		if err != nil {
			println("err")
		}

		bMap := make(map[string][]byte, 4)
		for k, v := range aMap {
			bMap[k], _ = CompressString(v)
		}

		conn, err := pool.Get(context.Background())
		if err != nil {
			t.Fatal(err)
			return
		}
		defer pool.Put(conn)

		tableName := "fsaq0"
		rowKey := id + ":0"
		println(rowKey)

		//data := make(map[string][]byte, 4)
		// 将获取到的数据放入map中

		extends := make(map[string][]byte, 1)
		ts := time.Now().Unix()
		tdata := utils.GetBytesForInt64(uint64(ts))
		extends[timeStampQualifier] = tdata

		m := make(map[string]map[string][]byte)
		m[archiveFamilyName] = bMap
		m[extendFamilyName] = extends

		err = conn.UpdateRow(SpaceName, tableName, rowKey, m)
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestDateFile(t *testing.T) {
	file.ReadLine("/Users/yeahyf/datadelete.txt", func(line *string) {
		l := *line
		println(l)
	})
}

func TestExistRow(t *testing.T) {
	file := "/Users/yeahyf/workproject/archive_service_system/conf/zap.json"
	// cfg.Load(&file)
	log.SetLogConf(&file)
	pool := NewPoolByCfg(conf)
	defer pool.Close()
	conn, err := pool.Get(context.Background())
	if err != nil {
		t.Fatal(err)
		return
	}
	defer pool.Put(conn)

	tableName := "oncn1"
	rowKey := "4le91psw1p1:1"

	exist, err := conn.ExistRow(SpaceName, tableName, rowKey)
	if err != nil {
		t.Fatal(err)
		t.Log(exist)
	} else {
		println(exist)
	}
}

func TestDeleteColumns(t *testing.T) {
	file := "/Users/yeahyf/workproject/archive_service_system/conf/zap.json"
	// cfg.Load(&file)
	log.SetLogConf(&file)
	pool := NewPoolByCfg(conf)
	defer pool.Close()
	conn, err := pool.Get(context.Background())
	if err != nil {
		t.Fatal(err)
		return
	}
	defer pool.Put(conn)

	tableName := "oncn1"
	rowKey := "41uectphb5w:1"

	m := make(map[string][]string)
	s := []string{"aes"}
	// t1 := []string{"t"}
	m["a"] = s
	// m["e"] = t1

	err = conn.DeleteColumns("", tableName, rowKey, m)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log("delete successful")
	}
}

func TestFetchRowWithVersion(t *testing.T) {
	logfile := "/Users/yeahyf/workproject/archive_service_system/conf/zap.json"
	// cfg.Load(&file)
	log.SetLogConf(&logfile)
	pool := NewPoolByCfg(conf)
	defer pool.Close()

	file, err := os.Open("/Users/yeahyf/黑名单数据.csv")
	if err != nil {
		fmt.Println("open file err")
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	for {
		time.Sleep(100 * time.Millisecond)
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("open file err")
		}
		conn, err := pool.Get(context.Background())
		if err != nil {
			t.Fatal(err)
			return
		}
		tableName := record[4]
		userid := record[0]
		rowKey := userid[:len(userid)-1] + ":" + userid[len(userid)-1:]
		m, err := conn.FetchRowByVer(SpaceName, tableName, rowKey, nil, 11)
		if err != nil {
			t.Fatal(err)
		}

		if v, ok := m["abnormal_count"]; ok {
			println("userid", utils.GetInt64FromBytes(v))
		} else {
			println("=====")
		}
		pool.Put(conn)
	}
}

func TestUpdateAbnormalCount(t *testing.T) {
	logFile := "/Users/yeahyf/workproject/archive_service_system/conf/zap.json"
	// cfg.Load(&file)
	log.SetLogConf(&logFile)
	pool := NewPoolByCfg(conf)
	defer pool.Close()

	file, err := os.Open("/Users/yeahyf/黑名单数据.csv")
	if err != nil {
		fmt.Println("open file err")
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	for {
		time.Sleep(100 * time.Millisecond)
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("open file err")
		}

		// fmt.Println(record)
		// time.Sleep(300 * time.Millisecond)

		conn, err := pool.Get(context.Background())
		if err != nil {
			t.Fatal(err)
			return
		}
		//	defer pool.Put(conn)

		// 从 excel 中读取数据
		tableName := record[4]
		userid := record[0]
		rowKey := userid[:len(userid)-1] + ":" + userid[len(userid)-1:]
		values := make(map[string]map[string][]byte, 1)
		v := make(map[string][]byte, 1)
		v["abnormal_count"] = utils.GetBytesForInt64(uint64(1))
		values["a"] = v

		fmt.Println(tableName, rowKey, values)
		// time.Sleep(3 * time.Minute)
		err = conn.UpdateRow(SpaceName, tableName, rowKey, values)
		if err != nil {
			fmt.Println(record)
		}
		pool.Put(conn)
	}
}
