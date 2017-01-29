package pgmsg

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
)

const (
	authSubTypeOk                = 0
	authSubTypeCleartextPassword = 3
	authSubTypeMD5Password       = 5
)

type AuthenticationOk struct{}

func (*AuthenticationOk) Backend() {}

type AuthenticationCleartextPassword struct{}

func (*AuthenticationCleartextPassword) Backend() {}

type AuthenticationMD5Password struct {
	Salt [4]byte
}

func (*AuthenticationMD5Password) Backend() {}

func ParseAuthentication(buf []byte) (BackendMessage, error) {
	subType := binary.BigEndian.Uint32(buf[:4])

	switch subType {
	case authSubTypeOk:
		return &AuthenticationOk{}, nil
	case authSubTypeCleartextPassword:
		return &AuthenticationCleartextPassword{}, nil
	case authSubTypeMD5Password:
		var msg AuthenticationMD5Password
		copy(msg.Salt[:], buf[4:8])
		return &msg, nil
	default:
		return nil, fmt.Errorf("unknown authentication subtype: %d", subType)
	}
}

func (a *AuthenticationOk) Encode() ([]byte, error) {
	var bigEndian BigEndianBuf
	buf := &bytes.Buffer{}
	buf.WriteByte('R')
	buf.Write(bigEndian.Uint32(8))
	buf.Write(bigEndian.Uint32(authSubTypeOk))
	return buf.Bytes(), nil
}

func (a *AuthenticationOk) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type string
	}{
		Type: "AuthenticationOk",
	})
}

func (a *AuthenticationCleartextPassword) Encode() ([]byte, error) {
	var bigEndian BigEndianBuf
	buf := &bytes.Buffer{}
	buf.WriteByte('R')
	buf.Write(bigEndian.Uint32(8))
	buf.Write(bigEndian.Uint32(authSubTypeCleartextPassword))
	return buf.Bytes(), nil
}

func (a *AuthenticationMD5Password) Encode() ([]byte, error) {
	var bigEndian BigEndianBuf
	buf := &bytes.Buffer{}
	buf.WriteByte('R')
	buf.Write(bigEndian.Uint32(12))
	buf.Write(bigEndian.Uint32(authSubTypeMD5Password))
	buf.Write(a.Salt[:])
	return buf.Bytes(), nil
}
