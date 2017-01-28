package pgmsg

import (
	"bytes"
	"encoding/json"
)

type PasswordMessage struct {
	Password string
}

func ParsePasswordMessage(body []byte) (*PasswordMessage, error) {
	var pm PasswordMessage

	buf := bytes.NewBuffer(body)

	b, err := buf.ReadBytes(0)
	if err != nil {
		return nil, err
	}
	pm.Password = string(b[:len(b)-1])

	return &pm, nil
}

func (pm *PasswordMessage) Encode() ([]byte, error) {
	var bigEndian BigEndianBuf
	buf := &bytes.Buffer{}
	buf.WriteByte('p')
	buf.Write(bigEndian.Uint32(uint32(4 + len(pm.Password) + 1)))
	buf.WriteString(pm.Password)
	buf.WriteByte(0)
	return buf.Bytes(), nil
}

func (pm *PasswordMessage) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type     string
		Password string
	}{
		Type:     "PasswordMessage",
		Password: pm.Password,
	})
}
