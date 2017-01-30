package pgmsg

import (
	"encoding/json"
)

type BindComplete struct{}

func (*BindComplete) Backend() {}

func ParseBindComplete(body []byte) (*BindComplete, error) {
	if len(body) != 0 {
		return nil, &invalidMessageLenErr{messageType: "BindComplete", expectedLen: 0, actualLen: len(body)}
	}

	return &BindComplete{}, nil
}

func (t *BindComplete) Encode() ([]byte, error) {
	return []byte{'2', 0, 0, 0, 4}, nil
}

func (t *BindComplete) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type string
	}{
		Type: "BindComplete",
	})
}
