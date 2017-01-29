package pgmsg

import (
	"bytes"
	"encoding/binary"
	"strconv"
)

type ErrorResponse struct {
	Severity         string
	Code             string
	Message          string
	Detail           string
	Hint             string
	Position         int32
	InternalPosition int32
	InternalQuery    string
	Where            string
	SchemaName       string
	TableName        string
	ColumnName       string
	DataTypeName     string
	ConstraintName   string
	File             string
	Line             int32
	Routine          string

	UnknownFields map[byte]string
}

func (*ErrorResponse) Backend() {}

func ParseErrorResponse(rawBuf []byte) (*ErrorResponse, error) {
	var errResp ErrorResponse

	buf := bytes.NewBuffer(rawBuf)

	for {
		k, err := buf.ReadByte()
		if err != nil {
			return nil, err
		}
		if k == 0 {
			break
		}

		vb, err := buf.ReadBytes(0)
		if err != nil {
			return nil, err
		}
		v := string(vb[:len(vb)-1])

		switch k {
		case 'S':
			errResp.Severity = v
		case 'C':
			errResp.Code = v
		case 'M':
			errResp.Message = v
		case 'D':
			errResp.Detail = v
		case 'H':
			errResp.Hint = v
		case 'P':
			s := v
			n, _ := strconv.ParseInt(s, 10, 32)
			errResp.Position = int32(n)
		case 'p':
			s := v
			n, _ := strconv.ParseInt(s, 10, 32)
			errResp.InternalPosition = int32(n)
		case 'q':
			errResp.InternalQuery = v
		case 'W':
			errResp.Where = v
		case 's':
			errResp.SchemaName = v
		case 't':
			errResp.TableName = v
		case 'c':
			errResp.ColumnName = v
		case 'd':
			errResp.DataTypeName = v
		case 'n':
			errResp.ConstraintName = v
		case 'F':
			errResp.File = v
		case 'L':
			s := v
			n, _ := strconv.ParseInt(s, 10, 32)
			errResp.Line = int32(n)
		case 'R':
			errResp.Routine = v

		default:
			if errResp.UnknownFields == nil {
				errResp.UnknownFields = make(map[byte]string)
			}
			errResp.UnknownFields[k] = v
		}
	}

	return &errResp, nil
}

func (er *ErrorResponse) Encode() ([]byte, error) {
	var bigEndian BigEndianBuf
	buf := &bytes.Buffer{}

	buf.WriteByte('E')
	buf.Write(bigEndian.Uint32(0))

	if er.Severity != "" {
		buf.WriteString(er.Severity)
		buf.WriteByte(0)
	}
	if er.Code != "" {
		buf.WriteString(er.Code)
		buf.WriteByte(0)
	}
	if er.Message != "" {
		buf.WriteString(er.Message)
		buf.WriteByte(0)
	}
	if er.Detail != "" {
		buf.WriteString(er.Detail)
		buf.WriteByte(0)
	}
	if er.Hint != "" {
		buf.WriteString(er.Hint)
		buf.WriteByte(0)
	}
	if er.Position != 0 {
		buf.WriteString(strconv.Itoa(int(er.Position)))
		buf.WriteByte(0)
	}
	if er.InternalPosition != 0 {
		buf.WriteString(strconv.Itoa(int(er.InternalPosition)))
		buf.WriteByte(0)
	}
	if er.InternalQuery != "" {
		buf.WriteString(er.InternalQuery)
		buf.WriteByte(0)
	}
	if er.Where != "" {
		buf.WriteString(er.Where)
		buf.WriteByte(0)
	}
	if er.SchemaName != "" {
		buf.WriteString(er.SchemaName)
		buf.WriteByte(0)
	}
	if er.TableName != "" {
		buf.WriteString(er.TableName)
		buf.WriteByte(0)
	}
	if er.ColumnName != "" {
		buf.WriteString(er.ColumnName)
		buf.WriteByte(0)
	}
	if er.DataTypeName != "" {
		buf.WriteString(er.DataTypeName)
		buf.WriteByte(0)
	}
	if er.ConstraintName != "" {
		buf.WriteString(er.ConstraintName)
		buf.WriteByte(0)
	}
	if er.File != "" {
		buf.WriteString(er.File)
		buf.WriteByte(0)
	}
	if er.Line != 0 {
		buf.WriteString(strconv.Itoa(int(er.Line)))
		buf.WriteByte(0)
	}
	if er.Routine != "" {
		buf.WriteString(er.Routine)
		buf.WriteByte(0)
	}

	for k, v := range er.UnknownFields {
		buf.WriteByte(k)
		buf.WriteByte(0)
		buf.WriteString(v)
		buf.WriteByte(0)
	}
	buf.WriteByte(0)

	binary.BigEndian.PutUint32(buf.Bytes()[1:5], uint32(buf.Len()-1))

	return buf.Bytes(), nil
}
