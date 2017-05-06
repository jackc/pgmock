package pgmock

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/jackc/pgx/pgproto3"
)

type Proxy struct {
	frontend *pgproto3.Frontend
	backend  *pgproto3.Backend

	frontendConn net.Conn
	backendConn  net.Conn
}

func NewProxy(frontendConn, backendConn net.Conn) (*Proxy, error) {
	backend, err := pgproto3.NewBackend(frontendConn, frontendConn)
	if err != nil {
		return nil, err
	}
	frontend, err := pgproto3.NewFrontend(backendConn, backendConn)
	if err != nil {
		return nil, err
	}

	proxy := &Proxy{
		frontend: frontend,
		backend:  backend,

		frontendConn: frontendConn,
		backendConn:  backendConn,
	}

	return proxy, nil
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

			err = p.frontend.Send(msg)
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

			err = p.backend.Send(msg)
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
	startupMessage, err := p.backend.ReceiveStartupMessage()
	if err != nil {
		errChan <- err
		return
	}

	msgChan <- startupMessage

	for {
		msg, err := p.backend.Receive()
		if err != nil {
			errChan <- err
			return
		}

		msgChan <- msg
	}
}

func (p *Proxy) readServerConn(msgChan chan pgproto3.BackendMessage, errChan chan error) {
	for {
		msg, err := p.frontend.Receive()
		if err != nil {
			errChan <- err
			return
		}

		msgChan <- msg
	}
}
