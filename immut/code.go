///系统代码表
package immut

const (
	//数据格式错误 1000-1999
	CodeExHttpMethod uint32 = 1000 //http请求方法错误
	CodeExVersion    uint32 = 1001 //版本号错误
	CodeExNonce      uint32 = 1002 //nonce错误
	CodeExTs         uint32 = 1003 //时间戳格式错误
	CodeExSignature  uint32 = 1004 //签名错误
	CodeExRepeatReq  uint32 = 1005 //随机数重复
	CodeExAppkey    uint32 = 1008 //appkey错误

	CodeExProtobufUn uint32 = 1006 //请求参数
	CodeExProtobufMa uint32 = 1007 //请求参数

	CodeExDdbMa uint32 = 1100 //Ddb序列化错误
	CodeExDdbUn uint32 = 1101 //Ddb反序列化错误

	/// 业务错误 2000-2999 用户系统
	/// 业务错误 3000-3999 存档系统
	/// 业务错误 4000-4999 排行榜系统
	/// 业务错误 5000-5999 防沉迷系统

	///服务器内部错误 9000-9999
	CodeExRedis  uint32 = 9000 //Redis交互异常
	CodeExJsonUn uint32 = 9001 //json反序列化错误
	CodeExJsonMa uint32 = 9002 //json序列化错误

	CodeExReadIO        uint32 = 9005 //读IO异常
	CodeExWriteIO       uint32 = 9006 //写IO异常
	CodeExNetTimeout    uint32 = 9007 //网络超时
	CodeExUrlNotExist   uint32 = 9008 //网络资源不存在
	CodeExDdbPutItem    uint32 = 9009 //DDB插入数据错误
	CodeExDDBGetItem    uint32 = 9010 //DDB读取数据错误
	CodeExDdbUpdateItem uint32 = 9011 //DDB更新数据错误
	CodeExS3Donwload    uint32 = 9012 //DDB更新数据错误

	CodeExRDSInert  uint32 = 9100 //插入数据库错误
	CodeExRDSUpdate uint32 = 9101 //更新数据库错误
	CodeExRDSSelect uint32 = 9102 //查询数据库错误
	CodeExRDSDelete uint32 = 9103 //删除数据库错误
)
