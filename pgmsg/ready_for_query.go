package pgmsg

import (
	"encoding/json"
)

type ReadyForQuery struct {
	TxStatus byte
}

func ParseReadyForQuery(body []byte) (*ReadyForQuery, error) {
	if len(body) != 1 {
		return nil, &invalidMessageLenErr{messageType: "ReadyForQuery", expectedLen: 1, actualLen: len(body)}
	}

	return &ReadyForQuery{TxStatus: body[0]}, nil
}

func (rfq *ReadyForQuery) Encode() ([]byte, error) {
	return []byte{'Z', 0, 0, 0, 5, rfq.TxStatus}, nil
}

func (rfq *ReadyForQuery) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type     string
		TxStatus string
	}{
		Type:     "ReadyForQuery",
		TxStatus: string(rfq.TxStatus),
	})
}
