package build

import (
	"encoding/binary"
	"io"
	"time"
)

func GetNowStr(isClient bool) string {
	var msg string
	layout := "01/02 15:04:05.000000"
	msg += time.Now().Format(layout)
	if isClient {
		msg += "| cli -> ser |"
	}else{
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

func ReadInt16(r io.Reader) (n int16) {
	binary.Read(r, binary.BigEndian, &n)
	return
}

func ReadInt32(r io.Reader) (n int32) {
	binary.Read(r, binary.BigEndian, &n)
	return
}

func ReadInt64(r io.Reader) (n int64) {
	binary.Read(r, binary.BigEndian, &n)
	return
}

func ReadString(r io.Reader) (string, int) {

	l := int(ReadInt16(r))

	//-1 => null
	if l == -1 {
		return " ",1
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

	var result []byte
	var b = make([]byte, l)
	for i:=0;i<l;i++ {

		_, err := r.Read(b)

		if err != nil {
			panic(err)
		}

		result = append(result, b[0])
	}

	return result
}
