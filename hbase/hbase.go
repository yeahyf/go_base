package hbase

import (
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

// convertError 将TIOError 转为 普通error
func convertError(err error) error {
	if err == nil {
		return nil
	}
	var tio *th.TIOError
	if errors.As(err, &tio) {
		return errors.New(*tio.Message)
	}
	return err
}

// buildTableName 构建表名，使用指定的命名空间
func (hb *ThriftHbaseConn) buildTableName(namespace, tableName string) []byte {
	// 预分配大小以提高性能
	buf := make([]byte, 0, len(namespace)+1+len(tableName))
	buf = append(buf, namespace...)
	buf = append(buf, ':')
	buf = append(buf, tableName...)
	return buf
}

// buildTTableName 构建 TTableName，使用指定的命名空间
func (hb *ThriftHbaseConn) buildTTableName(namespace, tableName string) *th.TTableName {
	return &th.TTableName{Ns: []byte(namespace), Qualifier: []byte(tableName)}
}

// CreateNameSpace 创建命名空间
func (hb *ThriftHbaseConn) CreateNameSpace(namespace string) error {
	descriptor, err := hb.ServiceClient.GetNamespaceDescriptor(context.Background(), namespace)
	if err != nil {
		// 查询出错，可能是命名空间不存在，尝试创建
		createErr := hb.ServiceClient.CreateNamespace(context.Background(),
			&th.TNamespaceDescriptor{Name: namespace})
		if createErr != nil {
			// 创建失败，返回原始查询错误（可能包含更多信息）
			return convertError(err)
		}
		return convertError(createErr)
	}
	//说明该命名空间已经存在过了
	if descriptor != nil {
		return NSExistErr
	}
	// descriptor 为 nil 且 err 为 nil，说明命名空间不存在，创建它
	err = hb.ServiceClient.CreateNamespace(context.Background(),
		&th.TNamespaceDescriptor{Name: namespace})
	return convertError(err)
}

// DeleteNameSpace 删除命名空间
func (hb *ThriftHbaseConn) DeleteNameSpace(namespace string) error {
	descriptor, err := hb.ServiceClient.GetNamespaceDescriptor(context.Background(), namespace)
	if err != nil {
		return convertError(err)
	}
	if descriptor == nil {
		return NSNotExistErr
	}
	// 直接删除,注意删除需要所有的表都被删除掉才可以删掉命名空间
	err = hb.ServiceClient.DeleteNamespace(context.Background(), namespace)
	return convertError(err)
}

// CreateTable 创建表,不带版本,只存储最新的数据
func (hb *ThriftHbaseConn) CreateTable(namespace, tableName string, familyNames []string) error {
	return hb.CreateTableWithVer(namespace, tableName, familyNames, 0)
}

// ExistTable 判断表是否存在
func (hb *ThriftHbaseConn) ExistTable(namespace, tableName string) (bool, error) {
	tbName := hb.buildTTableName(namespace, tableName)
	result, err := hb.ServiceClient.TableExists(context.Background(), tbName)
	if err != nil {
		return false, convertError(err)
	}
	return result, nil
}

// CreateTableWithVer 创建表，增加历史版本，一般情况下是不需要直接调用该接口的
// maxVersion 可以保留的最多的版本数，每次修改都会生成一个新的版本，并且必须是全部所有字段统一更新
func (hb *ThriftHbaseConn) CreateTableWithVer(namespace, tableName string, familyNames []string, maxVersion int32) error {
	tbName := hb.buildTTableName(namespace, tableName)
	result, err := hb.ServiceClient.TableExists(context.Background(), tbName)
	if err != nil {
		return convertError(err)
	}
	if result {
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
	err = hb.ServiceClient.CreateTable(context.Background(),
		&th.TTableDescriptor{
			TableName: tbName,
			Columns:   columnFamilyDescriptor,
		}, nil)
	return convertError(err)
}

// DisableTable 停用表
func (hb *ThriftHbaseConn) DisableTable(namespace, tableName string) error {
	tbName := hb.buildTTableName(namespace, tableName)
	//先判断表是否存在
	exist, err := hb.ServiceClient.TableExists(context.Background(), tbName)
	if err != nil {
		return convertError(err)
	}
	if !exist {
		return TableNotExistErr
	}
	enabled, err := hb.ServiceClient.IsTableEnabled(context.Background(), tbName)
	if err != nil {
		return convertError(err)
	}
	if enabled {
		return hb.ServiceClient.DisableTable(context.Background(), tbName)
	}
	// 表已经是 disabled 状态，直接返回成功
	return nil
}

// EnableTable 启用表
func (hb *ThriftHbaseConn) EnableTable(namespace, tableName string) error {
	tbName := hb.buildTTableName(namespace, tableName)
	//先判断表是否存在
	exist, err := hb.ServiceClient.TableExists(context.Background(), tbName)
	if err != nil {
		return convertError(err)
	}
	if !exist {
		return TableNotExistErr
	}
	disabled, err := hb.ServiceClient.IsTableDisabled(context.Background(), tbName)
	if err != nil {
		return convertError(err)
	}
	if disabled {
		return hb.ServiceClient.EnableTable(context.Background(), tbName)
	}
	// 表已经是 enabled 状态，直接返回成功
	return nil
}

// DeleteTable 删除表 必须要具备的条件，1. 表存在，2 表是disabled
func (hb *ThriftHbaseConn) DeleteTable(namespace, tableName string) error {
	tbName := hb.buildTTableName(namespace, tableName)
	//先判断表是否存在
	exist, err := hb.ServiceClient.TableExists(context.Background(), tbName)
	if err != nil {
		return convertError(err)
	}
	if !exist {
		//表不存在
		return TableNotExistErr
	}
	//存在,删除
	disabled, err := hb.ServiceClient.IsTableDisabled(context.Background(), tbName)
	if err != nil {
		return convertError(err)
	}
	// 表不是enable的,才能够删除
	if disabled {
		return hb.ServiceClient.DeleteTable(context.Background(), tbName)
	}
	// 表是enable的,不能删除
	return TableEnabledErr
}

// ListAllTable 列出空间中所有的表名
func (hb *ThriftHbaseConn) ListAllTable(namespace string) ([]string, error) {
	list, err := hb.ServiceClient.GetTableNamesByNamespace(context.Background(), namespace)
	if err != nil {
		return nil, convertError(err)
	}
	tList := make([]string, 0, len(list))
	for _, v := range list {
		tList = append(tList, string(v.Qualifier))
	}
	return tList, nil
}

// UpdateRow 更新row
func (hb *ThriftHbaseConn) UpdateRow(namespace, tableName, rowKey string, values map[string]map[string][]byte) error {
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
	tbName := hb.buildTableName(namespace, tableName)
	err := hb.ServiceClient.Put(context.Background(), tbName, tPut)
	return convertError(err)
}

// FetchRow 获取一条Row
func (hb *ThriftHbaseConn) FetchRow(namespace, tableName, rowKey string, columnKeys map[string][]string) (map[string][]byte, error) {
	//做DML操作时，表名参数为bytes，表名的规则是namespace + 冒号 + 表名
	var tGet *th.TGet
	//根据参数获取不同的数据
	number := 0
	for _, v := range columnKeys {
		number += len(v)
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
	tbName := hb.buildTableName(namespace, tableName)
	result, err := hb.ServiceClient.Get(context.Background(), tbName, tGet)
	if err != nil {
		return nil, convertError(err)
	}
	m := make(map[string][]byte, len(result.ColumnValues))
	for _, v := range result.ColumnValues {
		m[string(v.Qualifier)] = v.Value
	}
	return m, nil
}

// FetchRowByVer 按照版本获取一条Row,最新的版本号最小,从1开始（在创建表的时候需要设置版本信息）
func (hb *ThriftHbaseConn) FetchRowByVer(namespace, tableName, rowKey string, columnKeys map[string][]string, maxVer int32) (map[string][]byte, error) {
	//做DML操作时，表名参数为bytes，表名的规则是namespace + 冒号 + 表名
	number := 0
	for _, v := range columnKeys {
		number += len(v)
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
	tbName := hb.buildTableName(namespace, tableName)
	result, err := hb.ServiceClient.Get(context.Background(), tbName, tGet)
	if err != nil {
		return nil, convertError(err)
	}
	m := make(map[string][]byte, len(result.ColumnValues))
	for _, v := range result.ColumnValues {
		m[string(v.Qualifier)] = v.Value
	}
	return m, nil
}

// ExistRow 判断某行数据是否存在
func (hb *ThriftHbaseConn) ExistRow(namespace, tableName string, rowKey string) (bool, error) {
	tbName := hb.buildTTableName(namespace, tableName)
	exist, err := hb.ServiceClient.TableExists(context.Background(), tbName)
	if err != nil {
		return false, convertError(err)
	}
	if !exist {
		return false, TableNotExistErr
	}
	tGet := &th.TGet{
		Row: []byte(rowKey),
	}
	tbNameBytes := hb.buildTableName(namespace, tableName)
	exist, err = hb.ServiceClient.Exists(context.Background(), tbNameBytes, tGet)
	return exist, convertError(err)
}

// DeleteRow 删除某行数据
func (hb *ThriftHbaseConn) DeleteRow(namespace, tableName, rowKey string) error {
	//先判断是否存在再删除
	exist, err := hb.ExistRow(namespace, tableName, rowKey)
	if err != nil {
		return convertError(err)
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
	tbName := hb.buildTableName(namespace, tableName)
	err = hb.ServiceClient.DeleteSingle(context.Background(), tbName, tDelete)
	return convertError(err)
}

// DeleteColumns 删除某些列
func (hb *ThriftHbaseConn) DeleteColumns(namespace, tableName, rowKey string, columnKeys map[string][]string) error {
	//如果columnKeys位空,则删除所有,直接使用DeleteRow代替
	if columnKeys == nil {
		return hb.DeleteRow(namespace, tableName, rowKey)
	}
	//先判断是否存在再删除
	exist, err := hb.ExistRow(namespace, tableName, rowKey)
	if err != nil {
		return convertError(err)
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
	tbName := hb.buildTableName(namespace, tableName)
	err = hb.ServiceClient.DeleteSingle(context.Background(), tbName, tDelete)
	return convertError(err)
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
			log.Errorf("failed to close hbase connection: %v", err)
		}
	}
}

// IsOverdue 是否超过最大生命周期
func (hb *ThriftHbaseConn) isOverdue(t time.Duration) bool {
	return time.Since(hb.CreateTime) > t
}

// ThriftHbaseConn 链接封装
type ThriftHbaseConn struct {
	HttpClient    *thrift.THttpClient
	ServiceClient *th.THBaseServiceClient
	CreateTime    time.Time
	//SpaceName     string
}

// thriftHBaseConnFactory 用于产生连接的工厂
func thriftHBaseConnFactory(url, user, passwd string) (Connection, error) {
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
		//SpaceName:     spaceName, //命名空间
	}
	return hbaseConn, nil
}
