package build

import (
	"fmt"
	"io"
	"time"
)

type Message struct {
	Key          []byte
	Value        []byte
	Offset       int64
	Crc          uint32
	Magic        byte
	CompressCode byte
	Topic        string
	Partition    int32
	TipOffset    int64
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

func ReadProduceRequest(r io.Reader, version int16) *ProduceReq {
	// version == 1

	produceReq := ProduceReq{}

	if int(version) >= ApiV3 {
		produceReq.TransactionalID, _ = ReadString(r)
		fmt.Println(produceReq.TransactionalID)
	}

	produceReq.RequiredAcks = ReadInt16(r)
	produceReq.Timeout = time.Duration(ReadInt32(r)) * time.Millisecond

	l := ReadInt32(r)
	produceReq.Topics = make([]ProduceReqTopic, l)

	for ti := range produceReq.Topics {
		var topic = &produceReq.Topics[ti]
		topic.Name, _ = ReadString(r)

		l := ReadInt32(r)
		topic.Partitions = make([]ProduceReqPartition, l)

		for idx := 0; idx < int(l); idx++ {
			topic.Partitions[idx].ID = ReadInt32(r)
			_ = ReadInt32(r) // partitions size
			topic.Partitions[idx].Messages = ReadMessages(r, version)
		}
	}

	return &produceReq
}

type ProduceRspPartitions struct {
	PartitionID int32
	Error       int16
	Offset      int64
}

type ProduceRspTopic struct {
	TopicName    string
	Partitions   []ProduceRspPartitions
	ThrottleTime int32
}

type ProduceRsp struct {
	Topics []ProduceRspTopic
}

func ReadProduceResponse(r io.Reader, version int16) *ProduceRsp {
	// version == 1
	produceRsp := ProduceRsp{}
	l := ReadInt32(r)
	produceRsp.Topics = make([]ProduceRspTopic, 0)
	for i := 0; i < int(l); i++ {
		topic := ProduceRspTopic{}
		topic.TopicName, _ = ReadString(r)
		pl := ReadInt32(r)
		topic.Partitions = make([]ProduceRspPartitions, 0)
		for j := 0; j < int(pl); j++ {
			pt := ProduceRspPartitions{}
			pt.PartitionID = ReadInt32(r)
			pt.Error = ReadInt16(r)
			_, pt.Offset = ReadInt64(r)
			topic.Partitions = append(topic.Partitions, pt)
		}
		produceRsp.Topics = append(produceRsp.Topics, topic)
	}
	return &produceRsp
}

type MetadataReq struct {
	TopicNames []string
}

func ReadMetadataRequest(r io.Reader, version int16) *MetadataReq {
	// version == 0
	metadataReq := MetadataReq{}

	l := ReadInt32(r)
	for i := 0; i < int(l); i++ {
		topicName, _ := ReadString(r)
		metadataReq.TopicNames = append(metadataReq.TopicNames, topicName)
	}

	return &metadataReq
}

type Broker struct {
	NodeID int32
	Host   string
	Port   int32
}

type PartitionMetada struct {
	ErrorCode      int16
	PartitionIndex int32
	LeaderID       int32
	ReplicaNodes   []int32
	IsrNodes       []int32
}

type TopicMetadata struct {
	ErrorCode  int16
	Name       string
	Partitions []PartitionMetada
}

type MetadataRsp struct {
	Brokers []Broker
	Topics  []TopicMetadata
}

func ReadMetadataResponse(r io.Reader, version int16) *MetadataRsp {
	// version == 0
	metadataRsp := MetadataRsp{}

	// read brokers
	metadataRsp.Brokers = make([]Broker, 0)
	l := ReadInt32(r)
	for i := 0; i < int(l); i++ {
		broker := Broker{}
		broker.NodeID = ReadInt32(r)
		broker.Host, _ = ReadString(r)
		broker.Port = ReadInt32(r)
		metadataRsp.Brokers = append(metadataRsp.Brokers, broker)
	}

	// read topics
	metadataRsp.Topics = make([]TopicMetadata, 0)
	l = ReadInt32(r)
	for i := 0; i < int(l); i++ {
		topicMetadata := TopicMetadata{}
		topicMetadata.ErrorCode = ReadInt16(r)
		topicMetadata.Name, _ = ReadString(r)
		pl := ReadInt32(r)
		topicMetadata.Partitions = make([]PartitionMetada, 0)
		for j := 0; j < int(pl); j++ {
			pm := PartitionMetada{}
			pm.ErrorCode = ReadInt16(r)
			pm.PartitionIndex = ReadInt32(r)
			pm.LeaderID = ReadInt32(r)

			pm.ReplicaNodes = make([]int32, 0)
			replicaLen := ReadInt32(r)
			for ri := 0; ri < int(replicaLen); ri++ {
				pm.ReplicaNodes = append(pm.ReplicaNodes, ReadInt32(r))
			}

			pm.IsrNodes = make([]int32, 0)
			isrLen := ReadInt32(r)
			for ri := 0; ri < int(isrLen); ri++ {
				pm.IsrNodes = append(pm.IsrNodes, ReadInt32(r))
			}
			topicMetadata.Partitions = append(topicMetadata.Partitions, pm)
		}
		metadataRsp.Topics = append(metadataRsp.Topics, topicMetadata)
	}

	return &metadataRsp
}

type Action struct {
	Request    string
	Direction  string
	ApiVersion int16
	Message    interface{}
}
