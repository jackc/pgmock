package pgmsg

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
)

const (
	TextFormat   = 0
	BinaryFormat = 1
)

type FieldDescription struct {
	Name                 string
	TableOID             uint32
	TableAttributeNumber uint16
	DataTypeOID          uint32
	DataTypeSize         int16
	TypeModifier         uint32
	Format               int16
}

type RowDescription struct {
	Fields []FieldDescription
}

func (*RowDescription) Backend() {}

func ParseRowDescription(body []byte) (*RowDescription, error) {
	buf := bytes.NewBuffer(body)

	if buf.Len() < 2 {
		return nil, &invalidMessageFormatErr{messageType: "RowDescription"}
	}
	fieldCount := int(binary.BigEndian.Uint16(buf.Next(2)))

	rd := &RowDescription{Fields: make([]FieldDescription, fieldCount)}

	for i := 0; i < fieldCount; i++ {
		var fd FieldDescription
		bName, err := buf.ReadBytes(0)
		if err != nil {
			return nil, err
		}
		fd.Name = string(bName[:len(bName)-1])

		// Since buf.Next() doesn't return an error if we hit the end of the buffer
		// check Len ahead of time
		if buf.Len() < 18 {
			return nil, &invalidMessageFormatErr{messageType: "RowDescription"}
		}

		fd.TableOID = binary.BigEndian.Uint32(buf.Next(4))
		fd.TableAttributeNumber = binary.BigEndian.Uint16(buf.Next(2))
		fd.DataTypeOID = binary.BigEndian.Uint32(buf.Next(4))
		fd.DataTypeSize = int16(binary.BigEndian.Uint16(buf.Next(2)))
		fd.TypeModifier = binary.BigEndian.Uint32(buf.Next(4))
		fd.Format = int16(binary.BigEndian.Uint16(buf.Next(2)))

		rd.Fields[i] = fd
	}

	return rd, nil
}

func (rd *RowDescription) Encode() ([]byte, error) {
	var bigEndian BigEndianBuf
	buf := &bytes.Buffer{}

	buf.WriteByte('T')
	buf.Write(bigEndian.Uint32(0))

	buf.Write(bigEndian.Uint16(uint16(len(rd.Fields))))

	for _, fd := range rd.Fields {
		buf.WriteString(fd.Name)
		buf.WriteByte(0)

		buf.Write(bigEndian.Uint32(fd.TableOID))
		buf.Write(bigEndian.Uint16(fd.TableAttributeNumber))
		buf.Write(bigEndian.Uint32(fd.DataTypeOID))
		buf.Write(bigEndian.Uint16(uint16(fd.DataTypeSize)))
		buf.Write(bigEndian.Uint32(fd.TypeModifier))
		buf.Write(bigEndian.Uint16(uint16(fd.Format)))
	}

	binary.BigEndian.PutUint32(buf.Bytes()[1:5], uint32(buf.Len()-1))

	return buf.Bytes(), nil
}

func (rd *RowDescription) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type   string
		Fields []FieldDescription
	}{
		Type:   "RowDescription",
		Fields: rd.Fields,
	})
}
