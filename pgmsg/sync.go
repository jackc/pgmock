package pgmsg

import (
	"encoding/json"
)

type Sync struct{}

func (*Sync) Frontend() {}

func ParseSync(body []byte) (*Sync, error) {
	if len(body) != 0 {
		return nil, &invalidMessageLenErr{messageType: "Sync", expectedLen: 0, actualLen: len(body)}
	}

	return &Sync{}, nil
}

func (t *Sync) Encode() ([]byte, error) {
	return []byte{'S', 0, 0, 0, 4}, nil
}

func (t *Sync) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type string
	}{
		Type: "Sync",
	})
}
