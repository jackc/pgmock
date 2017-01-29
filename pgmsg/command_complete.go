package pgmsg

import (
	"bytes"
	"encoding/json"
)

type CommandComplete struct {
	CommandTag string
}

func (*CommandComplete) Backend() {}

func ParseCommandComplete(body []byte) (*CommandComplete, error) {
	var cc CommandComplete

	buf := bytes.NewBuffer(body)

	b, err := buf.ReadBytes(0)
	if err != nil {
		return nil, err
	}
	cc.CommandTag = string(b[:len(b)-1])

	return &cc, nil
}

func (cc *CommandComplete) Encode() ([]byte, error) {
	var bigEndian BigEndianBuf
	buf := &bytes.Buffer{}

	buf.WriteByte('C')
	buf.Write(bigEndian.Uint32(uint32(4 + len(cc.CommandTag) + 1)))

	buf.WriteString(cc.CommandTag)
	buf.WriteByte(0)

	return buf.Bytes(), nil
}

func (cc *CommandComplete) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type       string
		CommandTag string
	}{
		Type:       "CommandComplete",
		CommandTag: cc.CommandTag,
	})
}
