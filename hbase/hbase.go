package hbase

import (
	"bytes"
	"context"
	"errors"
	"time"

	th "github.com/yeahyf/go_base/hbase/t2hbase"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/yeahyf/go_base/log"
)

var (
	NSExistErr       = errors.New("namespace already existed")
	NSNotExistErr    = errors.New("namespace not existed")
	TableExistErr    = errors.New("table already existed")
	TableNotExistErr = errors.New("table not existed")
	TableEnabledErr  = errors.New("table enabled")
	RowNotFoundErr   = errors.New("row not found")
)

// CreateNameSpace 创建命名空间
// 如果是已经创建过,返回false,如果查询有异常,就再次创建
func (hb *ThriftHbaseConn) CreateNameSpace() error {
	descriptor, err := hb.ServiceClient.GetNamespaceDescriptor(context.Background(), hb.SpaceName)
	//说明该命名空间已经存在过了
	if descriptor != nil {
		return NSExistErr
	}
	//有异常都重新创建
	err = hb.ServiceClient.CreateNamespace(context.Background(),
		&th.TNamespaceDescriptor{Name: hb.SpaceName})
	return err
}

// DeleteNameSpace 删除命名空间
func (hb *ThriftHbaseConn) DeleteNameSpace() error {
	descriptor, _ := hb.ServiceClient.GetNamespaceDescriptor(context.Background(), hb.SpaceName)
	if descriptor == nil {
		return NSNotExistErr
	}
	// 直接构建
	return hb.ServiceClient.DeleteNamespace(context.Background(), hb.SpaceName)
}

// CreateTable 创建表,不带版本,只存储最新的数据
func (hb *ThriftHbaseConn) CreateTable(tableName string, familyNames []string) error {
	return hb.CreateTableWithVer(tableName, familyNames, 0)
}

// CreateTableWithVer 创建表，增加历史版本，一般情况下是不需要直接调用该接口的
// maxVersion 可以保留的最多的版本数，每次修改都会生成一个新的版本，并且必须是全部所有字段统一更新
func (hb *ThriftHbaseConn) CreateTableWithVer(tableName string, familyNames []string, maxVersion int32) error {
	tbName := &th.TTableName{Ns: []byte(hb.SpaceName), Qualifier: []byte(tableName)}
	result, err := hb.ServiceClient.TableExists(context.Background(), tbName)
	if err == nil && result {
		return TableExistErr
	}
	columnFamilyDescriptor := make([]*th.TColumnFamilyDescriptor, 0, len(familyNames))
	for _, v := range familyNames {
		if maxVersion > 0 {
			columnFamilyDescriptor = append(columnFamilyDescriptor,
				&th.TColumnFamilyDescriptor{
					Name:        []byte(v),
					MaxVersions: &maxVersion,
					//MinVersions: &minVersion,
				})
		} else {
			columnFamilyDescriptor = append(columnFamilyDescriptor,
				&th.TColumnFamilyDescriptor{
					Name: []byte(v),
				})
		}
	}
	return hb.ServiceClient.CreateTable(context.Background(),
		&th.TTableDescriptor{
			TableName: tbName,
			Columns:   columnFamilyDescriptor,
		}, nil)
}

// DisableTable 停用表
func (hb *ThriftHbaseConn) DisableTable(tableName string) error {
	tbName := &th.TTableName{Ns: []byte(hb.SpaceName), Qualifier: []byte(tableName)}
	//先判断表是否存在
	exist, err := hb.ServiceClient.TableExists(context.Background(), tbName)
	if err == nil {
		if exist {
			enabled, err := hb.ServiceClient.IsTableEnabled(context.Background(), tbName)
			if err == nil && enabled {
				return hb.ServiceClient.DisableTable(context.Background(), tbName)
			} else {
				return err
			}
		} else {
			return TableNotExistErr
		}
	}
	return err
}

// EnableTable 启用表
func (hb *ThriftHbaseConn) EnableTable(tableName string) error {
	tbName := &th.TTableName{Ns: []byte(hb.SpaceName), Qualifier: []byte(tableName)}
	//先判断表是否存在
	exist, err := hb.ServiceClient.TableExists(context.Background(), tbName)
	if err == nil {
		if exist {
			disabled, err := hb.ServiceClient.IsTableDisabled(context.Background(), tbName)
			if err == nil && disabled {
				return hb.ServiceClient.EnableTable(context.Background(), tbName)
			} else {
				return err
			}
		} else {
			return TableNotExistErr
		}
	}
	return err
}

// DeleteTable 删除表 必须要具备的条件，1. 表存在，2 表是disabled
func (hb *ThriftHbaseConn) DeleteTable(tableName string) error {
	tbName := &th.TTableName{Ns: []byte(hb.SpaceName), Qualifier: []byte(tableName)}
	//先判断表是否存在
	exist, err := hb.ServiceClient.TableExists(context.Background(), tbName)

	if err == nil {
		//存在,删除
		if exist {
			disabled, err := hb.ServiceClient.IsTableDisabled(context.Background(), tbName)
			if err == nil {
				// 表不是enable的,才能够删除
				if disabled {
					return hb.ServiceClient.DeleteTable(context.Background(), tbName)
				} else {
					// 表是是enable的,不能删除
					return TableEnabledErr
				}
			}
		} else {
			//表不存在
			return TableNotExistErr
		}
	}
	return err
}

// ListAllTable 列出空间中所有的表名
func (hb *ThriftHbaseConn) ListAllTable() ([]string, error) {
	list, err := hb.ServiceClient.GetTableNamesByNamespace(context.Background(), hb.SpaceName)
	if err != nil {
		return nil, err
	}
	tList := make([]string, 0, len(list))
	for _, v := range list {
		tList = append(tList, string(v.Qualifier))
	}
	return tList, nil
}

// UpdateRow 更新row
func (hb *ThriftHbaseConn) UpdateRow(tableName, rowKey string, values map[string]map[string][]byte) error {
	//做DML操作时，表名参数为bytes，表名的规则是namespace + 冒号 + 表名  []byte("ass:tableName")
	//先计算需要更新的Column的数量
	number := 0
	for _, v := range values {
		number += len(v)
	}
	cv := make([]*th.TColumnValue, 0, number)
	for k, v := range values {
		for kk, vv := range v {
			tc := &th.TColumnValue{Family: []byte(k), Qualifier: []byte(kk), Value: vv}
			cv = append(cv, tc)
		}
	}
	tPut := &th.TPut{
		Row:          []byte(rowKey),
		ColumnValues: cv,
	}
	//此处需要注意，需要增加NameSpace前缀
	tbName := bytes.Buffer{}
	tbName.WriteString(hb.SpaceName)
	tbName.WriteByte(':')
	tbName.WriteString(tableName)
	return hb.ServiceClient.Put(context.Background(), tbName.Bytes(), tPut)
}

// FetchRow 获取一条Row
func (hb *ThriftHbaseConn) FetchRow(tableName, rowKey string, columnKeys map[string][]string) (map[string][]byte, error) {
	//做DML操作时，表名参数为bytes，表名的规则是namespace + 冒号 + 表名
	var tGet *th.TGet
	//根据参数获取不同的数据
	number := 0
	if columnKeys != nil {
		for _, v := range columnKeys {
			number += len(v)
		}
	}
	if number > 0 {
		tColumns := make([]*th.TColumn, 0, number)
		for k, v := range columnKeys {
			for _, c := range v {
				tColumn := &th.TColumn{Family: []byte(k), Qualifier: []byte(c)}
				tColumns = append(tColumns, tColumn)
			}
		}
		tGet = &th.TGet{
			Row:     []byte(rowKey),
			Columns: tColumns,
		}
	} else {
		tGet = &th.TGet{
			Row: []byte(rowKey),
		}
	}
	//此处需要注意，需要增加NameSpace前缀
	tbName := bytes.Buffer{}
	tbName.WriteString(hb.SpaceName)
	tbName.WriteByte(':')
	tbName.WriteString(tableName)
	result, err := hb.ServiceClient.Get(context.Background(), tbName.Bytes(), tGet)
	if err != nil {
		return nil, err
	}
	m := make(map[string][]byte, len(result.ColumnValues))
	for _, v := range result.ColumnValues {
		m[string(v.Qualifier)] = v.Value
	}
	return m, nil
}

// FetchRowByVer 按照版本获取一条Row,最新的版本号最小,从1开始（在创建表的时候需要设置版本信息）
func (hb *ThriftHbaseConn) FetchRowByVer(tableName, rowKey string, columnKeys map[string][]string, maxVer int32) (map[string][]byte, error) {
	//做DML操作时，表名参数为bytes，表名的规则是namespace + 冒号 + 表名
	number := 0
	if columnKeys != nil {
		for _, v := range columnKeys {
			number += len(v)
		}
	}
	var tGet *th.TGet
	if number > 0 {
		tColumns := make([]*th.TColumn, 0, number)
		for k, v := range columnKeys {
			for _, c := range v {
				tColumn := &th.TColumn{Family: []byte(k), Qualifier: []byte(c)}
				tColumns = append(tColumns, tColumn)
			}
		}
		tGet = &th.TGet{
			Row:         []byte(rowKey),
			Columns:     tColumns,
			MaxVersions: &maxVer,
		}
	} else {
		tGet = &th.TGet{
			Row:         []byte(rowKey),
			MaxVersions: &maxVer,
			//TimeRange: &tr,
		}
	}
	//此处需要注意，需要增加NameSpace前缀
	tbName := bytes.Buffer{}
	tbName.WriteString(hb.SpaceName)
	tbName.WriteByte(':')
	tbName.WriteString(tableName)
	result, err := hb.ServiceClient.Get(context.Background(), tbName.Bytes(), tGet)
	if err != nil {
		return nil, err
	}
	m := make(map[string][]byte, len(result.ColumnValues))
	for _, v := range result.ColumnValues {
		m[string(v.Qualifier)] = v.Value
	}
	return m, nil
}

// ExistRow 判断某行数据是否存在
func (hb *ThriftHbaseConn) ExistRow(tableName string, rowKey string) (bool, error) {
	tGet := &th.TGet{
		Row: []byte(rowKey),
	}
	tbName := bytes.Buffer{}
	tbName.WriteString(hb.SpaceName)
	tbName.WriteByte(':')
	tbName.WriteString(tableName)
	return hb.ServiceClient.Exists(context.Background(), tbName.Bytes(), tGet)
}

// DeleteRow 删除某行数据
func (hb *ThriftHbaseConn) DeleteRow(tableName, rowKey string) error {
	//先判断是否存在再删除
	exist, err := hb.ExistRow(tableName, rowKey)
	if err != nil {
		return err
	}
	//要删除的row不存在
	if !exist {
		return RowNotFoundErr
	}
	tDelete := &th.TDelete{
		Row:        []byte(rowKey),
		DeleteType: th.TDeleteType_DELETE_FAMILY, //删除整个Row
	}
	//此处需要注意，需要增加NameSpace前缀
	tbName := bytes.Buffer{}
	tbName.WriteString(hb.SpaceName)
	tbName.WriteByte(':')
	tbName.WriteString(tableName)
	return hb.ServiceClient.DeleteSingle(context.Background(), tbName.Bytes(), tDelete)
}

// DeleteColumns 删除某些列
func (hb *ThriftHbaseConn) DeleteColumns(tableName, rowKey string, columnKeys map[string][]string) error {
	//如果columnKeys位空,则删除所有,直接使用DeleteRow代替
	if columnKeys == nil {
		return hb.DeleteRow(tableName, rowKey)
	}
	//先判断是否存在再删除
	exist, err := hb.ExistRow(tableName, rowKey)
	if err != nil {
		return err
	}
	//要删除的row不存在
	if !exist {
		return RowNotFoundErr
	}
	number := 0
	for _, v := range columnKeys {
		number += len(v)
	}

	tColumns := make([]*th.TColumn, 0, number)
	for k, v := range columnKeys {
		for _, vv := range v {
			tColumns = append(tColumns, &th.TColumn{
				Family:    []byte(k),
				Qualifier: []byte(vv),
			})
		}
	}

	tDelete := &th.TDelete{
		Row:        []byte(rowKey),
		Columns:    tColumns,
		DeleteType: th.TDeleteType_DELETE_COLUMNS, //删除部分columns
	}
	//此处需要注意，需要增加NameSpace前缀
	tbName := bytes.Buffer{}
	tbName.WriteString(hb.SpaceName)
	tbName.WriteByte(':')
	tbName.WriteString(tableName)
	return hb.ServiceClient.DeleteSingle(context.Background(), tbName.Bytes(), tDelete)
}

// IsOpen 是否处于打开状态
func (hb *ThriftHbaseConn) isOpen() bool {
	return hb.HttpClient.IsOpen()
}

// Open 打开状态
func (hb *ThriftHbaseConn) open() error {
	return hb.HttpClient.Open()
}

// Close 关闭
func (hb *ThriftHbaseConn) close() {
	if hb.HttpClient != nil {
		err := hb.HttpClient.Close()
		if err != nil {
			//Todo
		}
	}
}

// IsOverdue 是否超过最大生命周期
func (hb *ThriftHbaseConn) isOverdue(t time.Duration) bool {
	return time.Now().Sub(hb.CreateTime) > t
}

// ThriftHbaseConn 链接封装
type ThriftHbaseConn struct {
	HttpClient    *thrift.THttpClient
	ServiceClient *th.THBaseServiceClient
	CreateTime    time.Time
	SpaceName     string
}

// thriftHBaseConnFactory 用于产生连接的工厂
func thriftHBaseConnFactory(url, user, passwd, spaceName string) (Connection, error) {
	//部分基础配置
	conf := &thrift.TConfiguration{
		ConnectTimeout: time.Second, //连接超时时间
		SocketTimeout:  time.Second, //通讯超时时间
		//MaxMessageSize:     1024 * 1024 * 256,
		MaxFrameSize:       1024 * 1024 * 256,    //数据帧大小
		TBinaryStrictRead:  thrift.BoolPtr(true), //二进制严格读
		TBinaryStrictWrite: thrift.BoolPtr(true), //二进制严格写
	}
	//协议工厂
	protocolFactory := thrift.NewTBinaryProtocolFactoryConf(conf)

	//生成通讯链路
	transport, err := thrift.NewTHttpClient(url)
	if err != nil {
		log.Errorf("create transport error! %v", err)
		return nil, err
	}
	// TTransport 是一个接口, THttpClient是具体的实现
	// 设置用户名密码
	httpClient := transport.(*thrift.THttpClient)
	httpClient.SetHeader("ACCESSKEYID", user)
	httpClient.SetHeader("ACCESSSIGNATURE", passwd)

	//使用通讯链路生成交互的客户端
	serviceClient := th.NewTHBaseServiceClientFactory(httpClient, protocolFactory)

	hbaseConn := &ThriftHbaseConn{
		HttpClient:    httpClient,    //底层通讯链路
		ServiceClient: serviceClient, //业务接口封装
		CreateTime:    time.Now(),
		SpaceName:     spaceName, //命令空间
	}
	return hbaseConn, nil
}
