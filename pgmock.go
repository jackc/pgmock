package pgmock

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

const (
	protocolVersionNumber = 196608 // 3.0
	sslRequestNumber      = 80877103
)

type MockConn struct {
	pgConn io.ReadWriteCloser
}

type startupMsg map[string]string

func NewMockConn(pgConn io.ReadWriteCloser) *MockConn {
	return &MockConn{
		pgConn: pgConn,
	}
}

func (mc *MockConn) Run() error {
	msg, err := mc.acceptStartupMessage()
	if err != nil {
		return err
	}

	fmt.Println("startupMsg", msg)

	mc.respondWithAuthenticationOk()

	mc.pgConn.Write([]byte("foobar"))

	return nil

}

func (mc *MockConn) acceptStartupMessage() (startupMsg, error) {
	buf := make([]byte, 8)

	_, err := io.ReadFull(mc.pgConn, buf)
	if err != nil {
		return nil, err
	}

	fmt.Println(buf)

	msgSize := int32(binary.BigEndian.Uint32(buf[:4]))
	clientProtocalVersionNumber := int32(binary.BigEndian.Uint32(buf[4:8]))

	if clientProtocalVersionNumber == sslRequestNumber {
		return nil, fmt.Errorf("can't handle ssl connection request")
	}

	if clientProtocalVersionNumber != protocolVersionNumber {
		return nil, fmt.Errorf("Bad startup message version number. Expected %d, got %d", protocolVersionNumber, clientProtocalVersionNumber)
	}

	buf = make([]byte, int(msgSize-8))
	_, err = io.ReadFull(mc.pgConn, buf)
	if err != nil {
		return nil, err
	}

	reader := bytes.NewBuffer(buf)

	startupMsg := make(startupMsg)
	for {
		key, err := reader.ReadBytes(0)
		if err != nil {
			return nil, err
		}
		value, err := reader.ReadBytes(0)
		if err != nil {
			return nil, err
		}

		startupMsg[string(key[:len(key)-1])] = string(value[:len(value)-1])

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

	return startupMsg, nil
}

func (mc *MockConn) respondWithAuthenticationOk() error {
	wbuf := newWriteBuf('R') // Authentication
	wbuf.WriteInt32(0)       // Success
	wbuf.closeMsg()

	_, err := mc.pgConn.Write(wbuf.buf)
	if err != nil {
		return err
	}

	return nil
}
