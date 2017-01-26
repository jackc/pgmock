package pgmsg

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	protocolVersionNumber = 196608 // 3.0
	sslRequestNumber      = 80877103
)

type StartupMessage struct {
	ProtocolVersion int32
	Parameters      map[string]string
}

func ParseStartupMessage(buf []byte) (*StartupMessage, error) {
	var msg StartupMessage

	msg.ProtocolVersion = int32(binary.BigEndian.Uint32(buf[:4]))
	if msg.ProtocolVersion == sslRequestNumber {
		return nil, fmt.Errorf("can't handle ssl connection request")
	}

	if msg.ProtocolVersion != protocolVersionNumber {
		return nil, fmt.Errorf("Bad startup message version number. Expected %d, got %d", protocolVersionNumber, msg.ProtocolVersion)
	}

	reader := bytes.NewBuffer(buf[4:])

	msg.Parameters = make(map[string]string)
	for {
		key, err := reader.ReadBytes(0)
		if err != nil {
			return nil, err
		}
		value, err := reader.ReadBytes(0)
		if err != nil {
			return nil, err
		}

		msg.Parameters[string(key[:len(key)-1])] = string(value[:len(value)-1])

		if reader.Len() == 1 {
			b, err := reader.ReadByte()
			if err != nil {
				return nil, err
			}
			if b != 0 {
				return nil, fmt.Errorf("Bad startup message last byte. Expected 0, got %d", b)
			}
			break
		}
	}

	return &msg, nil
}
