package build

const (
	ComQueryRequestPacket string     = "【查询】"
	OkPacket string                  = "【正确】"
	ErrorPacket string               = "【错误】"
	PreparePacket string             = "【预处理】"
	SendClientHandshakePacket string = "【用户认证】"
	SendServerHandshakePacket string = "【登录认证】"
)

const (
	COM_SLEEP byte          = 0
	COM_QUIT                = 1
	COM_INIT_DB             = 2
	COM_QUERY               = 3
	COM_FIELD_LIST          = 4
	COM_CREATE_DB           = 5
	COM_DROP_DB             = 6
	COM_REFRESH             = 7
	COM_SHUTDOWN            = 8
	COM_STATISTICS          = 9
	COM_PROCESS_INFO        = 10
	COM_CONNECT             = 11
	COM_PROCESS_KILL        = 12
	COM_DEBUG               = 13
	COM_PING                = 14
	COM_TIME                = 15
	COM_DELAYED_INSERT      = 16
	COM_CHANGE_USER         = 17
	COM_BINLOG_DUMP         = 18
	COM_TABLE_DUMP          = 19
	COM_CONNECT_OUT         = 20
	COM_REGISTER_SLAVE      = 21
	COM_STMT_PREPARE        = 22
	COM_STMT_EXECUTE        = 23
	COM_STMT_SEND_LONG_DATA = 24
	COM_STMT_CLOSE          = 25
	COM_STMT_RESET          = 26
	COM_SET_OPTION          = 27
	COM_STMT_FETCH          = 28
	COM_DAEMON              = 29
	COM_BINLOG_DUMP_GTID    = 30
	COM_RESET_CONNECTION    = 31
)

const (
	MYSQL_TYPE_DECIMAL byte = 0
	MYSQL_TYPE_TINY         = 1
	MYSQL_TYPE_SHORT        = 2
	MYSQL_TYPE_LONG         = 3
	MYSQL_TYPE_FLOAT        = 4
	MYSQL_TYPE_DOUBLE       = 5
	MYSQL_TYPE_NULL         = 6
	MYSQL_TYPE_TIMESTAMP    = 7
	MYSQL_TYPE_LONGLONG     = 8
	MYSQL_TYPE_INT24        = 9
	MYSQL_TYPE_DATE         = 10
	MYSQL_TYPE_TIME         = 11
	MYSQL_TYPE_DATETIME     = 12
	MYSQL_TYPE_YEAR         = 13
	MYSQL_TYPE_NEWDATE      = 14
	MYSQL_TYPE_VARCHAR      = 15
	MYSQL_TYPE_BIT          = 16
)

const (
	MYSQL_TYPE_JSON byte = iota + 0xf5
	MYSQL_TYPE_NEWDECIMAL
	MYSQL_TYPE_ENUM
	MYSQL_TYPE_SET
	MYSQL_TYPE_TINY_BLOB
	MYSQL_TYPE_MEDIUM_BLOB
	MYSQL_TYPE_LONG_BLOB
	MYSQL_TYPE_BLOB
	MYSQL_TYPE_VAR_STRING
	MYSQL_TYPE_STRING
	MYSQL_TYPE_GEOMETRY
)
