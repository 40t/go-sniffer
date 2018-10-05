package build

import (
	"bytes"
	"fmt"
	"github.com/google/gopacket"
	"io"
	"strconv"
	"sync"
)

const (
	Port = 9092
	Version = "0.1"
	CmdPort = "-p"
)

type Kafka struct {
	port    int
	version string
	source  map[string]*stream
}

type stream struct {
	packets chan *packet
}

type packet struct {

	isClientFlow  bool   //客户端->服务器端流
	messageSize   int32

	requestHeader
	responseHeader

	payload       io.Reader
}

type requestHeader struct {
	apiKey        int16
	apiVersion    int16
	correlationId int32
	clientId      string
}

type responseHeader struct {
	correlationId int32
}

type messageSet struct {
	offset      int64
	messageSize int32
}
func newMessageSet(r io.Reader) messageSet {
	messageSet := messageSet{}
	messageSet.offset      = ReadInt64(r)
	messageSet.messageSize = ReadInt32(r)

	return messageSet
}

type message struct {
	crc           int32
	magicByte     int8
	attributes    int8
	key           []byte
	value         []byte
}

var kafkaInstance *Kafka
var once sync.Once
func NewInstance() *Kafka {
	once.Do(func() {
		kafkaInstance = &Kafka{
			port   :Port,
			version:Version,
			source: make(map[string]*stream),
		}
	})
	return kafkaInstance
}

func (m *Kafka) SetFlag(flg []string)  {
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
			kafkaInstance.port = p
			if p < 0 || p > 65535 {
				panic("参数不正确: 端口范围(0-65535)")
			}
			break
		default:
			panic("参数不正确")
		}
	}
}

func (m *Kafka) BPFFilter() string {
	return "tcp and port "+strconv.Itoa(m.port);
}

func (m *Kafka) Version() string {
	return m.version
}

func (m *Kafka) ResolveStream(net, transport gopacket.Flow, buf io.Reader) {

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

func (m *Kafka) newPacket(net, transport gopacket.Flow, r io.Reader) *packet {

	//read packet
	pk := packet{}

	//read messageSize
	pk.messageSize = ReadInt32(r)

	//set flow direction
	if transport.Src().String() == strconv.Itoa(m.port) {
		pk.isClientFlow = false

		respHeader := responseHeader{}
		respHeader.correlationId = ReadInt32(r)
		pk.responseHeader = respHeader

		var buf bytes.Buffer
		if _, err := io.CopyN(&buf, r, int64(pk.messageSize-4)); err != nil {
			if err == io.EOF {
				fmt.Println(net, transport, " 关闭")
				return nil
			}
			fmt.Println("流解析错误", net, transport, ":", err)
			return nil
		}

		pk.payload = &buf

	}else{
		pk.isClientFlow = true

		var clientIdLen = 0
		reqHeader := requestHeader{}
		reqHeader.apiKey        = ReadInt16(r)
		reqHeader.apiVersion    = ReadInt16(r)
		reqHeader.correlationId = ReadInt32(r)
		reqHeader.clientId, clientIdLen   = ReadString(r)
		pk.requestHeader = reqHeader
		var buf bytes.Buffer
		if _, err := io.CopyN(&buf, r, int64(pk.messageSize-10) - int64(clientIdLen)); err != nil {
			if err == io.EOF {
				fmt.Println(net, transport, " 关闭")
				return nil
			}
			fmt.Println("流解析错误", net, transport, ":", err)
			return nil
		}
		pk.payload = &buf
	}

	return &pk
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
	payload := pk.payload

	fmt.Println("apiKey:")
	fmt.Println(pk.apiKey)

	switch int(pk.apiKey) {
	case ProduceRequest:
		msg = ReadProduceRequest(payload, pk.apiVersion)
	}
	_=msg
	//fmt.Println(msg)
}