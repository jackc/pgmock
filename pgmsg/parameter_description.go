package pgmsg

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
)

type ParameterDescription struct {
	ParameterOIDs []uint32
}

func (*ParameterDescription) Backend() {}

func ParseParameterDescription(body []byte) (*ParameterDescription, error) {
	buf := bytes.NewBuffer(body)

	if buf.Len() < 2 {
		return nil, &invalidMessageFormatErr{messageType: "ParameterDescription"}
	}
	parameterCount := int(binary.BigEndian.Uint16(buf.Next(2)))
	if buf.Len() != parameterCount*4 {
		return nil, &invalidMessageFormatErr{messageType: "ParameterDescription"}
	}

	pd := &ParameterDescription{ParameterOIDs: make([]uint32, parameterCount)}

	for i := 0; i < parameterCount; i++ {
		pd.ParameterOIDs[i] = binary.BigEndian.Uint32(buf.Next(4))
	}

	return pd, nil
}

func (pd *ParameterDescription) Encode() ([]byte, error) {
	var bigEndian BigEndianBuf
	buf := &bytes.Buffer{}

	buf.WriteByte('t')
	buf.Write(bigEndian.Uint32(uint32(4 + 2 + 4*len(pd.ParameterOIDs))))

	buf.Write(bigEndian.Uint16(uint16(len(pd.ParameterOIDs))))

	for _, oid := range pd.ParameterOIDs {
		buf.Write(bigEndian.Uint32(oid))
	}

	return buf.Bytes(), nil
}

func (pd *ParameterDescription) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type          string
		ParameterOIDs []uint32
	}{
		Type:          "ParameterDescription",
		ParameterOIDs: pd.ParameterOIDs,
	})
}
