package build

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"

	"github.com/google/gopacket"
)

const (
	Port    = 1433
	Version = "0.1"
	CmdPort = "-p"
)

type Mssql struct {
	port    int
	version string
	source  map[string]*stream
}

type stream struct {
	packets chan *packet
}

type packet struct {
	isClientFlow bool
	status       int
	packetType   int
	length       int
	payload      []byte
}

// packet types
// https://msdn.microsoft.com/en-us/library/dd304214.aspx
const (
	packSQLBatch   = 1
	packRPCRequest = 3
	packReply      = 4
	packAttention  = 6

	packBulkLoadBCP = 7
	packTransMgrReq = 14
	packNormal      = 15
	packLogin7      = 16
	packSSPIMessage = 17
	packPrelogin    = 18
)

var mssql *Mssql
var once sync.Once

func NewInstance() *Mssql {

	once.Do(func() {
		mssql = &Mssql{
			port:    Port,
			version: Version,
			source:  make(map[string]*stream),
		}
	})
	return mssql
}

func (m *Mssql) Version() string {
	return m.version
}

func (m *Mssql) BPFFilter() string {
	return "tcp and port " + strconv.Itoa(m.port)
}

func (m *Mssql) SetFlag(flg []string) {
	c := len(flg)

	if c == 0 {
		return
	}

	if c>>1 == 0 {
		fmt.Println("ERR : Mssql Number of parameters")
		os.Exit(1)
	}

	for i := 0; i < c; i = i + 2 {
		key := flg[i]
		val := flg[i+1]

		switch key {
		case CmdPort:
			port, err := strconv.Atoi(val)
			m.port = port
			if err != nil {
				panic("ERR : port")
			}
			if port < 0 || port > 65535 {
				panic("ERR : port(0-65535)")
			}
			break
		default:
			panic("ERR : mssql's params")

		}

	}
}

// ResolveStream ...
func (m *Mssql) ResolveStream(net, transport gopacket.Flow, buf io.Reader) {
	//uuid
	uuid := fmt.Sprintf("%v:%v", net.FastHash(), transport.FastHash())

	// log.Println(uuid)

	if _, ok := m.source[uuid]; !ok {
		var newStream = &stream{
			packets: make(chan *packet, 100),
		}
		m.source[uuid] = newStream
		go newStream.resolve()
	}

	for {

		// log.Println("ssss")

		newPacket := m.newPacket(net, transport, buf)
		if newPacket == nil {
			return
		}
		m.source[uuid].packets <- newPacket

	}
	// log.Println('ddd')
}

func (m *Mssql) newPacket(net, transport gopacket.Flow, r io.Reader) *packet {
	// read packet
	var packet *packet
	var err error
	packet, err = readStream(r)

	//stream close
	if err == io.EOF {
		fmt.Println(net, transport, " close")
		return nil
	} else if err != nil {
		fmt.Println("ERR : Unknown stream", net, transport, ":", err)
		return nil
	}

	//set flow direction
	if transport.Src().String() == strconv.Itoa(m.port) {
		packet.isClientFlow = false
	} else {
		packet.isClientFlow = true
	}
	return packet
}

func (m *stream) resolve() {
	for {
		select {
		case packet := <-m.packets:
			if packet.isClientFlow {
				m.resolveClientPacket(packet)
			} else {
				m.resolveServerPacket(packet)
			}
		}
	}
}

func readStream(r io.Reader) (*packet, error) {

	var buffer bytes.Buffer

	header := make([]byte, 8)
	p := &packet{}
	if _, err := io.ReadFull(r, header); err != nil {
		return nil, err
	}

	p.packetType = int(uint32(header[0]))
	p.status = int(uint32(header[1]))
	p.length = int(binary.BigEndian.Uint16(header[2:4]))

	if p.length > 0 {
		io.CopyN(&buffer, r, int64(p.length-8))
	}
	p.payload = buffer.Bytes()
	return p, nil
}

func (m *stream) resolveClientPacket(p *packet) {

	var msg string

	switch p.packetType {
	case 1:
		headerLength := int(binary.LittleEndian.Uint32(p.payload[0:4]))
		// fmt.Printf("headers %d\n", headerLength)
		if headerLength > 22 {
			//not exists headers
			msg = fmt.Sprintf("【query】 %s", string(p.payload))

		} else {
			//tds 7.2+
			msg = fmt.Sprintf("【query】 %s", string(p.payload[headerLength:]))
		}

	case 4:
		msg = fmt.Sprintf("【query】 %s", "Tabular result")

	}

	fmt.Println(GetNowStr(true), msg)
}

func (m *stream) resolveServerPacket(p *packet) {

	var msg string
	switch p.packetType {
	case 4:
		var b = int32(p.payload[0])
		msg = fmt.Sprintf("【OK】%d", b)

	}

	// parseToken(p.payload)
	fmt.Println(GetNowStr(false), msg)
}
