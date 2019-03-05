package build

import (
	"encoding/binary"
)

type token byte

// token ids
const (
	tokenReturnStatus token = 121 // 0x79
	tokenColMetadata  token = 129 // 0x81
	tokenOrder        token = 169 // 0xA9
	tokenError        token = 170 // 0xAA
	tokenInfo         token = 171 // 0xAB
	tokenReturnValue  token = 0xAC
	tokenLoginAck     token = 173 // 0xad
	tokenRow          token = 209 // 0xd1
	tokenNbcRow       token = 210 // 0xd2
	tokenEnvChange    token = 227 // 0xE3
	tokenSSPI         token = 237 // 0xED
	tokenDone         token = 253 // 0xFD
	tokenDoneProc     token = 254
	tokenDoneInProc   token = 255
)

func parseToken(buf []byte) string {

	var pos = 0
	length := 0
	for {
		if len(buf) < pos+1 {
			break
		}
		token := token(buf[pos])
		switch token {
		case tokenSSPI:
			pos += 1
			length = int(binary.LittleEndian.Uint16(buf[pos+1 : pos+2]))
			pos += length

		}
		break

	}
	return ""
}
