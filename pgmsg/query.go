package pgmsg

import (
	"bytes"
	"encoding/json"
)

type Query struct {
	String string
}

func (*Query) Frontend() {}

func ParseQuery(body []byte) (*Query, error) {
	var q Query

	buf := bytes.NewBuffer(body)

	b, err := buf.ReadBytes(0)
	if err != nil {
		return nil, err
	}
	q.String = string(b[:len(b)-1])

	return &q, nil
}

func (q *Query) Encode() ([]byte, error) {
	var bigEndian BigEndianBuf
	buf := &bytes.Buffer{}
	buf.WriteByte('Q')
	buf.Write(bigEndian.Uint32(uint32(4 + len(q.String) + 1)))
	buf.WriteString(q.String)
	buf.WriteByte(0)
	return buf.Bytes(), nil
}

func (q *Query) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type   string
		String string
	}{
		Type:   "Query",
		String: q.String,
	})
}
