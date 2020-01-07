///系统代码表
package immut

const (
	//数据格式错误 1000-1999
	Code_Ex_HttpMethod uint32 = 1000 //http请求方法错误
	Code_Ex_Version    uint32 = 1001 //版本号错误
	Code_Ex_Nonce      uint32 = 1002 //nonce错误
	Code_Ex_TS         uint32 = 1003 //时间戳格式错误
	Code_Ex_Signature  uint32 = 1004 //签名错误
	Code_Ex_Repeat_Req uint32 = 1005 //随机数重复

	Code_Ex_ProtobufUn uint32 = 1006 //请求参数
	Code_Ex_ProtobufMa uint32 = 1007 //请求参数

	Code_Ex_DdbMa uint32 = 1100 //Ddb序列化错误
	Code_Ex_DdbUn uint32 = 1101 //Ddb反序列化错误

	/// 业务错误 2000-2999 用户系统
	/// 业务错误 3000-3999 存档系统
	/// 业务错误 4000-4999 排行榜系统
	/// 业务错误 5000-5999 防沉迷系统

	///服务器内部错误 9000-9999
	Code_Ex_Redis  uint32 = 9000 //Redis交互异常
	Code_Ex_JsonUn uint32 = 9001 //json反序列化错误
	Code_Ex_JsonMa uint32 = 9002 //json序列化错误

	Code_Ex_ReadIO         uint32 = 9005 //读IO异常
	Code_Ex_WriteIO        uint32 = 9006 //写IO异常
	Code_Ex_Net_Timeout    uint32 = 9007 //网络超时
	Code_Ex_URL_NotExist   uint32 = 9008 //网络资源不存在
	Code_Ex_DDB_PutItem    uint32 = 9009 //DDB插入数据错误
	Code_Ex_DDB_GetItem    uint32 = 9010 //DDB读取数据错误
	Code_Ex_DDB_UpdateItem uint32 = 9011 //DDB更新数据错误
)
