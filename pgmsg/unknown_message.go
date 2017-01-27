package pgmsg

import (
	"bytes"
	"encoding/json"
	"io"
)

type UnknownMessage struct {
	Type    byte
	Payload []byte
}

func ParseUnknownMessage(t byte, buf []byte) (*UnknownMessage, error) {
	return &UnknownMessage{Type: t, Payload: buf}, nil
}

func (um *UnknownMessage) Encode() ([]byte, error) {
	var bigEndian BigEndianBuf
	buf := &bytes.Buffer{}
	buf.WriteByte(um.Type)
	buf.Write(bigEndian.Uint32(uint32(len(um.Payload) + 4)))
	buf.Write(um.Payload)
	return buf.Bytes(), nil
}

func (um *UnknownMessage) WriteTo(w io.Writer) (int64, error) {
	buf, err := um.Encode()
	if err != nil {
		return 0, err
	}

	n, err := w.Write(buf)
	return int64(n), err
}

func (um *UnknownMessage) MarshalJSON() ([]byte, error) {
	return json.Marshal(um)
}
