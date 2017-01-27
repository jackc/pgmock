package pgmock

import (
	"bufio"
	"bytes"
	"encoding/binary"
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

type DialFunc func() (net.Conn, error)

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
	clientErrChan := make(chan error)
	go p.readClientConn(clientErrChan)

	serverErrChan := make(chan error)
	go p.readServerConn(serverErrChan)

	var clientErr, serverErr error

	select {
	case clientErr = <-clientErrChan:
		p.Close()
		serverErr = <-serverErrChan
	case serverErr = <-serverErrChan:
		p.Close()
		clientErr = <-clientErrChan
	}

	if clientErr != nil {
		return clientErr
	}
	return serverErr
}

func (p *Proxy) Close() error {
	clientCloseErr := p.clientConn.Close()
	serverCloseErr := p.serverConn.Close()

	if clientCloseErr != nil {
		return clientCloseErr
	}
	return serverCloseErr
}

func (p *Proxy) readClientConn(errChan chan error) {
	startupMessage, err := p.acceptStartupMessage()
	if err != nil {
		errChan <- err
		return
	}

	_, err = startupMessage.WriteTo(p.serverConn)
	if err != nil {
		errChan <- err
		return
	}

	p.relay(p.serverConn, p.clientReader, errChan)
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
func (p *Proxy) readServerConn(errChan chan error) {
	p.relay(p.clientConn, p.serverReader, errChan)
}

func (p *Proxy) relay(dst io.Writer, src io.Reader, errChan chan error) {
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

		switch header[0] {
		default:
			um, err := pgmsg.ParseUnknownMessage(header[0], payload.Bytes())
			if err != nil {
				errChan <- err
				return
			}

			_, err = um.WriteTo(dst)
			if err != nil {
				errChan <- err
				return
			}
		}

		payload.Reset()
	}
}
