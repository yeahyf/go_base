package hbase

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/yeahyf/go_base/log"
)

//
//  https://bbs.csdn.net/topics/607460480
//  https://qa.1r1g.com/sf/ask/3590832381/
//  https://www.jianshu.com/p/5e7bb33e8e12

var PoolClosedErr = errors.New("connection pool closed")

// Connection 连接接口，按照业务需求定制
type Connection interface {
	CreateNameSpace() error //创建表空间
	DeleteNameSpace() error //删除表空间

	CreateTable(tableName string, familyNames []string) error                          //创建表
	CreateTableWithVer(tableName string, familyNames []string, maxVersion int32) error //创建表
	DisableTable(tableName string) error                                               //停用表
	EnableTable(tableName string) error                                                //停用表
	DeleteTable(tableName string) error                                                //删除表
	ListAllTable() ([]string, error)                                                   //列出所有的表名

	UpdateRow(tableName, rowKey string, values map[string]map[string][]byte) error                                   //更新存档
	FetchRow(tableName, rowKey string, columnKeys map[string][]string) (map[string][]byte, error)                    //获取存档
	FetchRowByVer(tableName, rowKey string, columnKeys map[string][]string, maxVer int32) (map[string][]byte, error) //获取存档
	DeleteRow(tableName, rowKey string) error                                                                        //删除存档
	DeleteColumns(tableName, rowKey string, columnKeys map[string][]string) error                                    //删除存档中的一些Key
	ExistRow(tableName string, rowKey string) (bool, error)

	isOpen() bool                   //连接是否
	open() error                    //打开连接
	close()                         //关闭连接
	isOverdue(t time.Duration) bool //是否超期
}

// ConnFactory 创建连接资源的工厂方法
type ConnFactory func(url, user, passwd, spaceName string) (Connection, error)

// CommonConn 连接结构体，
type CommonConn struct {
	conn     Connection //连接资源
	idleTime time.Time  //开始空闲的时间
}

// ConnectionPool 连接池
type ConnectionPool struct {
	mutex       sync.Mutex       // 互斥量，用于并发访问控制
	cons        chan *CommonConn // 实际的“池”，用一个通道保存连接
	connFactory ConnFactory      // 创建连接的工厂方法
	closed      bool             // 连接池是否关闭
	conf        *PoolConf        // 连接池的配置参数
	inUsed      int32            //正在被用的连接数
	notify      chan struct{}    //获取不到连接时候的通知
}

func NewConnPool(factory func(url, user, passwd, spaceName string) (Connection, error), conf *PoolConf) *ConnectionPool {
	if conf.MaxOpenSize <= 0 {
		conf.MaxOpenSize = 50
	}
	if conf.MinIdleSize <= 0 {
		conf.MinIdleSize = 5
	}
	// 如果最大空闲时间等于0,则无须考虑这个值
	if conf.MaxIdleTime < 0 {
		conf.MaxIdleTime = 600
	}
	if conf.MinIdleSize >= conf.MaxOpenSize {
		conf.MinIdleSize = conf.MaxOpenSize
	}
	// 参数检查后，设置连接池属性
	cp := &ConnectionPool{
		mutex:       sync.Mutex{},
		cons:        make(chan *CommonConn, conf.MaxOpenSize),
		connFactory: factory,
		closed:      false, //连接池是否关闭状态
		conf:        conf,
		notify:      make(chan struct{}),
	}
	//初始化连接池中的基本连接
	for i := 0; i < conf.MinIdleSize; i++ {
		connRes, err := cp.connFactory(conf.Address, conf.User, conf.Passwd, conf.SpaceName)
		//如果启动的时候都无法创建连接,说明问题严重
		if err != nil {
			cp.Close()
			panic("error in NewConnPool while calling connFactory")
		}
		cp.cons <- &CommonConn{conn: connRes, idleTime: time.Now()} // 连接放入池中
	}
	go cp.balanceControl()
	//go cp.addNewConnection()
	return cp
}

// addNewConnection 异步增加一个新的连接
//func (cp *ConnectionPool) addNewConnection() {
//	for {
//		<-cp.notify
//		if cp.closed {
//			break
//		}
//		used := int(cp.inUsed)
//		idled := len(cp.cons)
//		//未超过最大限制
//		if used+idled < cp.conf.MaxOpenSize {
//			connRes, err := cp.connFactory(cp.conf.Address, cp.conf.User, cp.conf.Passwd)
//			if err != nil {
//				log.Errorf("error in NewConnPool while calling connFactory %v", err)
//				continue
//			}
//			if cp.closed {
//				connRes.close() //不要忘记关闭
//				break
//			}
//			cp.cons <- &CommonConn{conn: connRes, idleTime: time.Now()} // 连接放入池中
//		}
//	}
//}

// balanceControl 当发现总体连接降低到最小值的时候,补充新的连接
func (pool *ConnectionPool) balanceControl() {
	for {
		select {
		case <-time.After(5 * time.Second):
		case <-pool.notify: //紧急通知需要增加连接
		}
		//如果连接池已经关闭,则直接退出
		if pool.closed {
			break
		}
		used := int(pool.inUsed)
		idled := len(pool.cons)
		//未超过最大限制 宾缺
		if used+idled < pool.conf.MaxOpenSize && idled < pool.conf.MinIdleSize {
			if log.IsDebug() {
				log.Debugf("current idled too little: %d", idled)
				log.Debugf("current used: %d", used)
			}
			connRes, err := pool.connFactory(pool.conf.Address,
				pool.conf.User, pool.conf.Passwd, pool.conf.SpaceName)
			if err != nil {
				log.Errorf("error in NewConnPool while calling connFactory %v", err)
				continue
			}
			if pool.closed {
				connRes.close() //不要忘记关闭
				break
			}
			pool.cons <- &CommonConn{conn: connRes, idleTime: time.Now()} // 连接放入池中
			continue
		}
		if idled > pool.conf.MaxIdleSize {
			if log.IsDebug() {
				log.Debugf("current idled too much: %d", idled)
				log.Debugf("current used: %d", used)
			}
			hbaseConn, _ := <-pool.cons
			hbaseConn.conn.close()
		}
	}
}

// Get 从连接池中获取一个连接
func (pool *ConnectionPool) Get(ctx context.Context) (conn Connection, err error) {
	if pool.closed {
		return nil, PoolClosedErr
	}
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		//如果连接池中有连接可以用
		case hbaseConn, ok := <-pool.cons:
			if !ok { //判断存放连接的chan是否关闭
				return nil, PoolClosedErr
			}
			// 拿到的连接已经超时，将其关闭，继续取下一个
			if time.Now().Sub(hbaseConn.idleTime) > pool.conf.MaxIdleTime ||
				(pool.conf.MaxIdleTime != 0 && hbaseConn.conn.isOverdue(pool.conf.MaxIdleTime)) {
				hbaseConn.conn.close()
				continue
			}
			//成功从连接池中获取到一个连接,增加一个在用的连接
			atomic.AddInt32(&pool.inUsed, 1)
			conn, err = hbaseConn.conn, nil
			goto prepare
		case <-time.After(50 * time.Millisecond):
			//超时后发送通知,告知平衡控制需要增加连接
			{
				pool.notify <- struct{}{}
			}
		}
	}
prepare: // 统一在此处进行链路的处理
	if conn != nil {
		//判断通讯链路是否是打开的
		if !conn.isOpen() {
			if err = conn.open(); err != nil {
				return
			}
		}
	}
	return
}

// Put 用完归还一个连接到连接池
func (pool *ConnectionPool) Put(conn Connection) error {
	if pool.closed {
		return PoolClosedErr
	}
	//在用的连接减少
	atomic.AddInt32(&pool.inUsed, -1)
	select {
	//将连接放回到池子中，更新时间
	case pool.cons <- &CommonConn{conn: conn, idleTime: time.Now()}:
		return nil
	default:
		conn.close() // 连接池已满，无法放入资源，将这个连接关闭
		return nil
	}
}

// Close 关闭连接池
func (pool *ConnectionPool) Close() {
	if pool.closed {
		return
	}
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	pool.closed = true
	close(pool.notify)
	close(pool.cons) // 关闭通道，即连接“池”
	for conn := range pool.cons {
		conn.conn.close()
	}
}

// PoolConf 连接池配置参数
// 原则上一个连接池是连接到一个Hbase服务器的，与命名空间无关
// 但是为了简化处理，将一个命名空间封装到Pool的配置中
type PoolConf struct {
	SpaceName string //命名空间
	Address   string //hbase地址
	User      string //用户名
	Passwd    string //密码

	MinIdleSize int           //最小空闲数
	MaxIdleSize int           //最大空闲数
	MaxOpenSize int           //最大连接数,总体不能超过这个
	MaxIdleTime time.Duration //最大空闲时间
	MaxLifeTime time.Duration //最大生命周期
}

func NewPoolByParam(spaceName, address, user, passwd string, minIdleSize, maxIdleSize,
	maxOpenSize int, maxIdleTime, maxLifeTime time.Duration) *ConnectionPool {
	//spaceName 不能为空
	if spaceName == "" {
		panic("hbase spaceName not set")
	}
	poolConf := &PoolConf{
		SpaceName: spaceName,
		Address:   address,
		User:      user,
		Passwd:    passwd,

		MinIdleSize: minIdleSize,
		MaxIdleSize: maxIdleSize,
		MaxOpenSize: maxOpenSize,
		MaxIdleTime: maxIdleTime,
		MaxLifeTime: maxLifeTime,
	}
	return NewPoolByCfg(poolConf)
}

// NewPoolByCfg 构建一个新的Hbase Connection Pool
func NewPoolByCfg(poolCfg *PoolConf) *ConnectionPool {
	//SpaceName = cfg.GetString("hbase.namespace")
	//if SpaceName == "" {
	//	panic("hbase namespace not set!!!")
	//}
	//poolConf := &PoolConf{
	//	Address: cfg.GetString("hbase.url.address"),
	//	User:    cfg.GetString("hbase.user.name"),
	//	Passwd:  cfg.GetString("hbase.user.passwd"),
	//
	//	MinIdleSize: cfg.GetInt("hbase.min.idle.size"),
	//	MaxIdleSize: cfg.GetInt("hbase.max.idle.size"),
	//	MaxOpenSize: cfg.GetInt("hbase.max.open.size"),
	//	MaxIdleTime: time.Duration(cfg.GetInt("hbase.max.idle.time")) * time.Second,
	//	MaxLifeTime: time.Duration(cfg.GetInt("hbase.max.life.time")) * time.Second,
	//}
	return NewConnPool(thriftHBaseConnFactory, poolCfg)
}

func (pool *ConnectionPool) GetConn(ctx context.Context) (Connection, error) {
	return pool.Get(ctx)
}

func (pool *ConnectionPool) ReleaseConn(conn Connection) {
	_ = pool.Put(conn)
}

func (pool *ConnectionPool) CloseConnPool() {
	_ = pool.Close
}
