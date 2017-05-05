package pgmock

import (
	"encoding/binary"
	"io"

	"github.com/jackc/pgx/pgproto3"
)

type Mock struct {
	backend *pgproto3.Backend
	script  []interface{}
}

func NewMock(frontendConn io.ReadWriteCloser) (*Mock, error) {
	backend, err := pgproto3.NewBackend(frontendConn, frontendConn)
	if err != nil {
		return nil, err
	}
	m := &Mock{
		backend: backend,
	}

	_, err = m.acceptStartupMessage()
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (m *Mock) Send(msg pgproto3.BackendMessage) error {
	return m.backend.Send(msg)
}

func (m *Mock) Receive() (pgproto3.FrontendMessage, error) {
	return m.backend.Receive()
}

func (m *Mock) acceptStartupMessage() (*pgproto3.StartupMessage, error) {
	buf := make([]byte, 4)

	_, err := io.ReadFull(m.frontendReader, buf)
	if err != nil {
		return nil, err
	}

	msgSize := int(binary.BigEndian.Uint32(buf[:4]))

	buf = make([]byte, msgSize-4)
	_, err = io.ReadFull(m.frontendReader, buf)
	if err != nil {
		return nil, err
	}

	return pgproto3.ParseStartupMessage(buf)
}

func (m *Mock) respondWithAuthenticationOk() error {
	var msg pgproto3.AuthenticationOk

	buf, err := msg.MarshalBinary()
	if err != nil {
		return err
	}

	_, err = m.frontendWriter.Write(buf)
	return err
}
