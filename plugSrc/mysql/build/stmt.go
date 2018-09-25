package build

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"strings"

	"errors"
)

type Stmt struct {
	ID         uint32
	Query      string
	ParamCount uint16
	FieldCount uint16

	Args []interface{}
}

func (stmt *Stmt) WriteToText() []byte {

	var buf bytes.Buffer

	str := fmt.Sprintf("预处理编号[%d]: '%s';\n", stmt.ID, stmt.Query)
	buf.WriteString(str)

	for i := 0; i < int(stmt.ParamCount); i++ {
		var str string
		switch stmt.Args[i].(type) {
		case nil:
			str = fmt.Sprintf("set @p%v = NULL;\n", i)
		case []byte:
			param := string(stmt.Args[i].([]byte))
			str = fmt.Sprintf("set @p%v = '%s';\n", i, strings.TrimSpace(param))
		default:
			str = fmt.Sprintf("set @p%v = %v;\n", i, stmt.Args[i])
		}
		buf.WriteString(str)
	}

	str = fmt.Sprintf("执行预处理[%d]: ", stmt.ID)
	buf.WriteString(str)
	for i := 0; i < int(stmt.ParamCount); i++ {
		if i == 0 {
			buf.WriteString(" using ")
		}
		if i > 0 {
			buf.WriteString(", ")
		}
		str := fmt.Sprintf("@p%v", i)
		buf.WriteString(str)
	}
	buf.WriteString(";\n")

	str = fmt.Sprintf("丢弃预处理[%d];\n", stmt.ID)
	buf.WriteString(str)

	return buf.Bytes()
}

func (stmt *Stmt) BindArgs(nullBitmap, paramTypes, paramValues []byte) error {

	args := stmt.Args
	pos := 0

	var v []byte
	var n = 0
	var isNull bool
	var err error

	for i := 0; i < int(stmt.ParamCount); i++ {

		//判断参数是否为null
		if nullBitmap[i>>3]&(1<<(uint(i)%8)) > 0 {
			args[i] = nil
			continue
		}

		//参数类型
		typ := paramTypes[i<<1]
		unsigned := (paramTypes[(i<<1)+1] & 0x80) > 0

		switch typ {
		case MYSQL_TYPE_NULL:
			args[i] = nil
			continue

		case MYSQL_TYPE_TINY:

			value := paramValues[pos]
			if unsigned {
				args[i] = uint8(value)
			} else {
				args[i] = int8(value)
			}

			pos++
			continue

		case MYSQL_TYPE_SHORT, MYSQL_TYPE_YEAR:

			value := binary.LittleEndian.Uint16(paramValues[pos : pos+2])
			if unsigned {
				args[i] = uint16(value)
			} else {
				args[i] = int16(value)
			}
			pos += 2
			continue

		case MYSQL_TYPE_INT24, MYSQL_TYPE_LONG:

			value := binary.LittleEndian.Uint32(paramValues[pos : pos+4])
			if unsigned {
				args[i] = uint32(value)
			} else {
				args[i] = int32(value)
			}
			pos += 4
			continue

		case MYSQL_TYPE_LONGLONG:

			value := binary.LittleEndian.Uint64(paramValues[pos : pos+8])
			if unsigned {
				args[i] = value
			} else {
				args[i] = int64(value)
			}
			pos += 8
			continue

		case MYSQL_TYPE_FLOAT:

			value := math.Float32frombits(binary.LittleEndian.Uint32(paramValues[pos : pos+4]))
			args[i] = float32(value)
			pos += 4
			continue

		case MYSQL_TYPE_DOUBLE:

			value := math.Float64frombits(binary.LittleEndian.Uint64(paramValues[pos : pos+8]))
			args[i] = value
			pos += 8
			continue

		case MYSQL_TYPE_DECIMAL, MYSQL_TYPE_NEWDECIMAL,
			MYSQL_TYPE_VARCHAR, MYSQL_TYPE_BIT,
			MYSQL_TYPE_ENUM, MYSQL_TYPE_SET,
			MYSQL_TYPE_TINY_BLOB, MYSQL_TYPE_MEDIUM_BLOB, MYSQL_TYPE_LONG_BLOB, MYSQL_TYPE_BLOB,
			MYSQL_TYPE_VAR_STRING, MYSQL_TYPE_STRING,
			MYSQL_TYPE_GEOMETRY,
			MYSQL_TYPE_DATE, MYSQL_TYPE_NEWDATE, MYSQL_TYPE_TIMESTAMP, MYSQL_TYPE_DATETIME, MYSQL_TYPE_TIME:

			v, isNull, n, err = LengthEncodedString(paramValues[pos:])
			pos += n
			if err != nil {
				return err
			}

			if !isNull {
				args[i] = v
				continue
			} else {
				args[i] = nil
				continue
			}
		default:
			return errors.New(fmt.Sprintf("预处理未知类型 %d", typ))
		}
	}
	return nil
}
