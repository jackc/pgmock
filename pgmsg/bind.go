package pgmsg

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
)

type Bind struct {
	DestinationPortal    string
	PreparedStatement    string
	ParameterFormatCodes []int16
	Parameters           [][]byte
	ResultFormatCodes    []int16
}

func (*Bind) Frontend() {}

func ParseBind(body []byte) (*Bind, error) {
	var b Bind
	buf := bytes.NewBuffer(body)

	b2, err := buf.ReadBytes(0)
	if err != nil {
		return nil, err
	}
	b.DestinationPortal = string(b2[:len(b2)-1])

	b2, err = buf.ReadBytes(0)
	if err != nil {
		return nil, err
	}
	b.PreparedStatement = string(b2[:len(b2)-1])

	if buf.Len() < 2 {
		return nil, &invalidMessageFormatErr{messageType: "Bind"}
	}
	parameterFormatCodeCount := int(binary.BigEndian.Uint16(buf.Next(2)))

	b.ParameterFormatCodes = make([]int16, parameterFormatCodeCount)

	for i := 0; i < parameterFormatCodeCount; i++ {
		if buf.Len() < 2 {
			return nil, &invalidMessageFormatErr{messageType: "Bind"}
		}

		b.ParameterFormatCodes[i] = int16(binary.BigEndian.Uint16(buf.Next(2)))
	}

	if buf.Len() < 2 {
		return nil, &invalidMessageFormatErr{messageType: "Bind"}
	}
	parameterCount := int(binary.BigEndian.Uint16(buf.Next(2)))

	b.Parameters = make([][]byte, parameterCount)

	for i := 0; i < parameterCount; i++ {
		if buf.Len() < 4 {
			return nil, &invalidMessageFormatErr{messageType: "Bind"}
		}

		msgSize := int(int32(binary.BigEndian.Uint32(buf.Next(4))))

		// null
		if msgSize == -1 {
			continue
		}

		value := make([]byte, msgSize)
		_, err := buf.Read(value)
		if err != nil {
			return nil, err
		}

		b.Parameters[i] = value
	}

	if buf.Len() < 2 {
		return nil, &invalidMessageFormatErr{messageType: "Bind"}
	}
	resultFormatCodeCount := int(binary.BigEndian.Uint16(buf.Next(2)))

	b.ResultFormatCodes = make([]int16, resultFormatCodeCount)

	for i := 0; i < resultFormatCodeCount; i++ {
		if buf.Len() < 2 {
			return nil, &invalidMessageFormatErr{messageType: "Bind"}
		}

		b.ResultFormatCodes[i] = int16(binary.BigEndian.Uint16(buf.Next(2)))
	}

	return &b, nil
}

func (b *Bind) Encode() ([]byte, error) {
	var bigEndian BigEndianBuf
	buf := &bytes.Buffer{}

	buf.WriteByte('B')
	buf.Write(bigEndian.Uint32(0))

	buf.WriteString(b.DestinationPortal)
	buf.WriteByte(0)
	buf.WriteString(b.PreparedStatement)
	buf.WriteByte(0)

	buf.Write(bigEndian.Uint16(uint16(len(b.ParameterFormatCodes))))

	for _, fc := range b.ParameterFormatCodes {
		buf.Write(bigEndian.Int16(fc))
	}

	buf.Write(bigEndian.Uint16(uint16(len(b.Parameters))))

	for _, p := range b.Parameters {
		if p == nil {
			buf.Write(bigEndian.Int32(-1))
			continue
		}

		buf.Write(bigEndian.Int32(int32(len(p))))
		buf.Write(p)
	}

	buf.Write(bigEndian.Uint16(uint16(len(b.ResultFormatCodes))))

	for _, fc := range b.ResultFormatCodes {
		buf.Write(bigEndian.Int16(fc))
	}

	binary.BigEndian.PutUint32(buf.Bytes()[1:5], uint32(buf.Len()-1))

	return buf.Bytes(), nil
}

func (b *Bind) MarshalJSON() ([]byte, error) {
	formattedParameters := make([]map[string]string, len(b.Parameters))
	for i, p := range b.Parameters {
		if p == nil {
			continue
		}

		if b.ParameterFormatCodes[i] == 0 {
			formattedParameters[i] = map[string]string{"text": string(p)}
		} else {
			formattedParameters[i] = map[string]string{"binary": hex.EncodeToString(p)}
		}
	}

	return json.Marshal(struct {
		Type                 string
		DestinationPortal    string
		PreparedStatement    string
		ParameterFormatCodes []int16
		Parameters           []map[string]string
		ResultFormatCodes    []int16
	}{
		Type:                 "Bind",
		DestinationPortal:    b.DestinationPortal,
		PreparedStatement:    b.PreparedStatement,
		ParameterFormatCodes: b.ParameterFormatCodes,
		Parameters:           formattedParameters,
		ResultFormatCodes:    b.ResultFormatCodes,
	})
}
