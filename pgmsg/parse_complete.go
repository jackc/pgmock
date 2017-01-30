package pgmsg

import (
	"encoding/json"
)

type ParseComplete struct{}

func (*ParseComplete) Backend() {}

func ParseParseComplete(body []byte) (*ParseComplete, error) {
	if len(body) != 0 {
		return nil, &invalidMessageLenErr{messageType: "ParseComplete", expectedLen: 0, actualLen: len(body)}
	}

	return &ParseComplete{}, nil
}

func (t *ParseComplete) Encode() ([]byte, error) {
	return []byte{'1', 0, 0, 0, 4}, nil
}

func (t *ParseComplete) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type string
	}{
		Type: "ParseComplete",
	})
}
