package pgmock

import (
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

	_, err = m.ReceiveStartupMessage()
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

func (m *Mock) ReceiveStartupMessage() (*pgproto3.StartupMessage, error) {
	return m.backend.ReceiveStartupMessage()
}
