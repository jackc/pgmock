package pgmsg

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
)

type Execute struct {
	Portal  string
	MaxRows uint32
}

func (*Execute) Frontend() {}

func ParseExecute(body []byte) (*Execute, error) {
	var e Execute

	buf := bytes.NewBuffer(body)

	b, err := buf.ReadBytes(0)
	if err != nil {
		return nil, err
	}
	e.Portal = string(b[:len(b)-1])

	if buf.Len() < 4 {
		return nil, &invalidMessageFormatErr{messageType: "Execute"}
	}
	e.MaxRows = binary.BigEndian.Uint32(buf.Next(4))

	return &e, nil
}

func (e *Execute) Encode() ([]byte, error) {
	var bigEndian BigEndianBuf
	buf := &bytes.Buffer{}

	buf.WriteByte('E')
	buf.Write(bigEndian.Uint32(0))

	buf.WriteString(e.Portal)
	buf.WriteByte(0)

	buf.Write(bigEndian.Uint32(e.MaxRows))

	binary.BigEndian.PutUint32(buf.Bytes()[1:5], uint32(buf.Len()-1))

	return buf.Bytes(), nil
}

func (e *Execute) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type    string
		Portal  string
		MaxRows uint32
	}{
		Type:    "Execute",
		Portal:  e.Portal,
		MaxRows: e.MaxRows,
	})
}
