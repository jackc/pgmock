package pgmsg

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
)

type BackendKeyData struct {
	ProcessID uint32
	SecretKey uint32
}

func (*BackendKeyData) Backend() {}

func ParseBackendKeyData(body []byte) (*BackendKeyData, error) {
	if len(body) != 8 {
		return nil, &invalidMessageLenErr{messageType: "BackendKeyData", expectedLen: 8, actualLen: len(body)}
	}

	var bkd BackendKeyData

	bkd.ProcessID = binary.BigEndian.Uint32(body[:4])
	bkd.SecretKey = binary.BigEndian.Uint32(body[4:])

	return &bkd, nil
}

func (bkd *BackendKeyData) Encode() ([]byte, error) {
	var bigEndian BigEndianBuf
	buf := &bytes.Buffer{}
	buf.WriteByte('K')
	buf.Write(bigEndian.Uint32(12))
	buf.Write(bigEndian.Uint32(bkd.ProcessID))
	buf.Write(bigEndian.Uint32(bkd.SecretKey))
	return buf.Bytes(), nil
}

func (bkd *BackendKeyData) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type      string
		ProcessID uint32
		SecretKey uint32
	}{
		Type:      "BackendKeyData",
		ProcessID: bkd.ProcessID,
		SecretKey: bkd.SecretKey,
	})
}
