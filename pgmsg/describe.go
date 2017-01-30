package pgmsg

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
)

type Describe struct {
	ObjectType byte // 'S' = prepared statement, 'P' = portal
	Name       string
}

func (*Describe) Frontend() {}

func ParseDescribe(body []byte) (*Describe, error) {
	var d Describe
	var err error

	buf := bytes.NewBuffer(body)

	d.ObjectType, err = buf.ReadByte()
	if err != nil {
		return nil, err
	}

	b, err := buf.ReadBytes(0)
	if err != nil {
		return nil, err
	}
	d.Name = string(b[:len(b)-1])

	return &d, nil
}

func (d *Describe) Encode() ([]byte, error) {
	var bigEndian BigEndianBuf
	buf := &bytes.Buffer{}

	buf.WriteByte('D')
	buf.Write(bigEndian.Uint32(0))

	buf.WriteByte(d.ObjectType)
	buf.WriteString(d.Name)
	buf.WriteByte(0)

	binary.BigEndian.PutUint32(buf.Bytes()[1:5], uint32(buf.Len()-1))

	return buf.Bytes(), nil
}

func (d *Describe) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type       string
		ObjectType string
		Name       string
	}{
		Type:       "Describe",
		ObjectType: string(d.ObjectType),
		Name:       d.Name,
	})
}
