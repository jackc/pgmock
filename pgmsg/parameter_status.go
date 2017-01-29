package pgmsg

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
)

type ParameterStatus struct {
	Name  string
	Value string
}

func (*ParameterStatus) Backend() {}

func ParseParameterStatus(rawBuf []byte) (*ParameterStatus, error) {
	var ps ParameterStatus

	buf := bytes.NewBuffer(rawBuf)

	b, err := buf.ReadBytes(0)
	if err != nil {
		return nil, err
	}
	ps.Name = string(b[:len(b)-1])

	b, err = buf.ReadBytes(0)
	if err != nil {
		return nil, err
	}
	ps.Value = string(b[:len(b)-1])

	return &ps, nil
}

func (ps *ParameterStatus) Encode() ([]byte, error) {
	var bigEndian BigEndianBuf
	buf := &bytes.Buffer{}

	buf.WriteByte('S')
	buf.Write(bigEndian.Uint32(0))

	buf.WriteString(ps.Name)
	buf.WriteByte(0)
	buf.WriteString(ps.Value)
	buf.WriteByte(0)

	binary.BigEndian.PutUint32(buf.Bytes()[1:5], uint32(buf.Len()-1))

	return buf.Bytes(), nil
}

func (ps *ParameterStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type  string
		Name  string
		Value string
	}{
		Type:  "ParameterStatus",
		Name:  ps.Name,
		Value: ps.Value,
	})
}
