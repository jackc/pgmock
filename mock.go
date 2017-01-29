package pgmock

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"

	"github.com/jackc/pgmock/pgmsg"
)

type Mock struct {
	frontendWriter io.ReadWriteCloser
	frontendReader *bufio.Reader
	script         []interface{}
}

func NewMock(frontendConn io.ReadWriteCloser) (*Mock, error) {
	m := &Mock{
		frontendWriter: frontendConn,
		frontendReader: bufio.NewReader(frontendConn),
	}

	_, err := m.acceptStartupMessage()
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (m *Mock) Send(msg pgmsg.Message) error {
	buf, err := msg.Encode()
	if err != nil {
		return err
	}

	_, err = m.frontendWriter.Write(buf)
	return err
}

func (m *Mock) Receive() (pgmsg.FrontendMessage, error) {
	header := make([]byte, 5)
	payload := &bytes.Buffer{}

	_, err := io.ReadFull(m.frontendReader, header)
	if err != nil {
		return nil, err
	}

	msgSize := int(binary.BigEndian.Uint32(header[1:])) - 4
	_, err = io.CopyN(payload, m.frontendReader, int64(msgSize))
	if err != nil {
		return nil, err
	}

	return pgmsg.ParseFrontend(header[0], payload.Bytes())
}

func (m *Mock) acceptStartupMessage() (*pgmsg.StartupMessage, error) {
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

	return pgmsg.ParseStartupMessage(buf)
}

func (m *Mock) respondWithAuthenticationOk() error {
	var msg pgmsg.AuthenticationOk

	buf, err := msg.Encode()
	if err != nil {
		return err
	}

	_, err = m.frontendWriter.Write(buf)
	return err
}
