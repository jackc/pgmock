package pgmsg

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
)

type Parse struct {
	Name          string
	Query         string
	ParameterOIDs []uint32
}

func (*Parse) Frontend() {}

func ParseParse(body []byte) (*Parse, error) {
	var p Parse

	buf := bytes.NewBuffer(body)

	b, err := buf.ReadBytes(0)
	if err != nil {
		return nil, err
	}
	p.Name = string(b[:len(b)-1])

	b, err = buf.ReadBytes(0)
	if err != nil {
		return nil, err
	}
	p.Query = string(b[:len(b)-1])

	if buf.Len() < 2 {
		return nil, &invalidMessageFormatErr{messageType: "Parse"}
	}
	parameterOIDCount := int(binary.BigEndian.Uint16(buf.Next(2)))

	for i := 0; i < parameterOIDCount; i++ {
		if buf.Len() < 4 {
			return nil, &invalidMessageFormatErr{messageType: "Parse"}
		}
		p.ParameterOIDs = append(p.ParameterOIDs, binary.BigEndian.Uint32(buf.Next(4)))
	}

	return &p, nil
}

func (p *Parse) Encode() ([]byte, error) {
	var bigEndian BigEndianBuf
	buf := &bytes.Buffer{}

	buf.WriteByte('P')
	buf.Write(bigEndian.Uint32(0))

	buf.WriteString(p.Name)
	buf.WriteByte(0)
	buf.WriteString(p.Query)
	buf.WriteByte(0)

	buf.Write(bigEndian.Uint16(uint16(len(p.ParameterOIDs))))

	for _, v := range p.ParameterOIDs {
		buf.Write(bigEndian.Uint32(v))
	}

	binary.BigEndian.PutUint32(buf.Bytes()[1:5], uint32(buf.Len()-1))

	return buf.Bytes(), nil
}

func (p *Parse) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type          string
		Name          string
		Query         string
		ParameterOIDs []uint32
	}{
		Type:          "Parse",
		Name:          p.Name,
		Query:         p.Query,
		ParameterOIDs: p.ParameterOIDs,
	})
}
