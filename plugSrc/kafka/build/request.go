package build

import (
	"fmt"
	"io"
	"time"
)


type Message struct {
	Key       []byte
	Value     []byte
	Offset    int64
	Crc       uint32
	Topic     string
	Partition int32
	TipOffset int64
}

/**
	Produce request Protocol
	v0, v1 (supported in 0.9.0 or later) and v2 (supported in 0.10.0 or later)
	ProduceRequest => RequiredAcks Timeout [TopicName [Partition MessageSetSize MessageSet]]
		RequiredAcks => int16
		Timeout => int32
		Partition => int32
		MessageSetSize => int32

 */
type ProduceReq struct {
	TransactionalID string
	RequiredAcks    int16
	Timeout         time.Duration
	Topics          []ProduceReqTopic
}

type ProduceReqTopic struct {
	Name       string
	Partitions []ProduceReqPartition
}

type ProduceReqPartition struct {
	ID       int32
	Messages []*Message
}

func ReadProduceRequest(r io.Reader, version int16) string {

	var msg string

	produceReq := ProduceReq{}

	if int(version) >= ApiV3 {
		produceReq.TransactionalID, _ = ReadString(r)
		fmt.Println(produceReq.TransactionalID)
	}

	produceReq.RequiredAcks   = ReadInt16(r)
	produceReq.Timeout        = time.Duration(ReadInt32(r)) * time.Millisecond

	l := ReadInt32(r)
	req := ProduceReq{}
	req.Topics = make([]ProduceReqTopic, l)

	for ti := range req.Topics {
		var topic = &req.Topics[ti]
		topic.Name,_ = ReadString(r)
		fmt.Println("msg")
		fmt.Println(topic.Name)

		l := ReadInt32(r)
		topic.Partitions = make([]ProduceReqPartition, l)

	}

	return msg
}


