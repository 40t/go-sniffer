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

const _PLP_NULL = 0xFFFFFFFFFFFFFFFF
const _UNKNOWN_PLP_LEN = 0xFFFFFFFFFFFFFFFE
const _PLP_TERMINATOR = 0x00000000

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
	Reader   func(column *columnStruct, buf []byte) int
}

func readTypeInfo(pos int, buf []byte, column *columnStruct) (count int) {
	typeId := buf[pos]

	count = 1
	pos++
	column.TypeId = int(typeId)
	// fmt.Printf("column TypeId %x %x\n", column.TypeId, buf)

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

		column.Reader = readFixedType
		// those are fixed length types
	default: // all others are VARLENTYPE
		count += readVarLen(int(typeId), pos, buf, column)
	}
	return count
}

func readFixedType(column *columnStruct, buf []byte) int {
	return column.Size
}

func readByteLenType(column *columnStruct, buf []byte) int {
	size := int(buf[0])
	return 1 + size
}

// partially length prefixed stream
// http://msdn.microsoft.com/en-us/library/dd340469.aspx
func readPLPType(column *columnStruct, buf []byte) int {
	size := binary.LittleEndian.Uint64(buf[0:8])
	valueLength := 0
	switch size {
	case _PLP_NULL:
		valueLength = 0
	case _UNKNOWN_PLP_LEN:
		valueLength = 1000
	default:
		valueLength = int(size)
	}
	return valueLength + 8
}

func readShortLenType(column *columnStruct, buf []byte) int {
	size := int(binary.LittleEndian.Uint16(buf[0:2]))

	return 2 + size
}

func readLongLenType(column *columnStruct, buf []byte) int {
	count := 1
	textptrsize := int(buf[0]) //textptrsize
	if textptrsize == 0 {
		return 1
	}

	count = textptrsize + 8 + 1
	//timestamp 8

	size := int(binary.LittleEndian.Uint32(buf[count : count+4]))
	if size == -1 {
		return count + 4
	}
	return count + 4 + size

}

// reads variant value
// http://msdn.microsoft.com/en-us/library/dd303302.aspx
func readVariantType(column *columnStruct, buf []byte) int {
	count := 0
	size := int(binary.LittleEndian.Uint32(buf[count : count+4]))
	count = 4
	if size == 0 {
		return count
	}
	vartype := int(buf[count])
	count += 1
	propbytes := int(buf[count])
	switch vartype {
	case typeGuid:

		count = count + size - 2 - propbytes
	case typeBit:
		count += 1
	case typeInt1:
		count += 1
	case typeInt2:
		count += 2
	case typeInt4:
		count += 4
	case typeInt8:
		count += 8
	case typeDateTime:
		count = count + size - 2 - propbytes

	case typeDateTim4:
		count = count + size - 2 - propbytes

	case typeFlt4:
		count = count + 4
	case typeFlt8:
		count = count + 8
	case typeMoney4:
		count = count + size - 2 - propbytes

	case typeMoney:
		count = count + size - 2 - propbytes

	case typeDateN:
		count = count + size - 2 - propbytes

	case typeTimeN:
		count += 1
		count = count + size - 2 - propbytes
	case typeDateTime2N:
		count += 1
		count = count + size - 2 - propbytes
	case typeDateTimeOffsetN:
		count += 1
		count = count + size - 2 - propbytes
	case typeBigVarBin, typeBigBinary:
		count += 2
		count = count + size - 2 - propbytes
	case typeDecimalN, typeNumericN:
		count += 2
		count = count + size - 2 - propbytes
	case typeBigVarChar, typeBigChar:
		count += 5
		count += 2 // max length, ignoring
		count = count + size - 2 - propbytes

	case typeNVarChar, typeNChar:
		count += 5
		count += 2 // max length, ignoring
		count = count + size - 2 - propbytes
	default:
		panic("Invalid variant typeid")
	}
	return count
}

func readVarLen(typeId int, pos int, buf []byte, column *columnStruct) (count int) {
	count = 0
	switch typeId {
	case typeDateN:
		column.Size = 3
		column.Reader = readByteLenType
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

		column.Reader = readByteLenType

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
		column.Reader = readByteLenType
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
		column.Reader = readPLPType
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

		column.Reader = readPLPType
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

		if column.Size == 0xffff {
			column.Reader = readPLPType
		} else {
			column.Reader = readShortLenType
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
			column.Reader = readLongLenType
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
			column.Reader = readLongLenType

		case typeVariant:
			column.Reader = readVariantType

		}
	default:
		count += 0
	}
	return count
}

func parseToken(buf []byte) (rowCount int, msg string) {

	var pos = 0
	length := 0
	rowCount = 0
	msg = ""

	var columns []columnStruct

	defer func() {
		if r := recover(); r != nil {
			msg = "parse tds result error"
		}
	}()

	currentBuf := buf[:]

	// fmt.Printf("buf len %x", currentBuf)
	for {

		if len(currentBuf) == 0 {
			break
		}

		token := token(currentBuf[0])
		// fmt.Printf("item... %x %d\n", currentBuf[0], currentBuf[0])
		currentBuf = currentBuf[1:]

		switch token {
		case tokenSSPI:
			length = int(binary.LittleEndian.Uint16(currentBuf[0:2]))
			currentBuf = currentBuf[length+2:]
		case tokenReturnStatus:
			currentBuf = currentBuf[3:]

		case tokenLoginAck:
			length = int(binary.LittleEndian.Uint16(currentBuf[0:2]))
			currentBuf = currentBuf[2+length:]

		case tokenOrder:
			length = int(binary.LittleEndian.Uint16(currentBuf[0:2]))
			currentBuf = currentBuf[2+length:]

		case tokenDoneInProc:
			currentBuf = currentBuf[4:]
			rowCount = int(binary.LittleEndian.Uint64(currentBuf[0:8]))
			currentBuf = currentBuf[8:]
		case tokenDone, tokenDoneProc:
			currentBuf = currentBuf[4:]
			rowCount = int(binary.LittleEndian.Uint64(currentBuf[0:8]))
			currentBuf = currentBuf[8:]

		case tokenError:
			currentBuf = currentBuf[8:] //length2+Number4+State1+Class1
			//message length
			msgLength := int(binary.LittleEndian.Uint16(currentBuf[0:2]))
			currentBuf = currentBuf[2:]
			msgLength = msgLength * 2

			msg, _ = ucs22str(currentBuf[0:msgLength])
			return
		case tokenColMetadata:

			//http://msdn.microsoft.com/en-us/library/dd357363.aspx
			count := int(binary.LittleEndian.Uint16(currentBuf[0:2]))
			currentBuf = currentBuf[2:]

			if count == 0xffff {
				break
			}
			columns = make([]columnStruct, count)

			if count > 0 {
				for i := 0; i < count; i++ {

					// fmt.Printf("colums %d %d", i, count)
					column := &columns[i]
					// x := pos

					currentBuf = currentBuf[6:]

					pos = readTypeInfo(0, currentBuf, column)

					currentBuf = currentBuf[pos:]

					//ColName
					l := int(currentBuf[0])

					currentBuf = currentBuf[1:]
					//name
					currentBuf = currentBuf[l*2:]
				}

				// fmt.Printf("tokenRow %x\n", currentBuf)

			}
		case tokenRow:
			count := 0

			for _, column := range columns {

				count = column.Reader(&column, currentBuf)
				currentBuf = currentBuf[count:]

			}
		case tokenNbcRow:
			bitlen := (len(columns) + 7) / 8

			pres := currentBuf[0:bitlen]
			currentBuf = currentBuf[bitlen:]
			count := 0

			for i, column := range columns {
				if pres[i/8]&(1<<(uint(i)%8)) != 0 {
					continue
				}
				count = column.Reader(&column, currentBuf)
				currentBuf = currentBuf[count:]

				// fmt.Printf("tokenNbcRow %d %x \n", i, currentBuf)

			}
		case tokenEnvChange:
			// http://msdn.microsoft.com/en-us/library/dd303449.aspx
			length = int(binary.LittleEndian.Uint16(currentBuf[0:2]))
			currentBuf = currentBuf[2+length:]
		case tokenInfo:
			// http://msdn.microsoft.com/en-us/library/dd304156.aspx
			length = int(binary.LittleEndian.Uint16(currentBuf[0:2]))
			currentBuf = currentBuf[2+length:]
		case tokenReturnValue:
			// https://msdn.microsoft.com/en-us/library/dd303881.aspx
			currentBuf = currentBuf[2:]
			nameLength := int(currentBuf[0])
			currentBuf = currentBuf[1:]
			currentBuf = currentBuf[nameLength*2:]
			currentBuf = currentBuf[7:] //1byte + 4 byte+2 byt
			col := columnStruct{}
			count := readTypeInfo(0, currentBuf, &col)
			currentBuf = currentBuf[count:] //readTypeInfo

			count = col.Reader(&col, currentBuf)
			currentBuf = currentBuf[count:] //column value

		default:
			// fmt.Printf("tokenNbcRow %x \n", currentBuf[0])
			return rowCount, "parse result error"
		}

	}
	return rowCount, msg
}
