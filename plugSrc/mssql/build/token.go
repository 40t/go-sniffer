package build

import (
	"encoding/binary"
	"fmt"
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

// fixed-length data types
// http://msdn.microsoft.com/en-us/library/dd341171.aspx
const (
	typeNull     = 0x1f
	typeInt1     = 0x30
	typeBit      = 0x32
	typeInt2     = 0x34
	typeInt4     = 0x38
	typeDateTim4 = 0x3a
	typeFlt4     = 0x3b
	typeMoney    = 0x3c
	typeDateTime = 0x3d
	typeFlt8     = 0x3e
	typeMoney4   = 0x7a
	typeInt8     = 0x7f
)

// variable-length data types
// http://msdn.microsoft.com/en-us/library/dd358341.aspx
const (
	// byte len types
	typeGuid            = 0x24
	typeIntN            = 0x26
	typeDecimal         = 0x37 // legacy
	typeNumeric         = 0x3f // legacy
	typeBitN            = 0x68
	typeDecimalN        = 0x6a
	typeNumericN        = 0x6c
	typeFltN            = 0x6d
	typeMoneyN          = 0x6e
	typeDateTimeN       = 0x6f
	typeDateN           = 0x28
	typeTimeN           = 0x29
	typeDateTime2N      = 0x2a
	typeDateTimeOffsetN = 0x2b
	typeChar            = 0x2f // legacy
	typeVarChar         = 0x27 // legacy
	typeBinary          = 0x2d // legacy
	typeVarBinary       = 0x25 // legacy

	// short length types
	typeBigVarBin  = 0xa5
	typeBigVarChar = 0xa7
	typeBigBinary  = 0xad
	typeBigChar    = 0xaf
	typeNVarChar   = 0xe7
	typeNChar      = 0xef
	typeXml        = 0xf1
	typeUdt        = 0xf0
	typeTvp        = 0xf3

	// long length types
	typeText    = 0x23
	typeImage   = 0x22
	typeNText   = 0x63
	typeVariant = 0x62
)

type columnStruct struct {
	UserType uint32
	Flags    uint16
	ColName  string
	Size     int
	TypeId   int
}

func readTypeInfo(pos int, buf []byte, column *columnStruct) (count int) {
	typeId := buf[pos+1]

	count = 1
	pos++
	column.TypeId = int(typeId)
	switch typeId {
	case typeNull, typeInt1, typeBit, typeInt2, typeInt4, typeDateTim4,
		typeFlt4, typeMoney, typeDateTime, typeFlt8, typeMoney4, typeInt8:
		count += 0
		switch typeId {
		case typeNull:
			column.Size = 0
		case typeInt1, typeBit:
			column.Size = 1
		case typeInt2:
			column.Size = 2
		case typeInt4, typeDateTim4, typeFlt4, typeMoney4:
			column.Size = 4
		case typeMoney, typeDateTime, typeFlt8, typeInt8:
			column.Size = 8
		}
		// those are fixed length types
	default: // all others are VARLENTYPE
		count += readVarLen(int(typeId), pos, buf, column)
	}
	return count
}

func readVarLen(typeId int, pos int, buf []byte, column *columnStruct) (count int) {
	count = 0
	switch typeId {
	case typeDateN:
		column.Size = 3
	case typeTimeN, typeDateTime2N, typeDateTimeOffsetN:

		pos += 1 //Scale
		count += 1

		scale := buf[pos]

		switch scale {
		case 0, 1, 2:
			column.Size = 3
		case 3, 4:
			column.Size = 4
		case 5, 6, 7:
			column.Size = 5
		}

		switch typeId {
		case typeDateTime2N:
			column.Size += 3
		case typeDateTimeOffsetN:
			column.Size += 5
		}

	case typeGuid, typeIntN, typeDecimal, typeNumeric,
		typeBitN, typeDecimalN, typeNumericN, typeFltN,
		typeMoneyN, typeDateTimeN, typeChar,
		typeVarChar, typeBinary, typeVarBinary:
		// byle len types

		pos += 1 //byle len types
		count += 1

		column.Size = int(buf[pos]) //size
		switch typeId {
		case typeDecimal, typeNumeric, typeDecimalN, typeNumericN:
			pos += 2 //byle len types
			count += 2
		}
	case typeXml:
		pos += 1 //byle len types
		count += 1
		schemaPresent := buf[pos]

		if schemaPresent != 0 {
			pos += 1 //byle len types
			count += 1
			l := int(buf[pos]) // dbname
			count += l
			pos += l

			pos += 1 // owning schema
			count += 1
			l = int(buf[pos]) // owning schema
			count += l
			pos += l

			// xml schema collection
			pos += 1
			l = int(binary.LittleEndian.Uint16(buf[pos : pos+2]))
			pos += 1
			count += 2
			pos += l * 2
			count += l * 2
		}
	case typeUdt:
		pos += 1
		l := int(binary.LittleEndian.Uint16(buf[pos : pos+2]))
		pos += 1
		count += 2
		//ti.Size
		column.Size = l

		//DBName
		pos += 1 // owning schema
		count += 1
		l = int(buf[pos])
		count += l
		pos += l

		//SchemaName
		pos += 1 // owning schema
		count += 1
		l = int(buf[pos])
		count += l
		pos += l

		//TypeName
		pos += 1 // owning schema
		count += 1
		l = int(buf[pos])
		count += l
		pos += l

		//AssemblyQualifiedName
		pos += 1
		l = int(binary.LittleEndian.Uint16(buf[pos : pos+2]))
		pos += 1
		count += 2
		pos += l * 2
		count += l * 2

	case typeBigVarBin, typeBigVarChar, typeBigBinary, typeBigChar,
		typeNVarChar, typeNChar:
		// short len types
		pos += 1
		l := int(binary.LittleEndian.Uint16(buf[pos : pos+2]))
		pos += 1
		count += 2

		column.Size = l

		switch typeId {
		case typeBigVarChar, typeBigChar, typeNVarChar, typeNChar:
			pos += 5
			count += 5
		}

	case typeText, typeImage, typeNText, typeVariant:
		// LONGLEN_TYPE

		l := int(binary.LittleEndian.Uint16(buf[pos+1 : pos+5]))
		column.Size = l

		pos += 4
		count += 4

		switch typeId {
		case typeText, typeNText:
			pos += 6
			count += 6
			// ignore tablenames
			numparts := int(buf[pos])
			for i := 0; i < numparts; i++ {
				pos += 1
				l := int(binary.LittleEndian.Uint16(buf[pos : pos+2]))
				pos += 1
				count += 2
				pos += l
				count += l
			}

		case typeImage:
			// ignore tablenames
			pos++
			count++
			numparts := int(buf[pos])
			for i := 0; i < numparts; i++ {
				pos += 1
				l := int(binary.LittleEndian.Uint16(buf[pos : pos+2]))
				pos += 1
				count += 2
				pos += l
				count += l
			}
		}
	default:
		count += 0
	}
	return count
}

func parseToken(buf []byte) int {

	var pos = 0
	length := 0
	rowCount := 0

	var columns []columnStruct

	for {

		token := token(buf[pos])

		switch token {
		case tokenSSPI:
			pos += 1
			length = int(binary.LittleEndian.Uint16(buf[pos : pos+2]))
			pos += 1
			pos += length
		case tokenReturnStatus:
			pos += 4
		case tokenLoginAck:
			pos += 1
			length = int(binary.LittleEndian.Uint16(buf[pos : pos+2]))
			pos += 1
			pos += length
		case tokenOrder:
			pos += 1
			length = int(binary.LittleEndian.Uint16(buf[pos : pos+2]))
			pos += 1
			pos += length
		case tokenDoneInProc:
			pos += 5
			rowCount = int(binary.LittleEndian.Uint64(buf[pos : pos+8]))
			pos += 8
			pos += rowCount
		case tokenDone, tokenDoneProc:
			pos += 5
			rowCount = int(binary.LittleEndian.Uint64(buf[pos : pos+8]))
			pos += 8
			pos += rowCount
		case tokenColMetadata:
			pos += 1
			//http://msdn.microsoft.com/en-us/library/dd357363.aspx
			count := int(binary.LittleEndian.Uint16(buf[pos : pos+2]))
			columns = make([]columnStruct, count)
			pos += 1

			if count > 0 {
				for i := 0; i < count; i++ {

					// fmt.Printf("colums %d %d", i, count)
					column := &columns[i]
					// x := pos
					pos += 4 //UserType
					pos += 2 //Flags
					pos += readTypeInfo(pos, buf, column)

					//ColName
					pos += 1 // owning schema
					l := int(buf[pos])
					pos += l * 2

					// fmt.Printf("%d, %d ,%x\n", x, pos, buf[x+1:pos+1])

					fmt.Print("%v", column)
				}
			}
		case tokenRow:

		}
		break

	}
	return rowCount
}
