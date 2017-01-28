package pgmock

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"

	"github.com/jackc/pgmock/pgmsg"
)

type Proxy struct {
	clientConn   net.Conn
	clientReader *bufio.Reader

	serverConn   net.Conn
	serverReader *bufio.Reader
}

func NewProxy(clientConn, serverConn net.Conn) *Proxy {
	proxy := &Proxy{
		clientConn:   clientConn,
		clientReader: bufio.NewReader(clientConn),
		serverConn:   serverConn,
		serverReader: bufio.NewReader(serverConn),
	}

	return proxy
}

func (p *Proxy) Run() error {
	defer p.Close()

	clientErrChan := make(chan error, 1)
	clientMsgChan := make(chan pgmsg.Message)
	go p.readClientConn(clientMsgChan, clientErrChan)

	serverErrChan := make(chan error, 1)
	serverMsgChan := make(chan pgmsg.Message)
	go p.readServerConn(serverMsgChan, serverErrChan)

	for {
		select {
		case msg := <-clientMsgChan:
			fmt.Print("client: ")
			buf, err := json.Marshal(msg)
			if err != nil {
				return err
			}
			fmt.Println(string(buf))

			buf, err = msg.Encode()
			if err != nil {
				return err
			}

			_, err = p.serverConn.Write(buf)
			if err != nil {
				return err
			}
		case msg := <-serverMsgChan:
			fmt.Print("server: ")
			buf, err := json.Marshal(msg)
			if err != nil {
				return err
			}
			fmt.Println(string(buf))

			buf, err = msg.Encode()
			if err != nil {
				return err
			}

			_, err = p.clientConn.Write(buf)
			if err != nil {
				return err
			}
		case err := <-clientErrChan:
			return err
		case err := <-serverErrChan:
			return err
		}
	}
}

func (p *Proxy) Close() error {
	clientCloseErr := p.clientConn.Close()
	serverCloseErr := p.serverConn.Close()

	if clientCloseErr != nil {
		return clientCloseErr
	}
	return serverCloseErr
}

func (p *Proxy) readClientConn(msgChan chan pgmsg.Message, errChan chan error) {
	startupMessage, err := p.acceptStartupMessage()
	if err != nil {
		errChan <- err
		return
	}

	msgChan <- startupMessage

	p.relay(p.clientReader, msgChan, errChan)
}

func (p *Proxy) acceptStartupMessage() (*pgmsg.StartupMessage, error) {
	buf := make([]byte, 4)

	_, err := io.ReadFull(p.clientReader, buf)
	if err != nil {
		return nil, err
	}

	msgSize := int(binary.BigEndian.Uint32(buf))

	buf = make([]byte, msgSize-4)
	_, err = io.ReadFull(p.clientReader, buf)
	if err != nil {
		return nil, err
	}

	return pgmsg.ParseStartupMessage(buf)
}

// TODO - probably can DRY main loop for readServerConn and readClientConn
func (p *Proxy) readServerConn(msgChan chan pgmsg.Message, errChan chan error) {
	p.relay(p.serverReader, msgChan, errChan)
}

func (p *Proxy) relay(src io.Reader, msgChan chan pgmsg.Message, errChan chan error) {
	header := make([]byte, 5)
	payload := &bytes.Buffer{}
	for {
		_, err := io.ReadFull(src, header)
		if err != nil {
			errChan <- err
			return
		}

		msgSize := int(binary.BigEndian.Uint32(header[1:])) - 4
		_, err = io.CopyN(payload, src, int64(msgSize))
		if err != nil {
			errChan <- err
			return
		}

		var msg pgmsg.Message
		switch header[0] {
		case 'R':
			msg, err = pgmsg.ParseAuthentication(header[0], payload.Bytes())
			if err != nil {
				errChan <- err
				return
			}
		case 'E':
			msg, err = pgmsg.ParseErrorResponse(header[0], payload.Bytes())
			if err != nil {
				errChan <- err
				return
			}
		case 'S':
			msg, err = pgmsg.ParseParameterStatus(header[0], payload.Bytes())
			if err != nil {
				errChan <- err
				return
			}
		default:
			msg, err = pgmsg.ParseUnknownMessage(header[0], payload.Bytes())
			if err != nil {
				errChan <- err
				return
			}
		}

		msgChan <- msg

		payload.Reset()
	}
}
