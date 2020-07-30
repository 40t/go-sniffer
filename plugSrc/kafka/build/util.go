package build

import (
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
	"time"
)

func GetNowStr(isClient bool) string {
	var msg string
	layout := "01/02 15:04:05.000000"
	msg += time.Now().Format(layout)
	if isClient {
		msg += "| cli -> ser |"
	} else {
		msg += "| ser -> cli |"
	}
	return msg
}

func IsEof(r io.Reader) bool {
	buf := make([]byte, 1)
	_, err := r.Read(buf)
	if err != nil {
		return true
	}
	return false
}

func ReadOnce() {

}

func ReadByte(r io.Reader) (n byte) {
	binary.Read(r, binary.BigEndian, &n)
	return
}

func ReadInt16(r io.Reader) (n int16) {
	binary.Read(r, binary.BigEndian, &n)
	return
}

func ReadInt32(r io.Reader) (n int32) {
	binary.Read(r, binary.BigEndian, &n)
	return
}

func ReadUint32(r io.Reader) (n uint32) {
	binary.Read(r, binary.BigEndian, &n)
	return
}

func ReadInt64(r io.Reader) (err error, n int64) {
	err = binary.Read(r, binary.BigEndian, &n)
	return
}

func ReadString(r io.Reader) (string, int) {

	l := int(ReadInt16(r))

	//-1 => null
	if l == -1 {
		return " ", 1
	}

	str := make([]byte, l)
	if _, err := io.ReadFull(r, str); err != nil {
		panic(err)
	}

	return string(str), l
}

//
//func TryReadInt16(r io.Reader) (n int16, err error) {
//
//	if err := binary.Read(r, binary.BigEndian, &n); err != nil {
//		if n == -1 {
//			return 1,nil
//		}
//		panic(err)
//	}
//}

func ReadBytes(r io.Reader) []byte {

	l := int(ReadInt32(r))
	result := make([]byte, 0)

	if l <= 0 {
		return result
	}

	var b = make([]byte, l)
	for i := 0; i < l; i++ {

		_, err := r.Read(b)

		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}

		result = append(result, b...)
	}

	return result
}

func ReadMessages(r io.Reader, version int16) []*Message {
	switch version {
	case 0:
		return ReadMessagesV1(r)
	case 1:
		return ReadMessagesV1(r)
	}

	return make([]*Message, 0)
}

func ReadMessagesV1(r io.Reader) []*Message {
	var err error
	messages := make([]*Message, 0)
	for {
		message := Message{}
		err, message.Offset = ReadInt64(r)
		if err != nil {
			if err == io.EOF {
				break
			}
			if err != io.ErrUnexpectedEOF {
				fmt.Printf("read message offset , err: %+v\n", err)
			}
			break
		}
		_ = ReadInt32(r) // message size
		message.Crc = ReadUint32(r)
		message.Magic = ReadByte(r)
		message.CompressCode = ReadByte(r)
		message.Key = ReadBytes(r)
		message.Value = ReadBytes(r)
		messages = append(messages, &message)
	}
	return messages
}

func GetRquestName(apiKey int16) string {
	if name, ok := RquestNameMap[apiKey]; ok {
		return name
	}
	return strconv.Itoa(int(apiKey))
}
