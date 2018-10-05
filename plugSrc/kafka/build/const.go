package build

const (
	ProduceRequest          = 0
	FetchRequest            = 1
	OffsetRequest           = 2
	MetadataRequest         = 3
	//Non-user facing control APIs = 4-7
	OffsetCommitRequest     = 8
	OffsetFetchRequest      = 9
	GroupCoordinatorRequest	= 10
	JoinGroupRequest        = 11
	HeartbeatRequest        = 12
	LeaveGroupRequest       = 13
	SyncGroupRequest        = 14
	DescribeGroupsRequest   = 15
	ListGroupsRequest       = 16
	APIVersionsReqKind      = 18
	CreateTopicsReqKind     = 19
)

const (
	ApiV0 = 0
	ApiV1 = 1
	ApiV2 = 2
	ApiV3 = 3
	ApiV4 = 4
	ApiV5 = 5
)
