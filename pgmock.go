package pgmock

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/jackc/pgmock/pgmsg"
)

const (
	protocolVersionNumber = 196608 // 3.0
	sslRequestNumber      = 80877103
)

type MockConn struct {
	pgConn io.ReadWriteCloser
}

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

func (mc *MockConn) acceptStartupMessage() (*pgmsg.StartupMessage, error) {
	buf := make([]byte, 4)

	_, err := io.ReadFull(mc.pgConn, buf)
	if err != nil {
		return nil, err
	}

	msgSize := int(binary.BigEndian.Uint32(buf[:4]))
	fmt.Println("msgSize", msgSize)

	buf = make([]byte, msgSize-4)
	_, err = io.ReadFull(mc.pgConn, buf)
	if err != nil {
		return nil, err
	}
	fmt.Println(buf)
	return pgmsg.ParseStartupMessage(buf)
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
