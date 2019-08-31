package proxy

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/jackc/pgproto3/v2"
)

type Proxy struct {
	frontend *pgproto3.Frontend
	backend  *pgproto3.Backend

	frontendConn net.Conn
	backendConn  net.Conn
}

func NewProxy(frontendConn, backendConn net.Conn) *Proxy {
	backend := pgproto3.NewBackend(pgproto3.NewChunkReader(frontendConn), frontendConn)
	frontend := pgproto3.NewFrontend(pgproto3.NewChunkReader(backendConn), backendConn)

	proxy := &Proxy{
		frontend: frontend,
		backend:  backend,

		frontendConn: frontendConn,
		backendConn:  backendConn,
	}

	return proxy
}

func (p *Proxy) Run() error {
	defer p.Close()

	frontendErrChan := make(chan error, 1)
	frontendMsgChan := make(chan pgproto3.FrontendMessage)
	frontendNextChan := make(chan struct{})
	go p.readClientConn(frontendMsgChan, frontendNextChan, frontendErrChan)

	backendErrChan := make(chan error, 1)
	backendMsgChan := make(chan pgproto3.BackendMessage)
	backendNextChan := make(chan struct{})
	go p.readServerConn(backendMsgChan, backendNextChan, backendErrChan)

	for {
		select {
		case msg := <-frontendMsgChan:
			buf, err := json.Marshal(msg)
			if err != nil {
				return err
			}
			fmt.Println("F", string(buf))

			err = p.frontend.Send(msg)
			if err != nil {
				return err
			}
			frontendNextChan <- struct{}{}
		case msg := <-backendMsgChan:
			buf, err := json.Marshal(msg)
			if err != nil {
				return err
			}
			fmt.Println("B", string(buf))

			err = p.backend.Send(msg)
			if err != nil {
				return err
			}
			backendNextChan <- struct{}{}
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

func (p *Proxy) readClientConn(msgChan chan pgproto3.FrontendMessage, nextChan chan struct{}, errChan chan error) {
	startupMessage, err := p.backend.ReceiveStartupMessage()
	if err != nil {
		errChan <- err
		return
	}

	msgChan <- startupMessage
	<-nextChan

	for {
		msg, err := p.backend.Receive()
		if err != nil {
			errChan <- err
			return
		}

		msgChan <- msg
		<-nextChan
	}
}

func (p *Proxy) readServerConn(msgChan chan pgproto3.BackendMessage, nextChan chan struct{}, errChan chan error) {
	for {
		msg, err := p.frontend.Receive()
		if err != nil {
			errChan <- err
			return
		}

		msgChan <- msg

		<-nextChan
	}
}
