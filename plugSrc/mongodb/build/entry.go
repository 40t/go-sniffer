package build

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/google/gopacket"
	"io"
	"strconv"
)

const (
	Port = 27017
	Version = "0.1"
	CmdPort = "-p"
)

type Mongodb struct {
	port    int
	version string
	source  map[string]*stream
}

type stream struct {
	packets chan *packet
}

type packet struct {

	isClientFlow  bool   //客户端->服务器端流

	messageLength int    //总消息大小
	requestID     int    //此消息的标识符
	responseTo    int    //从原始请求的requestID
	opCode        int 	 //请求类型

	payload       io.Reader
}

var mongodbInstance *Mongodb

func NewInstance() *Mongodb {
	if mongodbInstance == nil {
		mongodbInstance = &Mongodb{
			port   :Port,
			version:Version,
			source: make(map[string]*stream),
		}
	}
	return mongodbInstance
}

func (m *Mongodb) SetFlag(flg []string)  {
	c := len(flg)
	if c == 0 {
		return
	}
	if c >> 1 != 1 {
		panic("Mongodb参数数量不正确!")
	}
	for i:=0;i<c;i=i+2 {
		key := flg[i]
		val := flg[i+1]

		switch key {
		case CmdPort:
			p, err := strconv.Atoi(val);
			if err != nil {
				panic("端口数不正确")
			}
			mongodbInstance.port = p
			if p < 0 || p > 65535 {
				panic("参数不正确: 端口范围(0-65535)")
			}
			break
		default:
			panic("参数不正确")
		}
	}
}

func (m *Mongodb) BPFFilter() string {
	return "tcp and port "+strconv.Itoa(m.port);
}

func (m *Mongodb) Version() string {
	return m.version
}

func (m *Mongodb) ResolveStream(net, transport gopacket.Flow, buf io.Reader) {

	//uuid
	uuid := fmt.Sprintf("%v:%v", net.FastHash(), transport.FastHash())

	//resolve packet
	if _, ok := m.source[uuid]; !ok {

		var newStream = stream {
			packets:make(chan *packet, 100),
		}

		m.source[uuid] = &newStream
		go newStream.resolve()
	}

	//read bi-directional packet
	//server -> client || client -> server
	for {

		newPacket := m.newPacket(net, transport, buf)
		if newPacket == nil {
			return
		}

		m.source[uuid].packets <- newPacket
	}
}

func (m *Mongodb) newPacket(net, transport gopacket.Flow, r io.Reader) *packet {

	//read packet
	var packet *packet
	var err error
	packet, err = readStream(r)

	//stream close
	if err == io.EOF {
		fmt.Println(net, transport, " 关闭")
		return nil
	} else if err != nil {
		fmt.Println("流解析错误", net, transport, ":", err)
		return nil
	}

	//set flow direction
	if transport.Src().String() == strconv.Itoa(m.port) {
		packet.isClientFlow = false
	}else{
		packet.isClientFlow = true
	}

	return packet
}

func (stm *stream) resolve() {
	for {
		select {
		case packet := <- stm.packets:
			if packet.isClientFlow {
				stm.resolveClientPacket(packet)
			} else {
				stm.resolveServerPacket(packet)
			}
		}
	}
}

func (stm *stream) resolveServerPacket(pk *packet) {
	return
}

func (stm *stream) resolveClientPacket(pk *packet) {

	var msg string
	switch pk.opCode {

	case OP_UPDATE:
		zero               := ReadInt32(pk.payload)
		fullCollectionName := ReadString(pk.payload)
		flags              := ReadInt32(pk.payload)
		selector           := ReadBson2Json(pk.payload)
		update             := ReadBson2Json(pk.payload)
		_ = zero
		_ = flags

		msg = fmt.Sprintf(" [更新] [集合:%s] 语句: %v %v",
			fullCollectionName,
			selector,
			update,
		)

	case OP_INSERT:
		flags              := ReadInt32(pk.payload)
		fullCollectionName := ReadString(pk.payload)
		command            := ReadBson2Json(pk.payload)
		_ = flags

		msg = fmt.Sprintf(" [插入] [集合:%s] 语句: %v",
			fullCollectionName,
			command,
		)

	case OP_QUERY:
		flags              := ReadInt32(pk.payload)
		fullCollectionName := ReadString(pk.payload)
		numberToSkip       := ReadInt32(pk.payload)
		numberToReturn     := ReadInt32(pk.payload)
		_ = flags
		_ = numberToSkip
		_ = numberToReturn

		command            := ReadBson2Json(pk.payload)
		selector           := ReadBson2Json(pk.payload)

		msg = fmt.Sprintf(" [查询] [集合:%s] 语句: %v %v",
			fullCollectionName,
			command,
			selector,
		)

	case OP_COMMAND:
		database           := ReadString(pk.payload)
		commandName        := ReadString(pk.payload)
		metaData           := ReadBson2Json(pk.payload)
		commandArgs        := ReadBson2Json(pk.payload)
		inputDocs          := ReadBson2Json(pk.payload)

		msg = fmt.Sprintf(" [命令] [数据库:%s] [命令名:%s] %v %v %v",
			database,
			commandName,
			metaData,
			commandArgs,
			inputDocs,
		)

	case OP_GET_MORE:
		zero               := ReadInt32(pk.payload)
		fullCollectionName := ReadString(pk.payload)
		numberToReturn     := ReadInt32(pk.payload)
		cursorId           := ReadInt64(pk.payload)
		_ = zero

		msg = fmt.Sprintf(" [查询更多] [集合:%s] [回复数量:%v] [游标:%v]",
			fullCollectionName,
			numberToReturn,
			cursorId,
		)

	case OP_DELETE:
		zero               := ReadInt32(pk.payload)
		fullCollectionName := ReadString(pk.payload)
		flags              := ReadInt32(pk.payload)
		selector           := ReadBson2Json(pk.payload)
		_ = zero
		_ = flags

		msg = fmt.Sprintf(" [删除] [集合:%s] 语句: %v",
			fullCollectionName,
			selector,
		)

	case OP_MSG:
		return
	default:
		return
	}

	fmt.Println(GetNowStr(true) + msg)
}

func readStream(r io.Reader) (*packet, error) {

	var buf bytes.Buffer
	p := &packet{}

	//header
	header := make([]byte, 16)
	if _, err := io.ReadFull(r, header); err != nil {
		return nil,err
	}

	// message length
	payloadLen := binary.LittleEndian.Uint32(header[0:4]) - 16
	p.messageLength = int(payloadLen)

	// opCode
	p.opCode = int(binary.LittleEndian.Uint32(header[12:]))

	if p.messageLength != 0 {
		io.CopyN(&buf, r, int64(payloadLen))
	}

	p.payload = bytes.NewReader(buf.Bytes())

	return p, nil
}
