package pgmsg

import (
	"encoding/json"
)

type Terminate struct{}

func ParseTerminate(body []byte) (*Terminate, error) {
	if len(body) != 0 {
		return nil, &invalidMessageLenErr{messageType: "Terminate", expectedLen: 0, actualLen: len(body)}
	}

	return &Terminate{}, nil
}

func (t *Terminate) Encode() ([]byte, error) {
	return []byte{'X', 0, 0, 0, 4}, nil
}

func (t *Terminate) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type string
	}{
		Type: "Terminate",
	})
}
