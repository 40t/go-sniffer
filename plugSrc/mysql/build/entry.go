package build

import (
	"github.com/google/gopacket"
	"io"
	"bytes"
	"errors"
	"log"
	"strconv"
	"sync"
	"time"
	"fmt"
	"encoding/binary"
	"strings"
	"os"
)

const (
	Port              = 3306
	Version           = "0.1"
	CmdPort           = "-p"
)

type Mysql struct {
	port       int
	version    string
	source     map[string]*stream
}

type stream struct {
	packets chan *packet
	stmtMap map[uint32]*Stmt
}

type packet struct {
	isClientFlow bool
	seq        int
	length     int
	payload   []byte
}

var mysql *Mysql
var once sync.Once
func NewInstance() *Mysql {

	once.Do(func() {
		mysql = &Mysql{
			port   :Port,
			version:Version,
			source: make(map[string]*stream),
		}
	})

	return mysql
}

func (m *Mysql) ResolveStream(net, transport gopacket.Flow, buf io.Reader) {

	//uuid
	uuid := fmt.Sprintf("%v:%v", net.FastHash(), transport.FastHash())

	//generate resolve's stream
	if _, ok := m.source[uuid]; !ok {

		var newStream = stream{
			packets:make(chan *packet, 100),
			stmtMap:make(map[uint32]*Stmt),
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

func (m *Mysql) BPFFilter() string {
	return "tcp and port "+strconv.Itoa(m.port);
}

func (m *Mysql) Version() string {
	return Version
}

func (m *Mysql) SetFlag(flg []string)  {

	c := len(flg)

	if c == 0 {
		return
	}
	if c >> 1 == 0 {
		fmt.Println("ERR : Mysql Number of parameters")
		os.Exit(1)
	}
	for i:=0;i<c;i=i+2 {
		key := flg[i]
		val := flg[i+1]

		switch key {
		case CmdPort:
			port, err := strconv.Atoi(val);
			m.port = port
			if err != nil {
				panic("ERR : port")
			}
			if port < 0 || port > 65535 {
				panic("ERR : port(0-65535)")
			}
			break
		default:
			panic("ERR : mysql's params")
		}
	}
}

func (m *Mysql) newPacket(net, transport gopacket.Flow, r io.Reader) *packet {

	//read packet
	var payload bytes.Buffer
	var seq uint8
	var err error
	if seq, err = m.resolvePacketTo(r, &payload); err != nil {
		return nil
	}

	//close stream
	if err == io.EOF {
		fmt.Println(net, transport, " close")
		return nil
	} else if err != nil {
		fmt.Println("ERR : Unknown stream", net, transport, ":", err)
	}

	//generate new packet
	var pk = packet{
		seq: int(seq),
		length:payload.Len(),
		payload:payload.Bytes(),
	}
	if transport.Src().String() == strconv.Itoa(Port) {
		pk.isClientFlow = false
	}else{
		pk.isClientFlow = true
	}

	return &pk
}

func (m *Mysql) resolvePacketTo(r io.Reader, w io.Writer) (uint8, error) {

	header := make([]byte, 4)
	if n, err := io.ReadFull(r, header); err != nil {
		if n == 0 && err == io.EOF {
			return 0, io.EOF
		}
		return 0, errors.New("ERR : Unknown stream")
	}

	length := int(uint32(header[0]) | uint32(header[1])<<8 | uint32(header[2])<<16)

	var seq uint8
	seq = header[3]

	if n, err := io.CopyN(w, r, int64(length)); err != nil {
		return 0, errors.New("ERR : Unknown stream")
	} else if n != int64(length) {
		return 0, errors.New("ERR : Unknown stream")
	} else {
		return seq, nil
	}

	return seq, nil
}

func (stm *stream) resolve() {
	for {
		select {
		case packet := <- stm.packets:
			if packet.length != 0 {
				if packet.isClientFlow {
					stm.resolveClientPacket(packet.payload, packet.seq)
				} else {
					stm.resolveServerPacket(packet.payload, packet.seq)
				}
			}
		}
	}
}

func (stm *stream) findStmtPacket (srv chan *packet, seq int) *packet {
	for {
		select {
		case packet, ok := <- stm.packets:
			if !ok {
				return nil
			}
			if packet.seq == seq {
				return packet
			}
		case <-time.After(5 * time.Second):
			return nil
		}
	}
}

func (stm *stream) resolveServerPacket(payload []byte, seq int) {

	var msg = ""
	if len(payload) == 0 {
		return
	}
	switch payload[0] {

		case 0xff:
			errorCode  := int(binary.LittleEndian.Uint16(payload[1:3]))
			errorMsg,_ := ReadStringFromByte(payload[4:])

			msg = GetNowStr(false)+"%s Err code:%s,Err msg:%s"
			msg  = fmt.Sprintf(msg, ErrorPacket, strconv.Itoa(errorCode), strings.TrimSpace(errorMsg))

		case 0x00:
			var pos = 1
			l,_ := LengthBinary(payload[pos:])
			affectedRows := int(l)

			msg += GetNowStr(false)+"%s Effect Row:%s"
			msg = fmt.Sprintf(msg, OkPacket, strconv.Itoa(affectedRows))

		default:
			return
	}

	fmt.Println(msg)
}

func (stm *stream) resolveClientPacket(payload []byte, seq int) {

	var msg string
	switch payload[0] {

	case COM_INIT_DB:

		msg = fmt.Sprintf("USE %s;\n", payload[1:])
	case COM_DROP_DB:

		msg = fmt.Sprintf("Drop DB %s;\n", payload[1:])
	case COM_CREATE_DB, COM_QUERY:

		statement := string(payload[1:])
		msg = fmt.Sprintf("%s %s", ComQueryRequestPacket, statement)
	case COM_STMT_PREPARE:

		serverPacket := stm.findStmtPacket(stm.packets, seq+1)
		if serverPacket == nil {
			log.Println("ERR : Not found stm packet")
			return
		}

		//fetch stm id
		stmtID := binary.LittleEndian.Uint32(serverPacket.payload[1:5])
		stmt := &Stmt{
			ID:    stmtID,
			Query: string(payload[1:]),
		}

		//record stm sql
		stm.stmtMap[stmtID] = stmt
		stmt.FieldCount = binary.LittleEndian.Uint16(serverPacket.payload[5:7])
		stmt.ParamCount = binary.LittleEndian.Uint16(serverPacket.payload[7:9])
		stmt.Args       = make([]interface{}, stmt.ParamCount)

		msg = PreparePacket+stmt.Query
	case COM_STMT_SEND_LONG_DATA:

		stmtID   := binary.LittleEndian.Uint32(payload[1:5])
		paramId  := binary.LittleEndian.Uint16(payload[5:7])
		stmt, _  := stm.stmtMap[stmtID]

		if stmt.Args[paramId] == nil {
			stmt.Args[paramId] = payload[7:]
		} else {
			if b, ok := stmt.Args[paramId].([]byte); ok {
				b = append(b, payload[7:]...)
				stmt.Args[paramId] = b
			}
		}
		return
	case COM_STMT_RESET:

		stmtID := binary.LittleEndian.Uint32(payload[1:5])
		stmt, _:= stm.stmtMap[stmtID]
		stmt.Args = make([]interface{}, stmt.ParamCount)
		return
	case COM_STMT_EXECUTE:

		var pos = 1
		stmtID := binary.LittleEndian.Uint32(payload[pos : pos+4])
		pos += 4
		var stmt *Stmt
		var ok bool
		if stmt, ok = stm.stmtMap[stmtID]; ok == false {
			log.Println("ERR : Not found stm id", stmtID)
			return
		}

		//params
		pos += 5
		if stmt.ParamCount > 0 {

			//（Null-Bitmap，len = (paramsCount + 7) / 8 byte）
			step := int((stmt.ParamCount + 7) / 8)
			nullBitmap := payload[pos : pos+step]
			pos += step

			//Parameter separator
			flag := payload[pos]

			pos++

			var pTypes  []byte
			var pValues []byte

			//if flag == 1
			//n （len = paramsCount * 2 byte）
			if flag == 1 {
				pTypes = payload[pos : pos+int(stmt.ParamCount)*2]
				pos += int(stmt.ParamCount) * 2
				pValues = payload[pos:]
			}

			//bind params
			err := stmt.BindArgs(nullBitmap, pTypes, pValues)
			if err != nil {
				log.Println("ERR : Could not bind params", err)
			}
		}
		msg = string(stmt.WriteToText())
	default:
		return
	}

	fmt.Println(GetNowStr(true) + msg)
}

