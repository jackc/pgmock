package pgmock

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"

	"github.com/jackc/pgx/pgproto3"
)

type Proxy struct {
	frontendConn   net.Conn
	frontendReader *bufio.Reader

	backendConn   net.Conn
	backendReader *bufio.Reader
}

func NewProxy(frontendConn, backendConn net.Conn) *Proxy {
	proxy := &Proxy{
		frontendConn:   frontendConn,
		frontendReader: bufio.NewReader(frontendConn),
		backendConn:    backendConn,
		backendReader:  bufio.NewReader(backendConn),
	}

	return proxy
}

func (p *Proxy) Run() error {
	defer p.Close()

	frontendErrChan := make(chan error, 1)
	frontendMsgChan := make(chan pgproto3.FrontendMessage)
	go p.readClientConn(frontendMsgChan, frontendErrChan)

	backendErrChan := make(chan error, 1)
	backendMsgChan := make(chan pgproto3.BackendMessage)
	go p.readServerConn(backendMsgChan, backendErrChan)

	for {
		select {
		case msg := <-frontendMsgChan:
			fmt.Print("frontend: ")
			buf, err := json.Marshal(msg)
			if err != nil {
				return err
			}
			fmt.Println(string(buf))

			buf, err = msg.Encode()
			if err != nil {
				return err
			}

			_, err = p.backendConn.Write(buf)
			if err != nil {
				return err
			}
		case msg := <-backendMsgChan:
			fmt.Print("backend: ")
			buf, err := json.Marshal(msg)
			if err != nil {
				return err
			}
			fmt.Println(string(buf))

			buf, err = msg.Encode()
			if err != nil {
				return err
			}

			_, err = p.frontendConn.Write(buf)
			if err != nil {
				return err
			}
		case err := <-frontendErrChan:
			return err
		case err := <-backendErrChan:
			return err
		}
	}
}

func (p *Proxy) Close() error {
	frontendCloseErr := p.frontendConn.Close()
	backendCloseErr := p.backendConn.Close()

	if frontendCloseErr != nil {
		return frontendCloseErr
	}
	return backendCloseErr
}

func (p *Proxy) readClientConn(msgChan chan pgproto3.FrontendMessage, errChan chan error) {
	startupMessage, err := p.acceptStartupMessage()
	if err != nil {
		errChan <- err
		return
	}

	msgChan <- startupMessage

	header := make([]byte, 5)
	payload := &bytes.Buffer{}
	for {
		_, err := io.ReadFull(p.frontendReader, header)
		if err != nil {
			errChan <- err
			return
		}

		msgSize := int(binary.BigEndian.Uint32(header[1:])) - 4
		_, err = io.CopyN(payload, p.frontendReader, int64(msgSize))
		if err != nil {
			errChan <- err
			return
		}

		msg, err := pgproto3.ParseFrontend(header[0], payload.Bytes())
		if err != nil {
			errChan <- err
			return
		}

		msgChan <- msg

		payload.Reset()
	}
}

func (p *Proxy) acceptStartupMessage() (*pgproto3.StartupMessage, error) {
	buf := make([]byte, 4)

	_, err := io.ReadFull(p.frontendReader, buf)
	if err != nil {
		return nil, err
	}

	msgSize := int(binary.BigEndian.Uint32(buf))

	buf = make([]byte, msgSize-4)
	_, err = io.ReadFull(p.frontendReader, buf)
	if err != nil {
		return nil, err
	}

	return pgproto3.ParseStartupMessage(buf)
}

func (p *Proxy) readServerConn(msgChan chan pgproto3.BackendMessage, errChan chan error) {
	header := make([]byte, 5)
	payload := &bytes.Buffer{}
	for {
		_, err := io.ReadFull(p.backendReader, header)
		if err != nil {
			errChan <- err
			return
		}

		msgSize := int(binary.BigEndian.Uint32(header[1:])) - 4
		_, err = io.CopyN(payload, p.backendReader, int64(msgSize))
		if err != nil {
			errChan <- err
			return
		}

		msg, err := pgproto3.ParseBackend(header[0], payload.Bytes())
		if err != nil {
			errChan <- err
			return
		}

		msgChan <- msg

		payload.Reset()
	}
}
